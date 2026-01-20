package orchestration

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/adevcorn/ensemble/internal/server/agent"
	"github.com/adevcorn/ensemble/internal/server/provider"
)

// Synthesizer merges agent contributions into final response
type Synthesizer struct {
	coordinator *agent.Agent
}

// NewSynthesizer creates a new synthesizer
func NewSynthesizer(coordinator *agent.Agent) *Synthesizer {
	return &Synthesizer{
		coordinator: coordinator,
	}
}

// Synthesize creates a coherent final response from collaboration
// Returns:
// - Summary of what was accomplished
// - List of artifacts (files created/modified)
// - Error if synthesis fails
func (s *Synthesizer) Synthesize(
	ctx context.Context,
	task string,
	messages []protocol.Message,
) (string, []string, error) {
	// Extract agent contributions
	var contributions []string
	artifacts := make(map[string]bool)

	for _, msg := range messages {
		if msg.Role == protocol.MessageRoleAssistant && msg.Agent != "" {
			contributions = append(contributions,
				fmt.Sprintf("[%s]: %s", msg.Agent, msg.Content))
		}

		// Extract artifacts from collaborate tool calls
		for _, toolCall := range msg.ToolCalls {
			if toolCall.ToolName == "collaborate" {
				var input protocol.CollaborateInput
				if err := json.Unmarshal(toolCall.Arguments, &input); err == nil {
					for _, artifact := range input.Artifacts {
						artifacts[artifact] = true
					}
				}
			}
		}

		// Extract file paths from write_file tool calls
		for _, toolCall := range msg.ToolCalls {
			if toolCall.ToolName == "write_file" {
				var input map[string]interface{}
				if err := json.Unmarshal(toolCall.Arguments, &input); err == nil {
					if path, ok := input["path"].(string); ok {
						artifacts[path] = true
					}
				}
			}
		}
	}

	// Build synthesis prompt
	contributionText := strings.Join(contributions, "\n\n")
	if len(contributionText) > 4000 {
		// Truncate if too long
		contributionText = contributionText[:4000] + "\n...(truncated)"
	}

	prompt := fmt.Sprintf(`You are synthesizing the results of a multi-agent collaboration. The team worked together on this task:

Task: %s

Agent contributions:
%s

Please provide a concise summary (2-3 paragraphs) of:
1. What was accomplished
2. Key decisions made
3. Any remaining concerns or next steps

Be specific and factual. Focus on outcomes, not process.`,
		task,
		contributionText)

	// Call coordinator to synthesize
	req := &provider.CompletionRequest{
		Messages: []protocol.Message{
			{
				Role:    protocol.MessageRoleUser,
				Content: prompt,
			},
		},
	}

	resp, err := s.coordinator.Complete(ctx, req)
	if err != nil {
		return "", nil, fmt.Errorf("synthesis failed: %w", err)
	}

	// Convert artifacts map to slice
	artifactList := make([]string, 0, len(artifacts))
	for artifact := range artifacts {
		artifactList = append(artifactList, artifact)
	}

	return resp.Content, artifactList, nil
}
