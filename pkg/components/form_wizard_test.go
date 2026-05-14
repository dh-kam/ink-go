package components_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/dh-kam/goink.go/pkg/components"
)

// wizardFixture returns a 3-step wizard whose middle step starts populated
// (so step 1 fails Required, step 2 succeeds, step 3 has no requirements).
func wizardFixture() *components.FormWizard {
	step1 := components.NewFormState([]components.FormField{
		{Name: "name", Label: "Name", Kind: components.FormFieldText, Required: true},
	})
	step2 := components.NewFormState([]components.FormField{
		{Name: "env", Label: "Env", Kind: components.FormFieldTab, Required: true,
			Options: []components.SelectItem{
				{Label: "Dev", Value: "dev"},
				{Label: "Prod", Value: "prod"},
			}, Default: "dev"},
	})
	step3 := components.NewFormState([]components.FormField{
		{Name: "notes", Label: "Notes", Kind: components.FormFieldText},
	})
	return components.NewFormWizard([]components.FormWizardStep{
		{Title: "Profile", State: step1},
		{Title: "Environment", State: step2},
		{Title: "Notes", State: step3},
	})
}

func TestFormWizardCurrentStepStartsAtZero(t *testing.T) {
	w := wizardFixture()
	if got := w.CurrentStep(); got != 0 {
		t.Errorf("CurrentStep() = %d, want 0", got)
	}
	if got := w.TotalSteps(); got != 3 {
		t.Errorf("TotalSteps() = %d, want 3", got)
	}
	if !w.IsFirst() {
		t.Errorf("IsFirst() should be true on fresh wizard")
	}
	if w.IsLast() {
		t.Errorf("IsLast() should be false on fresh wizard with 3 steps")
	}
}

func TestFormWizardNextAdvancesWhenValid(t *testing.T) {
	w := wizardFixture()
	w.State().SetValue("name", "Alice")
	if err := w.Next(); err != nil {
		t.Fatalf("Next() with valid step 1 returned %v", err)
	}
	if w.CurrentStep() != 1 {
		t.Errorf("after Next, CurrentStep = %d, want 1", w.CurrentStep())
	}
}

func TestFormWizardNextValidationGateBlocksAdvance(t *testing.T) {
	w := wizardFixture()
	// step 1 has empty required name — should fail.
	err := w.Next()
	if err == nil {
		t.Fatal("Next() should return an error when current step has validation failures")
	}
	if w.CurrentStep() != 0 {
		t.Errorf("Next failure changed CurrentStep to %d, want 0", w.CurrentStep())
	}
	if !strings.Contains(err.Error(), "Name") {
		t.Errorf("expected error to reference field, got %q", err.Error())
	}
	if !w.HasErrors() {
		t.Errorf("wizard should report HasErrors after failed Next")
	}
}

func TestFormWizardPrevPreservesState(t *testing.T) {
	w := wizardFixture()
	w.State().SetValue("name", "Alice")
	if err := w.Next(); err != nil {
		t.Fatalf("setup: Next() returned %v", err)
	}
	// On step 2, change the env value.
	w.State().SetValue("env", "prod")
	if err := w.Prev(); err != nil {
		t.Fatalf("Prev() returned %v", err)
	}
	if w.CurrentStep() != 0 {
		t.Errorf("after Prev, CurrentStep = %d, want 0", w.CurrentStep())
	}
	// Step 1's value should still be Alice.
	if got := w.State().Values["name"]; got != "Alice" {
		t.Errorf("step 1 name not preserved: %q", got)
	}
	// Step 2's mutated value must also still be there.
	if got := w.Steps[1].State.Values["env"]; got != "prod" {
		t.Errorf("step 2 env not preserved across Prev: %q", got)
	}
	// Prev on step 0 is a no-op.
	if err := w.Prev(); err != nil {
		t.Fatalf("Prev() on step 0 returned %v", err)
	}
	if w.CurrentStep() != 0 {
		t.Errorf("Prev on step 0 changed step to %d", w.CurrentStep())
	}
}

