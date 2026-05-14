package ink

import (
	"strings"
	"sync"
)

// AnnouncementUrgency values mirror the WAI-ARIA `aria-live` token set used
// by the static announcer regions in the renderer. Anything other than the
// two recognized values is normalized to "polite" so the runtime path stays
// in sync with the rendered output.
const (
	AnnouncementPolite    = "polite"
	AnnouncementAssertive = "assertive"
)

// Announcement is a single queued screen-reader notification emitted through
// the runtime announcer channel. It is the unit produced by Announcer.Drain
// and consumed by tests and the render path.
type Announcement struct {
	Message string
	Urgency string
}

// Announcer is the session-scoped runtime aria-live channel. Components call
// Announce (via the UseAnnounce hook) to enqueue a message; the session
// rotates the queue once per render so messages dispatched during render N
// land in render N+1's screen-reader output and are then dropped.
//
// Queue policy:
//   - No deduplication: dispatching the same message twice queues two
//     entries. This matches "default to no-dedupe" from the design brief and
//     keeps the runtime channel symmetric with the static aria-live regions
//     (which also re-narrate identical text when re-rendered).
//   - Single-frame retention: announcements are emitted in exactly one
//     render frame and then dropped. Both polite and assertive use the same
//     retention since upstream Ink has no announcer of its own to mirror.
//   - Pending vs active: announcements collected during render or effects
//     accumulate in `pending`. BeginRender rotates pending into `active`,
//     which the renderer reads. Capturing this snapshot before component
//     execution makes the "land in next frame" guarantee deterministic.
type Announcer struct {
	mu            sync.Mutex
	pending       []Announcement
	active        []Announcement
	requestRender func()
}

// newAnnouncer constructs an Announcer wired to the supplied rerender hook.
// requestRender may be nil when the announcer is detached from a session
// (for example in tests that exercise the queue directly).
func newAnnouncer(requestRender func()) *Announcer {
	return &Announcer{requestRender: requestRender}
}

// Announce enqueues a message at the requested urgency. Unknown urgency
// strings collapse to "polite" so the queue never carries values the static
// renderer wouldn't accept.
//
// When the announcer is attached to a session, queueing a message also
// requests a re-render so the announcement actually surfaces. Without the
// hook a producer that called Announce outside a render pass would have to
// wait for some other state change before its message ever reached output.
func (a *Announcer) Announce(message, urgency string) {
	if a == nil {
		return
	}

	urgency = normalizeAnnouncementUrgency(urgency)

	a.mu.Lock()
	a.pending = append(a.pending, Announcement{Message: message, Urgency: urgency})
	hook := a.requestRender
	a.mu.Unlock()

	if hook != nil {
		hook()
	}
}

// BeginRender rotates the pending queue into the active slot and returns
// the slice of announcements the current render should emit. Called by the
// session immediately before each render pass so that messages dispatched
// inside the upcoming render body or its effects accumulate into the next
// pending queue, not the active one.
func (a *Announcer) BeginRender() []Announcement {
	if a == nil {
		return nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	a.active = a.pending
	a.pending = nil
	return cloneAnnouncements(a.active)
}

// Active returns a copy of the current frame's active announcements without
// rotating the queue. Used by tests and any out-of-band consumer that wants
// to inspect what would emit on the most recent render.
func (a *Announcer) Active() []Announcement {
	if a == nil {
		return nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	return cloneAnnouncements(a.active)
}

// Pending returns a copy of the queue staged for the next render. Useful in
// tests to assert that an Announce call landed in the queue without
// triggering a render rotation.
func (a *Announcer) Pending() []Announcement {
	if a == nil {
		return nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	return cloneAnnouncements(a.pending)
}

// Clear drops both the active and pending queues. Called from
// Instance.Unmount so a session that is being torn down cannot continue to
// surface stale announcements.
func (a *Announcer) Clear() {
	if a == nil {
		return
	}

	a.mu.Lock()
	a.active = nil
	a.pending = nil
	a.mu.Unlock()
}

// renderToText formats the active queue as the renderer-style announcer
// block ("[assertive] msg" lines first, then "[polite] msg" lines). Returns
// "" when the queue is empty so callers can short-circuit append logic.
func (a *Announcer) renderToText() string {
	if a == nil {
		return ""
	}

	a.mu.Lock()
	active := a.active
	a.mu.Unlock()
	if len(active) == 0 {
		return ""
	}

	var builder strings.Builder
	// Match the static renderAnnouncerRegions ordering so a tree whose
	// only announcer content comes from the runtime channel produces the
	// same shape as one whose content lives in the vdom.
	for _, urgency := range []string{AnnouncementAssertive, AnnouncementPolite} {
		for _, entry := range active {
			if entry.Urgency != urgency {
				continue
			}
			if entry.Message == "" {
				continue
			}
			if builder.Len() > 0 {
				builder.WriteString("\n")
			}
			builder.WriteString("[")
			builder.WriteString(urgency)
			builder.WriteString("] ")
			builder.WriteString(entry.Message)
		}
	}

	return builder.String()
}

func normalizeAnnouncementUrgency(urgency string) string {
	if urgency == AnnouncementAssertive {
		return AnnouncementAssertive
	}
	return AnnouncementPolite
}

func cloneAnnouncements(in []Announcement) []Announcement {
	if len(in) == 0 {
		return nil
	}
	out := make([]Announcement, len(in))
	copy(out, in)
	return out
}

// UseAnnounce returns a dispatcher bound to the currently rendering app's
// announcer. Calling the returned function is equivalent to invoking
// `currentApp.Announcer().Announce(message, urgency)`. The hook itself is
// idempotent across renders — it re-resolves the current app each time
// because the announcer pointer is stable for the lifetime of an app.
func UseAnnounce() func(message, urgency string) {
	app := requireCurrentApp("UseAnnounce")
	announcer := app.Announcer()
	return func(message, urgency string) {
		announcer.Announce(message, urgency)
	}
}
