package components_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/dh-kam/goink.go/pkg/components"
)

func basicFields() []components.FormField {
	return []components.FormField{
		{Name: "name", Label: "Name", Kind: components.FormFieldText, Required: true},
		{Name: "pw", Label: "Password", Kind: components.FormFieldPassword, Required: true},
		{Name: "role", Label: "Role", Kind: components.FormFieldSelect, Options: []components.SelectItem{
			{Label: "Admin", Value: "admin"},
			{Label: "User", Value: "user"},
		}},
	}
}
func TestFormRendersAllFieldKinds(t *testing.T) {
	fields := basicFields()
	state := components.NewFormState(fields)
	node := components.Form(components.FormProps{
		Fields:       state.Fields,
		Values:       state.Values,
		Errors:       state.Errors,
		FocusedIndex: state.Focused,
	})
	if node == nil {
		t.Fatal("Form returned nil node")
	}
	// Children layout: field, sep, field, sep, field → 2*len-1
	want := 2*len(fields) - 1
	if got := len(node.Children); got != want {
		t.Fatalf("row count = %d, want %d", got, want)
	}
}

func TestFormEmptyFields(t *testing.T) {
	node := components.Form(components.FormProps{})
	if node == nil || len(node.Children) != 0 {
		t.Fatalf("empty Form should render no children, got %d", len(node.Children))
	}
}

func TestFormShowsErrorRowsBelowFields(t *testing.T) {
	fields := basicFields()
	props := components.FormProps{
		Fields: fields,
		Values: map[string]string{"name": "", "pw": "", "role": "admin"},
		Errors: map[string]string{
			"name": "Name is required",
			"pw":   "Password is required",
		},
	}
	node := components.Form(props)
	// 3 fields + 2 separators + 2 error rows = 7
	if got := len(node.Children); got != 7 {
		t.Fatalf("rows = %d, want 7 (3 fields + 2 sep + 2 errors)", got)
	}
}

func TestFormFocusMarkerAppliesToFocusedRowOnly(t *testing.T) {
	fields := basicFields()
	props := components.FormProps{
		Fields:       fields,
		Values:       map[string]string{"name": "Alice", "pw": "", "role": "admin"},
		FocusedIndex: 0,
	}
	node := components.Form(props)
	first := node.Children[0].Text
	// children[1] is the inter-row separator newline; children[2] is the next field.
	second := node.Children[2].Text
	if !strings.HasPrefix(first, "> ") {
		t.Errorf("focused row prefix = %q, want '> '", first[:2])
	}
	if !strings.HasPrefix(second, "  ") {
		t.Errorf("unfocused row prefix = %q, want two spaces", second[:2])
	}
}

func TestFormSelectFieldUsesValueAsSelected(t *testing.T) {
	fields := []components.FormField{
		{Name: "role", Label: "Role", Kind: components.FormFieldSelect, Options: []components.SelectItem{
			{Label: "Admin", Value: "admin"},
			{Label: "User", Value: "user"},
			{Label: "Guest", Value: "guest"},
		}},
	}
	props := components.FormProps{
		Fields: fields,
		Values: map[string]string{"role": "guest"},
	}
	node := components.Form(props)
	row := node.Children[0].Text
	// Selected row "Guest" should be present in the rendered text.
	if !strings.Contains(row, "Guest") {
		t.Errorf("expected row to contain 'Guest', got %q", row)
	}
}

func TestNewFormStateInitialisesValues(t *testing.T) {
	state := components.NewFormState(basicFields())
	if state.Values["name"] != "" {
		t.Errorf("text default = %q, want empty", state.Values["name"])
	}
	if state.Values["role"] != "admin" {
		t.Errorf("select default = %q, want 'admin' (first option)", state.Values["role"])
	}
	if state.Focused != 0 {
		t.Errorf("Focused = %d, want 0", state.Focused)
	}
}

func TestNewFormStateSelectWithoutOptionsDefaultsEmpty(t *testing.T) {
	state := components.NewFormState([]components.FormField{
		{Name: "x", Label: "X", Kind: components.FormFieldSelect},
	})
	if state.Values["x"] != "" {
		t.Errorf("Values[x] = %q, want empty", state.Values["x"])
	}
}

