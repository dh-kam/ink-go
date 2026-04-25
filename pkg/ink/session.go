package ink

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	inkinput "github.com/dh-kam/goink.go/pkg/input"
)

// RenderMetrics describes one completed render pass.
type RenderMetrics struct {
	RenderTime time.Duration
}

// RenderOptions configures a mounted Ink session.
type RenderOptions struct {
	AppOptions
	Debug                bool
	MaxFPS               int
	IncrementalRendering bool
	OnRender             func(RenderMetrics)
}

type scheduledTimer interface {
	Stop() bool
}

type sessionStreamProxy struct {
	target          io.Writer
	restoreToStdout bool
	instance        *Instance
}

func (proxy *sessionStreamProxy) Write(data []byte) (int, error) {
	if proxy == nil || proxy.target == nil {
		return 0, nil
	}

	if proxy.instance == nil {
		return proxy.target.Write(data)
	}

	return proxy.instance.writeHookOutputLocked(proxy.target, string(data), proxy.restoreToStdout)
}

func (proxy *sessionStreamProxy) IsTTY() bool {
	if ttyWriter, ok := proxy.target.(ttyAwareWriter); ok {
		return ttyWriter.IsTTY()
	}

	return false
}

func (proxy *sessionStreamProxy) Columns() int {
	if columnsWriter, ok := proxy.target.(columnsAwareWriter); ok {
		return columnsWriter.Columns()
	}

	return 0
}

func (proxy *sessionStreamProxy) Rows() int {
	if rowsWriter, ok := proxy.target.(rowsAwareWriter); ok {
		return rowsWriter.Rows()
	}

	return 0
}

func (proxy *sessionStreamProxy) SubscribeResize(handler func()) func() {
	if subscriber, ok := proxy.target.(resizeSubscriber); ok {
		return subscriber.SubscribeResize(handler)
	}

	return nil
}

func (proxy *sessionStreamProxy) Flush() error {
	if flushable, ok := proxy.target.(flushableWriter); ok {
		return flushable.Flush()
	}

	return nil
}

func (proxy *sessionStreamProxy) Wait() error {
	if waitable, ok := proxy.target.(waitableWriter); ok {
		return waitable.Wait()
	}

	return nil
}

type sessionStreamProxyWithUintptrFD struct {
	sessionStreamProxy
	fd uintptrFD
}

func (proxy *sessionStreamProxyWithUintptrFD) Fd() uintptr {
	return proxy.fd.Fd()
}

type sessionStreamProxyWithIntFD struct {
	sessionStreamProxy
	fd intFD
}

func (proxy *sessionStreamProxyWithIntFD) Fd() int {
	return proxy.fd.Fd()
}

// Instance manages a mounted Ink app session.
type Instance struct {
	mu                     sync.Mutex
	app                    *App
	stdout                 io.Writer
	stderr                 io.Writer
	onRender               func(RenderMetrics)
	debug                  bool
	maxFPS                 int
	renderThrottle         time.Duration
	incrementalRendering   bool
	fullStaticOutput       string
	staticCounts           []int
	previousLogicalOutput  string
	previousOutput         string
	previousLines          []string
	previousLineCount      int
	previousCursorPosition *CursorPosition
	cursorWasShown         bool
	cursorHidden           bool
	unmounted              bool
	exitErr                error
	exited                 chan struct{}
	exitOnce               sync.Once
	registry               *instanceRegistry
	registryKey            any
	now                    func() time.Time
	afterFunc              func(time.Duration, func()) scheduledTimer
	pendingTimer           scheduledTimer
	pendingRender          *preparedRender
	unsubscribeResize      func()
	unsubscribeInput       func()
	lastRenderAt           time.Time
}

type renderState struct {
	logicalOutput  string
	output         string
	lines          []string
	lineCount      int
	cursorPosition *CursorPosition
	cursorWasShown bool
}

type preparedRender struct {
	logicalOutput       string
	renderedOutput      string
	staticOutput        string
	cursorPosition      *CursorPosition
	shouldWrite         bool
	shouldClearTerminal bool
	renderTime          time.Duration
	exitRequested       bool
	exitErr             error
}

// Mount creates a mounted Ink session and immediately renders it.
func Mount(component ComponentFunc) (*Instance, error) {
	return MountWithOptions(component, RenderOptions{})
}

// MountWithOptions creates a mounted Ink session with explicit render options.
func MountWithOptions(component ComponentFunc, options RenderOptions) (*Instance, error) {
	app := NewAppWithOptions(component, options.AppOptions)
	instance := &Instance{
		app:                  app,
		onRender:             options.OnRender,
		debug:                options.Debug,
		maxFPS:               normalizeMaxFPS(options.MaxFPS),
		renderThrottle:       throttleDuration(options.MaxFPS),
		incrementalRendering: options.IncrementalRendering,
		exited:               make(chan struct{}),
		now:                  time.Now,
		afterFunc: func(delay time.Duration, fn func()) scheduledTimer {
			return time.AfterFunc(delay, fn)
		},
	}

	instance.installManagedStreamsLocked(app.stdout, app.stderr)
	instance.configureResizeLocked()
	instance.configureInputLocked()

	if err := instance.renderInitialLocked(); err != nil {
		return nil, err
	}

	return instance, nil
}

