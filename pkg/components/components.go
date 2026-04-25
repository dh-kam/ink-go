package components

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/dh-kam/goink.go/pkg/vdom"
)

// TransformFunc transforms rendered text output.
type TransformFunc = func(children string, index int) string

// BorderStyle defines the style of border
type BorderStyle string

const (
	BorderSingle  BorderStyle = "single"
	BorderDouble  BorderStyle = "double"
	BorderRounded BorderStyle = "rounded"
	BorderBold    BorderStyle = "bold"
)

// BorderProps contains border properties
type BorderProps struct {
	Style  BorderStyle
	Top    bool
	Bottom bool
	Left   bool
	Right  bool
	Label  string
	Title  string // Alias for label
}

const publicComponentMarkerKey = "__inkPublicComponent"

func markPublicComponentProps(props vdom.Props) vdom.Props {
	if props == nil {
		props = make(vdom.Props)
	} else {
		props = cloneProps(props)
	}

	props[publicComponentMarkerKey] = true
	return props
}

// Box creates a box element (container)
// Props can include layout properties (to be implemented in Phase 2)
func Box(props vdom.Props, children ...*vdom.Node) *vdom.Node {
	return vdom.CreateElement("box", markPublicComponentProps(props), children...)
}

// Text creates a text element.
//
// Supported arguments:
//   - string values, which are converted to text nodes
//   - *vdom.Node children, including nested text nodes
//   - vdom.Props, in any position
//
// This keeps simple usage like Text("hello") while also allowing richer
// upstream-like nested text trees such as Text(props, child1, child2).
func Text(args ...any) *vdom.Node {
	props, children := collectNodeArgs("Text", args...)
	return vdom.CreateElement("text", markPublicComponentProps(props), children...)
}

// Newline creates a newline text node.
// When count is provided and greater than zero, that many newline characters are emitted.
func Newline(count ...int) *vdom.Node {
	lines := 1
	if len(count) > 0 && count[0] > 0 {
		lines = count[0]
	}

	return vdom.CreateTextNode(strings.Repeat("\n", lines))
}

// Space creates a space text node
func Space() *vdom.Node {
	return vdom.CreateTextNode(" ")
}

// Spacer creates a flexible box that expands along the main axis.
func Spacer() *vdom.Node {
	return Box(vdom.Props{"flexGrow": 1.0})
}

// Border creates a bordered box component
func Border(borderProps BorderProps, props vdom.Props, children ...*vdom.Node) *vdom.Node {
	if props == nil {
		props = make(vdom.Props)
	}

	// Merge border props into element props
	props["borderStyle"] = string(borderProps.Style)
	props["borderTop"] = borderProps.Top
	props["borderBottom"] = borderProps.Bottom
	props["borderLeft"] = borderProps.Left
	props["borderRight"] = borderProps.Right

	// Handle label/title
	label := borderProps.Label
	if label == "" {
		label = borderProps.Title
	}
	if label != "" {
		props["borderLabel"] = label
	}

	return vdom.CreateElement("border", props, children...)
}

// Static creates a static component.
//
// Supported forms:
//   - Static(props, children...)
//   - Static(items, renderFn, props)
//
// The render function may accept (item) or (item, index) and may return either
// *vdom.Node, string, or nil.
func Static(args ...any) *vdom.Node {
	if len(args) >= 2 && isStaticItemsInvocation(args[0], args[1]) {
		props := collectTrailingProps("Static", args[2:]...)
		props = withStaticItemCount(props, staticItemsCount(args[0]))
		children := renderStaticChildren(args[0], args[1])
		return createStaticElement(props, children...)
	}

	props, children := collectNodeArgs("Static", args...)
	return createStaticElement(props, children...)
}

// StaticItems is a typed wrapper around Static's items/render form.
func StaticItems[T any](items []T, render func(item T, index int) *vdom.Node, props ...vdom.Props) *vdom.Node {
	children := make([]*vdom.Node, 0, len(items))
	for index, item := range items {
		node := render(item, index)
		if node != nil {
			children = append(children, node)
		}
	}

	var mergedProps vdom.Props
	if len(props) > 0 {
		mergedProps = cloneProps(props[0])
	}

	mergedProps = withStaticItemCount(mergedProps, len(items))
	return createStaticElement(mergedProps, children...)
}

// StaticText creates a static text component
func StaticText(content string) *vdom.Node {
	props := vdom.Props{"static": true}
	textNode := vdom.CreateTextNode(content)
	return vdom.CreateElement("text", props, textNode)
}

