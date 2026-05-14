package ink_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/dh-kam/ink-go/pkg/components"
	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

// TestUseAnnounceReturnsDispatcher verifies the hook hands back a non-nil
// dispatcher that can be invoked without panicking. This guards the
// dispatcher contract independently of where announcements actually land.
func TestUseAnnounceReturnsDispatcher(t *testing.T) {
	var dispatcher func(message, urgency string)

	app := ink.NewAppWithOptions(func() *vdom.Node {
		dispatcher = ink.UseAnnounce()
		return components.Text("ready")
	}, ink.AppOptions{ScreenReaderEnabled: true})

	app.RenderOnce()

	if dispatcher == nil {
		t.Fatal("expected UseAnnounce to return a non-nil dispatcher")
	}

	// The dispatcher captures the announcer pointer at call time, so it
	// remains usable even after the hook context resets.
	dispatcher("ack", ink.AnnouncementPolite)

	pending := app.Announcer().Pending()
	if len(pending) != 1 || pending[0].Message != "ack" || pending[0].Urgency != ink.AnnouncementPolite {
		t.Fatalf("expected pending [{ack polite}], got %+v", pending)
	}
}

// TestUseAnnouncePanicsOutsideRender ensures the hook follows the same
// out-of-render contract as the other UseX hooks: there's no current app to
// resolve, so calling it panics rather than silently producing a nil
// dispatcher.
func TestUseAnnouncePanicsOutsideRender(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected UseAnnounce to panic outside of component render")
		}
	}()

	_ = ink.UseAnnounce()
}

// TestAnnounceDispatchedDuringRenderLandsInNextFrame is the central
// guarantee of the queue policy: a message dispatched while building frame
// N must appear in frame N+1, never in frame N. This protects against a
// component reading its own dispatch back inside the same render.
func TestAnnounceDispatchedDuringRenderLandsInNextFrame(t *testing.T) {
	calls := 0
	app := ink.NewAppWithOptions(func() *vdom.Node {
		calls++
		announce := ink.UseAnnounce()
		if calls == 1 {
			announce("hello", ink.AnnouncementPolite)
		}
		return components.Text("body")
	}, ink.AppOptions{ScreenReaderEnabled: true})

	first, _ := app.RenderSplitOnce()
	if strings.Contains(first, "hello") {
		t.Fatalf("frame 1 must not contain in-render announcement, got %q", first)
	}

	second, _ := app.RenderSplitOnce()
	if !strings.Contains(second, "[polite] hello") {
		t.Fatalf("frame 2 must contain the announcement, got %q", second)
	}

	third, _ := app.RenderSplitOnce()
	if strings.Contains(third, "hello") {
		t.Fatalf("frame 3 must drop one-frame retention announcement, got %q", third)
	}
}

// TestAnnouncerSharesQueueAcrossSubscribers verifies that several components
// in the same tree calling UseAnnounce all funnel into the single
// session-scoped queue rather than each receiving an isolated copy.
func TestAnnouncerSharesQueueAcrossSubscribers(t *testing.T) {
	var first, second func(message, urgency string)

	app := ink.NewAppWithOptions(func() *vdom.Node {
		first = ink.UseAnnounce()
		second = ink.UseAnnounce()
		return components.Text("body")
	}, ink.AppOptions{ScreenReaderEnabled: true})

	app.RenderOnce()

	first("from-a", ink.AnnouncementPolite)
	second("from-b", ink.AnnouncementPolite)

	pending := app.Announcer().Pending()
	if len(pending) != 2 {
		t.Fatalf("expected 2 pending entries shared between subscribers, got %d (%+v)", len(pending), pending)
	}
	if pending[0].Message != "from-a" || pending[1].Message != "from-b" {
		t.Fatalf("expected dispatch order [from-a, from-b], got %+v", pending)
	}
}