// Rerender replaces the root component and renders it again.
func (instance *Instance) Rerender(component ComponentFunc) error {
	instance.mu.Lock()
	defer instance.mu.Unlock()

	if instance.unmounted {
		return errors.New("ink instance is unmounted")
	}

	if instance.shouldThrottleRendersLocked() {
		return instance.scheduleRerenderLocked(component)
	}

	instance.cancelPendingRenderLocked()
	instance.app.component = component
	return instance.renderCurrentLocked()
}

// Clear removes the current rendered output while keeping the session active.
func (instance *Instance) Clear() error {
	instance.mu.Lock()
	defer instance.mu.Unlock()

	if instance.unmounted {
		return nil
	}

	return instance.clearLocked()
}

// Unmount clears the rendered output, restores runtime state, and ends the session.
func (instance *Instance) Unmount() error {
	instance.mu.Lock()
	defer instance.mu.Unlock()

	return instance.unmountLocked(nil)
}

// Cleanup detaches the instance from the managed render registry, if present.
func (instance *Instance) Cleanup() error {
	instance.mu.Lock()
	defer instance.mu.Unlock()

	if instance.registry != nil && instance.registryKey != nil {
		instance.registry.delete(instance.registryKey, instance)
		instance.registry = nil
		instance.registryKey = nil
	}

	return nil
}

// WaitUntilExit blocks until the session has been unmounted and returns the exit error, if any.
func (instance *Instance) WaitUntilExit() error {
	<-instance.exited

	instance.mu.Lock()
	defer instance.mu.Unlock()
	return instance.exitErr
}

// Sync updates the internal render state without redrawing the full output.
func (instance *Instance) Sync(output string, cursorPosition *CursorPosition) error {
	instance.mu.Lock()
	defer instance.mu.Unlock()

	if instance.unmounted {
		return nil
	}

	return instance.syncLocked(output, cursorPosition)
}

// WriteStdout writes external data to stdout while preserving the current Ink output.
func (instance *Instance) WriteStdout(data string) (int, error) {
	instance.mu.Lock()
	defer instance.mu.Unlock()

	if instance.unmounted || instance.stdout == nil {
		return 0, nil
	}

	return instance.writeExternalLocked(instance.stdout, data, true)
}

// WriteStderr writes external data to stderr while preserving the current Ink output.
func (instance *Instance) WriteStderr(data string) (int, error) {
	instance.mu.Lock()
	defer instance.mu.Unlock()

	if instance.unmounted || instance.stderr == nil {
		return 0, nil
	}

	return instance.writeExternalLocked(instance.stderr, data, true)
}

// HandleInput dispatches one raw stdin payload through the mounted input hooks.
func (instance *Instance) HandleInput(data string) error {
	instance.mu.Lock()
	defer instance.mu.Unlock()

	if instance.unmounted {
		return nil
	}

	return instance.handleInputLocked(data)
}

func (instance *Instance) renderCurrentLocked() (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			panicErr := recoveredPanicError(recovered)
			if unmountErr := instance.unmountLocked(panicErr); unmountErr != nil {
				err = unmountErr
				return
			}

			err = nil
		}
	}()

	prepared := instance.prepareRenderLocked()
	return instance.commitPreparedRenderLocked(prepared)
}

func (instance *Instance) renderInitialLocked() error {
	return instance.renderCurrentLocked()
}

func (instance *Instance) prepareRenderLocked() preparedRender {
	now := instance.now
	if now == nil {
		now = time.Now
	}

	start := now()
	output, staticOutput, nextStaticCounts := instance.app.RenderRuntimeOnce(instance.staticCounts)
	cursorPosition := cloneCursorPosition(instance.app.CursorPosition())
	instance.staticCounts = append(instance.staticCounts[:0], nextStaticCounts...)
	renderedOutput := output
	if !instance.debug && !instance.app.IsScreenReaderEnabled() && !instance.shouldRenderFullscreenLocked(output) {
		renderedOutput = ensureTrailingNewline(output)
	}

	return preparedRender{
		logicalOutput:       output,
		renderedOutput:      renderedOutput,
		staticOutput:        staticOutput,
		cursorPosition:      cursorPosition,
		shouldWrite:         instance.willWriteRenderLocked(output, renderedOutput, cursorPosition, staticOutput),
		shouldClearTerminal: instance.shouldClearTerminalRenderLocked(output, staticOutput),
		renderTime:          now().Sub(start),
		exitRequested:       instance.app.ExitRequested(),
		exitErr:             instance.app.ExitError(),
	}
}

