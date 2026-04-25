package focus

import (
	"fmt"
	"sync"
)

// FocusID uniquely identifies a focusable component
type FocusID string

// FocusManager manages focus state across components
type FocusManager struct {
	mu           sync.RWMutex
	focusedID    FocusID
	hasFocused   bool
	focusable    map[FocusID]bool
	active       map[FocusID]bool
	autoFocus    map[FocusID]bool
	focusOrder   []FocusID
	currentIndex int
	idCounter    int // For generating unique IDs
}

// NewFocusManager creates a new focus manager
func NewFocusManager() *FocusManager {
	return &FocusManager{
		focusable:  make(map[FocusID]bool),
		active:     make(map[FocusID]bool),
		autoFocus:  make(map[FocusID]bool),
		focusOrder: make([]FocusID, 0),
		idCounter:  0,
	}
}

// Register registers a component as focusable
func (fm *FocusManager) Register(id FocusID, autoFocus bool) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if !fm.focusable[id] {
		fm.focusable[id] = true
		fm.active[id] = true
		fm.focusOrder = append(fm.focusOrder, id)

		// Ink registers focusables from an effect, so auto-focus only applies on
		// first registration, not on every render of an existing component.
		if autoFocus && !fm.hasFocused {
			fm.focusedID = id
			fm.hasFocused = true
			fm.currentIndex = len(fm.focusOrder) - 1
		}
	}

	fm.autoFocus[id] = autoFocus
}

// Unregister removes a component from focus management
func (fm *FocusManager) Unregister(id FocusID) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	delete(fm.focusable, id)
	delete(fm.active, id)
	delete(fm.autoFocus, id)

	// Update focus order
	removedIndex := -1
	newOrder := make([]FocusID, 0, len(fm.focusOrder))
	for index, fid := range fm.focusOrder {
		if fid != id {
			newOrder = append(newOrder, fid)
		} else if removedIndex == -1 {
			removedIndex = index
		}
	}
	fm.focusOrder = newOrder

	// If we unregistered the focused component, clear focus
	if fm.hasFocused && fm.focusedID == id {
		fm.focusedID = ""
		fm.hasFocused = false
		fm.currentIndex = 0
		return
	}

	if removedIndex >= 0 && removedIndex <= fm.currentIndex && fm.currentIndex > 0 {
		fm.currentIndex--
	}
}

// Activate marks a registered component as focusable in keyboard navigation.
func (fm *FocusManager) Activate(id FocusID) bool {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if !fm.focusable[id] {
		return false
	}

	fm.active[id] = true
	return true
}

// Deactivate marks a registered component as inactive while keeping its position in the focus order.
func (fm *FocusManager) Deactivate(id FocusID) bool {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if !fm.focusable[id] {
		return false
	}

	fm.active[id] = false
	if fm.hasFocused && fm.focusedID == id {
		fm.focusedID = ""
		fm.hasFocused = false
		fm.currentIndex = 0
	}

	return true
}

// IsActive reports whether a registered component is currently active for focus navigation.
func (fm *FocusManager) IsActive(id FocusID) bool {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	return fm.active[id]
}

// Focus sets focus to the specified component if it is registered.
func (fm *FocusManager) Focus(id FocusID) bool {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	targetIndex := -1
	for index, fid := range fm.focusOrder {
		if fid != id {
			continue
		}

		targetIndex = index
		break
	}

	if targetIndex < 0 {
		return false
	}

	fm.focusedID = id
	fm.hasFocused = true
	fm.currentIndex = targetIndex
	return true
}

// Blur removes focus from the currently focused component
func (fm *FocusManager) Blur() {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.focusedID = ""
	fm.hasFocused = false
}

// IsFocused checks if the specified component is focused
func (fm *FocusManager) IsFocused(id FocusID) bool {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	return fm.hasFocused && fm.focusedID == id
}

// FocusedID returns the currently focused component ID
func (fm *FocusManager) FocusedID() FocusID {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	if !fm.hasFocused {
		return ""
	}

	return fm.focusedID
}