func TestFormWizardSubmitAllCleanSucceeds(t *testing.T) {
	w := wizardFixture()
	w.State().SetValue("name", "Alice")
	if err := w.Next(); err != nil {
		t.Fatalf("setup Next 1 returned %v", err)
	}
	if err := w.Next(); err != nil {
		t.Fatalf("setup Next 2 returned %v", err)
	}
	// Now on step 3 (last). Calling Next again is equivalent to Submit.
	if !w.IsLast() {
		t.Fatalf("expected to be on last step, got %d", w.CurrentStep())
	}
	if err := w.Submit(); err != nil {
		t.Fatalf("Submit returned %v", err)
	}
	if !w.HasSubmitted() {
		t.Error("HasSubmitted should be true after successful Submit")
	}
	values := w.Values()
	if values["name"] != "Alice" || values["env"] != "dev" {
		t.Errorf("Values() = %v, expected name=Alice env=dev", values)
	}
}

func TestFormWizardSubmitJumpsToFirstFailingStep(t *testing.T) {
	w := wizardFixture()
	// Move to step 3 directly via Next, but step 1 is still empty so the
	// first Next will fail. To exercise mid-step error semantics we
	// instead advance using valid values then mutate step 1 to invalid.
	w.State().SetValue("name", "Alice")
	if err := w.Next(); err != nil {
		t.Fatalf("setup Next 1: %v", err)
	}
	if err := w.Next(); err != nil {
		t.Fatalf("setup Next 2: %v", err)
	}
	// On step 3. Now go back and clear step 1.
	if err := w.Prev(); err != nil {
		t.Fatalf("Prev: %v", err)
	}
	if err := w.Prev(); err != nil {
		t.Fatalf("Prev: %v", err)
	}
	w.State().SetValue("name", "")
	// Jump back to step 3 and try Submit — should jump to step 1.
	if err := w.Next(); err == nil {
		t.Fatal("expected Next to fail on cleared step 1")
	}
	// Reset to step 3 by re-validating.
	w.State().SetValue("name", "Alice")
	if err := w.Next(); err != nil {
		t.Fatalf("Next after re-fix: %v", err)
	}
	if err := w.Next(); err != nil {
		t.Fatalf("Next: %v", err)
	}
	if !w.IsLast() {
		t.Fatalf("expected to be on last step")
	}
	// Now invalidate step 1 from the last step and call Submit.
	w.Steps[0].State.SetValue("name", "")
	err := w.Submit()
	if err == nil {
		t.Fatal("Submit should fail when an earlier step has invalid data")
	}
	if w.CurrentStep() != 0 {
		t.Errorf("Submit should jump to first failing step (0), got %d", w.CurrentStep())
	}
	if w.HasSubmitted() {
		t.Errorf("HasSubmitted should be false after failed Submit")
	}
	if !strings.Contains(err.Error(), "step 1") {
		t.Errorf("expected error to reference step 1, got %q", err.Error())
	}
}

func TestFormWizardErrorsAggregateAcrossSteps(t *testing.T) {
	w := wizardFixture()
	// Validate every step explicitly to populate errors.
	for _, step := range w.Steps {
		step.State.Validate()
	}
	got := w.Errors()
	if len(got) == 0 {
		t.Fatal("expected aggregated errors across steps")
	}
	// Step 1 has a Required name; step 2's default = "dev" is valid;
	// step 3 has no validators — so we expect exactly 1 error tagged
	// "Profile".
	if len(got) != 1 {
		t.Fatalf("expected 1 aggregated error, got %d (%v)", len(got), got)
	}
	if !strings.Contains(got[0].Error(), "Profile") {
		t.Errorf("expected error tagged with step Title, got %q", got[0].Error())
	}
	if !strings.Contains(got[0].Error(), "Name") {
		t.Errorf("expected field Label in error, got %q", got[0].Error())
	}
}