func (instance *Instance) commitPreparedRenderLocked(prepared preparedRender) error {
	output := prepared.renderedOutput
	nextFullStaticOutput := instance.fullStaticOutput
	if prepared.staticOutput != "" {
		nextFullStaticOutput += prepared.staticOutput
	}

	wrapped := false

	if instance.shouldSynchronizeRenderLocked(prepared.shouldWrite) {
		if err := writePayload(instance.stdout, bsu); err != nil {
			if instance.exitErr == nil {
				instance.exitErr = err
			}
			return err
		}

		wrapped = true
	}

	if wrapped {
		defer func() {
			_ = writePayload(instance.stdout, esu)
		}()
	}

	if instance.usesCIMountedOutputLocked() {
		if prepared.staticOutput != "" {
			if err := writePayload(instance.stdout, prepared.staticOutput); err != nil {
				if instance.exitErr == nil {
					instance.exitErr = err
				}
				return err
			}
		}

		instance.previousLogicalOutput = prepared.logicalOutput
		instance.previousOutput = prepared.renderedOutput
		instance.previousLines = splitOutputLines(prepared.renderedOutput)
		instance.previousLineCount = outputLineCount(prepared.renderedOutput)
		instance.previousCursorPosition = nil
		instance.cursorWasShown = false
		instance.cursorHidden = false
		goto afterWrite
	}

	if instance.debug {
		output = nextFullStaticOutput + prepared.logicalOutput
	} else if prepared.shouldClearTerminal {
		if err := instance.writeClearTerminalRenderLocked(prepared.logicalOutput, prepared.renderedOutput, prepared.cursorPosition, nextFullStaticOutput); err != nil {
			if instance.exitErr == nil {
				instance.exitErr = err
			}
			return err
		}
		goto afterWrite
	} else if prepared.staticOutput != "" {
		if err := instance.clearLocked(); err != nil {
			if instance.exitErr == nil {
				instance.exitErr = err
			}
			return err
		}

		if err := writePayload(instance.stdout, prepared.staticOutput); err != nil {
			if instance.exitErr == nil {
				instance.exitErr = err
			}
			return err
		}
	}

	if prepared.shouldWrite {
		if err := instance.writeRenderLocked(prepared.logicalOutput, output, prepared.cursorPosition); err != nil {
			if instance.exitErr == nil {
				instance.exitErr = err
			}
			return err
		}
	}

afterWrite:
	instance.fullStaticOutput = nextFullStaticOutput

	if instance.onRender != nil {
		instance.onRender(RenderMetrics{RenderTime: prepared.renderTime})
	}

	instance.lastRenderAt = instance.now()

	if prepared.exitRequested {
		return instance.unmountLocked(prepared.exitErr)
	}

	if instance.app.consumeStateChange() {
		return instance.renderCurrentLocked()
	}

	return nil
}

func (instance *Instance) rerenderLocked(component ComponentFunc) error {
	instance.app.component = component
	return instance.renderCurrentLocked()
}

func (instance *Instance) applyOptionsLocked(options RenderOptions) {
	modeChanged := instance.debug != options.Debug || instance.incrementalRendering != options.IncrementalRendering
	if modeChanged {
		if err := instance.clearModeChangeOutputLocked(); err != nil && instance.exitErr == nil {
			instance.exitErr = err
		}
	}

	instance.onRender = options.OnRender
	instance.app.stdin = options.Stdin
	instance.installManagedStreamsLocked(options.Stdout, options.Stderr)
	instance.app.SetScreenReaderEnabled(options.ScreenReaderEnabled)
	instance.debug = options.Debug
	instance.maxFPS = normalizeMaxFPS(options.MaxFPS)
	instance.renderThrottle = throttleDuration(options.MaxFPS)
	instance.incrementalRendering = options.IncrementalRendering

	if options.Width > 0 || options.Height > 0 {
		instance.app.SetSize(options.Width, options.Height)
	}

	instance.configureResizeLocked()
	instance.configureInputLocked()

	if !instance.shouldThrottleRendersLocked() {
		instance.cancelPendingRenderLocked()
	}

	if modeChanged {
		instance.fullStaticOutput = ""
		instance.staticCounts = nil
		instance.previousLogicalOutput = ""
		instance.previousOutput = ""
		instance.previousLines = nil
		instance.previousLineCount = 0
		instance.previousCursorPosition = nil
		instance.cursorWasShown = false
		instance.cursorHidden = false
		instance.lastRenderAt = time.Time{}
	}
}

func (instance *Instance) shouldRenderFullscreenLocked(output string) bool {
	if instance.debug || instance.app.IsScreenReaderEnabled() || instance.stdout == nil {
		return false
	}

	if !shouldSynchronize(instance.stdout) {
		return false
	}

	viewportHeight := terminalViewportHeight(instance.stdout, instance.app.Height())
	if viewportHeight <= 0 {
		return false
	}

	return visibleLineCount(output) >= viewportHeight
}

func (instance *Instance) configureResizeLocked() {
	if instance.unsubscribeResize != nil {
		instance.unsubscribeResize()
		instance.unsubscribeResize = nil
	}

	instance.app.RefreshSizeFromWriter()

	if isCIEnvironment() {
		return
	}

	unsubscribe := subscribeResize(instance.stdout, instance.handleResize)
	if unsubscribe != nil {
		instance.unsubscribeResize = unsubscribe
	}
}

func (instance *Instance) configureInputLocked() {
	if instance.unsubscribeInput != nil {
		instance.unsubscribeInput()
		instance.unsubscribeInput = nil
	}

	unsubscribe := subscribeInput(instance.app.Stdin(), instance.handleInput)
	if unsubscribe != nil {
		instance.unsubscribeInput = unsubscribe
	}
}

