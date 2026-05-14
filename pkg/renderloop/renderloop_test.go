package renderloop_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dh-kam/ink-go/pkg/renderloop"
)

// TestNewRenderLoop tests render loop creation
func TestNewRenderLoop(t *testing.T) {
	rl := renderloop.NewRenderLoop(60)

	if rl == nil {
		t.Fatal("Expected non-nil render loop")
	}
}

// TestRenderLoopStartStop tests basic start/stop
func TestRenderLoopStartStop(t *testing.T) {
	rl := renderloop.NewRenderLoop(60)
	ctx, cancel := context.WithCancel(context.Background())

	// Start should not block
	go rl.Start(ctx, func() error {
		return nil
	})

	// Stop after a short delay
	time.Sleep(50 * time.Millisecond)
	cancel()

	// Should stop gracefully
	time.Sleep(100 * time.Millisecond)
}

// TestRenderLoopFPS tests FPS throttling
func TestRenderLoopFPS(t *testing.T) {
	rl := renderloop.NewRenderLoop(10) // 10 FPS
	ctx, cancel := context.WithCancel(context.Background())

	var renderCount int32
	renderFn := func() error {
		atomic.AddInt32(&renderCount, 1)
		return nil
	}

	go rl.Start(ctx, renderFn)

	// Wait for ~500ms, should get ~5 renders
	time.Sleep(500 * time.Millisecond)
	cancel()

	count := atomic.LoadInt32(&renderCount)
	// Should be around 5, allow some tolerance
	if count < 3 || count > 8 {
		t.Errorf("Expected ~5 renders, got %d", count)
	}
}

// TestRenderLoopExitOnError tests exit on render error
func TestRenderLoopExitOnError(t *testing.T) {
	rl := renderloop.NewRenderLoop(60)
	ctx := context.Background()

	callCount := 0
	renderFn := func() error {
		callCount++
		if callCount >= 3 {
			return renderloop.ErrExitRequested
		}
		return nil
	}

	rl.Start(ctx, renderFn)

	if callCount != 3 {
		t.Errorf("Expected 3 renders before exit, got %d", callCount)
	}
}

// TestRenderLoopSetFPS tests changing FPS
func TestRenderLoopSetFPS(t *testing.T) {
	rl := renderloop.NewRenderLoop(30)

	rl.SetFPS(60)
	if rl.FPS() != 60 {
		t.Errorf("Expected FPS to be 60, got %d", rl.FPS())
	}
}

// TestRenderLoopIsRunning tests running state
func TestRenderLoopIsRunning(t *testing.T) {
	rl := renderloop.NewRenderLoop(60)

	if rl.IsRunning() {
		t.Error("Expected render loop to not be running initially")
	}
}

// TestRenderFrameTime tests frame time calculation
func TestRenderFrameTime(t *testing.T) {
	rl := renderloop.NewRenderLoop(60)

	// 60 FPS = ~16.67ms per frame
	expected := time.Duration(1000/60) * time.Millisecond
	actual := rl.FrameTime()

	// Allow 1ms tolerance
	if actual < expected-1*time.Millisecond || actual > expected+1*time.Millisecond {
		t.Errorf("Expected frame time ~%v, got %v", expected, actual)
	}
}

// TestRenderLoopWithCancel tests context cancellation
func TestRenderLoopWithCancel(t *testing.T) {
	rl := renderloop.NewRenderLoop(60)
	ctx, cancel := context.WithCancel(context.Background())

	var renderCount int32
	renderFn := func() error {
		atomic.AddInt32(&renderCount, 1)
		return nil
	}

	go rl.Start(ctx, renderFn)

	// Cancel after 100ms
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Wait for loop to stop
	time.Sleep(50 * time.Millisecond)

	count := atomic.LoadInt32(&renderCount)
	if count == 0 {
		t.Error("Expected at least one render")
	}
}

// TestErrExitRequested tests the exit error
func TestErrExitRequested(t *testing.T) {
	err := renderloop.ErrExitRequested
	if err == nil {
		t.Error("Expected ErrExitRequested to be non-nil")
	}
	if err.Error() == "" {
		t.Error("Expected error message")
	}
}

// TestRenderLoopZeroFPS tests zero FPS handling
func TestRenderLoopZeroFPS(t *testing.T) {
	rl := renderloop.NewRenderLoop(0)

	// Should default to a reasonable FPS
	if rl.FPS() <= 0 {
		t.Errorf("Expected positive FPS for zero input, got %d", rl.FPS())
	}
}

// TestRenderLoopNegativeFPS tests negative FPS handling
func TestRenderLoopNegativeFPS(t *testing.T) {
	rl := renderloop.NewRenderLoop(-10)

	// Should default to a reasonable FPS
	if rl.FPS() <= 0 {
		t.Errorf("Expected positive FPS for negative input, got %d", rl.FPS())
	}
}

