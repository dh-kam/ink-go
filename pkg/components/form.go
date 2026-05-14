package components

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dh-kam/ink-go/pkg/vdom"
)

// FormFieldKind discriminates how a FormField is rendered and edited.
//
// FormFieldText  → plain TextInput.
// FormFieldPassword → masked TextInput (PasswordInput).
// FormFieldSelect → Select list using Options as the choices.
// FormFieldConfirm → yes/no Confirm prompt; canonical value is "true"/"false".
// FormFieldMultiSelect → checkbox list; canonical value is a comma-joined CSV
// of selected option Values (empty string when nothing is checked).
// FormFieldTab → horizontal segmented control; canonical value is the Value
// of the currently active option.
// FormFieldQuickSearch → filter-as-you-type list; canonical value is the
// Value of the currently selected option.
type FormFieldKind int

const (
	// FormFieldText is a single-line free-text field.
	FormFieldText FormFieldKind = iota
	// FormFieldPassword is a masked free-text field.
	FormFieldPassword
	// FormFieldSelect is a single-choice dropdown driven by Options.
	FormFieldSelect
	// FormFieldConfirm is a yes/no prompt. The canonical Values entry is the
	// string "true" or "false" once an answer is recorded; "" means
	// unanswered. Required treats an unanswered field as missing.
	FormFieldConfirm
	// FormFieldMultiSelect is a checkbox list driven by Options. The
	// canonical Values entry is the comma-joined list of selected option
	// Values, in the order they appear in Options. Required treats the
	// empty CSV as missing. Use FormMultiSelectValues / EncodeFormMultiSelect
	// to round-trip between the CSV form and a []string.
	FormFieldMultiSelect
	// FormFieldTab is a horizontal segmented control driven by Options. The
	// canonical Values entry is the Value of the currently active option;
	// it defaults to the first option's Value when Default is empty (or
	// references an unknown option). Required treats an empty value as
	// missing — useful when Options is empty until populated dynamically.
	FormFieldTab
	// FormFieldQuickSearch is a filter-as-you-type list driven by Options.
	// The canonical Values entry is the Value of the currently selected
	// option. Required treats an empty value as missing. Validation matches
	// FormFieldSelect: validators receive the bare option Value.
	FormFieldQuickSearch
)

// formMultiSelectSeparator is the canonical separator used to encode
// MultiSelect values inside the FormState.Values map.
const formMultiSelectSeparator = ","

// FormField describes one row of a Form. Name is the stable key used to
// look up the field's value in Values / Errors maps. Label is shown to the
// user. Validate, when set, is invoked on Submit (after Required) and may
// return an error to surface as the field's error text. Options is used
// by FormFieldSelect, FormFieldMultiSelect, FormFieldTab, and
// FormFieldQuickSearch; ignored for the other kinds.
//
// Default seeds FormState.Values for kinds where no natural empty state
// exists. For Confirm pass "true" or "false"; for MultiSelect pass a CSV of
// option Values; for Tab/QuickSearch pass the Value of the option that
// should be selected initially (anything outside Options falls back to the
// first option). Ignored for Text/Password/Select.
type FormField struct {
	Name     string
	Label    string
	Kind     FormFieldKind
	Required bool
	Validate func(value string) error
	Options  []SelectItem
	Default  string
}

// FormProps is the controlled props bag for the pure Form render. Values
// and Errors are keyed by FormField.Name. FocusedIndex controls which row
// is highlighted (and is the only row whose input shows focus styling).
type FormProps struct {
	Fields       []FormField
	Values       map[string]string
	Errors       map[string]string
	FocusedIndex int
}

// formLabelWidth is the column width used to pad labels so input boxes
// align across rows. Long labels are truncated; short labels are padded.
const formLabelWidth = 14