func (instance *Instance) installManagedStreamsLocked(stdout io.Writer, stderr io.Writer) {
	instance.stdout = stdout
	instance.stderr = stderr
	instance.app.stdout = newSessionStreamProxy(instance, stdout, true)
	instance.app.stderr = newSessionStreamProxy(instance, stderr, true)
}

func newSessionStreamProxy(instance *Instance, target io.Writer, restoreToStdout bool) io.Writer {
	if target == nil {
		return nil
	}

	base := sessionStreamProxy{
		target:          target,
		restoreToStdout: restoreToStdout,
		instance:        instance,
	}

	switch typed := target.(type) {
	case uintptrFD:
		return &sessionStreamProxyWithUintptrFD{
			sessionStreamProxy: base,
			fd:                 typed,
		}
	case intFD:
		return &sessionStreamProxyWithIntFD{
			sessionStreamProxy: base,
			fd:                 typed,
		}
	default:
		return &base
	}
}

func (instance *Instance) writeHookOutputLocked(writer io.Writer, data string, restoreToStdout bool) (int, error) {
	if instance.unmounted {
		if writer == nil {
			return 0, nil
		}

		return io.WriteString(writer, data)
	}

	return instance.writeExternalLocked(writer, data, restoreToStdout)
}

func (instance *Instance) handleResize() {
	instance.mu.Lock()
	defer instance.mu.Unlock()

	if instance.unmounted {
		return
	}

	previousWidth := instance.app.Width()
	if !instance.app.RefreshSizeFromWriter() {
		return
	}

	instance.cancelPendingRenderLocked()

	if instance.app.Width() < previousWidth {
		if err := instance.clearLocked(); err != nil {
			if instance.exitErr == nil {
				instance.exitErr = err
			}
			return
		}
	}

	if err := instance.renderCurrentLocked(); err != nil && instance.exitErr == nil {
		instance.exitErr = err
	}
}

func (instance *Instance) handleInput(data string) {
	instance.mu.Lock()
	defer instance.mu.Unlock()

	if instance.unmounted {
		return
	}

	if err := instance.handleInputLocked(data); err != nil && instance.exitErr == nil {
		instance.exitErr = err
	}
}

func (instance *Instance) handleInputLocked(data string) error {
	inputValue, key, keys, err := inkinput.NormalizeHookInput(data)
	if err != nil {
		return err
	}

	if data == "\x1b" {
		instance.app.blurFocus()
	}

	hasShiftTab := false
	for _, keyName := range keys {
		if keyName == inkinput.KeyShiftTab {
			hasShiftTab = true
			break
		}
	}

	for _, keyName := range keys {
		switch keyName {
		case inkinput.KeyShiftTab:
			instance.app.focusPrevious()
		case inkinput.KeyTab:
			if !hasShiftTab {
				instance.app.focusNext()
			}
		}
	}

	if instance.app.hooksCtx.DispatchInput(inputValue, key, keys) {
		instance.app.Exit()
	}

	if instance.app.ExitRequested() {
		return instance.unmountLocked(instance.app.ExitError())
	}

	if instance.app.consumeStateChange() {
		return instance.renderCurrentLocked()
	}

	return nil
}

func (instance *Instance) writeRenderLocked(logicalOutput string, renderedOutput string, cursorPosition *CursorPosition) error {
	if instance.app.IsScreenReaderEnabled() {
		return instance.writeScreenReaderRenderLocked(logicalOutput, renderedOutput)
	}

	if instance.debug {
		return instance.writeDebugRenderLocked(logicalOutput, renderedOutput, cursorPosition)
	}

	if instance.incrementalRendering {
		return instance.writeIncrementalRenderLocked(logicalOutput, renderedOutput, cursorPosition)
	}

	return instance.writeStandardRenderLocked(logicalOutput, renderedOutput, cursorPosition)
}

func (instance *Instance) writeScreenReaderRenderLocked(logicalOutput string, renderedOutput string) error {
	if instance.stdout == nil {
		return nil
	}

	if renderedOutput == instance.previousOutput {
		return nil
	}

	payload := ""
	if instance.previousLineCount > 0 {
		payload += ansiEraseLines(instance.previousLineCount)
	}

	payload += renderedOutput
	if err := writePayload(instance.stdout, payload); err != nil {
		return err
	}

	instance.previousLogicalOutput = logicalOutput
	instance.previousOutput = renderedOutput
	instance.previousLines = splitOutputLines(renderedOutput)
	instance.previousLineCount = outputLineCount(renderedOutput)
	instance.previousCursorPosition = nil
	instance.cursorWasShown = false
	instance.cursorHidden = false

	return nil
}

