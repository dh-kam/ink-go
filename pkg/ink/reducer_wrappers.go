package ink

import "github.com/dh-kam/ink-go/pkg/hooks"

// UseReducer is the Ink-level wrapper around hooks.UseReducer. Returns
// (state, dispatch) where dispatch has stable identity across renders, so
// it is safe to depend on in UseEffect/UseMemo deps without forcing
// re-runs.
func UseReducer[S, A any](reducer func(S, A) S, initial S) (S, func(A)) {
	return hooks.UseReducer[S, A](requireHooksContext("UseReducer"), reducer, initial)
}
