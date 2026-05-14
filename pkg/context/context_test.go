package context_test

import (
	"sync"
	"sync/atomic"
	"testing"

	gocontext "github.com/dh-kam/goink.go/pkg/context"
)

func TestDefaultWhenEmpty(t *testing.T) {
	c := gocontext.New("light")
	if got := c.Current(); got != "light" {
		t.Fatalf("Current() = %q, want %q", got, "light")
	}
	if got := c.Default(); got != "light" {
		t.Fatalf("Default() = %q, want %q", got, "light")
	}
	if depth := c.Depth(); depth != 0 {
		t.Fatalf("Depth() = %d, want 0", depth)
	}
}

func TestPushPopSingle(t *testing.T) {
	c := gocontext.New(0)
	pop := c.Push(42)
	if got := c.Current(); got != 42 {
		t.Fatalf("Current() after Push = %d, want 42", got)
	}
	if c.Depth() != 1 {
		t.Fatalf("Depth() = %d, want 1", c.Depth())
	}
	pop()
	if got := c.Current(); got != 0 {
		t.Fatalf("Current() after pop = %d, want 0 (default)", got)
	}
	if c.Depth() != 0 {
		t.Fatalf("Depth() = %d, want 0", c.Depth())
	}
}

func TestPushPopNested(t *testing.T) {
	c := gocontext.New("a")
	pop1 := c.Push("b")
	pop2 := c.Push("c")
	if got := c.Current(); got != "c" {
		t.Fatalf("nested Current() = %q, want %q", got, "c")
	}
	pop2()
	if got := c.Current(); got != "b" {
		t.Fatalf("after pop2 Current() = %q, want %q", got, "b")
	}
	pop1()
	if got := c.Current(); got != "a" {
		t.Fatalf("after pop1 Current() = %q, want %q", got, "a")
	}
}

func TestPushClosersIdempotent(t *testing.T) {
	c := gocontext.New(0)
	pop := c.Push(1)
	pop()
	pop() // must be no-op
	pop()
	if c.Depth() != 0 {
		t.Fatalf("Depth() after triple pop = %d, want 0", c.Depth())
	}
}

func TestPushOutOfOrderPopTruncates(t *testing.T) {
	// Simulates a panicked Provider where pop2 never runs; calling pop1
	// must still leave the stack consistent (truncated to my slot).
	c := gocontext.New(0)
	pop1 := c.Push(1)
	c.Push(2) // pop2 leaked
	c.Push(3) // pop3 leaked
	if c.Depth() != 3 {
		t.Fatalf("setup Depth() = %d, want 3", c.Depth())
	}
	pop1() // truncates back to my slot, dropping 2 and 3 above me
	if c.Depth() != 0 {
		t.Fatalf("after out-of-order pop1 Depth() = %d, want 0", c.Depth())
	}
	if got := c.Current(); got != 0 {
		t.Fatalf("Current() = %d, want 0", got)
	}
}

func TestProviderHappyPath(t *testing.T) {
	c := gocontext.New("default")
	called := false
	c.Provider("override", func() {
		called = true
		if got := c.Current(); got != "override" {
			t.Errorf("inside Provider Current() = %q, want %q", got, "override")
		}
	})
	if !called {
		t.Fatal("Provider fn was not called")
	}
	if got := c.Current(); got != "default" {
		t.Fatalf("after Provider Current() = %q, want %q", got, "default")
	}
}

func TestProviderRecoversAndPops(t *testing.T) {
	c := gocontext.New("default")
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic to propagate")
		}
		if c.Depth() != 0 {
			t.Fatalf("Depth() after panic = %d, want 0", c.Depth())
		}
	}()
	c.Provider("inside", func() {
		panic("boom")
	})
}

func TestProviderNilFnPanics(t *testing.T) {
	c := gocontext.New(0)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on nil fn")
		}
	}()
	c.Provider(1, nil)
}

func TestReset(t *testing.T) {
	c := gocontext.New("d")
	c.Push("a")
	c.Push("b")
	c.Reset()
	if c.Depth() != 0 {
		t.Fatalf("Depth() after Reset = %d, want 0", c.Depth())
	}
	if got := c.Current(); got != "d" {
		t.Fatalf("Current() after Reset = %q, want %q", got, "d")
	}
}

func TestConcurrentPushPop(t *testing.T) {
	c := gocontext.New(0)
	var wg sync.WaitGroup
	const goroutines = 64
	const perG = 100
	var ops int64
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < perG; i++ {
				pop := c.Push(id*1000 + i)
				_ = c.Current()
				pop()
				atomic.AddInt64(&ops, 1)
			}
		}(g)
	}
	wg.Wait()
	if got := atomic.LoadInt64(&ops); got != int64(goroutines*perG) {
		t.Fatalf("ops = %d, want %d", got, goroutines*perG)
	}
	if c.Depth() != 0 {
		t.Fatalf("Depth() after concurrent run = %d, want 0", c.Depth())
	}
}

type theme struct {
	Name  string
	Color string
}

func TestStructValue(t *testing.T) {
	c := gocontext.New(theme{Name: "light", Color: "white"})
	c.Provider(theme{Name: "dark", Color: "black"}, func() {
		got := c.Current()
		if got.Name != "dark" || got.Color != "black" {
			t.Fatalf("Current() = %+v, want dark/black", got)
		}
	})
	if got := c.Current(); got.Name != "light" {
		t.Fatalf("Current() after Provider = %+v, want light/white", got)
	}
}
