package ink

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dh-kam/ink-go/pkg/vdom"
)

func TestRenderAllFramesIfCIEnvironmentVariableEqualsFalse(t *testing.T) {
	t.Setenv("CI", "false")

	stdout := &ttyRecordingWriter{}

	var instance *Instance
	for frame := 0; frame <= 5; frame++ {
		frame := frame

		next, err := RenderWithOptions(func() *vdom.Node {
			return vdom.CreateTextNode(fmt.Sprintf("Counter: %d", frame))
		}, RenderOptions{
			AppOptions: AppOptions{Stdout: stdout},
		})
		if err != nil {
			t.Fatalf("render %d failed: %v", frame, err)
		}

		if instance == nil {
			instance = next
			continue
		}

		if next != instance {
			t.Fatal("expected CI=false managed renders to reuse the same instance")
		}
	}

	if instance == nil {
		t.Fatal("expected CI=false managed render to create an instance")
	}
	defer instance.Unmount()

	output := stdout.joined()
	for frame := 0; frame <= 5; frame++ {
		expected := fmt.Sprintf("Counter: %d", frame)
		if !strings.Contains(output, expected) {
			t.Fatalf("expected CI=false output to include %q, got %#v", expected, stdout.writes)
		}
	}
}