func (instance *Instance) writeDebugRenderLocked(logicalOutput string, renderedOutput string, cursorPosition *CursorPosition) error {
	if instance.stdout == nil {
		return nil
	}

	if renderedOutput == instance.previousOutput {
		instance.previousCursorPosition = cloneCursorPosition(cursorPosition)
		instance.cursorWasShown = cursorPosition != nil
		return nil
	}

	if err := writePayload(instance.stdout, renderedOutput); err != nil {
		return err
	}

	instance.previousLogicalOutput = logicalOutput
	instance.previousOutput = renderedOutput
	instance.previousLines = splitOutputLines(renderedOutput)
	instance.previousLineCount = outputLineCount(renderedOutput)
	instance.previousCursorPosition = cloneCursorPosition(cursorPosition)
	instance.cursorWasShown = cursorPosition != nil

	return nil
}

func (instance *Instance) writeClearTerminalRenderLocked(logicalOutput string, renderedOutput string, cursorPosition *CursorPosition, fullStaticOutput string) error {
	if instance.stdout == nil {
		return nil
	}

	if !instance.cursorHidden {
		if err := writePayload(instance.stdout, hideCursorEscape); err != nil {
			return err
		}

		instance.cursorHidden = true
	}

	visibleLines := visibleLineCount(renderedOutput)
	payload := clearTerminalEscape + fullStaticOutput + logicalOutput + buildCursorSuffix(visibleLines, cursorPosition)
	if err := writePayload(instance.stdout, payload); err != nil {
		return err
	}

	instance.previousLogicalOutput = logicalOutput
	instance.previousOutput = renderedOutput
	instance.previousLines = splitOutputLines(renderedOutput)
	instance.previousLineCount = outputLineCount(renderedOutput)
	instance.previousCursorPosition = cloneCursorPosition(cursorPosition)
	instance.cursorWasShown = cursorPosition != nil

	return nil
}

func (instance *Instance) usesCIMountedOutputLocked() bool {
	return !instance.debug && isCIEnvironment()
}

func (instance *Instance) shouldThrottleRendersLocked() bool {
	return instance.renderThrottle > 0 && !instance.debug && !instance.app.IsScreenReaderEnabled()
}

func (instance *Instance) shouldSynchronizeRenderLocked(shouldWrite bool) bool {
	return shouldWrite && instance.shouldThrottleRendersLocked() && shouldSynchronize(instance.stdout)
}

func (instance *Instance) shouldClearTerminalRenderLocked(logicalOutput string, staticOutput string) bool {
	if instance.debug || instance.app.IsScreenReaderEnabled() || instance.stdout == nil {
		return false
	}

	if logicalOutput == instance.previousLogicalOutput && staticOutput == "" {
		return false
	}

	if !shouldSynchronize(instance.stdout) {
		return false
	}

	viewportHeight := terminalViewportHeight(instance.stdout, instance.app.Height())
	if viewportHeight <= 0 {
		return false
	}

	return visibleLineCount(instance.previousOutput) >= viewportHeight
}

func (instance *Instance) willWriteRenderLocked(logicalOutput string, output string, cursorPosition *CursorPosition, staticOutput string) bool {
	if staticOutput != "" {
		return true
	}

	if instance.debug {
		return output != instance.previousOutput
	}

	return logicalOutput != instance.previousLogicalOutput || cursorPositionChanged(cursorPosition, instance.previousCursorPosition)
}

func (instance *Instance) scheduleRerenderLocked(component ComponentFunc) error {
	now := instance.now()
	instance.app.component = component
	prepared := instance.prepareRenderLocked()
	if prepared.staticOutput != "" || prepared.exitRequested || instance.lastRenderAt.IsZero() || now.Sub(instance.lastRenderAt) >= instance.renderThrottle {
		instance.cancelPendingRenderLocked()
		return instance.commitPreparedRenderLocked(prepared)
	}

	instance.pendingRender = &prepared
	remaining := instance.renderThrottle - now.Sub(instance.lastRenderAt)
	if remaining <= 0 {
		remaining = time.Millisecond
	}

	if instance.pendingTimer == nil {
		instance.pendingTimer = instance.afterFunc(remaining, instance.runPendingRender)
	}

	return nil
}

func (instance *Instance) runPendingRender() {
	instance.mu.Lock()
	defer instance.mu.Unlock()

	instance.pendingTimer = nil
	if instance.unmounted || instance.pendingRender == nil {
		return
	}

	prepared := *instance.pendingRender
	instance.pendingRender = nil

	if err := instance.commitPreparedRenderLocked(prepared); err != nil && instance.exitErr == nil {
		instance.exitErr = err
	}
}

func (instance *Instance) cancelPendingRenderLocked() {
	if instance.pendingTimer != nil {
		instance.pendingTimer.Stop()
		instance.pendingTimer = nil
	}

	instance.pendingRender = nil
}

func (instance *Instance) flushPendingRenderLocked() error {
	if instance.pendingRender == nil {
		instance.cancelPendingRenderLocked()
		return nil
	}

	if instance.pendingTimer != nil {
		instance.pendingTimer.Stop()
		instance.pendingTimer = nil
	}

	prepared := *instance.pendingRender
	instance.pendingRender = nil
	return instance.commitPreparedRenderLocked(prepared)
}

