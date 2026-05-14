package hooks

// UseReducer is a derived hook implemented on top of UseState + UseRef. It
// returns (state, dispatch) where dispatch has stable identity across
// renders — important for using it in UseEffect deps without re-running on
// every render.
//
// The reducer is captured fresh each render (so closures over component
// state stay current), but the dispatch closure itself is allocated only
// once via a UseRef-anchored cell, mirroring React's invariant that the
// dispatch function never changes between renders.
//
// State equality short-circuits via the existing UseState DeepEqual guard
// in setState — passing a reducer that returns its input value will not
// trigger a re-render.
func UseReducer[S, A any](ctx *Context, reducer func(S, A) S, initial S) (S, func(A)) {
	if reducer == nil {
		panic("UseReducer: reducer must not be nil")
	}

	rawState, setState := UseState(ctx, initial)
	state, ok := rawState.(S)
	if !ok {
		// Should only happen if the slot was populated by a prior render with
		// a different type signature — recover with the zero value rather
		// than panic mid-render.
		var zero S
		state = zero
	}

	cellRef := UseRef(ctx, (*reducerCell[S, A])(nil))
	cell, _ := cellRef.Current().(*reducerCell[S, A])
	if cell == nil {
		cell = &reducerCell[S, A]{}
		cellRef.SetCurrent(cell)
	}

	cell.reducer = reducer
	cell.state = state
	cell.setState = setState

	if cell.dispatch == nil {
		c := cell // capture for stable closure
		cell.dispatch = func(action A) {
			next := c.reducer(c.state, action)
			// Mirror in-cell so consecutive synchronous dispatches in the same
			// tick observe the latest state. UseState's setter handles the
			// re-render request and the DeepEqual no-op guard.
			c.state = next
			c.setState(next)
		}
	}

	return state, cell.dispatch
}

// reducerCell holds the bookkeeping that survives between renders so that
// the dispatch closure can keep stable identity while its captured state /
// reducer / setter stay fresh.
type reducerCell[S, A any] struct {
	reducer  func(S, A) S
	state    S
	setState SetStateFunc
	dispatch func(A)
}