// TestNewApp tests app creation
func TestNewApp(t *testing.T) {
	renderFn := func() error { return nil }
	app := renderloop.NewApp(60, renderFn)

	if app == nil {
		t.Fatal("Expected non-nil app")
	}

	if app.FPS() != 60 {
		t.Errorf("Expected FPS 60, got %d", app.FPS())
	}
}

// TestAppStartStop tests app start and stop
func TestAppStartStop(t *testing.T) {
	renderCount := 0
	renderFn := func() error {
		renderCount++
		if renderCount >= 3 {
			return renderloop.ErrExitRequested
		}
		return nil
	}

	app := renderloop.NewApp(60, renderFn)

	done := make(chan error, 1)
	go func() {
		done <- app.Start()
	}()

	// Wait for app to complete (should be fast with only 3 frames)
	err := <-done
	if err != nil && err != renderloop.ErrExitRequested {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should have run exactly 3 times
	if renderCount != 3 {
		t.Errorf("Expected 3 renders, got %d", renderCount)
	}

	// Should not be running anymore
	if app.IsRunning() {
		t.Error("Expected app to not be running after completion")
	}
}

// TestAppStop tests stopping an app
func TestAppStop(t *testing.T) {
	renderFn := func() error {
		time.Sleep(10 * time.Millisecond)
		return nil
	}

	app := renderloop.NewApp(60, renderFn)

	done := make(chan error, 1)
	go func() {
		done <- app.Start()
	}()

	// Wait for app to start
	time.Sleep(50 * time.Millisecond)

	// Stop the app
	app.Stop()

	// Should exit with context.Canceled
	err := <-done
	if err != nil && err != context.Canceled && err != context.DeadlineExceeded {
		t.Logf("App stopped with: %v", err)
	}
}

// TestAppSetFPS tests changing FPS on app
func TestAppSetFPS(t *testing.T) {
	renderFn := func() error { return nil }
	app := renderloop.NewApp(30, renderFn)

	app.SetFPS(60)
	if app.FPS() != 60 {
		t.Errorf("Expected FPS 60, got %d", app.FPS())
	}
}

// TestAppRequestExit tests requesting exit
func TestAppRequestExit(t *testing.T) {
	exitTriggered := false
	renderFn := func() error {
		if exitTriggered {
			return renderloop.ErrExitRequested
		}
		return nil
	}

	app := renderloop.NewApp(60, renderFn)

	done := make(chan error, 1)
	go func() {
		done <- app.Start()
	}()

	// Wait for app to start
	time.Sleep(50 * time.Millisecond)

	// Request exit - this replaces render function
	app.RequestExit()
	exitTriggered = true

	// Should exit soon (within 100ms)
	select {
	case err := <-done:
		if err != renderloop.ErrExitRequested {
			t.Errorf("Expected ErrExitRequested, got %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("App did not exit in time")
	}
}

// TestAppDoubleStart tests starting an already running app
func TestAppDoubleStart(t *testing.T) {
	renderFn := func() error {
		time.Sleep(100 * time.Millisecond)
		return nil
	}

	app := renderloop.NewApp(60, renderFn)

	done := make(chan error, 1)
	go func() {
		done <- app.Start()
	}()

	// Wait for app to start
	time.Sleep(20 * time.Millisecond)

	// Try to start again - should fail
	err := app.Start()
	if err == nil {
		t.Error("Expected error when starting already running app")
	}

	// Cleanup
	app.Stop()
	<-done
}

// TestRenderLoopStopMethod tests the deprecated Stop method
func TestRenderLoopStopMethod(t *testing.T) {
	rl := renderloop.NewRenderLoop(60)

	// Should not panic
	rl.Stop()
}

// TestRenderLoopHighFPS tests high FPS clamping
func TestRenderLoopHighFPS(t *testing.T) {
	rl := renderloop.NewRenderLoop(500) // Above max of 240

	if rl.FPS() != 240 {
		t.Errorf("Expected FPS to be clamped to 240, got %d", rl.FPS())
	}
}

// TestRenderLoopSetFPSZero tests setting FPS to zero
func TestRenderLoopSetFPSZero(t *testing.T) {
	rl := renderloop.NewRenderLoop(60)
	rl.SetFPS(0)

	if rl.FPS() != 30 { // Should default to 30
		t.Errorf("Expected FPS 30 after setting to 0, got %d", rl.FPS())
	}
}

// TestRenderLoopSetFPSHigh tests setting FPS above max
func TestRenderLoopSetFPSHigh(t *testing.T) {
	rl := renderloop.NewRenderLoop(60)
	rl.SetFPS(500)

	if rl.FPS() != 240 { // Should be clamped to 240
		t.Errorf("Expected FPS 240 after setting to 500, got %d", rl.FPS())
	}
}