func (instance *Instance) writeStandardRenderLocked(logicalOutput string, output string, cursorPosition *CursorPosition) error {
	if instance.stdout == nil {
		return nil
	}

	if !instance.cursorHidden {
		if err := writePayload(instance.stdout, hideCursorEscape); err != nil {
			return err
		}

		instance.cursorHidden = true
	}

	visibleLines := visibleLineCount(output)
	cursorChanged := cursorPositionChanged(cursorPosition, instance.previousCursorPosition)

	if output == instance.previousOutput && cursorChanged {
		sequence := buildCursorOnlySequence(
			instance.cursorWasShown,
			instance.previousLineCount,
			instance.previousCursorPosition,
			visibleLines,
			cursorPosition,
		)

		if err := writePayload(instance.stdout, sequence); err != nil {
			return err
		}
	} else if output != instance.previousOutput {
		payload := buildReturnToBottomPrefix(
			instance.cursorWasShown,
			instance.previousLineCount,
			instance.previousCursorPosition,
		) +
			ansiEraseLines(instance.previousLineCount) +
			output +
			buildCursorSuffix(visibleLines, cursorPosition)

		if err := writePayload(instance.stdout, payload); err != nil {
			return err
		}
	}

	instance.previousLogicalOutput = logicalOutput
	instance.previousOutput = output
	instance.previousLines = splitOutputLines(output)
	instance.previousLineCount = outputLineCount(output)
	instance.previousCursorPosition = cloneCursorPosition(cursorPosition)
	instance.cursorWasShown = cursorPosition != nil

	return nil
}

func (instance *Instance) writeIncrementalRenderLocked(logicalOutput string, output string, cursorPosition *CursorPosition) error {
	if instance.stdout == nil {
		return nil
	}

	if !instance.cursorHidden {
		if err := writePayload(instance.stdout, hideCursorEscape); err != nil {
			return err
		}

		instance.cursorHidden = true
	}

	visibleLines := visibleLineCount(output)
	cursorChanged := cursorPositionChanged(cursorPosition, instance.previousCursorPosition)
	if output == instance.previousOutput && !cursorChanged {
		return nil
	}

	nextLines := splitOutputLines(output)
	if output == instance.previousOutput && cursorChanged {
		sequence := buildCursorOnlySequence(
			instance.cursorWasShown,
			len(instance.previousLines),
			instance.previousCursorPosition,
			visibleLines,
			cursorPosition,
		)
		if err := writePayload(instance.stdout, sequence); err != nil {
			return err
		}

		instance.previousCursorPosition = cloneCursorPosition(cursorPosition)
		instance.cursorWasShown = cursorPosition != nil
		return nil
	}

	returnPrefix := buildReturnToBottomPrefix(
		instance.cursorWasShown,
		len(instance.previousLines),
		instance.previousCursorPosition,
	)

	if output == "\n" || instance.previousOutput == "" {
		payload := returnPrefix + ansiEraseLines(len(instance.previousLines)) + output + buildCursorSuffix(visibleLines, cursorPosition)
		if err := writePayload(instance.stdout, payload); err != nil {
			return err
		}

		instance.previousLogicalOutput = logicalOutput
		instance.previousOutput = output
		instance.previousLines = nextLines
		instance.previousLineCount = outputLineCount(output)
		instance.previousCursorPosition = cloneCursorPosition(cursorPosition)
		instance.cursorWasShown = cursorPosition != nil
		return nil
	}

	previousVisible := visibleLineCount(instance.previousOutput)
	hasTrailingNewline := outputLineCount(output) > visibleLines
	buffer := make([]string, 0, visibleLines+8)
	buffer = append(buffer, returnPrefix)

	if visibleLines < previousVisible {
		extraSlot := 0
		if strings.HasSuffix(instance.previousOutput, "\n") {
			extraSlot = 1
		}

		buffer = append(buffer, ansiEraseLines(previousVisible-visibleLines+extraSlot))
		buffer = append(buffer, ansiCursorUp(visibleLines))
	} else {
		buffer = append(buffer, ansiCursorUp(previousVisible-1))
	}

	for index := 0; index < visibleLines; index++ {
		isLastLine := index == visibleLines-1

		if index < len(instance.previousLines) && index < len(nextLines) && nextLines[index] == instance.previousLines[index] {
			if !isLastLine || hasTrailingNewline {
				buffer = append(buffer, ansiCursorNextLine())
			}

			continue
		}

		line := ""
		if index < len(nextLines) {
			line = nextLines[index]
		}

		suffix := "\n"
		if isLastLine && !hasTrailingNewline {
			suffix = ""
		}

		buffer = append(buffer, ansiCursorTo(0)+line+ansiEraseEndLine()+suffix)
	}

	buffer = append(buffer, buildCursorSuffix(visibleLines, cursorPosition))
	if err := writePayload(instance.stdout, strings.Join(buffer, "")); err != nil {
		return err
	}

	instance.previousLogicalOutput = logicalOutput
	instance.previousOutput = output
	instance.previousLines = nextLines
	instance.previousLineCount = outputLineCount(output)
	instance.previousCursorPosition = cloneCursorPosition(cursorPosition)
	instance.cursorWasShown = cursorPosition != nil

	return nil
}

