package hooks_test

import (
	"reflect"
	"testing"

	"github.com/dh-kam/goink.go/pkg/hooks"
)

// renderOnce simulates a single render pass: reset hook indices then invoke
// the body. Returns dispatch + final state observed during the render.
func renderOnce[S, A any](
	ctx *hooks.Context,
	reducer func(S, A) S,
	initial S,
) (S, func(A)) {
	ctx.Reset()
	return hooks.UseReducer(ctx, reducer, initial)
}

type counterAction int

const (
	increment counterAction = iota
	decrement
	reset
)

func counterReducer(state int, action counterAction) int {
	switch action {
	case increment:
		return state + 1
	case decrement:
		return state - 1
	case reset:
		return 0
	}
	return state
}

func TestUseReducerInitialState(t *testing.T) {
	ctx := hooks.NewContext()
	state, dispatch := renderOnce(ctx, counterReducer, 5)
	if state != 5 {
		t.Fatalf("initial state = %d, want 5", state)
	}
	if dispatch == nil {
		t.Fatal("dispatch must not be nil")
	}
}

func TestUseReducerDispatchUpdatesState(t *testing.T) {
	ctx := hooks.NewContext()
	_, dispatch := renderOnce(ctx, counterReducer, 0)
	dispatch(increment)
	state, _ := renderOnce(ctx, counterReducer, 0)
	if state != 1 {
		t.Fatalf("state after increment = %d, want 1", state)
	}
}

func TestUseReducerConsecutiveDispatches(t *testing.T) {
	ctx := hooks.NewContext()
	_, dispatch := renderOnce(ctx, counterReducer, 0)
	dispatch(increment)
	dispatch(increment)
	dispatch(increment)
	state, _ := renderOnce(ctx, counterReducer, 0)
	if state != 3 {
		t.Fatalf("after 3 increments state = %d, want 3", state)
	}
}

func TestUseReducerStableDispatchIdentity(t *testing.T) {
	ctx := hooks.NewContext()
	_, d1 := renderOnce(ctx, counterReducer, 0)
	_, d2 := renderOnce(ctx, counterReducer, 0)
	if reflect.ValueOf(d1).Pointer() != reflect.ValueOf(d2).Pointer() {
		t.Fatal("dispatch identity changed between renders — expected stable closure")
	}
}

type formState struct {
	Name  string
	Email string
	Count int
}

type formAction struct {
	Field string
	Value any
}

func formReducer(state formState, action formAction) formState {
	switch action.Field {
	case "name":
		state.Name = action.Value.(string)
	case "email":
		state.Email = action.Value.(string)
	case "count":
		state.Count = action.Value.(int)
	}
	return state
}

func TestUseReducerStructState(t *testing.T) {
	ctx := hooks.NewContext()
	_, dispatch := renderOnce(ctx, formReducer, formState{})
	dispatch(formAction{Field: "name", Value: "alice"})
	dispatch(formAction{Field: "email", Value: "alice@example.com"})
	dispatch(formAction{Field: "count", Value: 7})
	state, _ := renderOnce(ctx, formReducer, formState{})
	want := formState{Name: "alice", Email: "alice@example.com", Count: 7}
	if state != want {
		t.Fatalf("state = %+v, want %+v", state, want)
	}
}

func TestUseReducerNilReducerPanics(t *testing.T) {
	ctx := hooks.NewContext()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on nil reducer")
		}
	}()
	hooks.UseReducer[int, int](ctx, nil, 0)
}