// Form renders the field list as a stack of label + input rows, with any
// validation error appended directly under its field. Pure render — pair
// it with FormState (or your own controller) to drive Values / Errors /
// FocusedIndex on each frame.
func Form(props FormProps) *vdom.Node {
	if len(props.Fields) == 0 {
		return vdom.CreateElement("form", nil)
	}

	values := props.Values
	if values == nil {
		values = map[string]string{}
	}
	errors := props.Errors
	if errors == nil {
		errors = map[string]string{}
	}

	rows := make([]*vdom.Node, 0, len(props.Fields)*2)
	for i, field := range props.Fields {
		focused := i == props.FocusedIndex
		if i > 0 {
			rows = append(rows, vdom.CreateTextNode("\n"))
		}
		rows = append(rows, renderFormField(field, values[field.Name], focused))
		if msg, ok := errors[field.Name]; ok && msg != "" {
			rows = append(rows, vdom.CreateTextNode("\n  ! "+msg))
		}
	}

	return vdom.CreateElement("form", nil, rows...)
}

func renderFormField(field FormField, value string, focused bool) *vdom.Node {
	label := padLabel(field.Label, formLabelWidth)
	marker := " "
	if focused {
		marker = ">"
	}
	prefix := marker + " " + label + " "

	switch field.Kind {
	case FormFieldPassword:
		input := PasswordInput(TextInputProps{
			Value:     value,
			Focus:     focused,
			CursorPos: len(value),
		})
		return joinPrefixed(prefix, input)
	case FormFieldSelect:
		items := field.Options
		selected := indexOfOption(items, value)
		input := Select(SelectProps{
			Items:    items,
			Selected: selected,
			Focused:  focused,
		})
		return joinPrefixed(prefix, input)
	case FormFieldConfirm:
		var answer *bool
		switch value {
		case "true":
			t := true
			answer = &t
		case "false":
			f := false
			answer = &f
		}
		input := Confirm(ConfirmProps{
			Question: "",
			Default:  field.Default == "true",
			Value:    answer,
		})
		return joinPrefixed(prefix, input)
	case FormFieldMultiSelect:
		items := field.Options
		multi := make([]MultiSelectItem, len(items))
		for i, it := range items {
			multi[i] = MultiSelectItem{Label: it.Label, Value: it.Value}
		}
		input := MultiSelect(MultiSelectProps{
			Items:    multi,
			Selected: FormMultiSelectValues(value),
			Focused:  focused,
		})
		return joinPrefixed(prefix, input)
	case FormFieldTab:
		items := field.Options
		tabs := make([]TabItem, len(items))
		for i, it := range items {
			tabs[i] = TabItem{Label: it.Label}
		}
		active := indexOfOption(items, value)
		input := Tabs(TabsProps{
			Items:   tabs,
			Active:  active,
			Focused: focused,
		})
		return joinPrefixed(prefix, input)
	case FormFieldQuickSearch:
		items := field.Options
		selected := indexOfOption(items, value)
		input := QuickSearch(QuickSearchProps{
			Items:    items,
			Query:    "",
			Selected: selected,
			Focused:  focused,
		})
		return joinPrefixed(prefix, input)
	default:
		input := TextInput(TextInputProps{
			Value:     value,
			Focus:     focused,
			CursorPos: len(value),
		})
		return joinPrefixed(prefix, input)
	}
}

// FormMultiSelectValues splits a canonical FormFieldMultiSelect CSV value
// into its component option Values. Empty input returns an empty slice
// (never nil). Whitespace around commas is trimmed.
func FormMultiSelectValues(csv string) []string {
	if csv == "" {
		return []string{}
	}
	parts := strings.Split(csv, formMultiSelectSeparator)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v == "" {
			continue
		}
		out = append(out, v)
	}
	return out
}

// EncodeFormMultiSelect joins values into the canonical CSV form expected
// by FormFieldMultiSelect. Empty values are dropped.
func EncodeFormMultiSelect(values []string) string {
	cleaned := make([]string, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		cleaned = append(cleaned, v)
	}
	return strings.Join(cleaned, formMultiSelectSeparator)
}

// joinPrefixed glues a string prefix in front of an input node by walking
// the node's text-child output. Keeps Form composable on top of the
// existing TextInput / Select primitives without reaching into renderer
// internals.
func joinPrefixed(prefix string, input *vdom.Node) *vdom.Node {
	var b strings.Builder
	b.WriteString(prefix)
	collectText(input, &b)
	return vdom.CreateTextNode(b.String())
}