func (instance *Instance) clearLocked() error {
	if instance.usesCIMountedOutputLocked() {
		return nil
	}

	if instance.debug {
		instance.previousLogicalOutput = ""
		instance.previousOutput = ""
		instance.previousLines = nil
		instance.previousLineCount = 0
		instance.previousCursorPosition = nil
		instance.cursorWasShown = false
		return nil
	}

	if instance.stdout == nil {
		instance.previousLogicalOutput = ""
		instance.previousOutput = ""
		instance.previousLines = nil
		instance.previousLineCount = 0
		instance.previousCursorPosition = nil
		instance.cursorWasShown = false
		return nil
	}

	lineCount := instance.previousLineCount
	if instance.incrementalRendering {
		lineCount = len(instance.previousLines)
	}

	payload := buildReturnToBottomPrefix(
		instance.cursorWasShown,
		lineCount,
		instance.previousCursorPosition,
	) + ansiEraseLines(lineCount)

	if err := writePayload(instance.stdout, payload); err != nil {
		return err
	}

	instance.previousOutput = ""
	instance.previousLogicalOutput = ""
	instance.previousLines = nil
	instance.previousLineCount = 0
	instance.previousCursorPosition = nil
	instance.cursorWasShown = false

	return nil
}

func (instance *Instance) clearModeChangeOutputLocked() error {
	if instance.usesCIMountedOutputLocked() || instance.stdout == nil {
		return nil
	}

	if instance.debug {
		payload := ansiEraseLines(outputLineCount(instance.previousOutput))
		if payload == "" {
			return nil
		}

		return writePayload(instance.stdout, payload)
	}

	totalOutput := instance.fullStaticOutput + instance.previousOutput
	dynamicLineCount := instance.previousLineCount
	if instance.incrementalRendering {
		dynamicLineCount = len(instance.previousLines)
	}

	payload := buildReturnToBottomPrefix(
		instance.cursorWasShown,
		dynamicLineCount,
		instance.previousCursorPosition,
	) + ansiEraseLines(outputLineCount(totalOutput))
	if payload == "" {
		return nil
	}

	return writePayload(instance.stdout, payload)
}

func (instance *Instance) syncLocked(output string, cursorPosition *CursorPosition) error {
	if instance.usesCIMountedOutputLocked() {
		instance.previousLogicalOutput = output
		instance.previousOutput = output
		instance.previousLines = splitOutputLines(output)
		instance.previousLineCount = outputLineCount(output)
		instance.previousCursorPosition = nil
		instance.cursorWasShown = false
		instance.cursorHidden = false
		return nil
	}

	if instance.app.IsScreenReaderEnabled() {
		instance.previousLogicalOutput = output
		instance.previousOutput = output
		instance.previousLines = splitOutputLines(output)
		instance.previousLineCount = outputLineCount(output)
		instance.previousCursorPosition = nil
		instance.cursorWasShown = false
		return nil
	}

	if instance.debug {
		instance.previousLogicalOutput = output
		instance.previousOutput = output
		instance.previousLines = splitOutputLines(output)
		instance.previousLineCount = outputLineCount(output)
		instance.previousCursorPosition = cloneCursorPosition(cursorPosition)
		instance.cursorWasShown = cursorPosition != nil
		return nil
	}

	if instance.stdout == nil {
		instance.previousLogicalOutput = output
		instance.previousOutput = output
		instance.previousLines = splitOutputLines(output)
		instance.previousLineCount = outputLineCount(output)
		instance.previousCursorPosition = cloneCursorPosition(cursorPosition)
		instance.cursorWasShown = cursorPosition != nil
		return nil
	}

	if cursorPosition == nil && instance.cursorWasShown {
		if err := writePayload(instance.stdout, hideCursorEscape); err != nil {
			return err
		}
	}

	if cursorPosition != nil {
		if err := writePayload(instance.stdout, buildCursorSuffix(visibleLineCount(output), cursorPosition)); err != nil {
			return err
		}
	}

	instance.previousLogicalOutput = output
	instance.previousOutput = output
	instance.previousLines = splitOutputLines(output)
	instance.previousLineCount = outputLineCount(output)
	instance.previousCursorPosition = cloneCursorPosition(cursorPosition)
	instance.cursorWasShown = cursorPosition != nil

	return nil
}

func (instance *Instance) writeExternalLocked(writer io.Writer, data string, restoreToStdout bool) (int, error) {
	if instance.usesCIMountedOutputLocked() {
		return io.WriteString(writer, data)
	}

	if instance.debug {
		written, err := io.WriteString(writer, data)
		if err != nil {
			return written, err
		}

		if restoreToStdout && instance.previousOutput != "" {
			if err := writePayload(instance.stdout, instance.previousOutput); err != nil {
				return written, err
			}
		}

		return written, nil
	}

	wrapped := shouldSynchronize(instance.stdout)
	if wrapped {
		if err := writePayload(instance.stdout, bsu); err != nil {
			return 0, err
		}

		defer func() {
			_ = writePayload(instance.stdout, esu)
		}()
	}

	saved := instance.captureRenderStateLocked()
	if saved.output != "" || saved.cursorWasShown {
		if err := instance.clearLocked(); err != nil {
			return 0, err
		}
	}

	written, err := io.WriteString(writer, data)
	if err != nil {
		return written, err
	}

	if restoreToStdout {
		if err := instance.restoreRenderStateLocked(saved); err != nil {
			return written, err
		}
	}

	return written, nil
}

