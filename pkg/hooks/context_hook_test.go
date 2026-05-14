package hooks_test

import (
	"testing"

	gocontext "github.com/dh-kam/goink.go/pkg/context"
	"github.com/dh-kam/goink.go/pkg/hooks"
)

func TestUseContextDefault(t *testing.T) {
	ctx := hooks.NewContext()
	c := gocontext.New("light")
	got := hooks.UseContext(ctx, c)
	if got != "light" {
		t.Fatalf("UseContext = %q, want %q", got, "light")
	}
}

func TestUseContextProvided(t *testing.T) {
	ctx := hooks.NewContext()
	c := gocontext.New("light")
	pop := c.Push("dark")
	defer pop()
	got := hooks.UseContext(ctx, c)
	if got != "dark" {
		t.Fatalf("UseContext = %q, want %q", got, "dark")
	}
}

func TestUseContextNested(t *testing.T) {
	ctx := hooks.NewContext()
	c := gocontext.New(0)
	pop1 := c.Push(1)
	defer pop1()
	pop2 := c.Push(2)
	defer pop2()
	if got := hooks.UseContext(ctx, c); got != 2 {
		t.Fatalf("UseContext = %d, want 2", got)
	}
}

func TestUseContextNilPanics(t *testing.T) {
	ctx := hooks.NewContext()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on nil context")
		}
	}()
	hooks.UseContext[int](ctx, nil)
}