func TestFocusNextWraps(t *testing.T) {
	state := components.NewFormState(basicFields())
	state.FocusNext()
	state.FocusNext()
	state.FocusNext() // wraps back to 0
	if state.Focused != 0 {
		t.Fatalf("Focused after 3x next = %d, want 0", state.Focused)
	}
}

func TestFocusPrevWraps(t *testing.T) {
	state := components.NewFormState(basicFields())
	state.FocusPrev() // wraps to last
	if state.Focused != len(state.Fields)-1 {
		t.Fatalf("Focused after prev from 0 = %d, want %d", state.Focused, len(state.Fields)-1)
	}
	state.FocusPrev()
	if state.Focused != len(state.Fields)-2 {
		t.Fatalf("Focused = %d, want %d", state.Focused, len(state.Fields)-2)
	}
}

func TestFocusEmptyFormNoOp(t *testing.T) {
	state := components.NewFormState(nil)
	state.FocusNext()
	state.FocusPrev()
	if state.Focused != 0 {
		t.Fatalf("Focused = %d on empty, want 0", state.Focused)
	}
}

func TestSetValueOnlyAcceptsKnownFields(t *testing.T) {
	state := components.NewFormState(basicFields())
	state.SetValue("name", "Alice")
	state.SetValue("unknown", "ignored")
	if state.Values["name"] != "Alice" {
		t.Errorf("Values[name] = %q, want Alice", state.Values["name"])
	}
	if _, ok := state.Values["unknown"]; ok {
		t.Errorf("Values[unknown] should not exist")
	}
}

func TestValidateRequiredPopulatesErrors(t *testing.T) {
	state := components.NewFormState(basicFields())
	if state.Validate() {
		t.Fatal("Validate should fail on empty required fields")
	}
	if state.Errors["name"] == "" {
		t.Error("expected error for required name field")
	}
	if state.Errors["pw"] == "" {
		t.Error("expected error for required password field")
	}
	if _, ok := state.Errors["role"]; ok {
		t.Error("role has a default value, should not error")
	}
}

func TestValidateRequiredRejectsWhitespace(t *testing.T) {
	state := components.NewFormState(basicFields())
	state.SetValue("name", "   ")
	state.SetValue("pw", "   ")
	if state.Validate() {
		t.Fatal("whitespace-only required values should fail validation")
	}
	if state.Errors["name"] == "" {
		t.Error("whitespace name should produce error")
	}
}

func TestValidateCustomValidator(t *testing.T) {
	fields := []components.FormField{
		{Name: "email", Label: "Email", Kind: components.FormFieldText, Required: true,
			Validate: func(v string) error {
				if !strings.Contains(v, "@") {
					return errors.New("invalid email")
				}
				return nil
			}},
	}
	state := components.NewFormState(fields)
	state.SetValue("email", "not-an-email")
	if state.Validate() {
		t.Fatal("custom validator should reject value")
	}
	if state.Errors["email"] != "invalid email" {
		t.Errorf("error = %q, want 'invalid email'", state.Errors["email"])
	}

	state.SetValue("email", "user@example.com")
	if !state.Validate() {
		t.Fatalf("Validate should succeed, errors=%v", state.Errors)
	}
}

func TestValidateCustomValidatorSkippedOnEmptyOptionalField(t *testing.T) {
	fields := []components.FormField{
		{Name: "nick", Label: "Nick", Kind: components.FormFieldText,
			Validate: func(v string) error { return errors.New("never reached") }},
	}
	state := components.NewFormState(fields)
	if !state.Validate() {
		t.Fatalf("optional empty field must pass validation, errors=%v", state.Errors)
	}
}

func TestValidateClearsPreviousErrors(t *testing.T) {
	state := components.NewFormState(basicFields())
	state.Validate() // populate errors
	if len(state.Errors) == 0 {
		t.Fatal("setup: expected errors after first Validate")
	}
	state.SetValue("name", "Alice")
	state.SetValue("pw", "secret")
	if !state.Validate() {
		t.Fatalf("expected pass, errors=%v", state.Errors)
	}
	if len(state.Errors) != 0 {
		t.Errorf("Errors not cleared, got %v", state.Errors)
	}
}

