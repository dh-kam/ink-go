package devtools

import (
	"math"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dh-kam/goink.go/pkg/vdom"
)

func TestProfilerEmptyStats(t *testing.T) {
	t.Parallel()

	p := NewProfiler()
	stats := p.Stats()

	if stats.TotalRenders != 0 {
		t.Fatalf("expected zero renders, got %d", stats.TotalRenders)
	}
	if stats.TotalTime != 0 {
		t.Fatalf("expected zero total time, got %s", stats.TotalTime)
	}
	if stats.AvgDuration != 0 || stats.MinDuration != 0 || stats.MaxDuration != 0 {
		t.Fatalf("expected zero durations, got avg=%s min=%s max=%s",
			stats.AvgDuration, stats.MinDuration, stats.MaxDuration)
	}
	if stats.FreshCount != 0 || stats.CachedCount != 0 {
		t.Fatalf("expected zero counts, got fresh=%d cached=%d", stats.FreshCount, stats.CachedCount)
	}
	if math.IsNaN(stats.HitRatio) || stats.HitRatio != 0 {
		t.Fatalf("expected hit ratio 0 (not NaN), got %v", stats.HitRatio)
	}

	if got := p.Profiles(); got != nil {
		t.Fatalf("expected nil profiles slice for empty profiler, got %#v", got)
	}
}

func TestProfilerWrapRecordsRender(t *testing.T) {
	t.Parallel()

	p := NewProfiler()

	tree := vdom.CreateElement("box", nil,
		vdom.CreateTextNode("hello"),
		vdom.CreateTextNode("world"),
	)

	called := 0
	render := p.Wrap(func(node *vdom.Node) string {
		called++
		// Sleep ensures Duration > 0 even on extremely fast machines /
		// low-resolution clocks.
		time.Sleep(2 * time.Millisecond)
		return "rendered:" + node.ElementType
	})

	out, profile := render(tree)
	if called != 1 {
		t.Fatalf("expected inner to be called once, got %d", called)
	}
	if out != "rendered:box" {
		t.Fatalf("unexpected render output: %q", out)
	}
	if profile.Duration <= 0 {
		t.Fatalf("expected duration > 0, got %s", profile.Duration)
	}
	if !profile.Fresh {
		t.Fatalf("expected Fresh=true for wrapped call")
	}
	if profile.NodeCount != 3 {
		t.Fatalf("expected 3 nodes (box + 2 text), got %d", profile.NodeCount)
	}
	if profile.StartedAt.IsZero() {
		t.Fatalf("expected non-zero StartedAt")
	}

	profiles := p.Profiles()
	if len(profiles) != 1 {
		t.Fatalf("expected 1 recorded profile, got %d", len(profiles))
	}
	if profiles[0].Duration != profile.Duration {
		t.Fatalf("recorded profile mismatches returned profile")
	}
}

func TestProfilerWrapNilInnerStillRecords(t *testing.T) {
	t.Parallel()

	p := NewProfiler()
	render := p.Wrap(nil)

	out, profile := render(vdom.CreateTextNode("x"))
	if out != "" {
		t.Fatalf("expected empty output when inner is nil, got %q", out)
	}
	if !profile.Fresh {
		t.Fatalf("expected Fresh=true even with nil inner")
	}
	if profile.NodeCount != 1 {
		t.Fatalf("expected node count 1 for single text node, got %d", profile.NodeCount)
	}
	if got := len(p.Profiles()); got != 1 {
		t.Fatalf("expected 1 profile recorded, got %d", got)
	}
}

func TestProfilerStatsAggregation(t *testing.T) {
	t.Parallel()

	p := NewProfiler()

	now := time.Now()
	p.Record(Profile{StartedAt: now, Duration: 10 * time.Millisecond, NodeCount: 4, Fresh: true})
	p.Record(Profile{StartedAt: now, Duration: 30 * time.Millisecond, NodeCount: 4, Fresh: true})
	p.Record(Profile{StartedAt: now, Duration: 20 * time.Millisecond, NodeCount: 4, Fresh: false})
	p.Record(Profile{StartedAt: now, Duration: 60 * time.Millisecond, NodeCount: 4, Fresh: false})

	stats := p.Stats()

	if stats.TotalRenders != 4 {
		t.Fatalf("expected 4 renders, got %d", stats.TotalRenders)
	}
	if want := 120 * time.Millisecond; stats.TotalTime != want {
		t.Fatalf("expected total time %s, got %s", want, stats.TotalTime)
	}
	if want := 30 * time.Millisecond; stats.AvgDuration != want {
		t.Fatalf("expected avg %s, got %s", want, stats.AvgDuration)
	}
	if want := 10 * time.Millisecond; stats.MinDuration != want {
		t.Fatalf("expected min %s, got %s", want, stats.MinDuration)
	}
	if want := 60 * time.Millisecond; stats.MaxDuration != want {
		t.Fatalf("expected max %s, got %s", want, stats.MaxDuration)
	}
	if stats.FreshCount != 2 || stats.CachedCount != 2 {
		t.Fatalf("expected 2/2 split, got fresh=%d cached=%d", stats.FreshCount, stats.CachedCount)
	}
	if stats.HitRatio != 0.5 {
		t.Fatalf("expected hit ratio 0.5, got %v", stats.HitRatio)
	}
}

