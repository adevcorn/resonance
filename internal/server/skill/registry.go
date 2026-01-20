package skill

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
)

// Registry manages all available skills and provides discovery
type Registry struct {
	skills      map[string]*Skill   // skill name -> skill
	agentSkills map[string][]string // agent name -> skill names
	loader      *Loader
	watcher     *fsnotify.Watcher
	mu          sync.RWMutex
	skillsDir   string
}

// NewRegistry creates a new skill registry
func NewRegistry(skillsDir string) (*Registry, error) {
	absPath, err := filepath.Abs(skillsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve skills directory: %w", err)
	}

	loader := NewLoader(absPath)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	r := &Registry{
		skills:      make(map[string]*Skill),
		agentSkills: make(map[string][]string),
		loader:      loader,
		watcher:     watcher,
		skillsDir:   absPath,
	}

	// Initial load
	if err := r.loadAllSkills(); err != nil {
		return nil, fmt.Errorf("failed to load skills: %w", err)
	}

	// Start watching for changes
	if err := r.watcher.Add(absPath); err != nil {
		return nil, fmt.Errorf("failed to watch skills directory: %w", err)
	}

	go r.watchForChanges()

	return r, nil
}

// loadAllSkills loads all skills from the skills directory
func (r *Registry) loadAllSkills() error {
	skills, err := r.loader.DiscoverSkills()
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Clear existing skills
	r.skills = make(map[string]*Skill)

	// Register discovered skills
	for _, skill := range skills {
		r.skills[skill.Metadata.Name] = skill
		log.Info().
			Str("skill", skill.Metadata.Name).
			Str("path", skill.Metadata.Path).
			Msg("Loaded skill")
	}

	log.Info().Int("count", len(skills)).Msg("Skills loaded")
	return nil
}

// watchForChanges watches for changes to skill files and reloads
func (r *Registry) watchForChanges() {
	debounce := time.NewTimer(500 * time.Millisecond)
	debounce.Stop()

	for {
		select {
		case event, ok := <-r.watcher.Events:
			if !ok {
				return
			}

			// Only reload on create, write, or remove events
			if event.Op&fsnotify.Write == fsnotify.Write ||
				event.Op&fsnotify.Create == fsnotify.Create ||
				event.Op&fsnotify.Remove == fsnotify.Remove {

				// Debounce: wait for more events
				debounce.Reset(500 * time.Millisecond)
			}

		case err, ok := <-r.watcher.Errors:
			if !ok {
				return
			}
			log.Error().Err(err).Msg("File watcher error")

		case <-debounce.C:
			// Reload all skills after debounce period
			log.Info().Msg("Skills directory changed, reloading...")
			if err := r.loadAllSkills(); err != nil {
				log.Error().Err(err).Msg("Failed to reload skills")
			}
		}
	}
}

// GetSkill retrieves a skill by name
func (r *Registry) GetSkill(name string) (*Skill, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skill, ok := r.skills[name]
	if !ok {
		return nil, fmt.Errorf("skill not found: %s", name)
	}

	return skill, nil
}

// ListSkills returns all available skills
func (r *Registry) ListSkills() []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skills := make([]*Skill, 0, len(r.skills))
	for _, skill := range r.skills {
		skills = append(skills, skill)
	}

	return skills
}

// ListSkillNames returns all skill names
func (r *Registry) ListSkillNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.skills))
	for name := range r.skills {
		names = append(names, name)
	}

	return names
}

// RegisterAgentSkills associates skills with an agent
func (r *Registry) RegisterAgentSkills(agentName string, skillNames []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Validate that all skills exist
	for _, skillName := range skillNames {
		if _, ok := r.skills[skillName]; !ok {
			return fmt.Errorf("skill not found: %s", skillName)
		}
	}

	r.agentSkills[agentName] = skillNames
	log.Info().
		Str("agent", agentName).
		Strs("skills", skillNames).
		Msg("Registered agent skills")

	return nil
}

// GetAgentSkills returns all skills for an agent
func (r *Registry) GetAgentSkills(agentName string) []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skillNames, ok := r.agentSkills[agentName]
	if !ok {
		return nil
	}

	skills := make([]*Skill, 0, len(skillNames))
	for _, name := range skillNames {
		if skill, ok := r.skills[name]; ok {
			skills = append(skills, skill)
		}
	}

	return skills
}

// GetAvailableSkillsXML generates XML for agent system prompts
func (r *Registry) GetAvailableSkillsXML(agentName string) string {
	skills := r.GetAgentSkills(agentName)
	if len(skills) == 0 {
		return ""
	}

	xml := "<available_skills>\n"
	for _, skill := range skills {
		xml += fmt.Sprintf(`  <skill>
    <name>%s</name>
    <description>%s</description>
    <location>%s/SKILL.md</location>
  </skill>
`, skill.Metadata.Name, skill.Metadata.Description, skill.Metadata.Path)
	}
	xml += "</available_skills>"

	return xml
}

// SearchResult represents a skill search result
type SearchResult struct {
	SkillName      string   `json:"skill_name"`
	Description    string   `json:"description"`
	Category       string   `json:"category"`
	Capabilities   []string `json:"capabilities,omitempty"`
	RelevanceScore float64  `json:"relevance_score"`
}

// Search finds skills matching the query string
func (r *Registry) Search(query string, maxResults int) []SearchResult {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if maxResults == 0 {
		maxResults = 5
	}

	queryLower := strings.ToLower(query)
	tokens := tokenize(queryLower)

	var results []SearchResult
	for name, skill := range r.skills {
		score := calculateScore(tokens, skill)
		if score > 0 {
			results = append(results, SearchResult{
				SkillName:      name,
				Description:    skill.Metadata.Description,
				Category:       skill.Metadata.Category,
				Capabilities:   skill.Metadata.Capabilities,
				RelevanceScore: score,
			})
		}
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].RelevanceScore > results[j].RelevanceScore
	})

	if len(results) > maxResults {
		results = results[:maxResults]
	}

	return results
}

// tokenize splits a string into search tokens
func tokenize(s string) []string {
	return strings.Fields(s)
}

// calculateScore computes relevance score for a skill given search tokens
func calculateScore(tokens []string, skill *Skill) float64 {
	score := 0.0

	// Check skill name (weight: 3.0)
	nameLower := strings.ToLower(skill.Metadata.Name)
	for _, token := range tokens {
		if strings.Contains(nameLower, token) {
			score += 3.0
		}
	}

	// Check description (weight: 2.0)
	descLower := strings.ToLower(skill.Metadata.Description)
	for _, token := range tokens {
		if strings.Contains(descLower, token) {
			score += 2.0
		}
	}

	// Check capabilities (weight: 2.5)
	for _, cap := range skill.Metadata.Capabilities {
		capLower := strings.ToLower(cap)
		for _, token := range tokens {
			if strings.Contains(capLower, token) {
				score += 2.5
			}
		}
	}

	// Check category (weight: 1.5)
	if skill.Metadata.Category != "" {
		categoryLower := strings.ToLower(skill.Metadata.Category)
		for _, token := range tokens {
			if strings.Contains(categoryLower, token) {
				score += 1.5
			}
		}
	}

	return score
}

// Close stops the file watcher
func (r *Registry) Close() error {
	return r.watcher.Close()
}