func TestSubmitReturnsCopyOnSuccess(t *testing.T) {
	state := components.NewFormState(basicFields())
	state.SetValue("name", "Alice")
	state.SetValue("pw", "secret")
	out, ok := state.Submit()
	if !ok {
		t.Fatalf("Submit should succeed, errors=%v", state.Errors)
	}
	if out["name"] != "Alice" || out["pw"] != "secret" || out["role"] != "admin" {
		t.Errorf("Submit returned %v", out)
	}
	// Mutation of returned map should not leak into state.
	out["name"] = "Bob"
	if state.Values["name"] != "Alice" {
		t.Error("Submit returned map shares storage with state")
	}
}

func TestFormLongLabelIsTruncated(t *testing.T) {
	fields := []components.FormField{
		{Name: "x", Label: "ThisLabelIsObnoxiouslyLong", Kind: components.FormFieldText},
	}
	node := components.Form(components.FormProps{
		Fields: fields,
		Values: map[string]string{"x": ""},
	})
	row := node.Children[0].Text
	// padLabel truncates to formLabelWidth (14); the full label must not appear.
	if strings.Contains(row, "ThisLabelIsObnoxiouslyLong") {
		t.Errorf("expected label to be truncated, got %q", row)
	}
}

func TestFormSelectFieldWithMissingValueDefaultsToFirst(t *testing.T) {
	fields := []components.FormField{
		{Name: "role", Label: "Role", Kind: components.FormFieldSelect, Options: []components.SelectItem{
			{Label: "Admin", Value: "admin"},
			{Label: "User", Value: "user"},
		}},
	}
	props := components.FormProps{
		Fields: fields,
		Values: map[string]string{"role": "no-such-value"},
	}
	node := components.Form(props)
	row := node.Children[0].Text
	// Falling back to index 0 means Admin remains highlighted.
	if !strings.Contains(row, "Admin") {
		t.Errorf("expected fallback selection 'Admin', got %q", row)
	}
}

func TestFormHandlesNilValuesAndErrors(t *testing.T) {
	fields := basicFields()
	// nil maps shouldn't panic; Form treats them as empty.
	node := components.Form(components.FormProps{Fields: fields})
	want := 2*len(fields) - 1
	if got := len(node.Children); got != want {
		t.Fatalf("rows = %d, want %d", got, want)
	}
}

func TestSubmitFailsAndReturnsNil(t *testing.T) {
	state := components.NewFormState(basicFields())
	out, ok := state.Submit()
	if ok || out != nil {
		t.Fatalf("Submit on invalid form returned ok=%v out=%v", ok, out)
	}
	if len(state.Errors) == 0 {
		t.Error("Submit failure should populate Errors")
	}
}

// ---------------------------------------------------------------------------
// Confirm + MultiSelect field kinds.
// ---------------------------------------------------------------------------

func TestNewFormStateConfirmHonoursDefault(t *testing.T) {
	fields := []components.FormField{
		{Name: "ack", Label: "Accept", Kind: components.FormFieldConfirm, Default: "true"},
		{Name: "later", Label: "Later", Kind: components.FormFieldConfirm}, // no default → unanswered
		{Name: "junk", Label: "Junk", Kind: components.FormFieldConfirm, Default: "maybe"},
	}
	state := components.NewFormState(fields)
	if state.Values["ack"] != "true" {
		t.Errorf("ack default = %q, want \"true\"", state.Values["ack"])
	}
	if state.Values["later"] != "" {
		t.Errorf("later default = %q, want unanswered (\"\")", state.Values["later"])
	}
	if state.Values["junk"] != "" {
		t.Errorf("invalid default should normalise to unanswered, got %q", state.Values["junk"])
	}
}

func TestNewFormStateMultiSelectHonoursDefault(t *testing.T) {
	fields := []components.FormField{
		{Name: "tags", Label: "Tags", Kind: components.FormFieldMultiSelect, Default: "  go, ts ,,rust"},
	}
	state := components.NewFormState(fields)
	if state.Values["tags"] != "go,ts,rust" {
		t.Errorf("tags default = %q, want \"go,ts,rust\"", state.Values["tags"])
	}
	got := components.FormMultiSelectValues(state.Values["tags"])
	if !reflect.DeepEqual(got, []string{"go", "ts", "rust"}) {
		t.Errorf("decoded values = %v", got)
	}
}

