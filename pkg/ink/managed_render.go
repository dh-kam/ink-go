package ink

import (
	"io"
	"os"
	"reflect"
	"sync"
)

type instanceRegistry struct {
	mu        sync.Mutex
	instances map[any]*Instance
}

var renderInstances = &instanceRegistry{
	instances: make(map[any]*Instance),
}

// RenderWithOptions renders a component through a managed session that is reused per stdout target.
func RenderWithOptions(component ComponentFunc, options RenderOptions) (*Instance, error) {
	normalized := normalizeRenderOptions(options)
	key, ok := renderTargetKey(normalized.Stdout)
	if !ok {
		return MountWithOptions(component, normalized)
	}

	existing := renderInstances.get(key)
	if existing != nil {
		existing.mu.Lock()
		if existing.unmounted {
			existing.mu.Unlock()
			renderInstances.delete(key, existing)
		} else {
			existing.applyOptionsLocked(normalized)
			err := existing.rerenderLocked(component)
			existing.mu.Unlock()
			return existing, err
		}
	}

	instance, err := MountWithOptions(component, normalized)
	if err != nil {
		return nil, err
	}

	instance.mu.Lock()
	if !instance.unmounted {
		instance.registry = renderInstances
		instance.registryKey = key
		renderInstances.set(key, instance)
	}
	instance.mu.Unlock()

	return instance, nil
}

func normalizeRenderOptions(options RenderOptions) RenderOptions {
	if options.Stdin == nil {
		options.Stdin = os.Stdin
	}

	if options.Stdout == nil {
		options.Stdout = os.Stdout
	}

	if options.Stderr == nil {
		options.Stderr = os.Stderr
	}

	return options
}

func renderTargetKey(writer io.Writer) (any, bool) {
	if writer == nil {
		return nil, false
	}

	writerType := reflect.TypeOf(writer)
	if writerType == nil || !writerType.Comparable() {
		return nil, false
	}

	return writer, true
}

func (registry *instanceRegistry) get(key any) *Instance {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	return registry.instances[key]
}

func (registry *instanceRegistry) set(key any, instance *Instance) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.instances[key] = instance
}

func (registry *instanceRegistry) delete(key any, instance *Instance) {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	current := registry.instances[key]
	if current == instance {
		delete(registry.instances, key)
	}
}
