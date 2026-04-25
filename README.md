# Goink

> React for CLI apps in Go - A Golang port of [Ink](https://github.com/vadimdemedes/ink)

Build interactive command-line applications with component-based architecture, just like React.

## Status

✅ **Phase 1 MVP Complete!**

## Features

- ✅ Component-based architecture
- ✅ Virtual DOM diffing
- ✅ Hooks (useState)
- ✅ Simple rendering engine
- ✅ Built-in components (Box, Text, Newline, Space)
- 🚧 Flexbox layouts (Phase 2)
- 🚧 ANSI styling and colors (Phase 2)
- 🚧 Input handling (Phase 2)

## Installation

```bash
go get github.com/dh-kam/goink.go
```

## Quick Start

### Hello World

```go
package main

import (
	"fmt"
	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

func App() *vdom.Node {
	return components.Box(nil,
		components.Text("Hello, Goink!"),
	)
}

func main() {
	app := ink.NewApp(App)
	fmt.Println(app.RenderOnce())
}
```

### Stateful Counter

```go
package main

import (
	"fmt"
	"github.com/dh-kam/goink.go/pkg/components"
	"github.com/dh-kam/goink.go/pkg/ink"
	"github.com/dh-kam/goink.go/pkg/vdom"
)

func Counter() *vdom.Node {
	count, setCount := ink.UseState(0)
	
	// Update state (normally triggered by user input)
	if count.(int) < 5 {
		setCount(count.(int) + 1)
	}
	
	return components.Box(nil,
		components.Text(fmt.Sprintf("Count: %d", count)),
	)
}

func main() {
	app := ink.NewApp(Counter)
	
	// Render multiple times to show updates
	for i := 0; i < 6; i++ {
		fmt.Println(app.RenderOnce())
	}
}
```

## Examples

Run the examples:

```bash
# Simple hello world
go run examples/hello/main.go

# Multi-line text example
go run examples/hello-advanced/main.go

# Auto-incrementing counter
go run examples/counter/main.go
```

## API Reference

### Core Functions

- `ink.NewApp(component ComponentFunc) *App` - Create a new app instance
- `app.RenderOnce() string` - Render the component once and return output
- `ink.UseState(initialValue interface{}) (value, setValue)` - State management hook

### Components

- `components.Box(props, children...)` - Container element
- `components.Text(args...)` - Text element with strings, props, and nested child nodes
- `components.Newline(count...)` - Newline character, optionally repeated
- `components.Space()` - Space character
- `components.Static(args...)` / `components.StaticItems(items, render, props...)` - Static output helpers

### Virtual DOM

- `vdom.CreateElement(type, props, children...)` - Create element node
- `vdom.CreateTextNode(text)` - Create text node

## Architecture

```
┌─────────────────────────────────────┐
│  Component (Go function)            │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  Hooks (useState, etc.)             │
│  - State management                 │
│  - Hook ordering                    │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  Virtual DOM                        │
│  - Node tree structure              │
│  - Props & children                 │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  Renderer                           │
│  - 2D buffer                        │
│  - Tree traversal                   │
└──────────────┬──────────────────────┘
               │
               ▼
           Terminal
```

## Development

This project follows **TDD (Test-Driven Development)** principles.

### Run Tests

```bash
# All tests
go test ./...

# With coverage
go test ./... -cover

# Specific package
go test ./pkg/vdom -v
```

### Test Coverage

```
Package                Coverage
─────────────────────────────────
internal/buffer        89.7%
internal/renderer      84.6%
pkg/components         87.5%
pkg/hooks             100.0%
pkg/ink                57.1%
pkg/vdom               79.2%
─────────────────────────────────
Average               ~83.0%
```

## Roadmap

### ✅ Phase 1: MVP (Complete)
- Virtual DOM
- Basic rendering
- useState hook
- Helper components
- Examples

### 🚧 Phase 2: Core Features (Next)
- Yoga layout engine (Flexbox)
- ANSI colors and styling
- Border rendering
- Input handling (useInput)
- Focus management

### 📋 Phase 3: Advanced Features
- Context API
- More hooks (useEffect, useMemo, etc.)
- Static component
- Error boundaries
- DevTools support

### 🎯 Phase 4: Production Ready
- Performance optimization
- Comprehensive documentation
- Migration guide from Ink (TS)
- Cross-platform testing

## Contributing

Contributions are welcome! Please ensure all tests pass and coverage remains high.

## License

MIT

---

Inspired by [Ink](https://github.com/vadimdemedes/ink) by Vadim Demedes
