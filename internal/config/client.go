package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// ClientConfig represents the client configuration
type ClientConfig struct {
	Client      ClientSettings      `mapstructure:"client"`
	Project     ProjectSettings     `mapstructure:"project"`
	Permissions PermissionsSettings `mapstructure:"permissions"`
	MCP         MCPSettings         `mapstructure:"mcp"`
	Logging     LoggingSettings     `mapstructure:"logging"`
}

// ClientSettings contains client connection settings
type ClientSettings struct {
	ServerURL string `mapstructure:"server_url"`
}

// ProjectSettings contains project detection settings
type ProjectSettings struct {
	AutoDetect bool `mapstructure:"auto_detect"`
}

// PermissionsSettings contains tool permission settings
type PermissionsSettings struct {
	File FilePermissions `mapstructure:"file"`
	Exec ExecPermissions `mapstructure:"exec"`
}

// FilePermissions defines file operation permissions
type FilePermissions struct {
	AllowedPaths []string `mapstructure:"allowed_paths"`
	DeniedPaths  []string `mapstructure:"denied_paths"`
}

// ExecPermissions defines command execution permissions
type ExecPermissions struct {
	AllowedCommands []string `mapstructure:"allowed_commands"`
	DeniedCommands  []string `mapstructure:"denied_commands"`
}

// MCPSettings contains MCP server configurations
type MCPSettings struct {
	Servers []MCPServer `mapstructure:"servers"`
}

// MCPServer represents a single MCP server configuration
type MCPServer struct {
	Name    string            `mapstructure:"name"`
	Command string            `mapstructure:"command"`
	Args    []string          `mapstructure:"args"`
	Env     map[string]string `mapstructure:"env"`
}

// LoadClientConfig loads the client configuration from a file
func LoadClientConfig(configPath string) (*ClientConfig, error) {
	v := viper.New()

	// Set config file path
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("client")
		v.SetConfigType("yaml")
		v.AddConfigPath("./config")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.ensemble")
	}

	// Enable environment variable substitution
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal config
	var config ClientConfig
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
