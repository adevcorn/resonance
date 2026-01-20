package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// ServerConfig represents the server configuration
type ServerConfig struct {
	Server    ServerSettings    `mapstructure:"server"`
	Storage   StorageSettings   `mapstructure:"storage"`
	Agents    AgentsSettings    `mapstructure:"agents"`
	Providers ProvidersSettings `mapstructure:"providers"`
	Defaults  DefaultSettings   `mapstructure:"defaults"`
	Logging   LoggingSettings   `mapstructure:"logging"`
}

// ServerSettings contains HTTP server configuration
type ServerSettings struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// StorageSettings contains storage configuration
type StorageSettings struct {
	Type string `mapstructure:"type"`
	Path string `mapstructure:"path"`
}

// AgentsSettings contains agent configuration
type AgentsSettings struct {
	Path  string `mapstructure:"path"`
	Watch bool   `mapstructure:"watch"`
}

// ProvidersSettings contains LLM provider configurations
type ProvidersSettings struct {
	Anthropic AnthropicProvider `mapstructure:"anthropic"`
	OpenAI    OpenAIProvider    `mapstructure:"openai"`
	Zai       ZaiProvider       `mapstructure:"zai"`
	Gemini    GeminiProvider    `mapstructure:"gemini"`
	Ollama    OllamaProvider    `mapstructure:"ollama"`
}

// AnthropicProvider settings
type AnthropicProvider struct {
	APIKey       string `mapstructure:"api_key"`
	DefaultModel string `mapstructure:"default_model"`
}

// OpenAIProvider settings
type OpenAIProvider struct {
	APIKey       string `mapstructure:"api_key"`
	DefaultModel string `mapstructure:"default_model"`
}

// ZaiProvider settings
type ZaiProvider struct {
	APIKey       string `mapstructure:"api_key"`
	BaseURL      string `mapstructure:"base_url"`
	DefaultModel string `mapstructure:"default_model"`
}

// GeminiProvider settings
type GeminiProvider struct {
	APIKey       string `mapstructure:"api_key"`
	DefaultModel string `mapstructure:"default_model"`
	UseCLI       bool   `mapstructure:"use_cli"`    // Use Gemini CLI bridge instead of direct API
	BridgeURL    string `mapstructure:"bridge_url"` // Node.js bridge URL (default: http://localhost:3001)
}

// OllamaProvider settings
type OllamaProvider struct {
	Host         string `mapstructure:"host"`
	DefaultModel string `mapstructure:"default_model"`
}

// DefaultSettings contains default LLM settings
type DefaultSettings struct {
	Provider    string  `mapstructure:"provider"`
	Temperature float64 `mapstructure:"temperature"`
	MaxTokens   int     `mapstructure:"max_tokens"`
}

// LoggingSettings contains logging configuration
type LoggingSettings struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// LoadServerConfig loads the server configuration from a file
func LoadServerConfig(configPath string) (*ServerConfig, error) {
	v := viper.New()

	// Set config file path
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("server")
		v.SetConfigType("yaml")
		v.AddConfigPath("./config")
		v.AddConfigPath(".")
	}

	// Enable environment variable substitution
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Expand environment variables in config values
	for _, key := range v.AllKeys() {
		value := v.GetString(key)
		if strings.Contains(value, "${") {
			expanded := os.ExpandEnv(value)
			v.Set(key, expanded)
		}
	}

	// Unmarshal config
	var config ServerConfig
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
