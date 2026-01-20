package protocol

// CollaborateAction defines the type of collaboration message
type CollaborateAction string

const (
	CollaborateBroadcast CollaborateAction = "broadcast" // Message to all team members
	CollaborateDirect    CollaborateAction = "direct"    // Message to specific agent
	CollaborateHelp      CollaborateAction = "help"      // Request help from another agent
	CollaborateComplete  CollaborateAction = "complete"  // Signal task completion
)

// CollaborateInput represents input to the collaborate tool
type CollaborateInput struct {
	Action    CollaborateAction `json:"action"`
	Message   string            `json:"message"`
	ToAgent   string            `json:"to_agent,omitempty"`  // For direct/help actions
	Artifacts []string          `json:"artifacts,omitempty"` // File paths, code snippets, etc.
}

// CollaborateOutput represents the result of a collaborate action
type CollaborateOutput struct {
	Delivered  bool     `json:"delivered"`
	Recipients []string `json:"recipients"`
}
