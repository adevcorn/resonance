package capability

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
)

// Registry manages all available capabilities
type Registry struct {
	mu           sync.RWMutex
	capabilities map[string]Capability
}

// NewRegistry creates a new capability registry
func NewRegistry() *Registry {
	return &Registry{
		capabilities: make(map[string]Capability),
	}
}

// Register adds a capability to the registry
func (r *Registry) Register(cap Capability) error {
	if cap == nil {
		return fmt.Errorf("cannot register nil capability")
	}

	name := cap.Name()
	if name == "" {
		return fmt.Errorf("capability name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.capabilities[name]; exists {
		return fmt.Errorf("capability %q already registered", name)
	}

	r.capabilities[name] = cap
	return nil
}

// Get retrieves a capability by name
func (r *Registry) Get(name string) (Capability, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cap, ok := r.capabilities[name]
	if !ok {
		return nil, fmt.Errorf("capability %q not found", name)
	}
	return cap, nil
}

// Execute runs a capability with the given parameters
func (r *Registry) Execute(ctx context.Context, name string, params json.RawMessage) (json.RawMessage, error) {
	cap, err := r.Get(name)
	if err != nil {
		return nil, err
	}

	return cap.Execute(ctx, params)
}

// Has checks if a capability exists
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.capabilities[name]
	return ok
}

// List returns all registered capability names, sorted alphabetically
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.capabilities))
	for name := range r.capabilities {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

// Count returns the number of registered capabilities
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.capabilities)
}
