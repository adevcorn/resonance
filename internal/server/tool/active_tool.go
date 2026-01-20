package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/adevcorn/ensemble/internal/server/capability"
	"github.com/adevcorn/ensemble/internal/server/skill"
)

// ActiveToolAction represents the type of action to perform
type ActiveToolAction string

const (
	SearchSkills ActiveToolAction = "search_skills"
	LoadSkill    ActiveToolAction = "load_skill"
	Execute      ActiveToolAction = "execute"
)

// ActiveToolInput is the input for the active_tool
type ActiveToolInput struct {
	Action     ActiveToolAction `json:"action"`
	Query      string           `json:"query,omitempty"`       // for search_skills
	MaxResults int              `json:"max_results,omitempty"` // for search_skills
	SkillName  string           `json:"skill_name,omitempty"`  // for load_skill
	Capability string           `json:"capability,omitempty"`  // for execute
	Parameters json.RawMessage  `json:"parameters,omitempty"`  // for execute
}

// ActiveTool is the unified tool for discovering and executing capabilities
type ActiveTool struct {
	skillRegistry      *skill.Registry
	capabilityRegistry *capability.Registry
}

// NewActiveTool creates a new active tool
func NewActiveTool(skills *skill.Registry, caps *capability.Registry) *ActiveTool {
	return &ActiveTool{
		skillRegistry:      skills,
		capabilityRegistry: caps,
	}
}

// Name returns the tool name
func (t *ActiveTool) Name() string {
	return "active_tool"
}

// Description returns the tool description
func (t *ActiveTool) Description() string {
	return `Universal tool for discovering and executing capabilities.

Actions:
1. search_skills - Find skills by query
   Example: {"action": "search_skills", "query": "read files", "max_results": 3}

2. load_skill - Get full skill details and instructions
   Example: {"action": "load_skill", "skill_name": "filesystem-operations"}

3. execute - Run a capability (after learning from a skill)
   Example: {"action": "execute", "capability": "read_file", "parameters": {"path": "go.mod"}}

Workflow: Search for skills → Load relevant skill → Learn how to use → Execute capability`
}

// Parameters returns the JSON schema for active_tool parameters
func (t *ActiveTool) Parameters() json.RawMessage {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"search_skills", "load_skill", "execute"},
				"description": "The action to perform: search_skills, load_skill, or execute",
			},
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Search query (for search_skills action)",
			},
			"max_results": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of search results (for search_skills action, default: 5)",
				"minimum":     1,
				"maximum":     10,
			},
			"skill_name": map[string]interface{}{
				"type":        "string",
				"description": "Name of the skill to load (for load_skill action)",
			},
			"capability": map[string]interface{}{
				"type":        "string",
				"description": "Name of the capability to execute (for execute action)",
			},
			"parameters": map[string]interface{}{
				"type":        "object",
				"description": "Parameters for the capability (for execute action)",
			},
		},
		"required": []string{"action"},
	}

	data, _ := json.Marshal(schema)
	return data
}

// Execute performs the active_tool action
func (t *ActiveTool) Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var req ActiveToolInput
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid active_tool input: %w", err)
	}

	switch req.Action {
	case SearchSkills:
		return t.handleSearch(req.Query, req.MaxResults)
	case LoadSkill:
		return t.handleLoad(req.SkillName)
	case Execute:
		return t.handleExecute(ctx, req.Capability, req.Parameters)
	default:
		return nil, fmt.Errorf("unknown action: %s (must be search_skills, load_skill, or execute)", req.Action)
	}
}

// handleSearch searches for skills matching the query
func (t *ActiveTool) handleSearch(query string, maxResults int) (json.RawMessage, error) {
	if query == "" {
		return nil, fmt.Errorf("query is required for search_skills action")
	}

	results := t.skillRegistry.Search(query, maxResults)

	data, err := json.Marshal(map[string]interface{}{
		"query":   query,
		"results": results,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search results: %w", err)
	}

	return data, nil
}

// handleLoad loads a skill and returns its full content
func (t *ActiveTool) handleLoad(skillName string) (json.RawMessage, error) {
	if skillName == "" {
		return nil, fmt.Errorf("skill_name is required for load_skill action")
	}

	sk, err := t.skillRegistry.GetSkill(skillName)
	if err != nil {
		return nil, fmt.Errorf("failed to load skill: %w", err)
	}

	// Build skill file info
	type SkillFileInfo struct {
		Name string `json:"name"`
		Path string `json:"path"`
	}

	scripts := make([]SkillFileInfo, 0, len(sk.Scripts))
	for name, path := range sk.Scripts {
		scripts = append(scripts, SkillFileInfo{Name: name, Path: path})
	}

	references := make([]SkillFileInfo, 0, len(sk.References))
	for name, path := range sk.References {
		references = append(references, SkillFileInfo{Name: name, Path: path})
	}

	assets := make([]SkillFileInfo, 0, len(sk.Assets))
	for name, path := range sk.Assets {
		assets = append(assets, SkillFileInfo{Name: name, Path: path})
	}

	output := map[string]interface{}{
		"skill_name":   sk.Metadata.Name,
		"description":  sk.Metadata.Description,
		"category":     sk.Metadata.Category,
		"capabilities": sk.Metadata.Capabilities,
		"instructions": sk.GetFullContent(),
		"scripts":      scripts,
		"references":   references,
		"assets":       assets,
		"metadata":     sk.Metadata.Metadata,
	}

	data, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal skill: %w", err)
	}

	return data, nil
}

// handleExecute executes a capability with the given parameters
func (t *ActiveTool) handleExecute(ctx context.Context, capabilityName string, params json.RawMessage) (json.RawMessage, error) {
	if capabilityName == "" {
		return nil, fmt.Errorf("capability is required for execute action")
	}

	// Get the capability
	cap, err := t.capabilityRegistry.Get(capabilityName)
	if err != nil {
		return nil, fmt.Errorf("capability %q not found. Use search_skills to discover available capabilities", capabilityName)
	}

	// Execute the capability
	result, err := cap.Execute(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to execute capability %q: %w", capabilityName, err)
	}

	return result, nil
}

// ExecutionLocation returns where this tool executes
// Note: The active_tool itself runs on the server, but it delegates to capabilities
// which may be client-side or server-side
func (t *ActiveTool) ExecutionLocation() protocol.ExecutionLocation {
	return protocol.ExecutionLocationServer
}
