package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adevcorn/ensemble/internal/protocol"
	"gopkg.in/yaml.v3"
)

// Loader loads agent definitions from YAML files
type Loader struct {
	agentsPath string
}

// NewLoader creates a new agent loader
func NewLoader(agentsPath string) *Loader {
	return &Loader{
		agentsPath: agentsPath,
	}
}

// LoadAll loads all agent definitions from the agents directory
func (l *Loader) LoadAll() ([]*protocol.AgentDefinition, error) {
	entries, err := os.ReadDir(l.agentsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read agents directory: %w", err)
	}

	var definitions []*protocol.AgentDefinition
	var errors []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process .yaml and .yml files
		if !strings.HasSuffix(entry.Name(), ".yaml") && !strings.HasSuffix(entry.Name(), ".yml") {
			continue
		}

		def, err := l.LoadOne(entry.Name())
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", entry.Name(), err))
			continue
		}

		definitions = append(definitions, def)
	}

	if len(errors) > 0 {
		return definitions, fmt.Errorf("failed to load some agents:\n%s", strings.Join(errors, "\n"))
	}

	return definitions, nil
}

// LoadOne loads a single agent definition by filename
func (l *Loader) LoadOne(filename string) (*protocol.AgentDefinition, error) {
	path := filepath.Join(l.agentsPath, filename)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var def protocol.AgentDefinition
	if err := yaml.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := l.Validate(&def); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &def, nil
}

// Validate validates an agent definition
func (l *Loader) Validate(def *protocol.AgentDefinition) error {
	// Check required fields
	if def.Name == "" {
		return fmt.Errorf("name is required")
	}

	if def.DisplayName == "" {
		return fmt.Errorf("display_name is required")
	}

	if def.SystemPrompt == "" {
		return fmt.Errorf("system_prompt is required")
	}

	// Validate model configuration
	if def.Model.Provider == "" {
		return fmt.Errorf("model.provider is required")
	}

	if def.Model.Name == "" {
		return fmt.Errorf("model.name is required")
	}

	if def.Model.Temperature < 0 || def.Model.Temperature > 2 {
		return fmt.Errorf("model.temperature must be between 0 and 2, got %f", def.Model.Temperature)
	}

	if def.Model.MaxTokens <= 0 {
		return fmt.Errorf("model.max_tokens must be greater than 0, got %d", def.Model.MaxTokens)
	}

	// Check capabilities
	if len(def.Capabilities) == 0 {
		return fmt.Errorf("at least one capability is required")
	}

	return nil
}
