package skill

import (
	"fmt"
	"strings"
)

// SkillMetadata represents the YAML frontmatter of a SKILL.md file
type SkillMetadata struct {
	Name          string            `yaml:"name"`
	Description   string            `yaml:"description"`
	Category      string            `yaml:"category,omitempty"`     // "capability" or "workflow"
	Capabilities  []string          `yaml:"capabilities,omitempty"` // List of capability names
	License       string            `yaml:"license,omitempty"`
	Compatibility string            `yaml:"compatibility,omitempty"`
	Metadata      map[string]string `yaml:"metadata,omitempty"`
	AllowedTools  string            `yaml:"allowed-tools,omitempty"`
	Path          string            `yaml:"-"` // Absolute path to skill directory
}

// Validate validates the skill metadata according to AgentSkills.io spec
func (m *SkillMetadata) Validate() error {
	// Name validation
	if m.Name == "" {
		return fmt.Errorf("skill name is required")
	}
	if len(m.Name) > 64 {
		return fmt.Errorf("skill name must be 64 characters or less, got %d", len(m.Name))
	}
	if !isValidSkillName(m.Name) {
		return fmt.Errorf("skill name must contain only lowercase letters, numbers, and hyphens, and cannot start/end with hyphen")
	}

	// Description validation
	if m.Description == "" {
		return fmt.Errorf("skill description is required")
	}
	if len(m.Description) > 1024 {
		return fmt.Errorf("skill description must be 1024 characters or less, got %d", len(m.Description))
	}

	// Compatibility validation (optional)
	if m.Compatibility != "" && len(m.Compatibility) > 500 {
		return fmt.Errorf("skill compatibility must be 500 characters or less, got %d", len(m.Compatibility))
	}

	return nil
}

// isValidSkillName checks if skill name follows the spec:
// - 1-64 characters
// - Only lowercase letters, numbers, and hyphens
// - Cannot start or end with hyphen
// - Cannot contain consecutive hyphens
func isValidSkillName(name string) bool {
	if len(name) == 0 || len(name) > 64 {
		return false
	}
	if name[0] == '-' || name[len(name)-1] == '-' {
		return false
	}
	if strings.Contains(name, "--") {
		return false
	}

	for _, ch := range name {
		if !((ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-') {
			return false
		}
	}
	return true
}

// Skill represents a complete skill with metadata, instructions, and resources
type Skill struct {
	Metadata   SkillMetadata
	Body       string            // Markdown content after frontmatter
	Scripts    map[string]string // Script name -> absolute path
	References map[string]string // Reference name -> absolute path
	Assets     map[string]string // Asset name -> absolute path
}

// GetFullContent returns the complete skill instructions including body
func (s *Skill) GetFullContent() string {
	return s.Body
}

// GetScriptPath returns the absolute path to a script by name
func (s *Skill) GetScriptPath(name string) (string, bool) {
	path, ok := s.Scripts[name]
	return path, ok
}

// GetReferencePath returns the absolute path to a reference file by name
func (s *Skill) GetReferencePath(name string) (string, bool) {
	path, ok := s.References[name]
	return path, ok
}

// GetAssetPath returns the absolute path to an asset by name
func (s *Skill) GetAssetPath(name string) (string, bool) {
	path, ok := s.Assets[name]
	return path, ok
}

// SkillActivation represents an agent activating a skill
type SkillActivation struct {
	SkillName string
	AgentName string
	SessionID string
	Timestamp string
}