func TestValidateConfirmRequired(t *testing.T) {
	fields := []components.FormField{
		{Name: "ack", Label: "Accept", Kind: components.FormFieldConfirm, Required: true},
	}
	state := components.NewFormState(fields)
	if state.Validate() {
		t.Fatal("unanswered required Confirm should fail validation")
	}
	if state.Errors["ack"] == "" {
		t.Errorf("expected required error for unanswered Confirm")
	}

	state.SetValue("ack", "false")
	if !state.Validate() {
		t.Fatalf("explicit no should pass required, errors=%v", state.Errors)
	}
}

func TestValidateConfirmCustomValidator(t *testing.T) {
	fields := []components.FormField{
		{Name: "ack", Label: "Accept", Kind: components.FormFieldConfirm, Required: true,
			Validate: func(value string) error {
				if value != "true" {
					return errors.New("must accept the terms")
				}
				return nil
			}},
	}
	state := components.NewFormState(fields)
	state.SetValue("ack", "false")
	if state.Validate() {
		t.Fatal("Confirm validator should reject \"false\"")
	}
	if state.Errors["ack"] != "must accept the terms" {
		t.Errorf("error = %q", state.Errors["ack"])
	}
	state.SetValue("ack", "true")
	if !state.Validate() {
		t.Fatalf("expected pass after accept, errors=%v", state.Errors)
	}
}

func TestValidateMultiSelectRequired(t *testing.T) {
	fields := []components.FormField{
		{Name: "tags", Label: "Tags", Kind: components.FormFieldMultiSelect, Required: true,
			Options: []components.SelectItem{{Label: "Go", Value: "go"}, {Label: "TS", Value: "ts"}}},
	}
	state := components.NewFormState(fields)
	if state.Validate() {
		t.Fatal("empty required MultiSelect should fail validation")
	}
	state.SetValue("tags", "  ,  ,") // CSV of only blanks → still empty
	if state.Validate() {
		t.Fatal("CSV of blanks should still count as empty for Required")
	}
	state.SetValue("tags", "go")
	if !state.Validate() {
		t.Fatalf("non-empty MultiSelect should pass, errors=%v", state.Errors)
	}
}

func TestValidateMultiSelectCustomValidator(t *testing.T) {
	fields := []components.FormField{
		{Name: "tags", Label: "Tags", Kind: components.FormFieldMultiSelect, Required: true,
			Options: []components.SelectItem{{Label: "Go", Value: "go"}, {Label: "TS", Value: "ts"}, {Label: "Rust", Value: "rust"}},
			Validate: func(value string) error {
				if len(components.FormMultiSelectValues(value)) > 2 {
					return errors.New("pick at most 2")
				}
				return nil
			}},
	}
	state := components.NewFormState(fields)
	state.SetValue("tags", "go,ts,rust")
	if state.Validate() {
		t.Fatal("validator should reject 3 selections")
	}
	if state.Errors["tags"] != "pick at most 2" {
		t.Errorf("error = %q", state.Errors["tags"])
	}
	state.SetValue("tags", components.EncodeFormMultiSelect([]string{"go", "rust"}))
	if !state.Validate() {
		t.Fatalf("validator should accept 2 selections, errors=%v", state.Errors)
	}
}

func TestFormRendersConfirmAndMultiSelectKinds(t *testing.T) {
	fields := []components.FormField{
		{Name: "ack", Label: "Accept", Kind: components.FormFieldConfirm, Default: "true"},
		{Name: "tags", Label: "Tags", Kind: components.FormFieldMultiSelect, Default: "go,rust",
			Options: []components.SelectItem{
				{Label: "Go", Value: "go"},
				{Label: "TS", Value: "ts"},
				{Label: "Rust", Value: "rust"},
			}},
	}
	state := components.NewFormState(fields)
	node := components.Form(components.FormProps{
		Fields:       state.Fields,
		Values:       state.Values,
		FocusedIndex: 0,
	})
	if node == nil {
		t.Fatal("Form returned nil")
	}
	out := collectText(node)
	// Confirm rendering with default=true puts [Y/n] into the prompt.
	if !strings.Contains(out, "[Y/n]") {
		t.Errorf("expected Confirm hint [Y/n] in output, got %q", out)
	}
	// MultiSelect should mark the two checked tags ("go" and "rust").
	if !strings.Contains(out, "Go") || !strings.Contains(out, "Rust") {
		t.Errorf("expected MultiSelect option labels in output, got %q", out)
	}
}

