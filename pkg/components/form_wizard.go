package components

import (
	"errors"
	"fmt"

	"github.com/dh-kam/goink.go/pkg/vdom"
)

// FormWizardKey enumerates the default keyboard bindings that
// FormWizard.HandleKey understands. Callers may map their own raw input
// strings to these constants and dispatch through HandleKey, or call the
// Next / Prev / Submit / FocusNext / FocusPrev methods directly.
type FormWizardKey int

const (
	// FormWizardKeyNone is the zero value; HandleKey treats it as a no-op.
	FormWizardKeyNone FormWizardKey = iota
	// FormWizardKeyTab cycles focus forward within the current step's
	// fields (FocusNext on the active FormState).
	FormWizardKeyTab
	// FormWizardKeyShiftTab cycles focus backward within the current step.
	FormWizardKeyShiftTab
	// FormWizardKeyEnter advances to the next step (validating the
	// current step first). When already on the last step it triggers
	// Submit.
	FormWizardKeyEnter
	// FormWizardKeyEsc moves back one step. No-op on the first step.
	FormWizardKeyEsc
)

// ErrFormWizardEmpty is returned by Next / Prev / Submit when the wizard
// holds no steps. Callers can use errors.Is to short-circuit empty-state
// rendering.
var ErrFormWizardEmpty = errors.New("form wizard has no steps")

// FormWizardStep groups a human-friendly title with a FormState. Title is
// optional; when empty the step is referred to by its 1-based index in
// error messages.
type FormWizardStep struct {
	Title string
	State *FormState
}

// FormWizard is a multi-step form controller. It owns N FormStates (one
// per step), the index of the active step, and the cumulative submission
// state. It is intentionally controlled — UI code reads CurrentStep() /
// Step() and renders the matching FormState; input is dispatched into the
// wizard which forwards it to the active step.
//
// State machine:
//
//   - Each call to Next runs Validate on the active step; advances only
//     when there are no errors.
//   - Prev always succeeds (preserves all step state).
//   - Submit runs Validate on the final step and, on success, marks the
//     wizard "submitted" (HasSubmitted returns true).
//   - Errors() aggregates validation errors across every step that has at
//     least one error currently registered, in step order.
type FormWizard struct {
	Steps []FormWizardStep
	// step is the active step index; clamped to [0, len(Steps)-1] at
	// construction and after every navigation method.
	step int
	// submitted is set when Submit runs successfully on the final step.
	submitted bool
}

// NewFormWizard wraps the supplied steps in a wizard starting at step 0.
// The slice is shallow-copied so callers may keep mutating their own
// FormState references — the wizard reads through the shared pointers.
func NewFormWizard(steps []FormWizardStep) *FormWizard {
	clone := make([]FormWizardStep, len(steps))
	copy(clone, steps)
	return &FormWizard{Steps: clone}
}

// CurrentStep returns the active step index (0-based).
func (w *FormWizard) CurrentStep() int {
	return w.step
}

// TotalSteps returns the number of steps registered.
func (w *FormWizard) TotalSteps() int {
	return len(w.Steps)
}

// Step returns the active FormWizardStep and true when the wizard is
// non-empty. Returns the zero value and false when there are no steps.
func (w *FormWizard) Step() (FormWizardStep, bool) {
	if len(w.Steps) == 0 {
		return FormWizardStep{}, false
	}
	return w.Steps[w.step], true
}

// State returns the active step's FormState, or nil when the wizard is
// empty. Convenience accessor for input handlers that want to reach
// straight into the controller.
func (w *FormWizard) State() *FormState {
	if len(w.Steps) == 0 {
		return nil
	}
	return w.Steps[w.step].State
}

// IsFirst reports whether the active step is step 0.
func (w *FormWizard) IsFirst() bool { return w.step == 0 }

// IsLast reports whether the active step is the final one (true also for
// an empty wizard, since len(Steps)-1 < 0).
func (w *FormWizard) IsLast() bool {
	if len(w.Steps) == 0 {
		return true
	}
	return w.step >= len(w.Steps)-1
}

// HasSubmitted reports whether Submit has already succeeded once. After
// that point Next / Submit are no-ops returning nil.
func (w *FormWizard) HasSubmitted() bool { return w.submitted }

// Next validates the active step. When validation fails, returns an error
// whose Error() carries the step's first OrderedErrors entry; the wizard
// remains on the current step. When validation passes and there are more
// steps remaining, the active step advances and nil is returned. When
// already on the last step, Next is equivalent to Submit.
func (w *FormWizard) Next() error {
	if len(w.Steps) == 0 {
		return ErrFormWizardEmpty
	}
	state := w.Steps[w.step].State
	if state == nil {
		return fmt.Errorf("step %d has no FormState", w.step+1)
	}
	if !state.Validate() {
		ordered := state.OrderedErrors()
		if len(ordered) == 0 {
			return fmt.Errorf("step %d: validation failed", w.step+1)
		}
		return fmt.Errorf("step %d: %s", w.step+1, ordered[0].Error())
	}
	if w.IsLast() {
		w.submitted = true
		return nil
	}
	w.step++
	return nil
}

