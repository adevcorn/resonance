package skill

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader_LoadSkill(t *testing.T) {
	// Create temporary skill directory
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "test-skill")
	require.NoError(t, os.Mkdir(skillDir, 0755))

	// Create SKILL.md
	skillContent := `---
name: test-skill
description: A test skill for unit testing
license: MIT
compatibility: Works with any agent
metadata:
  version: "1.0.0"
  author: "Test Author"
---

# Test Skill

This is the body of the test skill.

## Instructions

1. Do something
2. Do something else
`
	skillFile := filepath.Join(skillDir, "SKILL.md")
	require.NoError(t, os.WriteFile(skillFile, []byte(skillContent), 0644))

	// Create scripts directory
	scriptsDir := filepath.Join(skillDir, "scripts")
	require.NoError(t, os.Mkdir(scriptsDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(scriptsDir, "test.sh"), []byte("#!/bin/bash\necho test"), 0755))

	// Load skill
	loader := NewLoader(tmpDir)
	skill, err := loader.LoadSkill(skillDir)
	require.NoError(t, err)
	require.NotNil(t, skill)

	// Verify metadata
	assert.Equal(t, "test-skill", skill.Metadata.Name)
	assert.Equal(t, "A test skill for unit testing", skill.Metadata.Description)
	assert.Equal(t, "MIT", skill.Metadata.License)
	assert.Equal(t, "Works with any agent", skill.Metadata.Compatibility)
	assert.Equal(t, "1.0.0", skill.Metadata.Metadata["version"])
	assert.Equal(t, "Test Author", skill.Metadata.Metadata["author"])

	// Verify body
	assert.Contains(t, skill.Body, "# Test Skill")
	assert.Contains(t, skill.Body, "## Instructions")

	// Verify scripts
	assert.Len(t, skill.Scripts, 1)
	assert.Contains(t, skill.Scripts, "test.sh")
}

func TestLoader_LoadSkill_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "missing-skill")
	require.NoError(t, os.Mkdir(skillDir, 0755))

	loader := NewLoader(tmpDir)
	skill, err := loader.LoadSkill(skillDir)
	assert.Error(t, err)
	assert.Nil(t, skill)
	assert.Contains(t, err.Error(), "SKILL.md not found")
}

func TestLoader_LoadSkill_InvalidFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "invalid-skill")
	require.NoError(t, os.Mkdir(skillDir, 0755))

	// Create SKILL.md without frontmatter
	skillContent := `# Invalid Skill

No frontmatter here!
`
	skillFile := filepath.Join(skillDir, "SKILL.md")
	require.NoError(t, os.WriteFile(skillFile, []byte(skillContent), 0644))

	loader := NewLoader(tmpDir)
	skill, err := loader.LoadSkill(skillDir)
	assert.Error(t, err)
	assert.Nil(t, skill)
	assert.Contains(t, err.Error(), "must start with YAML frontmatter")
}

func TestLoader_LoadSkill_InvalidName(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "Invalid-Skill") // Capital letters not allowed
	require.NoError(t, os.Mkdir(skillDir, 0755))

	skillContent := `---
name: Invalid-Skill
description: This skill has an invalid name
---

Body content
`
	skillFile := filepath.Join(skillDir, "SKILL.md")
	require.NoError(t, os.WriteFile(skillFile, []byte(skillContent), 0644))

	loader := NewLoader(tmpDir)
	skill, err := loader.LoadSkill(skillDir)
	assert.Error(t, err)
	assert.Nil(t, skill)
	assert.Contains(t, err.Error(), "lowercase letters")
}

func TestLoader_LoadSkill_NameMismatch(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "dir-name")
	require.NoError(t, os.Mkdir(skillDir, 0755))

	skillContent := `---
name: different-name
description: Directory name doesn't match skill name
---

Body
`
	skillFile := filepath.Join(skillDir, "SKILL.md")
	require.NoError(t, os.WriteFile(skillFile, []byte(skillContent), 0644))

	loader := NewLoader(tmpDir)
	skill, err := loader.LoadSkill(skillDir)
	assert.Error(t, err)
	assert.Nil(t, skill)
	assert.Contains(t, err.Error(), "does not match")
}

func TestLoader_DiscoverSkills(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple skills
	createTestSkill(t, tmpDir, "skill-one")
	createTestSkill(t, tmpDir, "skill-two")

	// Create nested skill
	nestedDir := filepath.Join(tmpDir, "category")
	require.NoError(t, os.Mkdir(nestedDir, 0755))
	createTestSkill(t, nestedDir, "skill-three")

	loader := NewLoader(tmpDir)
	skills, err := loader.DiscoverSkills()
	require.NoError(t, err)
	assert.Len(t, skills, 3)

	// Verify skill names
	names := make(map[string]bool)
	for _, skill := range skills {
		names[skill.Metadata.Name] = true
	}
	assert.True(t, names["skill-one"])
	assert.True(t, names["skill-two"])
	assert.True(t, names["skill-three"])
}

func TestValidateSkillName(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
	}{
		{"valid-skill", true},
		{"skill123", true},
		{"my-skill-123", true},
		{"a", true},
		{"Invalid-Skill", false},  // uppercase
		{"-invalid", false},       // starts with hyphen
		{"invalid-", false},       // ends with hyphen
		{"invalid--skill", false}, // consecutive hyphens
		{"", false},               // empty
		{"skill_name", false},     // underscore not allowed
		{"skill.name", false},     // dot not allowed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := isValidSkillName(tt.name)
			assert.Equal(t, tt.valid, valid, "Expected '%s' to be valid=%v", tt.name, tt.valid)
		})
	}
}

func TestSkillMetadata_Validate(t *testing.T) {
	tests := []struct {
		name      string
		metadata  SkillMetadata
		wantError bool
		errorText string
	}{
		{
			name: "valid metadata",
			metadata: SkillMetadata{
				Name:        "test-skill",
				Description: "A test skill",
			},
			wantError: false,
		},
		{
			name: "missing name",
			metadata: SkillMetadata{
				Description: "A test skill",
			},
			wantError: true,
			errorText: "name is required",
		},
		{
			name: "missing description",
			metadata: SkillMetadata{
				Name: "test-skill",
			},
			wantError: true,
			errorText: "description is required",
		},
		{
			name: "name too long",
			metadata: SkillMetadata{
				Name:        "this-is-a-very-long-skill-name-that-exceeds-the-maximum-length-allowed-by-the-specification",
				Description: "A test skill",
			},
			wantError: true,
			errorText: "64 characters or less",
		},
		{
			name: "description too long",
			metadata: SkillMetadata{
				Name:        "test-skill",
				Description: string(make([]byte, 1025)), // 1025 characters
			},
			wantError: true,
			errorText: "1024 characters or less",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.metadata.Validate()
			if tt.wantError {
				assert.Error(t, err)
				if tt.errorText != "" {
					assert.Contains(t, err.Error(), tt.errorText)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function to create a test skill
func createTestSkill(t *testing.T, baseDir, name string) {
	skillDir := filepath.Join(baseDir, name)
	require.NoError(t, os.Mkdir(skillDir, 0755))

	skillContent := `---
name: ` + name + `
description: Test skill for ` + name + `
---

# ` + name + `

Test skill body.
`
	skillFile := filepath.Join(skillDir, "SKILL.md")
	require.NoError(t, os.WriteFile(skillFile, []byte(skillContent), 0644))
}
