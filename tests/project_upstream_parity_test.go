package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dh-kam/goink.go/internal/renderer"
	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

type projectCaseSpec struct {
	Name             string             `json:"name"`
	Columns          int                `json:"columns"`
	Mode             string             `json:"mode,omitempty"`
	ScreenReader     bool               `json:"screenReader,omitempty"`
	ANSI             bool               `json:"ansi,omitempty"`
	ExpectedError    string             `json:"expectedError,omitempty"`
	ExpectedContains []string           `json:"expectedContains,omitempty"`
	Env              map[string]string  `json:"env,omitempty"`
	Node             projectNodeSpec   `json:"node,omitempty"`
	Frames           []projectNodeSpec `json:"frames,omitempty"`
}

type projectNodeSpec struct {
	Type     string                 `json:"type"`
	Value    string                 `json:"value,omitempty"`
	Props    map[string]interface{} `json:"props,omitempty"`
	Children []projectNodeSpec     `json:"children,omitempty"`
	Preset   string                 `json:"preset,omitempty"`
	Count    int                    `json:"count,omitempty"`
	Items    []string               `json:"items,omitempty"`
	Template *projectNodeSpec      `json:"template,omitempty"`
}

func (spec *projectNodeSpec) UnmarshalJSON(data []byte) error {
	type projectNodeAlias struct {
		Type     string             `json:"type"`
		Value    string             `json:"value,omitempty"`
		Props    json.RawMessage    `json:"props,omitempty"`
		Children []projectNodeSpec `json:"children,omitempty"`
		Preset   string             `json:"preset,omitempty"`
		Count    int                `json:"count,omitempty"`
		Items    []string           `json:"items,omitempty"`
		Template *projectNodeSpec  `json:"template,omitempty"`
	}

	var alias projectNodeAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}

	props, err := decodeProjectProps(alias.Props)
	if err != nil {
		return err
	}

	spec.Type = alias.Type
	spec.Value = alias.Value
	spec.Props = props
	spec.Children = alias.Children
	spec.Preset = alias.Preset
	spec.Count = alias.Count
	spec.Items = alias.Items
	spec.Template = alias.Template
	return nil
}

func decodeProjectProps(raw json.RawMessage) (map[string]interface{}, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}

	delimiter, ok := token.(json.Delim)
	if !ok || delimiter != '{' {
		var props map[string]interface{}
		if err := json.Unmarshal(raw, &props); err != nil {
			return nil, err
		}

		return props, nil
	}

	props := make(map[string]interface{})
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			return nil, err
		}

		key, ok := keyToken.(string)
		if !ok {
			return nil, fmt.Errorf("unexpected props key token %T", keyToken)
		}

		var valueRaw json.RawMessage
		if err := decoder.Decode(&valueRaw); err != nil {
			return nil, err
		}

		if key == "aria-state" {
			ordered, err := decodeProjectOrderedAriaState(valueRaw)
			if err != nil {
				return nil, err
			}

			props[key] = ordered
			continue
		}

		var value interface{}
		if err := json.Unmarshal(valueRaw, &value); err != nil {
			return nil, err
		}

		props[key] = value
	}

	if _, err := decoder.Token(); err != nil {
		return nil, err
	}

	return props, nil
}

func decodeProjectOrderedAriaState(raw json.RawMessage) ([]string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}

	delimiter, ok := token.(json.Delim)
	if !ok || delimiter != '{' {
		return nil, nil
	}

	states := make([]string, 0)
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			return nil, err
		}

		key, ok := keyToken.(string)
		if !ok {
			return nil, fmt.Errorf("unexpected aria-state key token %T", keyToken)
		}

		var enabled bool
		if err := decoder.Decode(&enabled); err != nil {
			return nil, err
		}

		if enabled {
			states = append(states, key)
		}
	}

	if _, err := decoder.Token(); err != nil {
		return nil, err
	}

	return states, nil
}

