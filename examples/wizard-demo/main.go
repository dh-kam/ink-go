package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/ink"
)

// Demo: a 3-step signup wizard wired through FormWizard. The driver loop
// simulates user input by calling controller methods directly so the demo
// runs in any environment (no real TTY required) and exercises the
// validation gate, the cross-step Prev that preserves state, and the
// final aggregate Submit. The new FormFieldTab and FormFieldQuickSearch
// kinds are exercised in steps 2 and 3 respectively.
func main() {
	step1 := components.NewFormState([]components.FormField{
		{
			Name: "email", Label: "Email", Kind: components.FormFieldText, Required: true,
			Validate: func(v string) error {
				if !strings.Contains(v, "@") || !strings.Contains(v, ".") {
					return errors.New("must look like name@host.tld")
				}
				return nil
			},
		},
	})

	step2 := components.NewFormState([]components.FormField{
		{
			Name: "tier", Label: "Tier", Kind: components.FormFieldTab, Required: true,
			Options: []components.SelectItem{
				{Label: "Free", Value: "free"},
				{Label: "Pro", Value: "pro"},
				{Label: "Enterprise", Value: "ent"},
			},
			Default: "free",
		},
	})

	step3 := components.NewFormState([]components.FormField{
		{
			Name: "lang", Label: "Lang", Kind: components.FormFieldQuickSearch, Required: true,
			Options: []components.SelectItem{
				{Label: "Go", Value: "go"},
				{Label: "Rust", Value: "rust"},
				{Label: "TypeScript", Value: "ts"},
				{Label: "Python", Value: "py"},
			},
		},
	})

	wizard := components.NewFormWizard([]components.FormWizardStep{
		{Title: "Account", State: step1},
		{Title: "Plan", State: step2},
		{Title: "Stack", State: step3},
	})

	render := func(label string) {
		state := wizard.State()
		focused := -1
		if state != nil {
			focused = state.Focused
		}
		title := "(empty)"
		if step, ok := wizard.Step(); ok {
			title = step.Title
		}
		fmt.Printf("--- %s | step %d/%d (%s) focus=%d ---\n",
			label, wizard.CurrentStep()+1, wizard.TotalSteps(), title, focused)
		if state == nil {
			return
		}
		node := components.Form(components.FormProps{
			Fields:       state.Fields,
			Values:       state.Values,
			Errors:       state.Errors,
			FocusedIndex: state.Focused,
		})
		for _, child := range node.Children {
			out := ink.RenderToString(child)
			if out == "\n" {
				continue
			}
			fmt.Println(out)
		}
	}

	render("step 1 initial")

	// Try to advance with an invalid email — validator rejects.
	step1.SetValue("email", "not-an-email")
	if err := wizard.Next(); err != nil {
		fmt.Printf("Next blocked: %s\n", err.Error())
	}
	render("step 1 after invalid Next")

	// Fix and advance to step 2.
	step1.SetValue("email", "alice@example.com")
	if err := wizard.Next(); err != nil {
		fmt.Printf("unexpected: %s\n", err.Error())
	}
	render("step 2 (Tab field)")

	// Pick "pro" and advance.
	step2.SetValue("tier", "pro")
	if err := wizard.Next(); err != nil {
		fmt.Printf("unexpected: %s\n", err.Error())
	}
	render("step 3 (QuickSearch field)")

	// Try Submit with empty QuickSearch — validator rejects, wizard jumps
	// to first failing step (which is step 3 here).
	if err := wizard.Submit(); err != nil {
		fmt.Printf("Submit blocked: %s\n", err.Error())
	}

	// Pick "go" and submit.
	step3.SetValue("lang", "go")
	if err := wizard.Submit(); err != nil {
		fmt.Printf("Submit failed: %s\n", err.Error())
	} else {
		fmt.Println()
		fmt.Println("Wizard submitted. Aggregate values:")
		for k, v := range wizard.Values() {
			fmt.Printf("  %-8s %s\n", k+":", v)
		}
	}

	// Demonstrate Prev preserves state — go back to step 1 and confirm.
	_ = wizard.Prev()
	_ = wizard.Prev()
	render("returned to step 1 after Prev twice (state preserved)")
}