// Transform creates a text-like node that transforms its rendered output.
func Transform(transform TransformFunc, children ...*vdom.Node) *vdom.Node {
	props := vdom.Props{"transform": transform}
	return vdom.CreateElement("transform", props, children...)
}

func createStaticElement(props vdom.Props, children ...*vdom.Node) *vdom.Node {
	if props == nil {
		props = make(vdom.Props)
	}

	props["static"] = true
	return vdom.CreateElement("static", props, children...)
}

func withStaticItemCount(props vdom.Props, count int) vdom.Props {
	if props == nil {
		props = make(vdom.Props)
	}

	props["__staticItemsCount"] = count
	return props
}

func collectNodeArgs(componentName string, args ...any) (vdom.Props, []*vdom.Node) {
	var props vdom.Props
	children := make([]*vdom.Node, 0, len(args))

	for _, arg := range args {
		switch typed := arg.(type) {
		case nil:
			continue
		case vdom.Props:
			if props == nil {
				props = make(vdom.Props)
			}

			for key, value := range typed {
				props[key] = value
			}
		case *vdom.Node:
			children = append(children, typed)
		case []*vdom.Node:
			children = append(children, typed...)
		case string:
			children = append(children, vdom.CreateTextNode(typed))
		case []string:
			for _, value := range typed {
				children = append(children, vdom.CreateTextNode(value))
			}
		default:
			panic(fmt.Sprintf("%s does not support argument type %T", componentName, arg))
		}
	}

	return props, children
}

func collectTrailingProps(componentName string, args ...any) vdom.Props {
	var props vdom.Props

	for _, arg := range args {
		if arg == nil {
			continue
		}

		typed, ok := arg.(vdom.Props)
		if !ok {
			panic(fmt.Sprintf("%s expected trailing props, got %T", componentName, arg))
		}

		if props == nil {
			props = make(vdom.Props)
		}

		for key, value := range typed {
			props[key] = value
		}
	}

	return props
}

func cloneProps(props vdom.Props) vdom.Props {
	if props == nil {
		return nil
	}

	cloned := make(vdom.Props, len(props))
	for key, value := range props {
		cloned[key] = value
	}

	return cloned
}

func isStaticItemsInvocation(items any, render any) bool {
	itemsValue := reflect.ValueOf(items)
	if !itemsValue.IsValid() {
		return false
	}

	kind := itemsValue.Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		return false
	}

	renderType := reflect.TypeOf(render)
	return renderType != nil && renderType.Kind() == reflect.Func
}

func renderStaticChildren(items any, render any) []*vdom.Node {
	itemsValue := reflect.ValueOf(items)
	renderValue := reflect.ValueOf(render)
	renderType := renderValue.Type()

	if renderType.NumIn() != 1 && renderType.NumIn() != 2 {
		panic("Static render function must accept (item) or (item, index)")
	}

	if renderType.NumOut() != 1 {
		panic("Static render function must return one value")
	}

	children := make([]*vdom.Node, 0, itemsValue.Len())
	for index := 0; index < itemsValue.Len(); index++ {
		callArgs := make([]reflect.Value, 0, renderType.NumIn())
		itemValue := itemsValue.Index(index)
		callArgs = append(callArgs, adaptValue(itemValue, renderType.In(0), "Static item"))

		if renderType.NumIn() == 2 {
			callArgs = append(callArgs, adaptValue(reflect.ValueOf(index), renderType.In(1), "Static index"))
		}

		result := renderValue.Call(callArgs)[0].Interface()
		switch typed := result.(type) {
		case nil:
			continue
		case *vdom.Node:
			if typed != nil {
				children = append(children, typed)
			}
		case string:
			children = append(children, vdom.CreateTextNode(typed))
		default:
			panic(fmt.Sprintf("Static render function returned unsupported type %T", result))
		}
	}

	return children
}

func staticItemsCount(items any) int {
	value := reflect.ValueOf(items)
	if value.Kind() != reflect.Slice && value.Kind() != reflect.Array {
		return 0
	}

	return value.Len()
}

func adaptValue(value reflect.Value, target reflect.Type, label string) reflect.Value {
	if !value.IsValid() {
		panic(fmt.Sprintf("%s is invalid", label))
	}

	if value.Type().AssignableTo(target) {
		return value
	}

	if value.Type().ConvertibleTo(target) {
		return value.Convert(target)
	}

	if target.Kind() == reflect.Interface && value.Type().Implements(target) {
		return value
	}

	panic(fmt.Sprintf("%s of type %s is not assignable to %s", label, value.Type(), target))
}