// ---------------------------------------------------------------------------
// OrderedErrors / ErrorOverviewFromForm helpers.
// ---------------------------------------------------------------------------

func TestOrderedErrorsPreservesFieldOrder(t *testing.T) {
	state := components.NewFormState(basicFields())
	if state.Validate() {
		t.Fatal("setup: expected validation failure")
	}
	got := state.OrderedErrors()
	if len(got) < 2 {
		t.Fatalf("expected at least 2 errors, got %v", got)
	}
	// Order is name, pw, role (role passes; expect name first).
	if !strings.HasPrefix(got[0].Error(), "Name:") {
		t.Errorf("first error = %q, want Name prefix", got[0].Error())
	}
	if !strings.HasPrefix(got[1].Error(), "Password:") {
		t.Errorf("second error = %q, want Password prefix", got[1].Error())
	}
}

func TestOrderedErrorsEmptyWhenValid(t *testing.T) {
	state := components.NewFormState(basicFields())
	state.SetValue("name", "Alice")
	state.SetValue("pw", "secret")
	if !state.Validate() {
		t.Fatalf("setup: expected pass, errors=%v", state.Errors)
	}
	got := state.OrderedErrors()
	if len(got) != 0 {
		t.Errorf("expected no errors, got %v", got)
	}
	if got == nil {
		t.Errorf("OrderedErrors should return empty (non-nil) slice")
	}
}

func TestFieldErrorReturnsBareMessage(t *testing.T) {
	state := components.NewFormState(basicFields())
	state.Validate()
	err := state.FieldError("name")
	if err == nil {
		t.Fatal("expected error for required name field")
	}
	if strings.HasPrefix(err.Error(), "Name:") {
		t.Errorf("FieldError should return bare message, got %q", err.Error())
	}
	if state.FieldError("role") != nil {
		t.Errorf("role has default value, FieldError should be nil")
	}
	if state.FieldError("does-not-exist") != nil {
		t.Errorf("unknown name should return nil")
	}
}

func TestHasErrors(t *testing.T) {
	state := components.NewFormState(basicFields())
	if state.HasErrors() {
		t.Fatal("fresh state should have no errors")
	}
	state.Validate()
	if !state.HasErrors() {
		t.Fatal("after failing Validate, HasErrors should be true")
	}
	state.SetValue("name", "Alice")
	state.SetValue("pw", "secret")
	state.Validate()
	if state.HasErrors() {
		t.Fatal("after passing Validate, HasErrors should be false")
	}
}

func TestErrorOverviewFromFormPipesValidationErrors(t *testing.T) {
	state := components.NewFormState(basicFields())
	if state.Validate() {
		t.Fatal("setup: expected failure")
	}
	node := components.ErrorOverviewFromForm(state)
	if node == nil {
		t.Fatal("ErrorOverviewFromForm returned nil")
	}
	out := collectText(node)
	for _, want := range []string{"ERRORS", "Validation", "Name:", "Password:"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output, got %q", want, out)
		}
	}
	if strings.Contains(out, "Runtime") {
		t.Errorf("no runtime errors supplied; Runtime sub-section should be hidden, got %q", out)
	}
}

func TestErrorOverviewFromFormCarriesRuntimeErrors(t *testing.T) {
	state := components.NewFormState(basicFields())
	state.SetValue("name", "Alice")
	state.SetValue("pw", "secret")
	if !state.Validate() {
		t.Fatalf("setup: expected pass")
	}
	node := components.ErrorOverviewFromForm(state, errors.New("network timeout"))
	out := collectText(node)
	if !strings.Contains(out, "Runtime") {
		t.Errorf("expected Runtime sub-section in output, got %q", out)
	}
	if !strings.Contains(out, "network timeout") {
		t.Errorf("expected runtime error text, got %q", out)
	}
}

