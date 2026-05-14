package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/ink"
)

// Demo: a 3-field signup form (name + email + password) driven by a
// FormState controller. The driver loop simulates user input by calling
// state methods directly so the demo runs in any environment (no real
// TTY required) and exercises both Required and custom Validate paths.
func main() {
	fields := []components.FormField{
		{
			Name:     "name",
			Label:    "Name",
			Kind:     components.FormFieldText,
			Required: true,
		},
		{
			Name:     "email",
			Label:    "Email",
			Kind:     components.FormFieldText,
			Required: true,
			Validate: func(value string) error {
				if !strings.Contains(value, "@") || !strings.Contains(value, ".") {
					return errors.New("must look like name@host.tld")
				}
				return nil
			},
		},
		{
			Name:     "password",
			Label:    "Password",
			Kind:     components.FormFieldPassword,
			Required: true,
			Validate: func(value string) error {
				if len(value) < 6 {
					return errors.New("at least 6 characters")
				}
				return nil
			},
		},
	}

	state := components.NewFormState(fields)

	render := func(label string) {
		fmt.Printf("--- %s (focused=%d) ---\n", label, state.Focused)
		node := components.Form(components.FormProps{
			Fields:       state.Fields,
			Values:       state.Values,
			Errors:       state.Errors,
			FocusedIndex: state.Focused,
		})
		// Form returns a flat list of text-node children separated by
		// newline nodes; render each child on its own line so the demo
		// reads naturally without needing a real layout engine.
		for _, child := range node.Children {
			out := ink.RenderToString(child)
			if out == "\n" {
				continue
			}
			fmt.Println(out)
		}
	}

	render("initial (all empty)")

	// Try to submit empty — should populate Required errors.
	if _, ok := state.Submit(); !ok {
		render("submit attempt 1: required errors")
	}

	// Fill in invalid values to trip the custom validators.
	state.SetValue("name", "Alice")
	state.SetValue("email", "not-an-email")
	state.SetValue("password", "abc")
	state.FocusNext()
	if _, ok := state.Submit(); !ok {
		render("submit attempt 2: custom validator errors")
	}

	// Fix the invalid fields.
	state.SetValue("email", "alice@example.com")
	state.SetValue("password", "hunter2")
	state.FocusNext()

	if values, ok := state.Submit(); ok {
		render("after fix: clean submit")
		fmt.Println()
		fmt.Println("Submitted values:")
		for _, f := range fields {
			if f.Kind == components.FormFieldPassword {
				fmt.Printf("  %-10s %s\n", f.Name+":", strings.Repeat("*", len(values[f.Name])))
			} else {
				fmt.Printf("  %-10s %s\n", f.Name+":", values[f.Name])
			}
		}
	}
}
