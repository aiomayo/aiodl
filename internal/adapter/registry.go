package adapter

import (
	"fmt"
	"sort"
	"sync"
)

type Registry struct {
	mu       sync.RWMutex
	adapters map[string]Adapter
}

func NewRegistry() *Registry {
	return &Registry{adapters: make(map[string]Adapter)}
}

func (r *Registry) Register(adapter Adapter) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	name := adapter.Name()
	if _, exists := r.adapters[name]; exists {
		return fmt.Errorf("adapter %q already registered", name)
	}
	r.adapters[name] = adapter
	return nil
}

func (r *Registry) Get(name string) (Adapter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	adapter, ok := r.adapters[name]
	return adapter, ok
}

func (r *Registry) Find(url string) (Adapter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, adapter := range r.adapters {
		if adapter.Matches(url) {
			return adapter, true
		}
	}
	return nil, false
}

func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.adapters))
	for name := range r.adapters {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

var DefaultRegistry = NewRegistry()

func Register(adapter Adapter) error {
	return DefaultRegistry.Register(adapter)
}

func Get(name string) (Adapter, bool) {
	return DefaultRegistry.Get(name)
}

func Find(url string) (Adapter, bool) {
	return DefaultRegistry.Find(url)
}

func List() []string {
	return DefaultRegistry.List()
}