func TestErrorOverviewFromFormHandlesNilState(t *testing.T) {
	node := components.ErrorOverviewFromForm(nil, errors.New("boom"))
	out := collectText(node)
	if !strings.Contains(out, "boom") {
		t.Errorf("expected runtime error text even with nil state, got %q", out)
	}
}

func TestSubmitStaysOnFocusedFieldAfterFailure(t *testing.T) {
	state := components.NewFormState(basicFields())
	state.FocusNext() // focus = 1 (pw)
	if _, ok := state.Submit(); ok {
		t.Fatal("Submit should fail on empty required fields")
	}
	if state.Focused != 1 {
		t.Errorf("Submit failure should not change Focused, got %d", state.Focused)
	}
	if !state.HasErrors() {
		t.Errorf("Submit failure should leave errors populated")
	}
}

// ---------------------------------------------------------------------------
// Round-trip helpers for FormFieldMultiSelect.
// ---------------------------------------------------------------------------

// TestFormToErrorOverviewIntegration walks a realistic submit-fix-submit
// loop across mixed field kinds (Text, Confirm, MultiSelect) and asserts
// the ErrorOverviewGroup helper surfaces the correct failures at each
// step. This is the canonical "Form → ErrorOverviewGroup" pipeline.
func TestFormToErrorOverviewIntegration(t *testing.T) {
	fields := []components.FormField{
		{Name: "email", Label: "Email", Kind: components.FormFieldText, Required: true,
			Validate: func(v string) error {
				if !strings.Contains(v, "@") {
					return errors.New("must include @")
				}
				return nil
			}},
		{Name: "tos", Label: "Terms", Kind: components.FormFieldConfirm, Required: true,
			Validate: func(v string) error {
				if v != "true" {
					return errors.New("must accept the terms")
				}
				return nil
			}},
		{Name: "tags", Label: "Tags", Kind: components.FormFieldMultiSelect, Required: true,
			Options: []components.SelectItem{
				{Label: "Go", Value: "go"},
				{Label: "TS", Value: "ts"},
			}},
	}
	state := components.NewFormState(fields)

	// Step 1 — empty submit. All three Required errors should appear in
	// the rendered ErrorOverviewGroup, in field order.
	if _, ok := state.Submit(); ok {
		t.Fatal("step 1: empty Submit should fail")
	}
	out := collectText(components.ErrorOverviewFromForm(state))
	for _, want := range []string{"Email:", "Terms:", "Tags:"} {
		if !strings.Contains(out, want) {
			t.Errorf("step 1: expected %q in overview, got %q", want, out)
		}
	}
	emailIdx := strings.Index(out, "Email:")
	termsIdx := strings.Index(out, "Terms:")
	tagsIdx := strings.Index(out, "Tags:")
	if !(emailIdx < termsIdx && termsIdx < tagsIdx) {
		t.Errorf("step 1: errors not in field order: email=%d terms=%d tags=%d", emailIdx, termsIdx, tagsIdx)
	}

	// Step 2 — partial fix. Email value is invalid; Terms answered "no";
	// Tags has one selection.
	state.SetValue("email", "no-at-sign")
	state.SetValue("tos", "false")
	state.SetValue("tags", "go")
	if _, ok := state.Submit(); ok {
		t.Fatal("step 2: invalid Submit should still fail")
	}
	out = collectText(components.ErrorOverviewFromForm(state, errors.New("network blip")))
	if !strings.Contains(out, "must include @") {
		t.Errorf("step 2: expected email validator message, got %q", out)
	}
	if !strings.Contains(out, "must accept the terms") {
		t.Errorf("step 2: expected terms validator message, got %q", out)
	}
	if strings.Contains(out, "Tags:") {
		t.Errorf("step 2: tags has a value; should not be in errors, got %q", out)
	}
	if !strings.Contains(out, "Runtime") || !strings.Contains(out, "network blip") {
		t.Errorf("step 2: expected runtime sub-section with network blip, got %q", out)
	}

	// Step 3 — all green. Submit succeeds and the overview is empty.
	state.SetValue("email", "user@example.com")
	state.SetValue("tos", "true")
	values, ok := state.Submit()
	if !ok {
		t.Fatalf("step 3: Submit should succeed, errors=%v", state.OrderedErrors())
	}
	if values["email"] != "user@example.com" || values["tos"] != "true" || values["tags"] != "go" {
		t.Errorf("step 3: submitted values = %v", values)
	}
	out = collectText(components.ErrorOverviewFromForm(state))
	if !strings.Contains(out, "<no errors>") {
		t.Errorf("step 3: expected empty overview placeholder, got %q", out)
	}
}

