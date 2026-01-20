package agent

import (
	"context"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/adevcorn/ensemble/internal/server/provider"
)

// SkillRegistry interface for getting skill metadata
type SkillRegistry interface {
	GetAvailableSkillsXML(agentName string) string
}

// Agent represents a runtime agent instance
type Agent struct {
	definition    *protocol.AgentDefinition
	provider      provider.Provider
	skillRegistry SkillRegistry
}

// NewAgent creates a new agent with the given definition and provider
func NewAgent(def *protocol.AgentDefinition, prov provider.Provider) *Agent {
	return &Agent{
		definition:    def,
		provider:      prov,
		skillRegistry: nil, // Set later via SetSkillRegistry
	}
}

// SetSkillRegistry sets the skill registry for this agent
func (a *Agent) SetSkillRegistry(registry SkillRegistry) {
	a.skillRegistry = registry
}

// Name returns the agent's internal name
func (a *Agent) Name() string {
	return a.definition.Name
}

// DisplayName returns the agent's display name
func (a *Agent) DisplayName() string {
	return a.definition.DisplayName
}

// Description returns the agent's description
func (a *Agent) Description() string {
	return a.definition.Description
}

// Capabilities returns the agent's capabilities
func (a *Agent) Capabilities() []string {
	return a.definition.Capabilities
}

// SystemPrompt returns the agent's system prompt with skills injected
func (a *Agent) SystemPrompt() string {
	prompt := a.definition.SystemPrompt

	// Inject skill metadata if available
	if a.skillRegistry != nil && len(a.definition.Skills) > 0 {
		skillsXML := a.skillRegistry.GetAvailableSkillsXML(a.definition.Name)
		if skillsXML != "" {
			prompt += "\n\n" + skillsXML + "\n\n"
			prompt += "You can activate any of these skills using the activate_skill tool when you need detailed guidance."
		}
	}

	return prompt
}

// HasTool checks if a tool is in the agent's allowed list
func (a *Agent) HasTool(toolName string) bool {
	for _, allowed := range a.definition.Tools.Allowed {
		if allowed == toolName {
			return true
		}
	}
	return false
}

// IsToolAllowed checks if a tool is allowed and not denied
func (a *Agent) IsToolAllowed(toolName string) bool {
	// Check if explicitly denied
	for _, denied := range a.definition.Tools.Denied {
		if denied == toolName {
			return false
		}
	}

	// Check if in allowed list
	return a.HasTool(toolName)
}

// Complete performs a non-streaming completion using the agent's provider
func (a *Agent) Complete(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
	// Set model from agent configuration if not specified
	if req.Model == "" {
		req.Model = a.definition.Model.Name
	}
	if req.Temperature == 0 {
		req.Temperature = a.definition.Model.Temperature
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = a.definition.Model.MaxTokens
	}

	return a.provider.Complete(ctx, req)
}

// Stream performs a streaming completion using the agent's provider
func (a *Agent) Stream(ctx context.Context, req *provider.CompletionRequest) (<-chan provider.StreamEvent, error) {
	// Set model from agent configuration if not specified
	if req.Model == "" {
		req.Model = a.definition.Model.Name
	}
	if req.Temperature == 0 {
		req.Temperature = a.definition.Model.Temperature
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = a.definition.Model.MaxTokens
	}

	return a.provider.Stream(ctx, req)
}

// Definition returns the agent's definition
func (a *Agent) Definition() *protocol.AgentDefinition {
	return a.definition
}

// Provider returns the agent's provider
func (a *Agent) Provider() provider.Provider {
	return a.provider
}