// Prev moves back one step. State on every step is preserved. Returns
// ErrFormWizardEmpty when the wizard has no steps; otherwise returns nil.
// Calling Prev on step 0 is a no-op (returns nil) — the wizard never
// goes negative.
func (w *FormWizard) Prev() error {
	if len(w.Steps) == 0 {
		return ErrFormWizardEmpty
	}
	if w.step > 0 {
		w.step--
	}
	return nil
}

// Submit runs Validate on every step in declaration order, populating
// each step's Errors map. When all steps pass, Submit marks the wizard as
// submitted, leaves the active step pointing at the last step, and
// returns nil. On failure, Submit jumps the active step to the first
// failing step so callers can re-render the relevant fields immediately
// and returns an error summarising that step.
//
// Submit is safe to call on an already-submitted wizard; it re-validates
// and may un-submit if state has been mutated to an invalid form since
// the previous Submit.
func (w *FormWizard) Submit() error {
	if len(w.Steps) == 0 {
		return ErrFormWizardEmpty
	}
	w.submitted = false
	firstFail := -1
	for i, step := range w.Steps {
		if step.State == nil {
			if firstFail == -1 {
				firstFail = i
			}
			continue
		}
		if !step.State.Validate() && firstFail == -1 {
			firstFail = i
		}
	}
	if firstFail >= 0 {
		w.step = firstFail
		state := w.Steps[firstFail].State
		if state == nil {
			return fmt.Errorf("step %d has no FormState", firstFail+1)
		}
		ordered := state.OrderedErrors()
		if len(ordered) == 0 {
			return fmt.Errorf("step %d: validation failed", firstFail+1)
		}
		return fmt.Errorf("step %d: %s", firstFail+1, ordered[0].Error())
	}
	w.step = len(w.Steps) - 1
	w.submitted = true
	return nil
}

// Errors returns the aggregate of every step's currently-registered
// validation errors, in step + field declaration order. Each error is
// prefixed with the step Title (or "Step N" when Title is empty) so the
// flat list remains self-describing across step boundaries — useful as
// input to ErrorOverviewFromWizard or any other error surface.
//
// Returns an empty (non-nil) slice when no step has any errors.
func (w *FormWizard) Errors() []error {
	out := make([]error, 0)
	for i, step := range w.Steps {
		if step.State == nil {
			continue
		}
		title := step.Title
		if title == "" {
			title = fmt.Sprintf("Step %d", i+1)
		}
		for _, err := range step.State.OrderedErrors() {
			out = append(out, fmt.Errorf("%s — %s", title, err.Error()))
		}
	}
	return out
}

// HasErrors reports whether any step currently has at least one
// registered validation error. Equivalent to len(w.Errors()) > 0 but
// cheaper.
func (w *FormWizard) HasErrors() bool {
	for _, step := range w.Steps {
		if step.State != nil && step.State.HasErrors() {
			return true
		}
	}
	return false
}

// FocusNext forwards FocusNext to the active step's FormState. No-op
// when the wizard is empty.
func (w *FormWizard) FocusNext() {
	if state := w.State(); state != nil {
		state.FocusNext()
	}
}

// FocusPrev forwards FocusPrev to the active step's FormState. No-op
// when the wizard is empty.
func (w *FormWizard) FocusPrev() {
	if state := w.State(); state != nil {
		state.FocusPrev()
	}
}

// HandleKey dispatches a default key binding into the wizard. The
// mapping is opt-in: callers must pre-translate raw input into a
// FormWizardKey. Returns the same error contract as the underlying
// method (Next / Prev / Submit) — Tab/Shift-Tab return nil since
// focus movement never fails.
//
//	tab        → FocusNext
//	shift+tab  → FocusPrev
//	enter      → Next (or Submit on the last step)
//	esc        → Prev
func (w *FormWizard) HandleKey(key FormWizardKey) error {
	switch key {
	case FormWizardKeyTab:
		w.FocusNext()
		return nil
	case FormWizardKeyShiftTab:
		w.FocusPrev()
		return nil
	case FormWizardKeyEnter:
		return w.Next()
	case FormWizardKeyEsc:
		return w.Prev()
	}
	return nil
}

// Values returns a flat copy of every step's Values, keyed by field
// Name. When two steps share a field Name, the later step's value wins
// (callers should keep field names unique across steps for predictable
// behaviour). Returns an empty (non-nil) map for an empty wizard.
func (w *FormWizard) Values() map[string]string {
	out := make(map[string]string)
	for _, step := range w.Steps {
		if step.State == nil {
			continue
		}
		for k, v := range step.State.Values {
			out[k] = v
		}
	}
	return out
}

// ErrorOverviewFromWizard renders an ErrorOverviewGroup pre-populated
// with every step's currently-registered validation errors as the
// Validation section. Additional runtime errors (panics, downstream call
// failures) may be supplied via runtime — they appear under the Runtime
// sub-section. The returned node is safe to embed directly into any
// layout; it renders a "<no errors>" placeholder when both sources are
// empty.
func ErrorOverviewFromWizard(w *FormWizard, runtime ...error) *vdom.Node {
	var validation []error
	if w != nil {
		validation = w.Errors()
	}
	return ErrorOverviewGroup(ErrorOverviewGroupProps{
		Validation: validation,
		Runtime:    runtime,
	})
}
