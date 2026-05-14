// Package devtools provides debugging utilities for goink applications.
//
// The inspect helpers in this file convert a live vdom tree into a
// serializable, side-effect free Snapshot that can be marshalled to JSON
// or rendered as a human readable indented tree.
package devtools

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/dh-kam/goink.go/pkg/vdom"
)

// LayoutInfo mirrors the layout metadata stored on a vdom.Node.
type LayoutInfo struct {
	Left   int `json:"left"`
	Top    int `json:"top"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Snapshot is a serializable view of a vdom tree.
//
// It is fully decoupled from the live vdom — modifying a Snapshot has no
// effect on the source tree, and Inspect makes deep copies of mutable data
// such as Props and Children.
type Snapshot struct {
	Type     string                 `json:"type"`
	Text     string                 `json:"text,omitempty"`
	Key      string                 `json:"key,omitempty"`
	Props    map[string]interface{} `json:"props,omitempty"`
	Children []*Snapshot            `json:"children,omitempty"`
	Layout   *LayoutInfo            `json:"layout,omitempty"`
}

// Inspect converts a vdom tree to a Snapshot tree (deep, side-effect free).
//
// A nil input yields a nil Snapshot. Text nodes are tagged with the literal
// type "text"; element nodes use their ElementType (or "element" if blank).
// Layout is only attached when at least one of the metrics is non-zero so
// JSON output stays compact for unlaid-out trees.
func Inspect(node *vdom.Node) *Snapshot {
	if node == nil {
		return nil
	}

	if node.Type == vdom.TextNode {
		snap := &Snapshot{
			Type: "text",
			Text: node.Text,
			Key:  node.Key,
		}
		if layout := layoutFrom(node); layout != nil {
			snap.Layout = layout
		}
		return snap
	}

	typeName := node.ElementType
	if typeName == "" {
		typeName = "element"
	}

	snap := &Snapshot{
		Type: typeName,
		Key:  node.Key,
	}

	if len(node.Props) > 0 {
		snap.Props = make(map[string]interface{}, len(node.Props))
		for k, v := range node.Props {
			snap.Props[k] = v
		}
	}

	if len(node.Children) > 0 {
		snap.Children = make([]*Snapshot, 0, len(node.Children))
		for _, child := range node.Children {
			if child == nil {
				continue
			}
			snap.Children = append(snap.Children, Inspect(child))
		}
	}

	if layout := layoutFrom(node); layout != nil {
		snap.Layout = layout
	}

	return snap
}

func layoutFrom(node *vdom.Node) *LayoutInfo {
	l := node.Layout
	if l.Left == 0 && l.Top == 0 && l.Width == 0 && l.Height == 0 {
		return nil
	}
	return &LayoutInfo{
		Left:   l.Left,
		Top:    l.Top,
		Width:  l.Width,
		Height: l.Height,
	}
}

// JSON marshals the snapshot to indented JSON.
//
// Marshalling a nil snapshot yields the literal "null" so callers can rely
// on the result always being a valid JSON document.
func (s *Snapshot) JSON() string {
	if s == nil {
		return "null"
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		// json.Marshal on this struct can only fail for non-serializable
		// values inside Props (e.g. channels). Surface the error inline so
		// debugging output remains useful instead of silently dropping it.
		return fmt.Sprintf("{\"error\": %q}", err.Error())
	}
	return string(data)
}

// Tree formats the snapshot as a human-readable indented tree.
//
// Element nodes render as <type prop=value ...> wrappers around their
// children, while text nodes appear as <text>"..."</text>. Props are sorted
// alphabetically for deterministic output.
func (s *Snapshot) Tree() string {
	if s == nil {
		return ""
	}
	var builder strings.Builder
	s.writeTree(&builder, 0)
	return builder.String()
}

func (s *Snapshot) writeTree(builder *strings.Builder, depth int) {
	if s == nil {
		return
	}

	indent := strings.Repeat("  ", depth)

	if s.Type == "text" {
		builder.WriteString(indent)
		builder.WriteString("<text")
		if s.Key != "" {
			fmt.Fprintf(builder, " key=%q", s.Key)
		}
		fmt.Fprintf(builder, ">%q</text>\n", s.Text)
		return
	}

	builder.WriteString(indent)
	builder.WriteByte('<')
	builder.WriteString(s.Type)
	if s.Key != "" {
		fmt.Fprintf(builder, " key=%q", s.Key)
	}
	for _, key := range sortedKeys(s.Props) {
		fmt.Fprintf(builder, " %s=%v", key, s.Props[key])
	}

	if len(s.Children) == 0 {
		builder.WriteString("/>\n")
		return
	}

	builder.WriteString(">\n")
	for _, child := range s.Children {
		child.writeTree(builder, depth+1)
	}
	builder.WriteString(indent)
	builder.WriteString("</")
	builder.WriteString(s.Type)
	builder.WriteString(">\n")
}

func sortedKeys(m map[string]interface{}) []string {
	if len(m) == 0 {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// PrintTree writes the Tree() rendering of node to w.
//
// It is a convenience wrapper that performs the Inspect + Tree pipeline in a
// single call, returning any write error from the underlying writer.
func PrintTree(w io.Writer, node *vdom.Node) error {
	if w == nil {
		return fmt.Errorf("devtools: PrintTree requires a non-nil writer")
	}
	snap := Inspect(node)
	_, err := io.WriteString(w, snap.Tree())
	return err
}
