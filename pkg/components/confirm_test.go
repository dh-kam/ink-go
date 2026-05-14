package components_test

import (
	"strings"
	"testing"

	"github.com/dh-kam/ink-go/pkg/components"
)

func renderConfirm(t *testing.T, props components.ConfirmProps) string {
	t.Helper()
	node := components.Confirm(props)
	if node == nil || len(node.Children) == 0 {
		t.Fatalf("Confirm produced no children")
	}
	return node.Children[0].Text
}

func TestConfirmRenderDefaultYes(t *testing.T) {
	out := renderConfirm(t, components.ConfirmProps{
		Question: "Are you sure?",
		Default:  true,
	})
	if !strings.Contains(out, "Are you sure?") {
		t.Fatalf("output missing question: %q", out)
	}
	if !strings.Contains(out, "[Y/n]") {
		t.Fatalf("output missing [Y/n]: %q", out)
	}
}

func TestConfirmRenderDefaultNo(t *testing.T) {
	out := renderConfirm(t, components.ConfirmProps{
		Question: "Delete?",
		Default:  false,
	})
	if !strings.Contains(out, "Delete?") {
		t.Fatalf("output missing question: %q", out)
	}
	if !strings.Contains(out, "[y/N]") {
		t.Fatalf("output missing [y/N]: %q", out)
	}
}

func TestConfirmRenderEmptyQuestion(t *testing.T) {
	out := renderConfirm(t, components.ConfirmProps{Default: true})
	if !strings.Contains(out, "[Y/n]") {
		t.Fatalf("expected hint, got %q", out)
	}
}

func TestConfirmRenderShowsAnswerYes(t *testing.T) {
	yes := true
	out := renderConfirm(t, components.ConfirmProps{
		Question: "Q?",
		Default:  true,
		Value:    &yes,
	})
	if !strings.Contains(out, "y") {
		t.Fatalf("expected 'y' echoed, got %q", out)
	}
}

func TestConfirmRenderShowsAnswerNo(t *testing.T) {
	no := false
	out := renderConfirm(t, components.ConfirmProps{
		Question: "Q?",
		Default:  false,
		Value:    &no,
	})
	if !strings.Contains(out, "n") {
		t.Fatalf("expected 'n' echoed, got %q", out)
	}
}

func TestConfirmStateHandleKeyYesLowercase(t *testing.T) {
	s := components.NewConfirmState("Q?", false)
	resolved, v := s.HandleKey('y')
	if !resolved || v != true {
		t.Fatalf("'y' → resolved=%v v=%v, want true/true", resolved, v)
	}
	if s.Answer == nil || *s.Answer != true {
		t.Fatalf("Answer = %v, want *true", s.Answer)
	}
}

func TestConfirmStateHandleKeyYesUppercase(t *testing.T) {
	s := components.NewConfirmState("Q?", false)
	resolved, v := s.HandleKey('Y')
	if !resolved || v != true {
		t.Fatalf("'Y' → resolved=%v v=%v, want true/true", resolved, v)
	}
}

func TestConfirmStateHandleKeyNoLowercase(t *testing.T) {
	s := components.NewConfirmState("Q?", true)
	resolved, v := s.HandleKey('n')
	if !resolved || v != false {
		t.Fatalf("'n' → resolved=%v v=%v, want true/false", resolved, v)
	}
	if s.Answer == nil || *s.Answer != false {
		t.Fatalf("Answer = %v, want *false", s.Answer)
	}
}

func TestConfirmStateHandleKeyNoUppercase(t *testing.T) {
	s := components.NewConfirmState("Q?", true)
	resolved, v := s.HandleKey('N')
	if !resolved || v != false {
		t.Fatalf("'N' → resolved=%v v=%v, want true/false", resolved, v)
	}
}

func TestConfirmStateHandleKeyEnterDefaultTrue(t *testing.T) {
	s := components.NewConfirmState("Q?", true)
	resolved, v := s.HandleKey('\r')
	if !resolved || v != true {
		t.Fatalf("Enter (def=true) → resolved=%v v=%v, want true/true", resolved, v)
	}
	if s.Answer == nil || *s.Answer != true {
		t.Fatalf("Answer = %v, want *true", s.Answer)
	}
}

func TestConfirmStateHandleKeyEnterDefaultFalse(t *testing.T) {
	s := components.NewConfirmState("Q?", false)
	resolved, v := s.HandleKey('\n')
	if !resolved || v != false {
		t.Fatalf("Enter (def=false) → resolved=%v v=%v, want true/false", resolved, v)
	}
	if s.Answer == nil || *s.Answer != false {
		t.Fatalf("Answer = %v, want *false", s.Answer)
	}
}

func TestConfirmStateHandleKeyIgnoresOther(t *testing.T) {
	s := components.NewConfirmState("Q?", true)
	resolved, v := s.HandleKey('x')
	if resolved || v {
		t.Fatalf("'x' → resolved=%v v=%v, want false/false", resolved, v)
	}
	if s.Answer != nil {
		t.Fatalf("Answer = %v, want nil after ignored key", s.Answer)
	}
}

func TestConfirmStateReset(t *testing.T) {
	s := components.NewConfirmState("Q?", true)
	s.HandleKey('y')
	if s.Answer == nil {
		t.Fatal("setup: expected Answer to be set")
	}
	s.Reset()
	if s.Answer != nil {
		t.Fatalf("Answer = %v after Reset, want nil", s.Answer)
	}
}