func (instance *Instance) captureRenderStateLocked() renderState {
	return renderState{
		logicalOutput:  instance.previousLogicalOutput,
		output:         instance.previousOutput,
		lines:          append([]string(nil), instance.previousLines...),
		lineCount:      instance.previousLineCount,
		cursorPosition: cloneCursorPosition(instance.previousCursorPosition),
		cursorWasShown: instance.cursorWasShown,
	}
}

func (instance *Instance) restoreRenderStateLocked(state renderState) error {
	if state.output == "" {
		if state.cursorWasShown && state.cursorPosition != nil {
			if err := writePayload(instance.stdout, buildCursorSuffix(visibleLineCount(state.output), state.cursorPosition)); err != nil {
				return err
			}
		}

		instance.previousLogicalOutput = ""
		instance.previousOutput = ""
		instance.previousLines = nil
		instance.previousLineCount = 0
		instance.previousCursorPosition = cloneCursorPosition(state.cursorPosition)
		instance.cursorWasShown = state.cursorWasShown
		return nil
	}

	return instance.writeRenderLocked(state.logicalOutput, state.output, state.cursorPosition)
}

func (instance *Instance) unmountLocked(exitErr error) error {
	if exitErr != nil && instance.exitErr == nil {
		instance.exitErr = exitErr
	}

	if instance.unmounted {
		instance.closeExitedLocked()
		return nil
	}

	if err := instance.flushPendingRenderLocked(); err != nil && instance.exitErr == nil {
		instance.exitErr = err
	}
	if instance.unmounted {
		instance.closeExitedLocked()
		return nil
	}

	instance.unmounted = true
	instance.app.stdout = instance.stdout
	instance.app.stderr = instance.stderr
	if instance.unsubscribeResize != nil {
		instance.unsubscribeResize()
		instance.unsubscribeResize = nil
	}
	if instance.unsubscribeInput != nil {
		instance.unsubscribeInput()
		instance.unsubscribeInput = nil
	}

	var firstErr error
	if instance.usesCIMountedOutputLocked() {
		if instance.stdout != nil && instance.previousLogicalOutput != "" {
			if err := writePayload(instance.stdout, instance.previousLogicalOutput+"\n"); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	} else if instance.app.IsScreenReaderEnabled() {
		instance.previousLogicalOutput = ""
		instance.previousOutput = ""
		instance.previousLines = nil
		instance.previousLineCount = 0
		instance.previousCursorPosition = nil
		instance.cursorWasShown = false
	} else {
		if err := instance.clearLocked(); err != nil && firstErr == nil {
			firstErr = err
		}

		if instance.cursorHidden && instance.stdout != nil {
			if err := writePayload(instance.stdout, showCursorEscape); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	instance.cursorHidden = false

	instance.app.hooksCtx.RunCleanup()

	if err := instance.disableRawModeOnExitLocked(); err != nil && firstErr == nil {
		firstErr = err
	}

	if instance.registry != nil && instance.registryKey != nil {
		instance.registry.delete(instance.registryKey, instance)
		instance.registry = nil
		instance.registryKey = nil
	}

	if err := flushWriter(instance.stdout); err != nil && firstErr == nil {
		firstErr = err
	}

	if err := flushWriter(instance.stderr); err != nil && firstErr == nil {
		firstErr = err
	}

	if err := waitWriter(instance.stdout); err != nil && firstErr == nil {
		firstErr = err
	}

	if err := waitWriter(instance.stderr); err != nil && firstErr == nil {
		firstErr = err
	}

	if firstErr != nil && instance.exitErr == nil {
		instance.exitErr = firstErr
	}

	instance.closeExitedLocked()
	return firstErr
}

func (instance *Instance) disableRawModeOnExitLocked() error {
	if instance.app == nil {
		return nil
	}

	if instance.app.rawModeUsers == 0 && instance.app.rawState == nil {
		return nil
	}

	// Session shutdown must restore the terminal even if manual SetRawMode(true)
	// calls were nested without balanced cleanup callbacks.
	instance.app.rawModeUsers = 1
	return instance.app.SetRawMode(false)
}

func (instance *Instance) closeExitedLocked() {
	instance.exitOnce.Do(func() {
		close(instance.exited)
	})
}

func cloneCursorPosition(position *CursorPosition) *CursorPosition {
	if position == nil {
		return nil
	}

	copy := *position
	return &copy
}

func recoveredPanicError(recovered any) error {
	if err, ok := recovered.(error); ok {
		return err
	}

	return fmt.Errorf("%v", recovered)
}

func writePayload(writer io.Writer, payload string) error {
	if writer == nil {
		return nil
	}

	_, err := io.WriteString(writer, payload)
	return err
}

func normalizeMaxFPS(fps int) int {
	if fps <= 0 {
		return 0
	}

	return fps
}

func throttleDuration(fps int) time.Duration {
	normalized := normalizeMaxFPS(fps)
	if normalized == 0 {
		return 0
	}

	return time.Duration((1000+normalized-1)/normalized) * time.Millisecond
}
