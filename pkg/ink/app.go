package ink

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dh-kam/goink.go/internal/renderer"
	"github.com/dh-kam/goink.go/pkg/focus"
	"github.com/dh-kam/goink.go/pkg/hooks"
	"github.com/dh-kam/goink.go/pkg/terminal"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

const publicComponentMarkerKey = "__inkPublicComponent"

// CursorPosition stores a cursor location relative to the rendered Ink output.
type CursorPosition struct {
	X int
	Y int
}

// AppOptions configures a new Ink app instance.
type AppOptions struct {
	Width               int
	Height              int
	Stdin               io.Reader
	Stdout              io.Writer
	Stderr              io.Writer
	ScreenReaderEnabled bool
}

// App represents a running Ink application.
type App struct {
	component           ComponentFunc
	hooksCtx            *hooks.Context
	width               int
	height              int
	autoWidth           bool
	autoHeight          bool
	stdin               io.Reader
	stdout              io.Writer
	stderr              io.Writer
	rawState            *terminal.State
	rawModeUsers        int
	exitRequested       bool
	exitErr             error
	cursorPosition      *CursorPosition
	screenReaderEnabled bool
}

func (a *App) renderVNode() *vdom.Node {
	a.hooksCtx.Reset()
	a.cursorPosition = nil
	currentHooksContext = a.hooksCtx
	currentApp = a
	defer ResetCurrentHooksContext()

	if a.component == nil {
		return nil
	}

	return a.component()
}

func (a *App) finishRender() {
	a.hooksCtx.FinalizeRender()
	a.hooksCtx.RunEffects()
}

func (a *App) consumeStateChange() bool {
	return a.hooksCtx.ConsumeStateChange()
}

// NewApp creates a new Ink application with default options.
func NewApp(component ComponentFunc) *App {
	return NewAppWithOptions(component, AppOptions{})
}

// NewAppWithOptions creates a new Ink application with explicit runtime options.
func NewAppWithOptions(component ComponentFunc, options AppOptions) *App {
	width := options.Width
	height := options.Height
	autoWidth := width <= 0
	autoHeight := height <= 0
	if width <= 0 || height <= 0 {
		defaultWidth, defaultHeight := terminalViewportSize(options.Stdout, 80, 24)

		if width <= 0 {
			width = defaultWidth
		}

		if height <= 0 {
			height = defaultHeight
		}
	}

	stdin := options.Stdin
	if stdin == nil {
		stdin = os.Stdin
	}

	stdout := options.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}

	stderr := options.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	return &App{
		component:           component,
		hooksCtx:            hooks.NewContext(),
		width:               width,
		height:              height,
		autoWidth:           autoWidth,
		autoHeight:          autoHeight,
		stdin:               stdin,
		stdout:              stdout,
		stderr:              stderr,
		screenReaderEnabled: options.ScreenReaderEnabled,
	}
}

// RenderOnce renders the component once and returns the output.
func (a *App) RenderOnce() string {
	node := a.renderVNode()
	if node == nil {
		a.finishRender()
		return ""
	}

	validatePublicComponentTree(node)
	output := renderer.RenderWithLayout(node, a.width, a.height)
	a.finishRender()
	return output
}

// RenderSplitOnce renders the current tree into dynamic and static sections.
func (a *App) RenderSplitOnce() (output string, staticOutput string) {
	node := a.renderVNode()
	if node == nil {
		a.finishRender()
		return "", ""
	}

	validatePublicComponentTree(node)
	renderer.SyncComputedLayout(node, a.width, a.height)

	var sections renderer.RenderSections
	if a.IsScreenReaderEnabled() {
		sections = renderer.RenderScreenReaderSections(node)
	} else {
		sections = renderer.RenderWithLayoutSectionsMode(node, a.width, a.height, a.shouldRenderANSI())
	}

	a.finishRender()
	return sections.Output, sections.StaticOutput
}

// RenderRuntimeOnce renders the current tree into dynamic output plus newly appended static delta.
func (a *App) RenderRuntimeOnce(previousStaticCounts []int) (output string, staticDelta string, nextStaticCounts []int) {
	node := a.renderVNode()
	if node == nil {
		a.finishRender()
		return "", "", nil
	}

	validatePublicComponentTree(node)
	renderer.SyncComputedLayout(node, a.width, a.height)
	sections := renderer.RenderRuntimeSectionsMode(node, a.width, a.height, previousStaticCounts, a.IsScreenReaderEnabled(), a.shouldRenderANSI())
	a.finishRender()
	return sections.Output, sections.StaticDeltaOutput, sections.StaticCounts
}