func TestProfilerConcurrentRecord(t *testing.T) {
	t.Parallel()

	p := NewProfiler()

	const goroutines = 16
	const perGoroutine = 64

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		fresh := i%2 == 0
		go func(fresh bool) {
			defer wg.Done()
			for j := 0; j < perGoroutine; j++ {
				p.Record(Profile{
					StartedAt: time.Now(),
					Duration:  time.Microsecond,
					NodeCount: 1,
					Fresh:     fresh,
				})
			}
		}(fresh)
	}
	wg.Wait()

	stats := p.Stats()
	expected := goroutines * perGoroutine
	if stats.TotalRenders != expected {
		t.Fatalf("expected %d renders, got %d", expected, stats.TotalRenders)
	}
	if stats.FreshCount+stats.CachedCount != expected {
		t.Fatalf("fresh+cached (%d+%d) != total %d",
			stats.FreshCount, stats.CachedCount, expected)
	}
	// Half the goroutines record fresh, half cached.
	if stats.FreshCount != stats.CachedCount {
		t.Fatalf("expected even fresh/cached split, got fresh=%d cached=%d",
			stats.FreshCount, stats.CachedCount)
	}
}

func TestProfilerProfilesDefensiveCopy(t *testing.T) {
	t.Parallel()

	p := NewProfiler()
	p.Record(Profile{Duration: 5 * time.Millisecond, NodeCount: 2, Fresh: true})
	p.Record(Profile{Duration: 7 * time.Millisecond, NodeCount: 3, Fresh: false})

	first := p.Profiles()
	if len(first) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(first))
	}

	// Mutate the returned slice — internal state must not change.
	first[0].Duration = 9999 * time.Hour
	first[1].NodeCount = -1

	second := p.Profiles()
	if second[0].Duration != 5*time.Millisecond {
		t.Fatalf("internal state was mutated via returned slice (Duration)")
	}
	if second[1].NodeCount != 3 {
		t.Fatalf("internal state was mutated via returned slice (NodeCount)")
	}

	// Appending to the returned slice must also not affect the internal slice.
	first = append(first, Profile{Duration: time.Hour})
	if got := len(p.Profiles()); got != 2 {
		t.Fatalf("expected internal length 2 after external append, got %d", got)
	}
}

func TestProfilerReset(t *testing.T) {
	t.Parallel()

	p := NewProfiler()
	p.Record(Profile{Duration: time.Millisecond, Fresh: true})
	p.Record(Profile{Duration: 2 * time.Millisecond, Fresh: false})

	if len(p.Profiles()) != 2 {
		t.Fatalf("expected 2 profiles before reset")
	}

	p.Reset()

	if got := p.Profiles(); got != nil {
		t.Fatalf("expected nil profiles after reset, got %#v", got)
	}
	stats := p.Stats()
	if stats.TotalRenders != 0 || stats.TotalTime != 0 || stats.HitRatio != 0 {
		t.Fatalf("expected zeroed stats after reset, got %+v", stats)
	}
}

func TestStatsFormatContainsKeyMetrics(t *testing.T) {
	t.Parallel()

	stats := Stats{
		TotalRenders: 3,
		TotalTime:    30 * time.Millisecond,
		AvgDuration:  10 * time.Millisecond,
		MinDuration:  5 * time.Millisecond,
		MaxDuration:  20 * time.Millisecond,
		FreshCount:   2,
		CachedCount:  1,
		HitRatio:     1.0 / 3.0,
	}

	out := stats.Format()

	for _, want := range []string{
		"Renders",
		"Total time",
		"Avg duration",
		"Min duration",
		"Max duration",
		"Fresh",
		"Cached",
		"Hit ratio",
		"3",                         // TotalRenders
		"30ms",                      // TotalTime
		"10ms",                      // AvgDuration
		"5ms",                       // MinDuration
		"20ms",                      // MaxDuration
		"33.33%",                    // HitRatio formatted as percent
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("Format output missing %q\n---\n%s", want, out)
		}
	}
}

func TestCountNodes(t *testing.T) {
	t.Parallel()

	if got := countNodes(nil); got != 0 {
		t.Fatalf("expected 0 for nil node, got %d", got)
	}

	leaf := vdom.CreateTextNode("a")
	if got := countNodes(leaf); got != 1 {
		t.Fatalf("expected 1 for single node, got %d", got)
	}

	tree := vdom.CreateElement("box", nil,
		vdom.CreateTextNode("a"),
		vdom.CreateElement("box", nil,
			vdom.CreateTextNode("b"),
			vdom.CreateTextNode("c"),
		),
	)
	if got := countNodes(tree); got != 5 {
		t.Fatalf("expected 5 nodes in compound tree, got %d", got)
	}

	// Sparse children with a nil entry must not panic and must be skipped.
	sparse := vdom.CreateElement("box", nil, vdom.CreateTextNode("x"))
	sparse.Children = append(sparse.Children, nil)
	if got := countNodes(sparse); got != 2 {
		t.Fatalf("expected 2 for sparse tree (root + 1 text, nil skipped), got %d", got)
	}
}
