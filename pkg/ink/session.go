package ink

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	inkinput "github.com/dh-kam/ink-go/pkg/input"
)

// RenderMetrics describes one completed render pass.
type RenderMetrics struct {
	RenderTime time.Duration
}

// DefaultMaxFPS matches upstream Ink's default render update cap.
const DefaultMaxFPS = 30
const maxPendingStateSettles = 50

// RenderOptions configures a mounted Ink session.
type RenderOptions struct {
	AppOptions
	Debug       bool
	ExitOnCtrlC *bool
	// MaxFPS is the legacy frame cap. Values <= 0 disable throttling for
	// backwards compatibility. Prefer MaxFPSLimit when omitted vs explicit zero
	// needs to be distinguishable.
	MaxFPS int
	// MaxFPSLimit overrides MaxFPS when set. This provides tri-state semantics:
	// nil keeps MaxFPS legacy behavior, a pointer containing DefaultMaxFPS
	// matches upstream Ink's default, and a pointer containing 0 explicitly
	// disables throttling.
	MaxFPSLimit          *int
	IncrementalRendering bool
	// CellLevelDiff opts into cell-by-cell dirty-rect repaints layered on
	// top of IncrementalRendering. When enabled, frames are parsed into
	// per-cell records and only changed cells are emitted (with cursor
	// jumps and SGR transitions) instead of rewriting whole lines.
	//
	// Opt-in because it changes the exact bytes produced on the wire — many
	// existing tests assert on full-line rewrite shapes (cursorTo(0) +
	// payload + eraseEndLine), so flipping cell-level diff on by default
	// would break them. Falls back to the line-level path on frame size
	// changes or on any non-trivial ANSI input the parser cannot represent.
	CellLevelDiff bool
	OnRender      func(RenderMetrics)
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
	exitOnCtrlC            bool
	maxFPS                 int
	renderThrottle         time.Duration
	incrementalRendering   bool
	cellLevelDiff          bool
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
	renderCache            *renderTracker
	// pendingPaste accumulates the body of a bracketed-paste sequence whose
	// start marker (\x1b[200~) arrived in a TTY read that did not contain
	// the matching end marker (\x1b[201~). Subsequent reads append to this
	// buffer until the end marker is observed, at which point the full
	// payload is dispatched as a single paste event. pastePending tracks
	// whether a paste is currently in flight (independent of pendingPaste
	// length so that the empty-body case is preserved).
	pendingPaste []byte
	pastePending bool
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

type unmountOptions struct {
	clearOutput bool
}

// Mount creates a mounted Ink session and immediately renders it.
func Mount(component ComponentFunc) (*Instance, error) {
	return MountWithOptions(component, RenderOptions{})
}

// MountWithOptions creates a mounted Ink session with explicit render options.
func MountWithOptions(component ComponentFunc, options RenderOptions) (*Instance, error) {
	app := NewAppWithOptions(component, options.AppOptions)
	maxFPS := normalizeRenderMaxFPS(options)
	instance := &Instance{
		app:                  app,
		onRender:             options.OnRender,
		debug:                options.Debug,
		exitOnCtrlC:          normalizeExitOnCtrlC(options.ExitOnCtrlC),
		maxFPS:               maxFPS,
		renderThrottle:       throttleDuration(maxFPS),
		incrementalRendering: options.IncrementalRendering,
		cellLevelDiff:        options.CellLevelDiff,
		exited:               make(chan struct{}),
		now:                  time.Now,
		afterFunc: func(delay time.Duration, fn func()) scheduledTimer {
			return time.AfterFunc(delay, fn)
		},
		renderCache: newRenderTracker(RenderToString),
	}

	instance.installManagedStreamsLocked(app.stdout, app.stderr)
	app.hooksCtx.SetWorkScheduler(instance.scheduleHookWork)
	instance.configureResizeLocked()
	instance.configureInputLocked()

	if err := instance.renderInitialLocked(); err != nil {
		return nil, err
	}

	return instance, nil
}

func (instance *Instance) scheduleHookWork(work func()) {
	if work == nil {
		return
	}

	afterFunc := instance.afterFunc
	if afterFunc == nil {
		afterFunc = func(delay time.Duration, fn func()) scheduledTimer {
			return time.AfterFunc(delay, fn)
		}
	}

	afterFunc(time.Millisecond, func() {
		instance.runHookWork(work)
	})
}

func (instance *Instance) runHookWork(work func()) (err error) {
	instance.mu.Lock()
	defer instance.mu.Unlock()

	if instance.unmounted {
		return nil
	}

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

	work()

	if instance.app.ExitRequested() {
		return instance.unmountDoneLocked(instance.app.ExitError())
	}

	if instance.app.consumeStateChange() {
		return instance.renderCurrentLocked()
	}

	return nil
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

// Announcer returns the runtime aria-live channel attached to this
// session. Multiple subscribers (UseAnnounce callers, external producers)
// share this single instance, so dispatches from different parts of the
// app aggregate into the same per-frame queue.
func (instance *Instance) Announcer() *Announcer {
	if instance == nil || instance.app == nil {
		return nil
	}

	return instance.app.Announcer()
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

// WaitUntilExit blocks until the session has been explicitly unmounted or exited
// and returns the exit error, if any. Go has no Node beforeExit equivalent, so
// natural event-loop completion cannot be inferred by the runtime.
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
	sections, fresh := instance.app.renderRuntimeOnceWithCache(instance.staticCounts, instance.renderCache)
	output := sections.Output
	staticOutput := sections.StaticDeltaOutput
	nextStaticCounts := sections.StaticCounts
	cursorPosition := cloneCursorPosition(instance.app.CursorPosition())
	instance.staticCounts = append(instance.staticCounts[:0], nextStaticCounts...)
	renderedOutput := output
	if !instance.debug && !instance.app.IsScreenReaderEnabled() && !instance.shouldRenderFullscreenLocked(output) {
		renderedOutput = output + "\n"
	}

	prepared := preparedRender{
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
	if prepared.shouldClearTerminal {
		prepared.shouldWrite = true
	}

	// On a tracker cache hit (no patches relative to last render, identical
	// inputs, no exit/state changes pending) the rendered bytes are
	// guaranteed identical to what we already wrote — suppress the write
	// so idle ticks do not redraw the screen. We only do this when the
	// previously-committed output already matches the cached output: a
	// preceding Clear() or mode change invalidates that assumption and
	// we still need to repaint even though the tree is unchanged.
	//
	// Debug mode opts out of this suppression: upstream Ink emits one append
	// per render even for byte-identical frames, so callers tailing the
	// debug log can correlate writes 1:1 with onRender invocations.
	cursorMatches := (cursorPosition == nil && instance.previousCursorPosition == nil) ||
		(cursorPosition != nil && instance.previousCursorPosition != nil && *cursorPosition == *instance.previousCursorPosition)
	if !instance.debug && !fresh && !prepared.exitRequested && staticOutput == "" &&
		instance.previousLogicalOutput == output && cursorMatches && !prepared.shouldClearTerminal {
		prepared.shouldWrite = false
		prepared.shouldClearTerminal = false
	}

	return prepared
}

func (instance *Instance) commitPreparedRenderLocked(prepared preparedRender) error {
	// Mirror upstream Ink's `hasStaticOutput = staticOutput && staticOutput !== '\n'`
	// filter from ink/src/ink.tsx onRender: a static delta consisting solely
	// of a trailing newline carries no real content and must not grow the
	// fullStaticOutput buffer or trigger the clear+rewrite cycle that a real
	// static append would require.
	if prepared.staticOutput == "\n" {
		prepared.staticOutput = ""
		prepared.shouldWrite = instance.willWriteRenderLocked(prepared.logicalOutput, prepared.renderedOutput, prepared.cursorPosition, prepared.staticOutput)
		prepared.shouldClearTerminal = instance.shouldClearTerminalRenderLocked(prepared.logicalOutput, prepared.staticOutput)
		if prepared.shouldClearTerminal {
			prepared.shouldWrite = true
		}
	}

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
	modeChanged := instance.debug != options.Debug ||
		instance.incrementalRendering != options.IncrementalRendering ||
		instance.cellLevelDiff != options.CellLevelDiff
	if modeChanged {
		if err := instance.clearModeChangeOutputLocked(); err != nil && instance.exitErr == nil {
			instance.exitErr = err
		}
	}

	instance.onRender = options.OnRender
	instance.exitOnCtrlC = normalizeExitOnCtrlC(options.ExitOnCtrlC)
	instance.app.stdin = options.Stdin
	instance.installManagedStreamsLocked(options.Stdout, options.Stderr)
	instance.app.SetScreenReaderEnabled(options.ScreenReaderEnabled)
	instance.debug = options.Debug
	maxFPS := normalizeRenderMaxFPS(options)
	instance.maxFPS = maxFPS
	instance.renderThrottle = throttleDuration(maxFPS)
	instance.incrementalRendering = options.IncrementalRendering
	instance.cellLevelDiff = options.CellLevelDiff

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
		// Mode change invalidates everything we knew about the visible
		// output — drop the render cache so the next commit always writes.
		instance.renderCache.Reset()
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
	instance.renderCache.Reset()

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
	// Route SGR 1006 and legacy X10 mouse reports first so subscribers via
	// UseMouse get them before key normalization treats the bytes as
	// regular escape sequences. consumeMouseFrames peels off as many
	// leading mouse frames as it can find — this matters because raw TTY
	// reads frequently coalesce multiple mouse-move events, or stitch a
	// mouse frame onto a subsequent keypress, into a single chunk.
	mouseDispatched, leftover := consumeMouseFramesWithManager(data, instance.app.mouseManager)
	if mouseDispatched && leftover == "" {
		if instance.app.consumeStateChange() {
			return instance.renderCurrentLocked()
		}
		return nil
	}
	if mouseDispatched {
		// Keep the trailing non-mouse bytes flowing through key handling.
		// We may also need to render after dispatching the mouse event(s),
		// but renderCurrentLocked is tail-called below after key handling
		// completes, so a single render pass covers both.
		data = leftover
	}

	// Peel any bracketed-paste segments off the front of the chunk so the
	// pasted body is delivered as a single hook event with no synthetic
	// modifier flags, even when the kernel coalesces the paste with leading
	// or trailing keystrokes into a single TTY read. Upstream Ink does not
	// do this — pasted markers leak into the useInput callback's input
	// string verbatim there — but goink intentionally diverges so terminal
	// pastes round-trip cleanly.
	for len(data) > 0 {
		// If a paste is already in flight from an earlier chunk, accumulate
		// bytes until we observe the end marker. The end marker may itself
		// be split across chunks (e.g. "\x1b[201" in one read, "~" in the
		// next), but appending the raw bytes and re-scanning the buffer on
		// every iteration handles that case naturally.
		if instance.pastePending {
			// Concatenate so an end marker that straddles the chunk
			// boundary (e.g. "\x1b[20" in the buffer, "1~" in data) is
			// detected on the joined string.
			combined := string(instance.pendingPaste) + data
			endIdx := strings.Index(combined, bracketedPasteEndLiteral)
			if endIdx < 0 {
				instance.pendingPaste = []byte(combined)
				data = ""
				break
			}
			body := combined[:endIdx]
			rest := combined[endIdx+len(bracketedPasteEndLiteral):]
			instance.pendingPaste = nil
			instance.pastePending = false
			if err := instance.dispatchPasteLocked(body); err != nil {
				return err
			}
			if instance.unmounted {
				return nil
			}
			data = rest
			continue
		}

		consumed, payload, hadPaste, rest := splitBracketedPaste(data)
		if hadPaste {
			if consumed != "" {
				if err := instance.dispatchKeypressLocked(consumed); err != nil {
					return err
				}
				if instance.unmounted {
					return nil
				}
			}
			if err := instance.dispatchPasteLocked(payload); err != nil {
				return err
			}
			if instance.unmounted {
				return nil
			}
			data = rest
			continue
		}

		// Detect a start marker without a matching end marker in this
		// chunk — open a pending paste and buffer the trailing bytes
		// (everything after the start marker) until subsequent reads
		// deliver the end marker. Bytes preceding the start marker are
		// still dispatched as ordinary keypresses.
		if openIdx := strings.Index(data, bracketedPasteStartLiteral); openIdx >= 0 {
			leading := data[:openIdx]
			tail := data[openIdx+len(bracketedPasteStartLiteral):]
			if leading != "" {
				if err := instance.dispatchKeypressLocked(leading); err != nil {
					return err
				}
				if instance.unmounted {
					return nil
				}
			}
			instance.pastePending = true
			instance.pendingPaste = append(instance.pendingPaste[:0], []byte(tail)...)
			data = ""
			break
		}

		break
	}

	if data == "" {
		if instance.app.consumeStateChange() {
			return instance.renderCurrentLocked()
		}
		return nil
	}

	return instance.dispatchKeypressLocked(data)
}

// splitBracketedPaste finds the first complete bracketed-paste sequence in
// data and reports the bytes preceding the paste-start marker, the inner
// payload, whether a complete paste was found, and the bytes after the
// paste-end marker. When no complete paste is present, hadPaste is false and
// the other return values are zero.
func splitBracketedPaste(data string) (preceding string, payload string, hadPaste bool, rest string) {
	startIdx := strings.Index(data, bracketedPasteStartLiteral)
	if startIdx < 0 {
		return "", "", false, ""
	}

	tail := data[startIdx+len(bracketedPasteStartLiteral):]
	endIdx := strings.Index(tail, bracketedPasteEndLiteral)
	if endIdx < 0 {
		// No matching end marker in this chunk — leave the bytes intact
		// so the regular keypress dispatcher delivers them verbatim. We
		// intentionally avoid buffering across reads to keep the input
		// pipeline stateless; terminals that emit start without end in a
		// single chunk are rare enough that this falls back gracefully.
		return "", "", false, ""
	}

	return data[:startIdx], tail[:endIdx], true, tail[endIdx+len(bracketedPasteEndLiteral):]
}

const (
	bracketedPasteStartLiteral = "\x1b[200~"
	bracketedPasteEndLiteral   = "\x1b[201~"
)

func (instance *Instance) dispatchPasteLocked(payload string) error {
	if instance.app.hooksCtx.DispatchInput(payload, inkinput.HookKey{}, nil) {
		instance.app.Exit()
	}

	if instance.app.ExitRequested() {
		return instance.unmountDoneLocked(instance.app.ExitError())
	}

	if instance.app.consumeStateChange() {
		return instance.renderCurrentLocked()
	}

	return nil
}

func (instance *Instance) dispatchKeypressLocked(data string) error {
	if data == "\x03" && instance.exitOnCtrlC {
		instance.app.Exit()
		return instance.unmountDoneLocked(nil)
	}

	inputValue, key, keys, err := inkinput.NormalizeHookInput(data)
	if err != nil {
		return err
	}

	if data == "\x1b" {
		instance.app.clearFocus()
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
		return instance.unmountDoneLocked(instance.app.ExitError())
	}

	if instance.app.consumeStateChange() {
		return instance.renderCurrentLocked()
	}

	return nil
}

func (instance *Instance) writeRenderLocked(logicalOutput string, renderedOutput string, cursorPosition *CursorPosition) error {
	// Branch precedence mirrors upstream Ink's onRender: the debug branch is
	// checked first, before the screen-reader branch, so debug+screenReader
	// produces an append-only stream of plain accessible frames rather than
	// the screen-reader path's eraseLines+rewrite cycle. See ink/src/ink.tsx.
	if instance.debug {
		return instance.writeDebugRenderLocked(logicalOutput, renderedOutput, cursorPosition)
	}

	if instance.app.IsScreenReaderEnabled() {
		return instance.writeScreenReaderRenderLocked(logicalOutput, renderedOutput)
	}

	if instance.incrementalRendering {
		if instance.cellLevelDiff {
			return instance.writeCellLevelRenderLocked(logicalOutput, renderedOutput, cursorPosition)
		}
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

	// Upstream Ink's debug path always emits one append per render (see
	// onRender's `this.options.stdout.write(this.fullStaticOutput + output)`
	// with no equality short-circuit). Identical-frame skipping here would
	// silently swallow ticks of an animation or progress-log timeline that
	// callers explicitly opted into by enabling debug.
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

	if !instance.cursorHidden && shouldManageCursor(instance.stdout) {
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

	// Debug mode is an append-only frame log: every render must produce a
	// write so callers can step through the timeline. Upstream Ink's debug
	// branch unconditionally writes — see ink/src/ink.tsx onRender.
	if instance.debug {
		return true
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
	prepared = instance.settlePendingRenderLocked(prepared)

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
	prepared = instance.settlePendingRenderLocked(prepared)
	return instance.commitPreparedRenderLocked(prepared)
}

func (instance *Instance) settlePendingRenderLocked(prepared preparedRender) preparedRender {
	for settled := 0; settled < maxPendingStateSettles; settled++ {
		if !instance.app.consumeStateChange() {
			return prepared
		}

		prepared = instance.prepareRenderLocked()
		if prepared.exitRequested {
			return prepared
		}
	}

	return prepared
}

func (instance *Instance) writeStandardRenderLocked(logicalOutput string, output string, cursorPosition *CursorPosition) error {
	if instance.stdout == nil {
		return nil
	}

	if !instance.cursorHidden && shouldManageCursor(instance.stdout) {
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

		previousLine := ""
		if index < len(instance.previousLines) {
			previousLine = instance.previousLines[index]
		}

		suffix := "\n"
		if isLastLine && !hasTrailingNewline {
			suffix = ""
		}

		// Column-level dirty-rect optimization: when both the previous and
		// next line are pure plain text (no ANSI escapes) and share a
		// common visible prefix, move the cursor to the divergence column,
		// erase to end-of-line, and emit only the differing tail. Falls
		// back to a full-line rewrite when ANSI sequences are present —
		// styled output mixes display state with code points so a naïve
		// column count would corrupt the active SGR state.
		if commonCols, useColumnDiff := commonPlainPrefixWidth(previousLine, line); useColumnDiff {
			tail := line[len(previousLine[:commonByteOffsetForWidth(previousLine, commonCols)]):]
			buffer = append(buffer, ansiCursorTo(commonCols)+ansiEraseEndLine()+tail+suffix)
		} else {
			buffer = append(buffer, ansiCursorTo(0)+line+ansiEraseEndLine()+suffix)
		}
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
	return instance.unmountLockedWithOptions(exitErr, unmountOptions{clearOutput: true})
}

func (instance *Instance) unmountDoneLocked(exitErr error) error {
	return instance.unmountLockedWithOptions(exitErr, unmountOptions{clearOutput: false})
}

func (instance *Instance) unmountLockedWithOptions(exitErr error, options unmountOptions) error {
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
	// Discard any in-flight bracketed-paste body — the session is closing
	// before the terminal emitted the matching end marker, so the partial
	// payload should not be delivered as a paste event.
	instance.pendingPaste = nil
	instance.pastePending = false
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
	} else if options.clearOutput {
		if err := instance.clearLocked(); err != nil && firstErr == nil {
			firstErr = err
		}

		if instance.cursorHidden && instance.stdout != nil {
			if err := writePayload(instance.stdout, showCursorEscape); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	} else {
		instance.previousLogicalOutput = ""
		instance.previousOutput = ""
		instance.previousLines = nil
		instance.previousLineCount = 0
		instance.previousCursorPosition = nil
		instance.cursorWasShown = false
		if instance.cursorHidden && instance.stdout != nil {
			if err := writePayload(instance.stdout, showCursorEscape); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	instance.cursorHidden = false

	// Drop any queued announcements: the session is going away, so any
	// messages still pending would otherwise outlive the consumer that
	// produced them. Done after the final render write but before hook
	// cleanup so unmount-time cleanup can still observe an empty queue.
	if announcer := instance.app.Announcer(); announcer != nil {
		announcer.Clear()
	}

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

func normalizeRenderMaxFPS(options RenderOptions) int {
	if options.MaxFPSLimit != nil {
		return normalizeMaxFPS(*options.MaxFPSLimit)
	}

	return normalizeMaxFPS(options.MaxFPS)
}

func normalizeExitOnCtrlC(value *bool) bool {
	if value == nil {
		return true
	}

	return *value
}

func throttleDuration(fps int) time.Duration {
	normalized := normalizeMaxFPS(fps)
	if normalized == 0 {
		return 0
	}

	return time.Duration((1000+normalized-1)/normalized) * time.Millisecond
}
