package main

import (
	"fmt"
	"os"

	"github.com/adevcorn/ensemble/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show configuration",
	Long:  `Display the current configuration.`,
	RunE:  showConfig,
}

func showConfig(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.LoadClientConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Marshal to YAML for display
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	fmt.Println("Current Configuration:")
	fmt.Println()
	fmt.Println(string(data))

	// Show config file location
	cfgPath := configPath
	if cfgPath == "" {
		cfgPath = "config/client.yaml (default)"
	}
	fmt.Printf("Config file: %s\n", cfgPath)

	// Show environment variables
	fmt.Println("\nRelevant Environment Variables:")
	envVars := []string{
		"ENSEMBLE_CLIENT_SERVER_URL",
		"ENSEMBLE_PROJECT_AUTO_DETECT",
	}

	for _, envVar := range envVars {
		if val := os.Getenv(envVar); val != "" {
			fmt.Printf("  %s=%s\n", envVar, val)
		}
	}

	return nil
}