func TestFormWizardErrorsTitleFallsBackToStepIndex(t *testing.T) {
	step := components.NewFormState([]components.FormField{
		{Name: "x", Label: "X", Kind: components.FormFieldText, Required: true},
	})
	w := components.NewFormWizard([]components.FormWizardStep{
		{State: step},
	})
	step.Validate()
	got := w.Errors()
	if len(got) != 1 {
		t.Fatalf("expected 1 error, got %v", got)
	}
	if !strings.Contains(got[0].Error(), "Step 1") {
		t.Errorf("expected fallback \"Step 1\" prefix, got %q", got[0].Error())
	}
}

func TestErrorOverviewFromWizardAggregatesValidationAndRuntime(t *testing.T) {
	w := wizardFixture()
	for _, step := range w.Steps {
		step.State.Validate()
	}
	node := components.ErrorOverviewFromWizard(w, errors.New("network blip"))
	if node == nil {
		t.Fatal("ErrorOverviewFromWizard returned nil")
	}
	out := collectText(node)
	for _, want := range []string{"ERRORS", "Validation", "Profile", "Runtime", "network blip"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in overview, got %q", want, out)
		}
	}
}

func TestErrorOverviewFromWizardEmptyState(t *testing.T) {
	node := components.ErrorOverviewFromWizard(nil)
	out := collectText(node)
	if !strings.Contains(out, "<no errors>") {
		t.Errorf("expected placeholder for empty state, got %q", out)
	}
}

func TestFormWizardHandleKeyDispatchesDefaultBindings(t *testing.T) {
	w := wizardFixture()
	w.State().SetValue("name", "Alice")
	// Tab cycles focus inside the step. With one field, focus stays at 0.
	if err := w.HandleKey(components.FormWizardKeyTab); err != nil {
		t.Fatalf("Tab returned %v", err)
	}
	// Enter advances to next step.
	if err := w.HandleKey(components.FormWizardKeyEnter); err != nil {
		t.Fatalf("Enter returned %v", err)
	}
	if w.CurrentStep() != 1 {
		t.Errorf("Enter should advance step, got %d", w.CurrentStep())
	}
	// Esc moves back.
	if err := w.HandleKey(components.FormWizardKeyEsc); err != nil {
		t.Fatalf("Esc returned %v", err)
	}
	if w.CurrentStep() != 0 {
		t.Errorf("Esc should move back, got %d", w.CurrentStep())
	}
	// Shift-Tab cycles focus backward (single field → still 0).
	if err := w.HandleKey(components.FormWizardKeyShiftTab); err != nil {
		t.Fatalf("Shift-Tab returned %v", err)
	}
	// Unknown key is a no-op.
	if err := w.HandleKey(components.FormWizardKeyNone); err != nil {
		t.Fatalf("None key returned %v", err)
	}
}

func TestFormWizardHandleKeyEnterOnLastStepSubmits(t *testing.T) {
	w := wizardFixture()
	w.State().SetValue("name", "Alice")
	if err := w.HandleKey(components.FormWizardKeyEnter); err != nil {
		t.Fatalf("step 1 Enter: %v", err)
	}
	if err := w.HandleKey(components.FormWizardKeyEnter); err != nil {
		t.Fatalf("step 2 Enter: %v", err)
	}
	// On last step.
	if err := w.HandleKey(components.FormWizardKeyEnter); err != nil {
		t.Fatalf("final Enter: %v", err)
	}
	if !w.HasSubmitted() {
		t.Error("Enter on last step should trigger Submit")
	}
}