// GetVNode renders the component and returns the virtual DOM node.
func (a *App) GetVNode() *vdom.Node {
	return a.renderVNode()
}

func validatePublicComponentTree(node *vdom.Node) {
	validatePublicComponentNode(node, false, false)
}

func validatePublicComponentNode(node *vdom.Node, inPublicTree bool, inTextContext bool) {
	if node == nil {
		return
	}

	switch node.Type {
	case vdom.TextNode:
		if inPublicTree && !inTextContext {
			panic(fmt.Sprintf("Text string %q must be rendered inside <Text> component", node.Text))
		}
	case vdom.ElementNode:
		publicNode := isPublicComponentNode(node)
		textLike := isTextLikeElement(node)
		nextPublicTree := inPublicTree || publicNode
		nextTextContext := nextPublicTree && (inTextContext || textLike)

		if nextPublicTree && inTextContext && !textLike {
			panic(fmt.Sprintf("<%s> can’t be nested inside <Text> component", displayElementName(node.ElementType)))
		}

		for _, child := range node.Children {
			validatePublicComponentNode(child, nextPublicTree, nextTextContext)
		}
	}
}

func isPublicComponentNode(node *vdom.Node) bool {
	if node == nil || node.Props == nil {
		return false
	}

	value, ok := node.Props[publicComponentMarkerKey]
	if !ok {
		return false
	}

	enabled, _ := value.(bool)
	return enabled
}

func isTextLikeElement(node *vdom.Node) bool {
	if node == nil || node.Type != vdom.ElementNode {
		return false
	}

	return node.ElementType == "text" || node.ElementType == "transform"
}

func displayElementName(elementType string) string {
	if elementType == "" {
		return "Unknown"
	}

	return strings.ToUpper(elementType[:1]) + elementType[1:]
}

// Width returns the configured render width.
func (a *App) Width() int {
	return a.width
}

// Height returns the configured render height.
func (a *App) Height() int {
	return a.height
}

// SetSize updates the render area dimensions.
func (a *App) SetSize(width, height int) {
	if width > 0 {
		a.width = width
	}

	if height > 0 {
		a.height = height
	}
}

// RefreshSizeFromWriter updates any auto-sized dimensions from the configured stdout.
func (a *App) RefreshSizeFromWriter() bool {
	if !a.autoWidth && !a.autoHeight {
		return false
	}

	width, height := terminalViewportSize(a.stdout, a.width, a.height)
	changed := false

	if a.autoWidth && width > 0 && width != a.width {
		a.width = width
		changed = true
	}

	if a.autoHeight && height > 0 && height != a.height {
		a.height = height
		changed = true
	}

	return changed
}

// Stdin returns the configured stdin stream.
func (a *App) Stdin() io.Reader {
	return a.stdin
}

// Stdout returns the configured stdout stream.
func (a *App) Stdout() io.Writer {
	return a.stdout
}

// Stderr returns the configured stderr stream.
func (a *App) Stderr() io.Writer {
	return a.stderr
}

// WriteStdout writes directly to the configured stdout stream.
func (a *App) WriteStdout(data string) (int, error) {
	if a.stdout == nil {
		return 0, errors.New("stdout is not configured")
	}

	return io.WriteString(a.stdout, data)
}

// WriteStderr writes directly to the configured stderr stream.
func (a *App) WriteStderr(data string) (int, error) {
	if a.stderr == nil {
		return 0, errors.New("stderr is not configured")
	}

	return io.WriteString(a.stderr, data)
}

// IsRawModeSupported reports whether the configured stdin can be placed into raw mode.
func (a *App) IsRawModeSupported() bool {
	fd, ok := streamFD(a.stdin)
	if !ok {
		return false
	}

	return terminal.IsTerminal(fd)
}

// SetRawMode toggles raw mode on the configured stdin stream.
func (a *App) SetRawMode(enabled bool) error {
	if !enabled {
		if a.rawModeUsers > 0 {
			a.rawModeUsers--
		}

		if a.rawModeUsers > 0 {
			return nil
		}

		if a.rawState != nil {
			err := a.rawState.Restore()
			a.rawState = nil
			return err
		}

		return nil
	}

	if !a.IsRawModeSupported() {
		return nil
	}

	if a.rawState != nil {
		a.rawModeUsers++
		return nil
	}

	fd, _ := streamFD(a.stdin)
	state, err := terminal.MakeRaw(fd)
	if err != nil {
		return err
	}

	a.rawState = state
	a.rawModeUsers = 1
	return nil
}

