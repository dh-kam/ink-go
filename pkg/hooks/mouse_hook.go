package hooks

import (
	"sync"

	"github.com/dh-kam/goink.go/pkg/input"
)

// MouseCallback receives one decoded mouse event.
type MouseCallback func(input.MouseEvent)

// MouseManager owns registered UseMouse callbacks for one mounted app.
type MouseManager struct {
	mu        sync.RWMutex
	nextID    uint64
	callbacks map[uint64]MouseCallback
}

// NewMouseManager creates an isolated mouse callback registry.
func NewMouseManager() *MouseManager {
	return &MouseManager{
		callbacks: make(map[uint64]MouseCallback),
	}
}

var globalMouseManager = NewMouseManager()

// UseMouse registers cb to receive mouse events and returns a deregister closure.
func (manager *MouseManager) UseMouse(cb MouseCallback) func() {
	if manager == nil {
		return func() {}
	}
	if cb == nil {
		return func() {}
	}
	manager.mu.Lock()
	manager.nextID++
	id := manager.nextID
	manager.callbacks[id] = cb
	manager.mu.Unlock()

	var once sync.Once
	return func() {
		once.Do(func() {
			manager.mu.Lock()
			delete(manager.callbacks, id)
			manager.mu.Unlock()
		})
	}
}

// Dispatch fans out an event to every registered callback and reports whether
// at least one subscriber received it. Snapshot under RLock so user callbacks
// cannot deadlock by re-entering UseMouse.
func (manager *MouseManager) Dispatch(ev input.MouseEvent) bool {
	if manager == nil {
		return false
	}

	manager.mu.RLock()
	snapshot := make([]MouseCallback, 0, len(manager.callbacks))
	for _, cb := range manager.callbacks {
		snapshot = append(snapshot, cb)
	}
	manager.mu.RUnlock()

	if len(snapshot) == 0 {
		return false
	}

	for _, cb := range snapshot {
		cb(ev)
	}

	return true
}

// Count reports the number of registered callbacks.
func (manager *MouseManager) Count() int {
	if manager == nil {
		return 0
	}

	manager.mu.RLock()
	defer manager.mu.RUnlock()
	return len(manager.callbacks)
}

// Reset clears every registered callback.
func (manager *MouseManager) Reset() {
	if manager == nil {
		return
	}

	manager.mu.Lock()
	manager.callbacks = make(map[uint64]MouseCallback)
	manager.nextID = 0
	manager.mu.Unlock()
}

// UseMouse registers cb on the package-global compatibility manager.
func UseMouse(cb MouseCallback) func() {
	return globalMouseManager.UseMouse(cb)
}

// DispatchMouse dispatches through the package-global compatibility manager.
func DispatchMouse(ev input.MouseEvent) bool {
	return globalMouseManager.Dispatch(ev)
}

// MouseHookCount reports the number of global callbacks (test helper).
func MouseHookCount() int {
	return globalMouseManager.Count()
}

// ResetMouseHooks clears the package-global compatibility manager.
func ResetMouseHooks() {
	globalMouseManager.Reset()
}