type projectGolden struct {
	Name         string   `json:"name"`
	Columns      int      `json:"columns"`
	Mode         string   `json:"mode,omitempty"`
	ScreenReader bool     `json:"screenReader,omitempty"`
	ANSI         bool     `json:"ansi,omitempty"`
	Output       string   `json:"output,omitempty"`
	Error        string   `json:"error,omitempty"`
	Contains     []string `json:"contains,omitempty"`
}

func TestProjectUpstreamGoldenParity(t *testing.T) {
	cases := loadProjectCases(t)
	goldens := loadProjectGoldens(t)

	if len(cases) != len(goldens) {
		t.Fatalf("case count mismatch: %d cases vs %d goldens", len(cases), len(goldens))
	}

	goldenByName := make(map[string]projectGolden, len(goldens))
	for _, golden := range goldens {
		goldenByName[golden.Name] = golden
	}

	for _, testCase := range cases {
		testCase := testCase
		t.Run(testCase.Name, func(t *testing.T) {
			expected, ok := goldenByName[testCase.Name]
			if !ok {
				t.Fatalf("missing golden for case %q", testCase.Name)
			}

			for key, value := range testCase.Env {
				t.Setenv(key, value)
			}

			switch testCase.Mode {
			case "", "static":
				node, err := buildProjectNode(testCase.Node)
				if err != nil {
					t.Fatalf("build node: %v", err)
				}

				columns := testCase.Columns
				if columns <= 0 {
					columns = 100
				}

				actual := renderProjectParityOutput(node, columns, testCase.ScreenReader, testCase.ANSI)
				if actual != expected.Output {
					t.Fatalf("output mismatch\nexpected: %q\nactual:   %q", expected.Output, actual)
				}
			case "error":
				node, err := buildProjectNode(testCase.Node)
				if err != nil {
					t.Fatalf("build node: %v", err)
				}

				actual, ok := renderProjectParityError(node)
				if !ok {
					t.Fatalf("expected render panic for %q", testCase.Name)
				}

				if actual != expected.Error {
					t.Fatalf("error mismatch\nexpected: %q\nactual:   %q", expected.Error, actual)
				}
			case "managed-frames":
				actual, err := renderProjectManagedFrames(testCase)
				if err != nil {
					t.Fatalf("render managed frames: %v", err)
				}

				assertProjectContainsInOrder(t, actual, expected.Contains)
			default:
				t.Fatalf("unknown case mode %q", testCase.Mode)
			}
		})
	}
}

func TestProjectUpstreamCoverageCounts(t *testing.T) {
	targets := map[string]int{
		"gemini-cli":    8,
		"neovate-code":  11,
		"shopify-cli":   1,
		"tweakcc":       1,
		"nanocoder":     1,
	}

	counts := make(map[string]int, len(targets))
	for _, testCase := range loadProjectCases(t) {
		family, _, _ := strings.Cut(testCase.Name, "/")
		counts[family]++
	}

	for family, minimum := range targets {
		if counts[family] < minimum {
			t.Fatalf("expected at least %d cases for %s, got %d", minimum, family, counts[family])
		}
	}
}

func loadProjectCases(t *testing.T) []projectCaseSpec {
	t.Helper()

	path := filepath.Join("project_upstream", "cases.json")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read cases: %v", err)
	}

	var cases []projectCaseSpec
	if err := json.Unmarshal(content, &cases); err != nil {
		t.Fatalf("unmarshal cases: %v", err)
	}

	return cases
}

func loadProjectGoldens(t *testing.T) []projectGolden {
	t.Helper()

	path := filepath.Join("project_upstream", "goldens.json")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read goldens: %v", err)
	}

	var goldens []projectGolden
	if err := json.Unmarshal(content, &goldens); err != nil {
		t.Fatalf("unmarshal goldens: %v", err)
	}

	return goldens
}

func renderProjectParityOutput(node *vdom.Node, columns int, screenReader bool, ansi bool) string {
	if screenReader {
		sections := renderer.RenderScreenReaderSections(node)
		return sections.StaticOutput + sections.Output
	}

	sections := renderer.RenderWithLayoutSectionsMode(node, columns, 256, ansi)
	return sections.StaticOutput + sections.Output
}