// TestAnnouncerSeparatesPoliteFromAssertive checks that the rendered block
// matches the renderer's static announcer ordering: assertive lines emit
// first, polite second. Document order is preserved within each group.
func TestAnnouncerSeparatesPoliteFromAssertive(t *testing.T) {
	app := ink.NewAppWithOptions(func() *vdom.Node {
		return components.Text("body")
	}, ink.AppOptions{ScreenReaderEnabled: true})

	app.RenderOnce()

	app.Announcer().Announce("p1", ink.AnnouncementPolite)
	app.Announcer().Announce("a1", ink.AnnouncementAssertive)
	app.Announcer().Announce("p2", ink.AnnouncementPolite)
	app.Announcer().Announce("a2", ink.AnnouncementAssertive)

	output, _ := app.RenderSplitOnce()

	idxA1 := strings.Index(output, "[assertive] a1")
	idxA2 := strings.Index(output, "[assertive] a2")
	idxP1 := strings.Index(output, "[polite] p1")
	idxP2 := strings.Index(output, "[polite] p2")

	if idxA1 < 0 || idxA2 < 0 || idxP1 < 0 || idxP2 < 0 {
		t.Fatalf("expected all four announcements in output, got %q", output)
	}
	if !(idxA1 < idxA2 && idxA2 < idxP1 && idxP1 < idxP2) {
		t.Fatalf("expected order assertive(a1,a2) before polite(p1,p2) in output %q", output)
	}
}

// TestAnnounceWithoutScreenReaderModeDoesNotEmit guards the augmentation
// boundary: when screen-reader mode is off, the runtime queue still
// accepts dispatches but the rendered output stays free of announcer
// chrome that would corrupt a regular layout.
func TestAnnounceWithoutScreenReaderModeDoesNotEmit(t *testing.T) {
	app := ink.NewApp(func() *vdom.Node {
		return components.Text("body")
	})

	app.RenderOnce()

	app.Announcer().Announce("hidden", ink.AnnouncementAssertive)

	output, _ := app.RenderSplitOnce()
	if strings.Contains(output, "[assertive]") || strings.Contains(output, "hidden") {
		t.Fatalf("did not expect announcer text in non-screen-reader output, got %q", output)
	}
}

// TestUnmountClearsAnnouncerQueue verifies the unmount-time hook empties
// both the active and pending queues so a session that is being torn down
// cannot continue to surface stale announcements after teardown.
func TestUnmountClearsAnnouncerQueue(t *testing.T) {
	stdout := &bytes.Buffer{}
	instance, err := ink.MountWithOptions(func() *vdom.Node {
		return components.Text("hi")
	}, ink.RenderOptions{
		AppOptions: ink.AppOptions{
			Stdout:              stdout,
			ScreenReaderEnabled: true,
		},
	})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}

	announcer := instance.Announcer()
	announcer.Announce("pending-before-unmount", ink.AnnouncementPolite)

	if len(announcer.Pending()) == 0 {
		t.Fatal("expected pending queue populated before unmount")
	}

	if err := instance.Unmount(); err != nil {
		t.Fatalf("unmount returned error: %v", err)
	}

	if got := announcer.Pending(); len(got) != 0 {
		t.Fatalf("expected empty pending queue after unmount, got %+v", got)
	}
	if got := announcer.Active(); len(got) != 0 {
		t.Fatalf("expected empty active queue after unmount, got %+v", got)
	}
}

// TestAnnouncerNoDedupePolicy locks in the documented "no-dedupe" default:
// dispatching the same message twice must produce two separate queue
// entries so consumers that re-narrate identical text still see two
// announcements instead of one.
func TestAnnouncerNoDedupePolicy(t *testing.T) {
	app := ink.NewAppWithOptions(func() *vdom.Node {
		return components.Text("body")
	}, ink.AppOptions{ScreenReaderEnabled: true})

	app.RenderOnce()

	app.Announcer().Announce("same", ink.AnnouncementPolite)
	app.Announcer().Announce("same", ink.AnnouncementPolite)

	pending := app.Announcer().Pending()
	if len(pending) != 2 {
		t.Fatalf("expected 2 entries (no dedupe), got %d", len(pending))
	}

	output, _ := app.RenderSplitOnce()
	if got := strings.Count(output, "[polite] same"); got != 2 {
		t.Fatalf("expected 2 occurrences of '[polite] same' in output, got %d (%q)", got, output)
	}
}