// Exit marks the app as exited and stores an optional exit error.
func (a *App) Exit(err ...error) {
	a.exitRequested = true
	if len(err) > 0 {
		a.exitErr = err[0]
	}
}

// ExitRequested reports whether Exit has been called.
func (a *App) ExitRequested() bool {
	return a.exitRequested
}

// ExitError returns the error supplied to Exit, if any.
func (a *App) ExitError() error {
	return a.exitErr
}

// SetCursorPosition updates the current cursor position. Pass nil to hide the cursor.
func (a *App) SetCursorPosition(position *CursorPosition) {
	if position == nil {
		a.cursorPosition = nil
		return
	}

	copy := *position
	a.cursorPosition = &copy
}

// CursorPosition returns a copy of the current cursor position, if set.
func (a *App) CursorPosition() *CursorPosition {
	if a.cursorPosition == nil {
		return nil
	}

	copy := *a.cursorPosition
	return &copy
}

// SetScreenReaderEnabled updates the screen-reader mode flag.
func (a *App) SetScreenReaderEnabled(enabled bool) {
	a.screenReaderEnabled = enabled
}

// IsScreenReaderEnabled reports whether screen-reader mode is enabled.
func (a *App) IsScreenReaderEnabled() bool {
	return a.screenReaderEnabled
}

func (a *App) shouldRenderANSI() bool {
	return !a.IsScreenReaderEnabled() && shouldSynchronize(a.stdout)
}

func (a *App) setFocusEnabled(enabled bool) {
	if a.hooksCtx.FocusEnabled() == enabled {
		return
	}

	a.hooksCtx.SetFocusEnabled(enabled)
	a.hooksCtx.RequestRerender()
}

func (a *App) focusManager() *focus.FocusManager {
	return a.hooksCtx.FocusManager()
}

func (a *App) activeFocusID() *string {
	activeID := a.focusManager().ActiveID()
	if activeID == nil {
		return nil
	}

	id := string(*activeID)
	return &id
}

func (a *App) focusNext() bool {
	if !a.hooksCtx.FocusEnabled() {
		return false
	}

	if !a.focusManager().FocusNext() {
		return false
	}

	a.hooksCtx.RequestRerender()
	return true
}

func (a *App) focusPrevious() bool {
	if !a.hooksCtx.FocusEnabled() {
		return false
	}

	if !a.focusManager().FocusPrevious() {
		return false
	}

	a.hooksCtx.RequestRerender()
	return true
}

func (a *App) blurFocus() bool {
	if !a.hooksCtx.FocusEnabled() {
		return false
	}

	manager := a.focusManager()
	if !manager.HasFocus() {
		return false
	}

	manager.Blur()
	a.hooksCtx.RequestRerender()
	return true
}

func (a *App) focus(id string) bool {
	manager := a.focusManager()
	before := manager.FocusedID()
	beforeHasFocus := manager.HasFocus()
	focused := manager.Focus(focus.FocusID(id))
	if manager.FocusedID() != before || manager.HasFocus() != beforeHasFocus {
		a.hooksCtx.RequestRerender()
	}

	return focused
}

type uintptrFD interface {
	Fd() uintptr
}

type intFD interface {
	Fd() int
}

func streamFD(stream any) (int, bool) {
	switch typed := stream.(type) {
	case uintptrFD:
		return int(typed.Fd()), true
	case intFD:
		return typed.Fd(), true
	default:
		return 0, false
	}
}

// Global hook/runtime context for the currently rendering app.
var (
	currentHooksContext *hooks.Context
	currentApp          *App
)

// UseState is a convenience wrapper that uses the current app's hooks context.
func UseState(initialValue interface{}) (interface{}, hooks.SetStateFunc) {
	if currentHooksContext == nil {
		panic("UseState called outside of component render")
	}

	return hooks.UseState(currentHooksContext, initialValue)
}

// ResetCurrentHooksContext resets the global hooks/runtime context.
func ResetCurrentHooksContext() {
	currentHooksContext = nil
	currentApp = nil
}

func requireCurrentApp(hookName string) *App {
	if currentApp == nil {
		panic(hookName + " called outside of component render")
	}

	return currentApp
}

func requireHooksContext(hookName string) *hooks.Context {
	if currentHooksContext == nil {
		panic(hookName + " called outside of component render")
	}

	return currentHooksContext
}