func collectText(n *vdom.Node, b *strings.Builder) {
	if n == nil {
		return
	}
	if n.Text != "" {
		b.WriteString(n.Text)
	}
	for _, child := range n.Children {
		collectText(child, b)
	}
}

func padLabel(label string, width int) string {
	if len(label) >= width {
		return label[:width]
	}
	return label + strings.Repeat(" ", width-len(label))
}

func indexOfOption(items []SelectItem, value string) int {
	for i, item := range items {
		if item.Value == value {
			return i
		}
	}
	return 0
}

// pickOptionValue returns preferred when it matches one of the option
// Values; otherwise it returns the first option's Value (or "" when the
// option list is empty). Used by Tab / QuickSearch initialisers where an
// out-of-range Default should fall back to a stable canonical state.
func pickOptionValue(items []SelectItem, preferred string) string {
	for _, item := range items {
		if item.Value == preferred {
			return preferred
		}
	}
	if len(items) == 0 {
		return ""
	}
	return items[0].Value
}

// FormState is the controller half of the Form pattern. It owns the
// per-field values and errors plus the focused-row cursor. Wire its
// movement methods to your input loop and feed Values / Errors /
// FocusedIndex back into Form() on every render.
type FormState struct {
	Fields  []FormField
	Values  map[string]string
	Errors  map[string]string
	Focused int
}

// NewFormState builds a FormState with empty values for every field. The
// initial focus is the first field. Select fields default to their first
// option value when one is available; Confirm and MultiSelect fields use
// FormField.Default when set.
func NewFormState(fields []FormField) *FormState {
	values := make(map[string]string, len(fields))
	for _, f := range fields {
		switch f.Kind {
		case FormFieldSelect:
			if len(f.Options) > 0 {
				values[f.Name] = f.Options[0].Value
			} else {
				values[f.Name] = ""
			}
		case FormFieldConfirm:
			// Default values are normalised to the canonical "true"/"false"
			// form; anything else is treated as unanswered.
			switch f.Default {
			case "true", "false":
				values[f.Name] = f.Default
			default:
				values[f.Name] = ""
			}
		case FormFieldMultiSelect:
			values[f.Name] = EncodeFormMultiSelect(FormMultiSelectValues(f.Default))
		case FormFieldTab:
			// Tab requires a selected option at all times when Options
			// is non-empty: prefer Default when it matches an option,
			// otherwise fall back to the first option's Value.
			values[f.Name] = pickOptionValue(f.Options, f.Default)
		case FormFieldQuickSearch:
			// QuickSearch is single-choice like Select but allows the
			// initial value to be empty (no result selected yet) when
			// Default is unset.
			if f.Default != "" {
				values[f.Name] = pickOptionValue(f.Options, f.Default)
			} else {
				values[f.Name] = ""
			}
		default:
			values[f.Name] = ""
		}
	}
	return &FormState{
		Fields:  fields,
		Values:  values,
		Errors:  map[string]string{},
		Focused: 0,
	}
}

// FocusNext advances the focus cursor by one, wrapping at the end. No-op
// for an empty form.
func (s *FormState) FocusNext() {
	if len(s.Fields) == 0 {
		return
	}
	s.Focused = (s.Focused + 1) % len(s.Fields)
}

// FocusPrev moves the focus cursor back one, wrapping at the start.
// No-op for an empty form.
func (s *FormState) FocusPrev() {
	if len(s.Fields) == 0 {
		return
	}
	s.Focused = (s.Focused - 1 + len(s.Fields)) % len(s.Fields)
}

// SetValue stores value under name. Unknown names are silently ignored so
// callers can pipe arbitrary key strokes without pre-checking.
func (s *FormState) SetValue(name, value string) {
	for _, f := range s.Fields {
		if f.Name == name {
			s.Values[name] = value
			return
		}
	}
}

