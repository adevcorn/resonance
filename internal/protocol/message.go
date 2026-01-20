package protocol

import "time"

// MessageRole represents the role of the message sender
type MessageRole string

const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleSystem    MessageRole = "system"
	MessageRoleTool      MessageRole = "tool"
)

// Message represents a single message in a conversation
type Message struct {
	ID          string         `json:"id"`
	SessionID   string         `json:"session_id"`
	Role        MessageRole    `json:"role"`
	Agent       string         `json:"agent"` // Which agent sent this message
	Content     string         `json:"content"`
	ToolCalls   []ToolCall     `json:"tool_calls,omitempty"`
	ToolResults []ToolResult   `json:"tool_results,omitempty"`
	Timestamp   time.Time      `json:"timestamp"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	Tokens      *Usage         `json:"tokens,omitempty"`
}

// ConversationContext holds the shared context for all agents in a session
type ConversationContext struct {
	Messages    []Message    `json:"messages"`
	ActiveTeam  []string     `json:"active_team"`
	CurrentTask string       `json:"current_task"`
	ProjectInfo *ProjectInfo `json:"project_info,omitempty"`
}

// ProjectInfo contains information about the project being worked on
type ProjectInfo struct {
	Path      string            `json:"path"`
	GitBranch string            `json:"git_branch,omitempty"`
	GitRemote string            `json:"git_remote,omitempty"`
	Language  string            `json:"language,omitempty"`
	Framework string            `json:"framework,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// Usage contains token usage information from LLM providers
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}
