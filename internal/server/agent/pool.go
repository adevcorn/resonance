package agent

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/adevcorn/ensemble/internal/server/provider"
)

// Pool manages runtime agent instances
type Pool struct {
	mu       sync.RWMutex
	agents   map[string]*Agent
	registry *provider.Registry
}

// NewPool creates a new agent pool
func NewPool(registry *provider.Registry) *Pool {
	return &Pool{
		agents:   make(map[string]*Agent),
		registry: registry,
	}
}

// Load loads agents from definitions
func (p *Pool) Load(definitions []*protocol.AgentDefinition) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var errors []string

	for _, def := range definitions {
		// Get provider for this agent
		prov, err := p.registry.Get(def.Model.Provider)
		if err != nil {
			errors = append(errors, fmt.Sprintf("agent %s: %v", def.Name, err))
			continue
		}

		// Create agent instance
		agent := NewAgent(def, prov)
		p.agents[def.Name] = agent
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to load some agents: %v", errors)
	}

	return nil
}

// Get retrieves an agent by name
func (p *Pool) Get(name string) (*Agent, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	agent, ok := p.agents[name]
	if !ok {
		// Try to find a similar agent name
		suggestion := p.findSimilarAgent(name)
		if suggestion != "" {
			return nil, fmt.Errorf("agent %q not found - did you mean %q? Available agents: %v",
				name, suggestion, p.listUnsafe())
		}
		return nil, fmt.Errorf("agent %q not found - available agents: %v", name, p.listUnsafe())
	}
	return agent, nil
}

// findSimilarAgent suggests an agent name based on common aliases or fuzzy matching
// Must be called with read lock held
func (p *Pool) findSimilarAgent(name string) string {
	nameLower := strings.ToLower(name)

	// Common aliases/mappings
	aliases := map[string]string{
		"documentation":    "writer",
		"docs":             "writer",
		"doc":              "writer",
		"technical-writer": "writer",
		"dev":              "developer",
		"coder":            "developer",
		"programmer":       "developer",
		"ops":              "devops",
		"deploy":           "devops",
		"qa":               "tester",
		"test":             "tester",
		"testing":          "tester",
		"code-review":      "reviewer",
		"review":           "reviewer",
		"sec":              "security",
		"research":         "researcher",
		"architect":        "architect",
		"design":           "architect",
	}

	// Check aliases first
	if suggestion, ok := aliases[nameLower]; ok {
		// Verify the suggested agent actually exists
		if _, exists := p.agents[suggestion]; exists {
			return suggestion
		}
	}

	// Fuzzy matching: check if any agent name contains the search term
	// or if the search term contains an agent name
	for agentName := range p.agents {
		agentLower := strings.ToLower(agentName)
		if strings.Contains(nameLower, agentLower) || strings.Contains(agentLower, nameLower) {
			return agentName
		}
	}

	return ""
}

// listUnsafe returns all agent names without locking (must be called with lock held)
func (p *Pool) listUnsafe() []string {
	names := make([]string, 0, len(p.agents))
	for name := range p.agents {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// List returns all agent names, sorted alphabetically
func (p *Pool) List() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	names := make([]string, 0, len(p.agents))
	for name := range p.agents {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

// Has checks if agent exists
func (p *Pool) Has(name string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	_, ok := p.agents[name]
	return ok
}

// Reload reloads specific agent
func (p *Pool) Reload(def *protocol.AgentDefinition) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Get provider for this agent
	prov, err := p.registry.Get(def.Model.Provider)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	// Create new agent instance
	agent := NewAgent(def, prov)
	p.agents[def.Name] = agent

	return nil
}

// Remove removes an agent from the pool
func (p *Pool) Remove(name string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.agents[name]; !ok {
		return fmt.Errorf("agent %q not found", name)
	}

	delete(p.agents, name)
	return nil
}

// Count returns number of loaded agents
func (p *Pool) Count() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.agents)
}

// GetAll returns all agents (useful for iteration)
func (p *Pool) GetAll() []*Agent {
	p.mu.RLock()
	defer p.mu.RUnlock()

	agents := make([]*Agent, 0, len(p.agents))
	for _, agent := range p.agents {
		agents = append(agents, agent)
	}
	return agents
}
