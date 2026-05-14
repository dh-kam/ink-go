// Package devtools provides read-only introspection helpers for goink runtime
// internals. The helpers in this file expose a non-mutating snapshot of a
// hooks.Context's slot slices so that developers can debug, profile or render
// devtools panels without having to modify the hooks package itself.
//
// Because Context's slot slices (states, refs, effects, ...) are unexported,
// the snapshot uses the reflect + unsafe escape hatch to read them. No values
// are written back, and effect / callback function pointers are deliberately
// not exposed - calling them out-of-band would violate the hooks lifecycle.
package devtools

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"github.com/dh-kam/goink.go/pkg/hooks"
)

// HookInfo describes a single hook slot inside a hooks.Context.
type HookInfo struct {
	// Kind is one of: "state", "ref", "effect", "memo", "callback",
	// "input", "focus".
	Kind string
	// Index is the slot position inside the slice for this kind.
	Index int
	// Value is the stored value for "state" / "ref" / "memo" hooks.
	// For "effect" / "callback" / "input" / "focus" it is left nil because
	// invoking or even materialising those callbacks out of the render
	// pipeline is unsafe.
	Value interface{}
}

// ContextSnapshot is a point-in-time, read-only view of a hooks.Context.
// All slices are independent copies; mutating them does not affect the
// underlying Context.
type ContextSnapshot struct {
	States    []HookInfo
	Refs      []HookInfo
	Effects   []HookInfo // count only, callback intentionally hidden
	Memos     []HookInfo
	Callbacks []HookInfo
	Inputs    int
	Foci      int
}

// SnapshotContext returns a ContextSnapshot for ctx without mutating it.
//
// SnapshotContext is safe to call with a nil pointer; it returns a zero
// ContextSnapshot in that case. It uses reflect together with unsafe to read
// the unexported slot slices on Context, so it must be kept in lock-step with
// the field names declared in pkg/hooks/hooks.go.
func SnapshotContext(ctx *hooks.Context) ContextSnapshot {
	if ctx == nil {
		return ContextSnapshot{}
	}

	v := reflect.ValueOf(ctx).Elem()

	snap := ContextSnapshot{}

	if states, ok := readSliceField(v, "states"); ok {
		snap.States = make([]HookInfo, states.Len())
		for i := 0; i < states.Len(); i++ {
			snap.States[i] = HookInfo{
				Kind:  "state",
				Index: i,
				Value: states.Index(i).Interface(),
			}
		}
	}

	if refs, ok := readSliceField(v, "refs"); ok {
		snap.Refs = make([]HookInfo, refs.Len())
		for i := 0; i < refs.Len(); i++ {
			snap.Refs[i] = HookInfo{
				Kind:  "ref",
				Index: i,
				Value: refValueAt(refs.Index(i)),
			}
		}
	}

	if effects, ok := readSliceField(v, "effects"); ok {
		snap.Effects = make([]HookInfo, effects.Len())
		for i := 0; i < effects.Len(); i++ {
			snap.Effects[i] = HookInfo{
				Kind:  "effect",
				Index: i,
				Value: nil,
			}
		}
	}

	if memos, ok := readSliceField(v, "memos"); ok {
		snap.Memos = make([]HookInfo, memos.Len())
		for i := 0; i < memos.Len(); i++ {
			snap.Memos[i] = HookInfo{
				Kind:  "memo",
				Index: i,
				Value: memoValueAt(memos.Index(i)),
			}
		}
	}

	if callbacks, ok := readSliceField(v, "callbacks"); ok {
		snap.Callbacks = make([]HookInfo, callbacks.Len())
		for i := 0; i < callbacks.Len(); i++ {
			snap.Callbacks[i] = HookInfo{
				Kind:  "callback",
				Index: i,
				Value: nil,
			}
		}
	}

	if inputs, ok := readSliceField(v, "inputs"); ok {
		snap.Inputs = inputs.Len()
	}

	if foci, ok := readSliceField(v, "foci"); ok {
		snap.Foci = foci.Len()
	}

	return snap
}