// ---------------------------------------------------------------------------
// FormFieldTab + FormFieldQuickSearch field kinds.
// ---------------------------------------------------------------------------

func tabAndQuickSearchFields() []components.FormField {
	return []components.FormField{
		{Name: "env", Label: "Env", Kind: components.FormFieldTab, Required: true,
			Options: []components.SelectItem{
				{Label: "Dev", Value: "dev"},
				{Label: "Stage", Value: "stage"},
				{Label: "Prod", Value: "prod"},
			},
			Default: "stage",
		},
		{Name: "lang", Label: "Lang", Kind: components.FormFieldQuickSearch, Required: true,
			Options: []components.SelectItem{
				{Label: "Go", Value: "go"},
				{Label: "Rust", Value: "rust"},
				{Label: "TypeScript", Value: "ts"},
			}},
	}
}

func TestNewFormStateTabHonoursDefault(t *testing.T) {
	state := components.NewFormState(tabAndQuickSearchFields())
	if state.Values["env"] != "stage" {
		t.Errorf("env default = %q, want \"stage\"", state.Values["env"])
	}
}

func TestNewFormStateTabFallsBackToFirstOption(t *testing.T) {
	fields := []components.FormField{
		{Name: "env", Label: "Env", Kind: components.FormFieldTab,
			Options: []components.SelectItem{
				{Label: "Dev", Value: "dev"},
				{Label: "Prod", Value: "prod"},
			}},
		{Name: "bad", Label: "Bad", Kind: components.FormFieldTab, Default: "nope",
			Options: []components.SelectItem{
				{Label: "A", Value: "a"},
			}},
		{Name: "empty", Label: "Empty", Kind: components.FormFieldTab},
	}
	state := components.NewFormState(fields)
	if state.Values["env"] != "dev" {
		t.Errorf("env without default = %q, want \"dev\"", state.Values["env"])
	}
	if state.Values["bad"] != "a" {
		t.Errorf("bad default = %q, want fallback to first option \"a\"", state.Values["bad"])
	}
	if state.Values["empty"] != "" {
		t.Errorf("empty options Tab = %q, want \"\"", state.Values["empty"])
	}
}

func TestNewFormStateQuickSearchEmptyByDefault(t *testing.T) {
	state := components.NewFormState(tabAndQuickSearchFields())
	if state.Values["lang"] != "" {
		t.Errorf("QuickSearch should default to empty when Default unset, got %q", state.Values["lang"])
	}
}

func TestNewFormStateQuickSearchHonoursDefault(t *testing.T) {
	fields := []components.FormField{
		{Name: "lang", Label: "Lang", Kind: components.FormFieldQuickSearch, Default: "rust",
			Options: []components.SelectItem{
				{Label: "Go", Value: "go"},
				{Label: "Rust", Value: "rust"},
			}},
		{Name: "bad", Label: "Bad", Kind: components.FormFieldQuickSearch, Default: "nope",
			Options: []components.SelectItem{{Label: "A", Value: "a"}}},
	}
	state := components.NewFormState(fields)
	if state.Values["lang"] != "rust" {
		t.Errorf("QuickSearch default = %q, want \"rust\"", state.Values["lang"])
	}
	if state.Values["bad"] != "a" {
		t.Errorf("invalid default should fall back to first option, got %q", state.Values["bad"])
	}
}

func TestValidateTabRequired(t *testing.T) {
	fields := []components.FormField{
		{Name: "env", Label: "Env", Kind: components.FormFieldTab, Required: true,
			// No Options yet — Required should flag the empty value.
		},
	}
	state := components.NewFormState(fields)
	if state.Validate() {
		t.Fatal("Tab with no options should fail Required validation when value is empty")
	}
	if state.Errors["env"] == "" {
		t.Errorf("expected required error for empty Tab, got %v", state.Errors)
	}
}

