package components_test

import (
	"strings"
	"testing"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

func linkText(t *testing.T, p components.LinkProps) string {
	t.Helper()
	node := components.Link(p)
	if node == nil {
		t.Fatal("Link returned nil node")
	}
	if node.Type != vdom.TextNode {
		t.Fatalf("Link returned non-text node (type=%v)", node.Type)
	}
	return node.Text
}

func TestLinkURLAndText(t *testing.T) {
	got := linkText(t, components.LinkProps{URL: "https://example.com", Text: "Example"})
	want := "\x1b]8;;https://example.com\x1b\\Example\x1b]8;;\x1b\\"
	if got != want {
		t.Fatalf("Link mismatch:\n got:  %q\n want: %q", got, want)
	}
}

func TestLinkURLOnlyDefaultsTextToURL(t *testing.T) {
	url := "https://example.com/path?x=1"
	got := linkText(t, components.LinkProps{URL: url})
	// Visible label should be the URL itself when Text is unset.
	wantVisible := url
	wantPrefix := "\x1b]8;;" + url + "\x1b\\"
	wantSuffix := "\x1b]8;;\x1b\\"

	if !strings.HasPrefix(got, wantPrefix) {
		t.Fatalf("missing OSC 8 open prefix.\n got: %q\n want prefix: %q", got, wantPrefix)
	}
	if !strings.HasSuffix(got, wantSuffix) {
		t.Fatalf("missing OSC 8 close suffix.\n got: %q\n want suffix: %q", got, wantSuffix)
	}
	if !strings.Contains(got, wantVisible) {
		t.Fatalf("expected visible label %q in %q", wantVisible, got)
	}

	// The visible label appears between the open ST and the close OSC,
	// so the full byte string must equal prefix + URL + close.
	full := wantPrefix + url + wantSuffix
	if got != full {
		t.Fatalf("Link bytes mismatch:\n got:  %q\n want: %q", got, full)
	}
}

func TestLinkEmptyURLPlainText(t *testing.T) {
	got := linkText(t, components.LinkProps{Text: "click me"})
	if got != "click me" {
		t.Fatalf("expected plain text, got %q", got)
	}
	if strings.Contains(got, "\x1b") {
		t.Fatalf("expected no escape sequences, got %q", got)
	}
}

func TestLinkEmptyURLAndTextProducesEmpty(t *testing.T) {
	got := linkText(t, components.LinkProps{})
	if got != "" {
		t.Fatalf("expected empty string for empty Link, got %q", got)
	}
}

func TestLinkExactByteSequence(t *testing.T) {
	got := linkText(t, components.LinkProps{URL: "u", Text: "t"})
	// Spell the expected bytes out by hand to defend against a sneaky
	// regression in the OSC 8 framing constants.
	want := []byte{
		0x1b, ']', '8', ';', ';', 'u', 0x1b, '\\',
		't',
		0x1b, ']', '8', ';', ';', 0x1b, '\\',
	}
	if got != string(want) {
		t.Fatalf("byte mismatch.\n got:  % x\n want: % x", got, want)
	}
}