// readSliceField returns a reflect.Value that aliases the unexported slice
// field name on the struct value v. It uses reflect.NewAt + unsafe.Pointer to
// bypass the usual "obtained from unexported field" restriction so that
// .Len() / .Index() can be called on the result. The returned Value shares
// memory with the original field; the caller must treat it as read-only.
func readSliceField(v reflect.Value, name string) (reflect.Value, bool) {
	field := v.FieldByName(name)
	if !field.IsValid() {
		return reflect.Value{}, false
	}
	if field.Kind() != reflect.Slice {
		return reflect.Value{}, false
	}
	// Re-create an addressable, exported handle on the same memory.
	fp := unsafe.Pointer(field.UnsafeAddr())
	return reflect.NewAt(field.Type(), fp).Elem(), true
}

// refValueAt extracts the stored value from a refHook slice element. The
// element is expected to be a struct with a *Ref field whose underlying Ref
// has an unexported "value" field. If the layout does not match, nil is
// returned.
func refValueAt(elem reflect.Value) interface{} {
	// elem is hooks.refHook (struct). Find the *Ref field.
	for i := 0; i < elem.NumField(); i++ {
		f := elem.Field(i)
		if f.Kind() != reflect.Ptr {
			continue
		}
		if f.IsNil() {
			continue
		}
		target := f.Elem()
		if target.Kind() != reflect.Struct {
			continue
		}
		valueField := target.FieldByName("value")
		if !valueField.IsValid() {
			continue
		}
		fp := unsafe.Pointer(valueField.UnsafeAddr())
		return reflect.NewAt(valueField.Type(), fp).Elem().Interface()
	}
	return nil
}

// memoValueAt pulls the cached value out of a memoHook struct element.
func memoValueAt(elem reflect.Value) interface{} {
	valueField := elem.FieldByName("value")
	if !valueField.IsValid() {
		return nil
	}
	fp := unsafe.Pointer(valueField.UnsafeAddr())
	return reflect.NewAt(valueField.Type(), fp).Elem().Interface()
}

// Format renders the snapshot as a multi-line, human-readable summary. The
// exact layout is not part of the public API and may change; it is intended
// for logs and devtools panels, not machine consumption.
func (s ContextSnapshot) Format() string {
	var b strings.Builder

	fmt.Fprintf(&b, "States    (%d):", len(s.States))
	if len(s.States) == 0 {
		b.WriteString(" -")
	} else {
		parts := make([]string, len(s.States))
		for i, h := range s.States {
			parts[i] = fmt.Sprintf("[%d] %s", h.Index, formatValue(h.Value))
		}
		b.WriteString(" ")
		b.WriteString(strings.Join(parts, ", "))
	}
	b.WriteString("\n")

	fmt.Fprintf(&b, "Refs      (%d):", len(s.Refs))
	if len(s.Refs) == 0 {
		b.WriteString(" -")
	} else {
		parts := make([]string, len(s.Refs))
		for i, h := range s.Refs {
			parts[i] = fmt.Sprintf("[%d] %s", h.Index, formatValue(h.Value))
		}
		b.WriteString(" ")
		b.WriteString(strings.Join(parts, ", "))
	}
	b.WriteString("\n")

	fmt.Fprintf(&b, "Effects   (%d): (callbacks not displayed)\n", len(s.Effects))

	fmt.Fprintf(&b, "Memos     (%d):", len(s.Memos))
	if len(s.Memos) == 0 {
		b.WriteString(" -")
	} else {
		parts := make([]string, len(s.Memos))
		for i, h := range s.Memos {
			parts[i] = fmt.Sprintf("[%d] cached=%s", h.Index, formatValue(h.Value))
		}
		b.WriteString(" ")
		b.WriteString(strings.Join(parts, ", "))
	}
	b.WriteString("\n")

	fmt.Fprintf(&b, "Callbacks (%d): (callbacks not displayed)\n", len(s.Callbacks))
	fmt.Fprintf(&b, "Inputs    : %d\n", s.Inputs)
	fmt.Fprintf(&b, "Foci      : %d\n", s.Foci)

	return b.String()
}

// formatValue renders a single hook value with its concrete type so the
// output of Format is unambiguous.
func formatValue(v interface{}) string {
	if v == nil {
		return "nil"
	}
	switch typed := v.(type) {
	case string:
		return fmt.Sprintf("string=%q", typed)
	default:
		return fmt.Sprintf("%T=%v", v, v)
	}
}