func renderProjectParityError(node *vdom.Node) (string, bool) {
	var recovered interface{}

	func() {
		defer func() {
			recovered = recover()
		}()

		app := ink.NewApp(func() *vdom.Node {
			return node
		})
		_ = app.RenderOnce()
	}()

	if recovered == nil {
		return "", false
	}

	return fmt.Sprint(recovered), true
}

type projectParityRuntimeWriter struct {
	writes  []string
	columns int
}

func (writer *projectParityRuntimeWriter) Write(data []byte) (int, error) {
	writer.writes = append(writer.writes, string(data))
	return len(data), nil
}

func (writer *projectParityRuntimeWriter) Columns() int {
	return writer.columns
}

func (writer *projectParityRuntimeWriter) joined() string {
	return strings.Join(writer.writes, "")
}

func renderProjectManagedFrames(spec projectCaseSpec) (string, error) {
	stdout := &projectParityRuntimeWriter{columns: spec.Columns}
	var instance *ink.Instance

	for index, frame := range spec.Frames {
		node, err := buildProjectNode(frame)
		if err != nil {
			return "", fmt.Errorf("build frame %d: %w", index, err)
		}

		frameNode := node
		next, err := ink.RenderWithOptions(func() *vdom.Node {
			return frameNode
		}, ink.RenderOptions{
			AppOptions: ink.AppOptions{
				Width:               spec.Columns,
				Stdout:              stdout,
				ScreenReaderEnabled: spec.ScreenReader,
			},
		})
		if err != nil {
			return "", fmt.Errorf("frame %d: %w", index, err)
		}

		if instance == nil {
			instance = next
			continue
		}

		if next != instance {
			return "", fmt.Errorf("frame %d reused a different instance", index)
		}
	}

	if instance == nil {
		return "", nil
	}

	if err := instance.Unmount(); err != nil {
		return "", fmt.Errorf("unmount managed frames: %w", err)
	}

	return stdout.joined(), nil
}

func assertProjectContainsInOrder(t *testing.T, actual string, expected []string) {
	t.Helper()

	offset := 0
	for _, needle := range expected {
		index := strings.Index(actual[offset:], needle)
		if index < 0 {
			t.Fatalf("expected output to contain %q after offset %d, got %q", needle, offset, actual)
		}

		offset += index + len(needle)
	}
}

func buildProjectNode(spec projectNodeSpec) (*vdom.Node, error) {
	switch spec.Type {
	case "raw":
		return vdom.CreateTextNode(spec.Value), nil
	case "empty":
		return nil, nil
	case "newline":
		count := spec.Count
		if count <= 0 {
			count = 1
		}

		return components.Newline(count), nil
	case "spacer":
		return components.Spacer(), nil
	case "text":
		children, err := buildProjectChildren(spec.Children)
		if err != nil {
			return nil, err
		}

		args := make([]any, 0, len(children)+1)
		if props := toProjectProps(spec.Props); props != nil {
			args = append(args, props)
		}
		for _, child := range children {
			args = append(args, child)
		}

		return components.Text(args...), nil
	case "box":
		children, err := buildProjectChildren(spec.Children)
		if err != nil {
			return nil, err
		}

		return components.Box(toProjectProps(spec.Props), children...), nil
	case "static":
		template := spec.Template
		if template == nil {
			template = &projectNodeSpec{
				Type: "text",
				Children: []projectNodeSpec{
					{Type: "raw", Value: "{{item}}"},
				},
			}
		}

		return components.StaticItems(spec.Items, func(item string, index int) *vdom.Node {
			resolvedSpec := instantiateProjectTemplate(*template, item, index)
			node, err := buildProjectNode(resolvedSpec)
			if err != nil {
				panic(fmt.Sprintf("build static child: %v", err))
			}

			return node
		}, toProjectProps(spec.Props)), nil
	case "transform":
		children, err := buildProjectChildren(spec.Children)
		if err != nil {
			return nil, err
		}

		props := toProjectProps(spec.Props)
		if props == nil {
			props = vdom.Props{}
		}
		props["transform"] = projectTransformPreset(spec.Preset)

		return vdom.CreateElement("transform", props, children...), nil
	default:
		return nil, fmt.Errorf("unknown node type %q", spec.Type)
	}
}

