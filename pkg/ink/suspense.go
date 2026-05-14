package ink

import "github.com/dh-kam/goink.go/pkg/vdom"

type suspension struct {
	done <-chan struct{}
}

// SuspendUntil suspends the current render until done is closed. It must be
// called from inside Suspense's render function.
func SuspendUntil(done <-chan struct{}) {
	if done == nil {
		panic("SuspendUntil requires a non-nil done channel")
	}

	panic(suspension{done: done})
}

// Suspense renders fallback while its render function is suspended. When the
// suspended work resolves, the mounted runtime schedules a rerender.
func Suspense(fallback *vdom.Node, render func() *vdom.Node) (node *vdom.Node) {
	app := requireCurrentApp("Suspense")
	if render == nil {
		return fallback
	}

	previousCursorPosition := cloneCursorPosition(app.cursorPosition)
	defer func() {
		recovered := recover()
		if recovered == nil {
			return
		}

		suspended, ok := recovered.(suspension)
		if !ok {
			panic(recovered)
		}

		app.cursorPosition = previousCursorPosition
		app.watchSuspension(suspended.done)
		node = fallback
	}()

	return render()
}

func (a *App) watchSuspension(done <-chan struct{}) {
	go func() {
		<-done
		a.scheduleRuntimeWork(func() {
			a.hooksCtx.RequestRerender()
		})
	}()
}