// ActiveID returns the currently focused component ID, including an explicit
// empty-string focus target, or nil when nothing is focused.
func (fm *FocusManager) ActiveID() *FocusID {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	if !fm.hasFocused {
		return nil
	}

	id := fm.focusedID
	return &id
}

// HasFocus reports whether any component is currently focused, including an
// explicitly empty-string focus target.
func (fm *FocusManager) HasFocus() bool {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	return fm.hasFocused
}

// FocusNext moves focus to the next component
func (fm *FocusManager) FocusNext() bool {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if len(fm.focusOrder) == 0 {
		return false
	}

	if !fm.hasFocused {
		for index, id := range fm.focusOrder {
			if fm.active[id] {
				fm.currentIndex = index
				fm.focusedID = id
				fm.hasFocused = true
				return true
			}
		}

		return false
	}

	currentIndex := fm.currentIndex
	if currentIndex < 0 || currentIndex >= len(fm.focusOrder) || fm.focusOrder[currentIndex] != fm.focusedID {
		currentIndex = -1
		for index, id := range fm.focusOrder {
			if id == fm.focusedID {
				currentIndex = index
				break
			}
		}
	}

	for offset := 1; offset <= len(fm.focusOrder); offset++ {
		index := (currentIndex + offset + len(fm.focusOrder)) % len(fm.focusOrder)
		id := fm.focusOrder[index]
		if !fm.active[id] {
			continue
		}

		fm.currentIndex = index
		fm.focusedID = id
		fm.hasFocused = true
		return true
	}

	return false
}

// FocusPrevious moves focus to the previous component
func (fm *FocusManager) FocusPrevious() bool {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if len(fm.focusOrder) == 0 {
		return false
	}

	if !fm.hasFocused {
		for index := len(fm.focusOrder) - 1; index >= 0; index-- {
			id := fm.focusOrder[index]
			if fm.active[id] {
				fm.currentIndex = index
				fm.focusedID = id
				fm.hasFocused = true
				return true
			}
		}

		return false
	}

	currentIndex := fm.currentIndex
	if currentIndex < 0 || currentIndex >= len(fm.focusOrder) || fm.focusOrder[currentIndex] != fm.focusedID {
		currentIndex = -1
		for index, id := range fm.focusOrder {
			if id == fm.focusedID {
				currentIndex = index
				break
			}
		}
	}

	for offset := 1; offset <= len(fm.focusOrder); offset++ {
		index := (currentIndex - offset + len(fm.focusOrder)*2) % len(fm.focusOrder)
		id := fm.focusOrder[index]
		if !fm.active[id] {
			continue
		}

		fm.currentIndex = index
		fm.focusedID = id
		fm.hasFocused = true
		return true
	}

	return false
}

// FocusableCount returns the number of focusable components
func (fm *FocusManager) FocusableCount() int {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	return len(fm.focusOrder)
}

// FocusOrder returns the focus order
func (fm *FocusManager) FocusOrder() []FocusID {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	order := make([]FocusID, len(fm.focusOrder))
	copy(order, fm.focusOrder)
	return order
}

// Global focus manager instance
var globalManager = NewFocusManager()

// Global returns the global focus manager
func Global() *FocusManager {
	return globalManager
}

// Focusable represents a component that can be focused
type Focusable interface {
	ID() FocusID
	SetFocus(bool)
}

// Component is a simple focusable component implementation
type Component struct {
	id      FocusID
	focused bool
}

// NewComponent creates a new focusable component
func NewComponent(id string) *Component {
	return &Component{
		id: FocusID(id),
	}
}

// ID returns the component's focus ID
func (c *Component) ID() FocusID {
	return c.id
}

// SetFocus sets the component's focus state
func (c *Component) SetFocus(focused bool) {
	c.focused = focused
}

// IsFocused returns whether the component is focused
func (c *Component) IsFocused() bool {
	return c.focused
}

// GenerateID generates a unique focus ID
func GenerateID(prefix string) FocusID {
	globalManager.mu.Lock()
	defer globalManager.mu.Unlock()
	globalManager.idCounter++
	return FocusID(fmt.Sprintf("%s-%d", prefix, globalManager.idCounter))
}
