package skill

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Loader handles loading and parsing skills from the filesystem
type Loader struct {
	baseDir string
}

// NewLoader creates a new skill loader
func NewLoader(baseDir string) *Loader {
	return &Loader{
		baseDir: baseDir,
	}
}

// LoadSkill loads a single skill from a directory
func (l *Loader) LoadSkill(skillPath string) (*Skill, error) {
	// Ensure path is absolute
	absPath, err := filepath.Abs(skillPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve skill path: %w", err)
	}

	// Check if SKILL.md exists
	skillFile := filepath.Join(absPath, "SKILL.md")
	if _, err := os.Stat(skillFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("SKILL.md not found in %s", absPath)
	}

	// Parse SKILL.md
	metadata, body, err := l.parseSkillFile(skillFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SKILL.md: %w", err)
	}

	// Set the path
	metadata.Path = absPath

	// Validate metadata
	if err := metadata.Validate(); err != nil {
		return nil, fmt.Errorf("invalid skill metadata: %w", err)
	}

	// Verify directory name matches skill name
	dirName := filepath.Base(absPath)
	if dirName != metadata.Name {
		return nil, fmt.Errorf("directory name '%s' does not match skill name '%s'", dirName, metadata.Name)
	}

	// Discover scripts, references, and assets
	scripts, err := l.discoverFiles(filepath.Join(absPath, "scripts"))
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to discover scripts: %w", err)
	}

	references, err := l.discoverFiles(filepath.Join(absPath, "references"))
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to discover references: %w", err)
	}

	assets, err := l.discoverFiles(filepath.Join(absPath, "assets"))
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to discover assets: %w", err)
	}

	return &Skill{
		Metadata:   metadata,
		Body:       body,
		Scripts:    scripts,
		References: references,
		Assets:     assets,
	}, nil
}

// parseSkillFile parses a SKILL.md file and extracts frontmatter and body
func (l *Loader) parseSkillFile(filePath string) (SkillMetadata, string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return SkillMetadata{}, "", fmt.Errorf("failed to read file: %w", err)
	}

	// Check for frontmatter delimiter
	if !bytes.HasPrefix(content, []byte("---\n")) && !bytes.HasPrefix(content, []byte("---\r\n")) {
		return SkillMetadata{}, "", fmt.Errorf("SKILL.md must start with YAML frontmatter (---)")
	}

	// Split into frontmatter and body
	scanner := bufio.NewScanner(bytes.NewReader(content))
	var frontmatterLines []string
	var bodyLines []string
	inFrontmatter := false
	frontmatterClosed := false

	for scanner.Scan() {
		line := scanner.Text()

		if !inFrontmatter && strings.TrimSpace(line) == "---" {
			// First --- found
			inFrontmatter = true
			continue
		}

		if inFrontmatter && strings.TrimSpace(line) == "---" {
			// Second --- found, frontmatter is complete
			frontmatterClosed = true
			inFrontmatter = false
			continue
		}

		if inFrontmatter {
			frontmatterLines = append(frontmatterLines, line)
		} else if frontmatterClosed {
			bodyLines = append(bodyLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return SkillMetadata{}, "", fmt.Errorf("failed to scan file: %w", err)
	}

	if !frontmatterClosed {
		return SkillMetadata{}, "", fmt.Errorf("frontmatter not properly closed with ---")
	}

	// Parse YAML frontmatter
	frontmatterYAML := strings.Join(frontmatterLines, "\n")
	var metadata SkillMetadata
	if err := yaml.Unmarshal([]byte(frontmatterYAML), &metadata); err != nil {
		return SkillMetadata{}, "", fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}

	// Join body lines
	body := strings.Join(bodyLines, "\n")

	return metadata, strings.TrimSpace(body), nil
}

// discoverFiles discovers all files in a directory (non-recursive)
func (l *Loader) discoverFiles(dir string) (map[string]string, error) {
	files := make(map[string]string)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			name := entry.Name()
			absPath := filepath.Join(dir, name)
			files[name] = absPath
		}
	}

	return files, nil
}

// DiscoverSkills discovers all skills in the base directory
// Searches for directories containing SKILL.md files
func (l *Loader) DiscoverSkills() ([]*Skill, error) {
	var skills []*Skill

	// Walk the base directory
	err := filepath.Walk(l.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Look for SKILL.md files
		if !info.IsDir() && info.Name() == "SKILL.md" {
			skillDir := filepath.Dir(path)
			skill, err := l.LoadSkill(skillDir)
			if err != nil {
				// Log error but continue discovering other skills
				fmt.Printf("Warning: failed to load skill from %s: %v\n", skillDir, err)
				return nil
			}
			skills = append(skills, skill)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to discover skills: %w", err)
	}

	return skills, nil
}