func buildProjectChildren(children []projectNodeSpec) ([]*vdom.Node, error) {
	nodes := make([]*vdom.Node, 0, len(children))
	for _, child := range children {
		node, err := buildProjectNode(child)
		if err != nil {
			return nil, err
		}

		if node == nil {
			continue
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

func instantiateProjectTemplate(spec projectNodeSpec, item string, index int) projectNodeSpec {
	resolved := projectNodeSpec{
		Type:   replaceProjectTemplateString(spec.Type, item, index),
		Value:  replaceProjectTemplateString(spec.Value, item, index),
		Preset: replaceProjectTemplateString(spec.Preset, item, index),
		Count:  spec.Count,
	}

	if len(spec.Props) > 0 {
		resolved.Props = make(map[string]interface{}, len(spec.Props))
		for key, value := range spec.Props {
			resolved.Props[key] = replaceProjectTemplateValue(value, item, index)
		}
	}

	if len(spec.Children) > 0 {
		resolved.Children = make([]projectNodeSpec, 0, len(spec.Children))
		for _, child := range spec.Children {
			resolved.Children = append(resolved.Children, instantiateProjectTemplate(child, item, index))
		}
	}

	if spec.Template != nil {
		template := instantiateProjectTemplate(*spec.Template, item, index)
		resolved.Template = &template
	}

	if len(spec.Items) > 0 {
		resolved.Items = make([]string, len(spec.Items))
		for itemIndex, childItem := range spec.Items {
			resolved.Items[itemIndex] = replaceProjectTemplateString(childItem, item, index)
		}
	}

	return resolved
}

func replaceProjectTemplateValue(value interface{}, item string, index int) interface{} {
	switch typed := value.(type) {
	case string:
		return replaceProjectTemplateString(typed, item, index)
	case map[string]interface{}:
		resolved := make(map[string]interface{}, len(typed))
		for key, nestedValue := range typed {
			resolved[key] = replaceProjectTemplateValue(nestedValue, item, index)
		}

		return resolved
	case []interface{}:
		resolved := make([]interface{}, 0, len(typed))
		for _, nestedValue := range typed {
			resolved = append(resolved, replaceProjectTemplateValue(nestedValue, item, index))
		}

		return resolved
	default:
		return value
	}
}

func replaceProjectTemplateString(value string, item string, index int) string {
	value = strings.ReplaceAll(value, "{{item}}", item)
	value = strings.ReplaceAll(value, "{{index}}", fmt.Sprintf("%d", index))
	return value
}

func toProjectProps(props map[string]interface{}) vdom.Props {
	if len(props) == 0 {
		return nil
	}

	result := make(vdom.Props, len(props))
	for key, value := range props {
		result[key] = value
	}

	return result
}

func projectTransformPreset(name string) func(string, int) string {
	switch name {
	case "identity":
		return func(children string, _ int) string {
			return children
		}
	case "bracket_index":
		return func(children string, index int) string {
			return fmt.Sprintf("[%d: %s]", index, children)
		}
	case "brace_index":
		return func(children string, index int) string {
			return fmt.Sprintf("{%d: %s}", index, children)
		}
	case "angle":
		return func(children string, _ int) string {
			return fmt.Sprintf("<%s>", children)
		}
	case "reverse":
		return func(children string, _ int) string {
			runes := []rune(children)
			for left, right := 0, len(runes)-1; left < right; left, right = left+1, right-1 {
				runes[left], runes[right] = runes[right], runes[left]
			}

			return string(runes)
		}
	case "upper":
		return func(children string, _ int) string {
			return strings.ToUpper(children)
		}
	default:
		panic(fmt.Sprintf("unknown transform preset %q", name))
	}
}
