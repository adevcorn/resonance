package protocol

// AgentDefinition represents a complete agent configuration
type AgentDefinition struct {
	Name         string      `yaml:"name" json:"name"`
	DisplayName  string      `yaml:"display_name" json:"display_name"`
	Description  string      `yaml:"description" json:"description"`
	SystemPrompt string      `yaml:"system_prompt" json:"system_prompt"`
	Capabilities []string    `yaml:"capabilities" json:"capabilities"`
	Model        ModelConfig `yaml:"model" json:"model"`
	Tools        ToolsConfig `yaml:"tools" json:"tools"`
}

// ModelConfig defines the LLM model configuration for an agent
type ModelConfig struct {
	Provider    string  `yaml:"provider" json:"provider"`
	Name        string  `yaml:"name" json:"name"`
	Temperature float64 `yaml:"temperature" json:"temperature"`
	MaxTokens   int     `yaml:"max_tokens" json:"max_tokens"`
}

// ToolsConfig defines which tools an agent can use
type ToolsConfig struct {
	Allowed []string `yaml:"allowed" json:"allowed"`
	Denied  []string `yaml:"denied" json:"denied"`
}
