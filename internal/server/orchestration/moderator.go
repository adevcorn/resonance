package orchestration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/adevcorn/ensemble/internal/server/agent"
	"github.com/adevcorn/ensemble/internal/server/provider"
)

// Moderator manages turn-taking in free-form collaboration
type Moderator struct {
	coordinator *agent.Agent
}

// NewModerator creates a new moderator
func NewModerator(coordinator *agent.Agent) *Moderator {
	return &Moderator{
		coordinator: coordinator,
	}
}

// SelectNextAgent decides which agent should speak next
// Based on conversation context and current task
func (m *Moderator) SelectNextAgent(
	ctx context.Context,
	team []string,
	messages []protocol.Message,
	task string,
	pendingCollaboration map[string][]string,
) (string, error) {
	// If no messages yet, coordinator starts
	if len(messages) == 0 {
		return "coordinator", nil
	}

	// Check if task is complete (look for "complete" action in recent messages)
	if m.isTaskComplete(messages) {
		return "complete", nil
	}

	// Priority 1: Agents with pending collaboration requests should respond next
	for _, agent := range team {
		if requesters, hasPending := pendingCollaboration[agent]; hasPending && len(requesters) > 0 {
			fmt.Fprintf(os.Stderr, "[DEBUG] Moderator: Prioritizing %s (has pending collaboration from %v)\n", agent, requesters)
			return agent, nil
		}
	}

	// Count agent contributions
	agentTurns := make(map[string]int)
	for _, msg := range messages {
		if msg.Agent != "" {
			agentTurns[msg.Agent]++
		}
	}

	// Get last few messages for context
	recentMessages := messages
	if len(messages) > 10 {
		recentMessages = messages[len(messages)-10:]
	}

	// Build context summary
	var contextParts []string
	for _, msg := range recentMessages {
		if msg.Role == protocol.MessageRoleAssistant {
			contextParts = append(contextParts,
				fmt.Sprintf("[%s]: %s", msg.Agent, truncate(msg.Content, 200)))
		}
	}
	context := strings.Join(contextParts, "\n")

	// Build prompt for coordinator to select next agent
	var agentList []string
	for _, agentName := range team {
		turns := agentTurns[agentName]
		agentList = append(agentList,
			fmt.Sprintf("- %s (contributed %d times)", agentName, turns))
	}

	prompt := fmt.Sprintf(`You are moderating a multi-agent collaboration. Based on the recent conversation, decide which agent should speak next.

Task: %s

Team members:
%s

Recent conversation:
%s

Who should speak next to make progress on this task? Consider:
1. What needs to happen next
2. Which agent is best suited for that
3. Who hasn't contributed recently
4. Whether the task might be complete

Reply with ONLY the agent name, or "complete" if the task is done.`,
		task,
		strings.Join(agentList, "\n"),
		context)

	// Call coordinator to decide
	req := &provider.CompletionRequest{
		Messages: []protocol.Message{
			{
				Role:    protocol.MessageRoleUser,
				Content: prompt,
			},
		},
	}

	resp, err := m.coordinator.Complete(ctx, req)
	if err != nil {
		return "", fmt.Errorf("moderator failed to select next agent: %w", err)
	}

	// Parse response - should be just the agent name
	nextAgent := strings.TrimSpace(strings.ToLower(resp.Content))

	// Validate that it's a valid team member or "complete"
	if nextAgent == "complete" {
		return "complete", nil
	}

	for _, agent := range team {
		if agent == nextAgent {
			return nextAgent, nil
		}
	}

	// If we got an invalid agent, default to coordinator
	return "coordinator", nil
}

// ShouldContinue determines if collaboration should continue
func (m *Moderator) ShouldContinue(messages []protocol.Message) bool {
	// Check if we've hit a reasonable message limit
	const maxMessages = 50
	if len(messages) >= maxMessages {
		return false
	}

	// Check for completion signals
	if m.isTaskComplete(messages) {
		return false
	}

	return true
}

// isTaskComplete checks if any agent has signaled completion
func (m *Moderator) isTaskComplete(messages []protocol.Message) bool {
	// Look at recent messages for completion signals
	checkCount := 5
	if len(messages) < checkCount {
		checkCount = len(messages)
	}

	for i := len(messages) - checkCount; i < len(messages); i++ {
		msg := messages[i]

		// Check tool calls for collaborate complete action
		for _, toolCall := range msg.ToolCalls {
			if toolCall.ToolName == "collaborate" {
				var input protocol.CollaborateInput
				if err := json.Unmarshal(toolCall.Arguments, &input); err == nil {
					if input.Action == protocol.CollaborateComplete {
						return true
					}
				}
			}
		}

		// Check content for completion keywords
		content := strings.ToLower(msg.Content)
		if strings.Contains(content, "task complete") ||
			strings.Contains(content, "task is complete") ||
			strings.Contains(content, "completed successfully") {
			return true
		}
	}

	return false
}

// truncate truncates a string to maxLen characters
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
