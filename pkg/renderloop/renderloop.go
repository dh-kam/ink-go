package renderloop

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ErrExitRequested is returned when the render loop should exit
var ErrExitRequested = errors.New("exit requested")

// RenderFn is the function called each frame
type RenderFn func() error

// RenderLoop manages a continuous rendering loop with FPS throttling
type RenderLoop struct {
	fps        int
	frameTime  time.Duration
	running    atomic.Bool
	mu         sync.RWMutex
	cancelOnce sync.Once
}

// NewRenderLoop creates a new render loop with the target FPS
func NewRenderLoop(fps int) *RenderLoop {
	// Clamp FPS to reasonable values
	if fps <= 0 {
		fps = 30 // Default to 30 FPS
	}
	if fps > 240 {
		fps = 240 // Max 240 FPS
	}

	frameTime := time.Duration(1000/fps) * time.Millisecond

	return &RenderLoop{
		fps:       fps,
		frameTime: frameTime,
	}
}

// Start begins the render loop
// The loop will continue until the context is cancelled or RenderFn returns an error
func (rl *RenderLoop) Start(ctx context.Context, renderFn RenderFn) error {
	if !rl.running.CompareAndSwap(false, true) {
		return fmt.Errorf("render loop is already running")
	}
	defer rl.running.Store(false)

	ticker := time.NewTicker(rl.frameTime)
	defer ticker.Stop()

	// Initial render
	if err := renderFn(); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			if err := renderFn(); err != nil {
				return err
			}

		default:
			// Sleep for a short time to avoid busy-waiting
			time.Sleep(5 * time.Millisecond)
		}
	}
}

// Stop stops the render loop by cancelling the context
// Deprecated: Use context cancellation instead
func (rl *RenderLoop) Stop() {
	// This is a no-op - use context cancellation instead
}

// SetFPS updates the target FPS
func (rl *RenderLoop) SetFPS(fps int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Clamp FPS to reasonable values
	if fps <= 0 {
		fps = 30
	}
	if fps > 240 {
		fps = 240
	}

	rl.fps = fps
	rl.frameTime = time.Duration(1000/fps) * time.Millisecond
}

// FPS returns the current target FPS
func (rl *RenderLoop) FPS() int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return rl.fps
}

// FrameTime returns the current frame time duration
func (rl *RenderLoop) FrameTime() time.Duration {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return rl.frameTime
}

// IsRunning returns true if the render loop is currently running
func (rl *RenderLoop) IsRunning() bool {
	return rl.running.Load()
}

// RequestExit is a convenience function to exit the render loop
// by returning ErrExitRequested from the next render call
func RequestExit() error {
	return ErrExitRequested
}

// App manages the application lifecycle with rendering
type App struct {
	renderLoop *RenderLoop
	ctx        context.Context
	cancel     context.CancelFunc
	renderFn   RenderFn
	mu         sync.RWMutex
}

// NewApp creates a new app with the specified FPS and render function
func NewApp(fps int, renderFn RenderFn) *App {
	return &App{
		renderLoop: NewRenderLoop(fps),
		renderFn:   renderFn,
	}
}

// Start starts the app's render loop
func (a *App) Start() error {
	a.mu.Lock()
	if a.ctx != nil {
		a.mu.Unlock()
		return fmt.Errorf("app is already running")
	}
	a.ctx, a.cancel = context.WithCancel(context.Background())
	ctx := a.ctx
	a.mu.Unlock()

	defer func() {
		a.mu.Lock()
		a.ctx = nil
		a.cancel = nil
		a.mu.Unlock()
	}()

	return a.renderLoop.Start(ctx, a.renderFn)
}

// Stop stops the app
func (a *App) Stop() {
	a.mu.RLock()
	cancel := a.cancel
	a.mu.RUnlock()

	if cancel != nil {
		cancel()
	}
}

// IsRunning returns true if the app is running
func (a *App) IsRunning() bool {
	return a.renderLoop.IsRunning()
}

// SetFPS updates the target FPS
func (a *App) SetFPS(fps int) {
	a.renderLoop.SetFPS(fps)
}

// FPS returns the current target FPS
func (a *App) FPS() int {
	return a.renderLoop.FPS()
}

// RequestExit stops the app on the next frame
func (a *App) RequestExit() {
	// Replace the render function with one that returns ErrExitRequested
	a.mu.Lock()
	a.renderFn = func() error {
		return ErrExitRequested
	}
	a.mu.Unlock()
}