func TestFormWizardEmptyStepsErrors(t *testing.T) {
	w := components.NewFormWizard(nil)
	if err := w.Next(); !errors.Is(err, components.ErrFormWizardEmpty) {
		t.Errorf("Next on empty wizard = %v, want ErrFormWizardEmpty", err)
	}
	if err := w.Prev(); !errors.Is(err, components.ErrFormWizardEmpty) {
		t.Errorf("Prev on empty wizard = %v, want ErrFormWizardEmpty", err)
	}
	if err := w.Submit(); !errors.Is(err, components.ErrFormWizardEmpty) {
		t.Errorf("Submit on empty wizard = %v, want ErrFormWizardEmpty", err)
	}
	if w.State() != nil {
		t.Errorf("State on empty wizard should be nil")
	}
	if _, ok := w.Step(); ok {
		t.Errorf("Step on empty wizard should return false")
	}
	if !w.IsLast() {
		t.Errorf("IsLast on empty wizard should be true")
	}
}

// TestFormWizardEndToEndIntegration walks a realistic happy-then-fix flow
// across all three steps to validate the wizard's controller contract end
// to end. Mirrors the Form → ErrorOverviewGroup integration test pattern.
func TestFormWizardEndToEndIntegration(t *testing.T) {
	step1 := components.NewFormState([]components.FormField{
		{Name: "email", Label: "Email", Kind: components.FormFieldText, Required: true,
			Validate: func(v string) error {
				if !strings.Contains(v, "@") {
					return errors.New("must include @")
				}
				return nil
			}},
	})
	step2 := components.NewFormState([]components.FormField{
		{Name: "tier", Label: "Tier", Kind: components.FormFieldTab, Required: true,
			Options: []components.SelectItem{
				{Label: "Free", Value: "free"},
				{Label: "Pro", Value: "pro"},
			}},
	})
	step3 := components.NewFormState([]components.FormField{
		{Name: "lang", Label: "Lang", Kind: components.FormFieldQuickSearch, Required: true,
			Options: []components.SelectItem{
				{Label: "Go", Value: "go"},
				{Label: "Rust", Value: "rust"},
			}},
	})
	w := components.NewFormWizard([]components.FormWizardStep{
		{Title: "Account", State: step1},
		{Title: "Plan", State: step2},
		{Title: "Stack", State: step3},
	})

	// Step 1: invalid email → Next blocked.
	step1.SetValue("email", "no-at-sign")
	if err := w.Next(); err == nil {
		t.Fatal("step 1: invalid email should block advance")
	}
	if w.CurrentStep() != 0 {
		t.Errorf("step 1: should remain on step 0, got %d", w.CurrentStep())
	}
	// Fix and advance.
	step1.SetValue("email", "alice@example.com")
	if err := w.Next(); err != nil {
		t.Fatalf("step 1: fixed email should advance, got %v", err)
	}
	// Step 2: Tab default = first option "free" (set explicitly via the
	// state since NewFormState already seeded it).
	if w.State().Values["tier"] != "free" {
		t.Errorf("step 2 default tier = %q, want \"free\"", w.State().Values["tier"])
	}
	if err := w.Next(); err != nil {
		t.Fatalf("step 2: should advance with default tier, got %v", err)
	}
	// Step 3: empty QuickSearch → submit blocked.
	if err := w.Submit(); err == nil {
		t.Fatal("step 3: empty QuickSearch should block submit")
	}
	if w.CurrentStep() != 2 {
		t.Errorf("Submit failure should jump to first failing step 2, got %d", w.CurrentStep())
	}
	// Fix step 3 and submit.
	step3.SetValue("lang", "go")
	if err := w.Submit(); err != nil {
		t.Fatalf("Submit after fix: %v", err)
	}
	if !w.HasSubmitted() {
		t.Error("HasSubmitted should be true after final Submit")
	}
	all := w.Values()
	if all["email"] != "alice@example.com" || all["tier"] != "free" || all["lang"] != "go" {
		t.Errorf("Values aggregate = %v", all)
	}
	overview := collectText(components.ErrorOverviewFromWizard(w))
	if !strings.Contains(overview, "<no errors>") {
		t.Errorf("expected empty overview after clean submit, got %q", overview)
	}
}