func TestValidateTabCustomValidator(t *testing.T) {
	fields := []components.FormField{
		{Name: "env", Label: "Env", Kind: components.FormFieldTab, Required: true,
			Options: []components.SelectItem{
				{Label: "Dev", Value: "dev"},
				{Label: "Prod", Value: "prod"},
			},
			Default: "dev",
			Validate: func(v string) error {
				if v == "prod" {
					return errors.New("prod requires sign-off")
				}
				return nil
			}},
	}
	state := components.NewFormState(fields)
	if !state.Validate() {
		t.Fatalf("default \"dev\" should pass, errors=%v", state.Errors)
	}
	state.SetValue("env", "prod")
	if state.Validate() {
		t.Fatal("validator should reject \"prod\"")
	}
	if state.Errors["env"] != "prod requires sign-off" {
		t.Errorf("error = %q", state.Errors["env"])
	}
}

func TestValidateQuickSearchRequired(t *testing.T) {
	fields := []components.FormField{
		{Name: "lang", Label: "Lang", Kind: components.FormFieldQuickSearch, Required: true,
			Options: []components.SelectItem{{Label: "Go", Value: "go"}}},
	}
	state := components.NewFormState(fields)
	if state.Validate() {
		t.Fatal("empty QuickSearch should fail Required")
	}
	state.SetValue("lang", "go")
	if !state.Validate() {
		t.Fatalf("non-empty QuickSearch should pass, errors=%v", state.Errors)
	}
}

func TestValidateQuickSearchCustomValidator(t *testing.T) {
	fields := []components.FormField{
		{Name: "lang", Label: "Lang", Kind: components.FormFieldQuickSearch,
			Options: []components.SelectItem{
				{Label: "Go", Value: "go"},
				{Label: "Rust", Value: "rust"},
			},
			Validate: func(v string) error {
				if v != "go" {
					return errors.New("only go allowed")
				}
				return nil
			}},
	}
	state := components.NewFormState(fields)
	state.SetValue("lang", "rust")
	if state.Validate() {
		t.Fatal("validator should reject \"rust\"")
	}
	state.SetValue("lang", "go")
	if !state.Validate() {
		t.Fatalf("\"go\" should pass, errors=%v", state.Errors)
	}
}

func TestFormRendersTabAndQuickSearchKinds(t *testing.T) {
	state := components.NewFormState(tabAndQuickSearchFields())
	state.SetValue("lang", "go")
	node := components.Form(components.FormProps{
		Fields:       state.Fields,
		Values:       state.Values,
		FocusedIndex: 0,
	})
	if node == nil {
		t.Fatal("Form returned nil")
	}
	out := collectText(node)
	// Tab header should bracket the active option; default = "stage".
	if !strings.Contains(out, "[Stage]") {
		t.Errorf("expected Tab header to bracket active option, got %q", out)
	}
	// QuickSearch always renders the search prompt prefix.
	if !strings.Contains(out, "Search:") {
		t.Errorf("expected QuickSearch prompt, got %q", out)
	}
	// Selected QuickSearch row's label should appear.
	if !strings.Contains(out, "Go") {
		t.Errorf("expected QuickSearch option label, got %q", out)
	}
}

func TestEncodeAndDecodeFormMultiSelectRoundTrip(t *testing.T) {
	cases := []struct {
		in  []string
		csv string
	}{
		{nil, ""},
		{[]string{}, ""},
		{[]string{"go"}, "go"},
		{[]string{"go", "ts"}, "go,ts"},
		{[]string{"  go ", "", " ts"}, "go,ts"},
	}
	for _, tc := range cases {
		got := components.EncodeFormMultiSelect(tc.in)
		if got != tc.csv {
			t.Errorf("Encode(%v) = %q, want %q", tc.in, got, tc.csv)
		}
		decoded := components.FormMultiSelectValues(got)
		if got == "" {
			if len(decoded) != 0 {
				t.Errorf("Decode(\"\") = %v, want empty", decoded)
			}
			continue
		}
		// Re-encode to verify idempotence on round trip.
		if components.EncodeFormMultiSelect(decoded) != tc.csv {
			t.Errorf("round-trip mismatch for %q", tc.csv)
		}
	}
}