// Validate runs Required + Validate hooks across every field, populates
// Errors with any failures, and returns true when all fields pass. A
// previous Errors map is reset on each call.
//
// Required semantics by kind:
//
//   - Text/Password/Select: value must be non-empty after TrimSpace.
//   - Confirm: value must be "true" or "false" (i.e. the user must have
//     answered).
//   - MultiSelect: at least one option must be selected (CSV non-empty
//     after trimming whitespace tokens).
//
// Validate hooks for non-text kinds receive the canonical string form:
// "true"/"false" for Confirm, the full CSV for MultiSelect. Use
// FormMultiSelectValues to decode the CSV inside a validator.
func (s *FormState) Validate() bool {
	s.Errors = map[string]string{}
	ok := true
	for _, f := range s.Fields {
		val := s.Values[f.Name]
		if f.Required && !formValueIsPresent(f.Kind, val) {
			s.Errors[f.Name] = fmt.Sprintf("%s is required", f.Label)
			ok = false
			continue
		}
		if f.Validate != nil && formValueIsPresent(f.Kind, val) {
			if err := f.Validate(val); err != nil {
				s.Errors[f.Name] = err.Error()
				ok = false
			}
		}
	}
	return ok
}

// formValueIsPresent returns true when the field carries a meaningful
// value. The rules differ by kind so that Required validation lines up
// with user expectations across non-text inputs.
func formValueIsPresent(kind FormFieldKind, value string) bool {
	switch kind {
	case FormFieldConfirm:
		return value == "true" || value == "false"
	case FormFieldMultiSelect:
		return len(FormMultiSelectValues(value)) > 0
	case FormFieldTab, FormFieldQuickSearch:
		// Tab/QuickSearch carry the Value of an option; treat the empty
		// string as missing the same way Select / Text does.
		return strings.TrimSpace(value) != ""
	default:
		return strings.TrimSpace(value) != ""
	}
}

// OrderedErrors returns the current validation errors as a flat []error
// in field-declaration order. Each error message is prefixed with the
// field's Label (e.g. "Email: must look like name@host.tld") so the
// flat list remains self-describing when piped into ErrorOverviewGroup.
//
// The returned slice is freshly allocated; callers may mutate it freely.
// Returns an empty (non-nil) slice when there are no errors.
func (s *FormState) OrderedErrors() []error {
	out := make([]error, 0, len(s.Errors))
	for _, f := range s.Fields {
		msg, ok := s.Errors[f.Name]
		if !ok || msg == "" {
			continue
		}
		label := f.Label
		if label == "" {
			label = f.Name
		}
		out = append(out, fmt.Errorf("%s: %s", label, msg))
	}
	return out
}

// FieldError returns the current error registered against the named
// field, or nil when the field has no error (or does not exist). The
// returned error carries only the bare validator message — no Label
// prefix — so callers can compose their own surfaces.
func (s *FormState) FieldError(name string) error {
	msg, ok := s.Errors[name]
	if !ok || msg == "" {
		return nil
	}
	return errors.New(msg)
}

// HasErrors reports whether the most recent Validate run left any field
// errors behind. Equivalent to len(s.OrderedErrors()) > 0 but cheaper.
func (s *FormState) HasErrors() bool {
	for _, msg := range s.Errors {
		if msg != "" {
			return true
		}
	}
	return false
}

// ErrorOverviewFromForm renders an ErrorOverviewGroup pre-populated with
// the form's field-level errors as the Validation section. Additional
// runtime errors (panics, downstream call failures) may be supplied via
// runtime — they appear under a separate Runtime sub-section. The
// returned node is safe to embed directly into any layout; it renders a
// "<no errors>" placeholder when both sources are empty.
func ErrorOverviewFromForm(state *FormState, runtime ...error) *vdom.Node {
	var validation []error
	if state != nil {
		validation = state.OrderedErrors()
	}
	return ErrorOverviewGroup(ErrorOverviewGroupProps{
		Validation: validation,
		Runtime:    runtime,
	})
}

// Submit runs Validate and, on success, returns a copy of Values plus
// true. On failure, returns nil + false; the caller should re-render so
// the freshly populated Errors surface in the UI.
func (s *FormState) Submit() (map[string]string, bool) {
	if !s.Validate() {
		return nil, false
	}
	out := make(map[string]string, len(s.Values))
	for k, v := range s.Values {
		out[k] = v
	}
	return out, true
}
