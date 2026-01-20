package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/adevcorn/ensemble/internal/client"
	"github.com/adevcorn/ensemble/internal/config"
	"github.com/spf13/cobra"
)

var agentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Manage and view agents",
	Long:  `View available agents and their capabilities.`,
}

var agentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available agents",
	RunE:  listAgents,
}

var agentsShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show agent details",
	Args:  cobra.ExactArgs(1),
	RunE:  showAgent,
}

func init() {
	agentsCmd.AddCommand(agentsListCmd)
	agentsCmd.AddCommand(agentsShowCmd)
}

func listAgents(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.LoadClientConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create client
	c := client.NewClient(cfg.Client.ServerURL)

	// List agents
	ctx := context.Background()
	agents, err := c.ListAgents(ctx)
	if err != nil {
		return fmt.Errorf("failed to list agents: %w", err)
	}

	fmt.Printf("Available Agents (%d):\n\n", len(agents))

	for _, agent := range agents {
		fmt.Printf("🤖 %s (%s)\n", agent.DisplayName, agent.Name)
		fmt.Printf("   %s\n", agent.Description)
		if len(agent.Capabilities) > 0 {
			fmt.Printf("   Capabilities: %s\n", strings.Join(agent.Capabilities, ", "))
		}
		fmt.Println()
	}

	return nil
}

func showAgent(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Load configuration
	cfg, err := config.LoadClientConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create client
	c := client.NewClient(cfg.Client.ServerURL)

	// Get agent details
	ctx := context.Background()
	agent, err := c.GetAgent(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to get agent: %w", err)
	}

	fmt.Printf("🤖 %s (%s)\n\n", agent.DisplayName, agent.Name)
	fmt.Printf("Description:\n%s\n\n", agent.Description)

	if len(agent.Capabilities) > 0 {
		fmt.Printf("Capabilities:\n")
		for _, cap := range agent.Capabilities {
			fmt.Printf("  - %s\n", cap)
		}
		fmt.Println()
	}

	if agent.Model != nil {
		fmt.Printf("Model:\n")
		if provider, ok := agent.Model["provider"].(string); ok {
			fmt.Printf("  Provider: %s\n", provider)
		}
		if modelName, ok := agent.Model["name"].(string); ok {
			fmt.Printf("  Name: %s\n", modelName)
		}
		fmt.Println()
	}

	if agent.Tools != nil {
		if allowed, ok := agent.Tools["allowed"].([]interface{}); ok && len(allowed) > 0 {
			fmt.Printf("Allowed Tools:\n")
			for _, tool := range allowed {
				if toolName, ok := tool.(string); ok {
					fmt.Printf("  - %s\n", toolName)
				}
			}
			fmt.Println()
		}
	}

	fmt.Printf("System Prompt:\n%s\n", agent.SystemPrompt)

	return nil
}
