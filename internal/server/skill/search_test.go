package skill

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single word",
			input:    "test",
			expected: []string{"test"},
		},
		{
			name:     "multiple words",
			input:    "read file system",
			expected: []string{"read", "file", "system"},
		},
		{
			name:     "with punctuation",
			input:    "read, write, execute",
			expected: []string{"read,", "write,", "execute"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tokenize(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateScore(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []string
		skill    *Skill
		expected float64
	}{
		{
			name:   "exact name match",
			tokens: []string{"filesystem"},
			skill: &Skill{
				Metadata: SkillMetadata{
					Name:        "filesystem-operations",
					Description: "File system operations",
					Category:    "capability",
				},
			},
			expected: 3.0, // Name match weight
		},
		{
			name:   "description match",
			tokens: []string{"operations"},
			skill: &Skill{
				Metadata: SkillMetadata{
					Name:        "filesystem",
					Description: "File system operations",
					Category:    "capability",
				},
			},
			expected: 2.0, // Description match only
		},
		{
			name:   "capability match",
			tokens: []string{"read"},
			skill: &Skill{
				Metadata: SkillMetadata{
					Name:         "filesystem-operations",
					Description:  "File operations",
					Category:     "capability",
					Capabilities: []string{"read_file", "write_file"},
				},
			},
			expected: 2.5, // Capability match weight
		},
		{
			name:   "category match",
			tokens: []string{"capability"},
			skill: &Skill{
				Metadata: SkillMetadata{
					Name:        "filesystem-operations",
					Description: "File operations",
					Category:    "capability",
				},
			},
			expected: 1.5, // Category match weight
		},
		{
			name:   "multiple token matches",
			tokens: []string{"file", "read"},
			skill: &Skill{
				Metadata: SkillMetadata{
					Name:         "filesystem-operations",
					Description:  "File system operations",
					Category:     "capability",
					Capabilities: []string{"read_file", "write_file"},
				},
			},
			// "file" token:
			//   - matches name "filesystem-operations" -> 3.0
			//   - matches description "File system" -> 2.0
			//   - matches capability "read_file" -> 2.5
			//   - matches capability "write_file" -> 2.5
			// "read" token:
			//   - matches capability "read_file" -> 2.5
			// Total: 3.0 + 2.0 + 2.5 + 2.5 + 2.5 = 12.5
			expected: 12.5,
		},
		{
			name:   "no match",
			tokens: []string{"nonexistent"},
			skill: &Skill{
				Metadata: SkillMetadata{
					Name:        "filesystem-operations",
					Description: "File operations",
					Category:    "capability",
				},
			},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateScore(tt.tokens, tt.skill)
			// Allow for small floating point differences
			assert.InDelta(t, tt.expected, score, 0.1, "Score mismatch")
		})
	}
}

func TestSearch(t *testing.T) {
	// Create a registry with test skills
	reg := &Registry{
		skills: map[string]*Skill{
			"filesystem-operations": {
				Metadata: SkillMetadata{
					Name:         "filesystem-operations",
					Description:  "Read and write files on the filesystem",
					Category:     "capability",
					Capabilities: []string{"read_file", "write_file", "list_directory"},
				},
			},
			"shell-execution": {
				Metadata: SkillMetadata{
					Name:         "shell-execution",
					Description:  "Execute shell commands",
					Category:     "capability",
					Capabilities: []string{"execute_command"},
				},
			},
			"web-access": {
				Metadata: SkillMetadata{
					Name:         "web-access",
					Description:  "Fetch content from the web",
					Category:     "capability",
					Capabilities: []string{"fetch_url", "web_search"},
				},
			},
			"git-workflow": {
				Metadata: SkillMetadata{
					Name:        "git-workflow",
					Description: "Git version control workflow",
					Category:    "workflow",
				},
			},
		},
	}

	tests := []struct {
		name          string
		query         string
		maxResults    int
		expectedFirst string
		expectedCount int
		minScoreFirst float64
	}{
		{
			name:          "search for file operations",
			query:         "read file",
			maxResults:    5,
			expectedFirst: "filesystem-operations",
			expectedCount: 1,
			minScoreFirst: 7.0,
		},
		{
			name:          "search for shell",
			query:         "shell command",
			maxResults:    5,
			expectedFirst: "shell-execution",
			expectedCount: 1,
			minScoreFirst: 5.0,
		},
		{
			name:          "search for web",
			query:         "web",
			maxResults:    5,
			expectedFirst: "web-access",
			expectedCount: 1,
			minScoreFirst: 5.0,
		},
		{
			name:          "search for capability category",
			query:         "capability",
			maxResults:    5,
			expectedCount: 3, // filesystem, shell, web
			minScoreFirst: 1.5,
		},
		{
			name:          "limit results",
			query:         "capability",
			maxResults:    2,
			expectedCount: 2,
			minScoreFirst: 1.5,
		},
		{
			name:          "no matches",
			query:         "nonexistent",
			maxResults:    5,
			expectedCount: 0,
			minScoreFirst: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := reg.Search(tt.query, tt.maxResults)
			assert.Len(t, results, tt.expectedCount)

			if tt.expectedCount > 0 {
				if tt.expectedFirst != "" {
					assert.Equal(t, tt.expectedFirst, results[0].SkillName)
				}
				assert.GreaterOrEqual(t, results[0].RelevanceScore, tt.minScoreFirst)
			}
		})
	}
}

func TestSearchResultSorting(t *testing.T) {
	reg := &Registry{
		skills: map[string]*Skill{
			"exact-match": {
				Metadata: SkillMetadata{
					Name:        "file",
					Description: "Description",
					Category:    "test",
				},
			},
			"description-match": {
				Metadata: SkillMetadata{
					Name:        "other",
					Description: "This handles file operations",
					Category:    "test",
				},
			},
			"capability-match": {
				Metadata: SkillMetadata{
					Name:         "another",
					Description:  "Something else",
					Category:     "test",
					Capabilities: []string{"file_handler"},
				},
			},
		},
	}

	results := reg.Search("file", 10)
	assert.Len(t, results, 3)

	// exact-match should be first (name weight 3.0)
	assert.Equal(t, "exact-match", results[0].SkillName)
	assert.Greater(t, results[0].RelevanceScore, results[1].RelevanceScore)
}
