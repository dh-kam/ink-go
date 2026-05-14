package devtools

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dh-kam/ink-go/pkg/vdom"
)

// Profile is a single render measurement.
//
// A Profile records when a render started, how long the render-or-cache-lookup
// took, the size of the vdom tree that was processed, and whether the result
// was a fresh render (Fresh=true) or a cache hit (Fresh=false).
type Profile struct {
	StartedAt time.Time
	Duration  time.Duration
	NodeCount int  // tree size
	Fresh     bool // true if rendered, false if cache hit
}

// Profiler accumulates Profile entries thread-safely.
//
// All public methods are safe for concurrent use; internally a single mutex
// guards the underlying slice. The intended usage is to wrap a render
// function via Wrap and let it record automatically, while Record is provided
// for manual cases (e.g. counting cache-hit ticks that never call the render
// function at all).
type Profiler struct {
	mu       sync.Mutex
	profiles []Profile
}

// NewProfiler returns an empty Profiler ready for use.
func NewProfiler() *Profiler {
	return &Profiler{}
}

// Wrap returns a render function that records each call into the profiler.
//
// inner is the actual render func (e.g. ink.RenderToString). Each invocation
// of the returned function measures wall-clock duration, counts the input
// tree's nodes, marks the call as a fresh render, and appends a Profile
// entry to the receiver. The Profile is also returned alongside the rendered
// string so callers can inspect the measurement inline.
//
// To track cache-hit calls (which do not invoke inner) use Record instead.
func (p *Profiler) Wrap(inner func(*vdom.Node) string) func(*vdom.Node) (string, Profile) {
	return func(node *vdom.Node) (string, Profile) {
		start := time.Now()
		var output string
		if inner != nil {
			output = inner(node)
		}
		profile := Profile{
			StartedAt: start,
			Duration:  time.Since(start),
			NodeCount: countNodes(node),
			Fresh:     true,
		}
		p.Record(profile)
		return output, profile
	}
}

// Record manually appends a profile to the profiler.
//
// This is useful for cache-hit ticks where the render function is bypassed
// entirely but the runtime still wants to account for the work avoided.
func (p *Profiler) Record(profile Profile) {
	p.mu.Lock()
	p.profiles = append(p.profiles, profile)
	p.mu.Unlock()
}

// Profiles returns a defensive copy of recorded profiles.
//
// Mutating the returned slice has no effect on the profiler's internal state.
func (p *Profiler) Profiles() []Profile {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.profiles) == 0 {
		return nil
	}

	out := make([]Profile, len(p.profiles))
	copy(out, p.profiles)
	return out
}

// Stats summarises a Profiler's recorded measurements.
type Stats struct {
	TotalRenders int
	TotalTime    time.Duration
	AvgDuration  time.Duration
	MinDuration  time.Duration
	MaxDuration  time.Duration
	FreshCount   int
	CachedCount  int
	HitRatio     float64 // CachedCount / TotalRenders
}

// Stats returns aggregate metrics across all recorded profiles.
//
// For an empty profiler, every numeric field is zero and HitRatio is 0
// (never NaN), so callers can safely format the result without guarding
// against division-by-zero.
func (p *Profiler) Stats() Stats {
	p.mu.Lock()
	defer p.mu.Unlock()

	stats := Stats{}
	if len(p.profiles) == 0 {
		return stats
	}

	stats.TotalRenders = len(p.profiles)
	stats.MinDuration = p.profiles[0].Duration
	stats.MaxDuration = p.profiles[0].Duration

	for _, profile := range p.profiles {
		stats.TotalTime += profile.Duration
		if profile.Duration < stats.MinDuration {
			stats.MinDuration = profile.Duration
		}
		if profile.Duration > stats.MaxDuration {
			stats.MaxDuration = profile.Duration
		}
		if profile.Fresh {
			stats.FreshCount++
		} else {
			stats.CachedCount++
		}
	}

	stats.AvgDuration = time.Duration(int64(stats.TotalTime) / int64(stats.TotalRenders))
	stats.HitRatio = float64(stats.CachedCount) / float64(stats.TotalRenders)
	return stats
}

// Reset clears all recorded profiles.
func (p *Profiler) Reset() {
	p.mu.Lock()
	p.profiles = nil
	p.mu.Unlock()
}

// Format returns a multi-line tabular summary of the stats.
//
// The format is intended for human consumption (CLI debug output, log
// entries) and includes every Stats field so the output is self-describing.
func (s Stats) Format() string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Renders     : %d\n", s.TotalRenders)
	fmt.Fprintf(&builder, "Total time  : %s\n", s.TotalTime)
	fmt.Fprintf(&builder, "Avg duration: %s\n", s.AvgDuration)
	fmt.Fprintf(&builder, "Min duration: %s\n", s.MinDuration)
	fmt.Fprintf(&builder, "Max duration: %s\n", s.MaxDuration)
	fmt.Fprintf(&builder, "Fresh       : %d\n", s.FreshCount)
	fmt.Fprintf(&builder, "Cached      : %d\n", s.CachedCount)
	fmt.Fprintf(&builder, "Hit ratio   : %.2f%%\n", s.HitRatio*100)
	return builder.String()
}

// countNodes recursively counts the receiver and all descendants.
//
// nil children are skipped to mirror the rest of the vdom package, which
// tolerates sparse Children slices.
func countNodes(node *vdom.Node) int {
	if node == nil {
		return 0
	}

	count := 1
	for _, child := range node.Children {
		count += countNodes(child)
	}
	return count
}
