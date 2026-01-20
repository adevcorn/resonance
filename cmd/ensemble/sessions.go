package main

import (
	"context"
	"fmt"

	"github.com/adevcorn/ensemble/internal/client"
	"github.com/adevcorn/ensemble/internal/config"
	"github.com/spf13/cobra"
)

var sessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "Manage sessions",
	Long:  `List, view, and delete sessions.`,
}

var sessionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all sessions",
	RunE:  listSessions,
}

var sessionsShowCmd = &cobra.Command{
	Use:   "show [id]",
	Short: "Show session details",
	Args:  cobra.ExactArgs(1),
	RunE:  showSession,
}

var sessionsDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a session",
	Args:  cobra.ExactArgs(1),
	RunE:  deleteSession,
}

func init() {
	sessionsCmd.AddCommand(sessionsListCmd)
	sessionsCmd.AddCommand(sessionsShowCmd)
	sessionsCmd.AddCommand(sessionsDeleteCmd)
}

func listSessions(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.LoadClientConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create client
	c := client.NewClient(cfg.Client.ServerURL)

	// List sessions
	ctx := context.Background()
	sessions, err := c.ListSessions(ctx)
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions found.")
		return nil
	}

	fmt.Printf("Sessions (%d):\n\n", len(sessions))

	for _, session := range sessions {
		fmt.Printf("📋 %s\n", session.ID)
		fmt.Printf("   Project: %s\n", session.ProjectPath)
		fmt.Printf("   State: %s\n", session.State)
		fmt.Printf("   Created: %s\n", session.CreatedAt.Format("2006-01-02 15:04:05"))
		if len(session.ActiveTeam) > 0 {
			fmt.Printf("   Team: %v\n", session.ActiveTeam)
		}
		fmt.Println()
	}

	return nil
}

func showSession(cmd *cobra.Command, args []string) error {
	sessionID := args[0]

	// Load configuration
	cfg, err := config.LoadClientConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create client
	c := client.NewClient(cfg.Client.ServerURL)

	// Get session
	ctx := context.Background()
	session, err := c.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	fmt.Printf("📋 Session: %s\n\n", session.ID)
	fmt.Printf("Project Path: %s\n", session.ProjectPath)
	fmt.Printf("State: %s\n", session.State)
	fmt.Printf("Created: %s\n", session.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated: %s\n", session.UpdatedAt.Format("2006-01-02 15:04:05"))

	if len(session.ActiveTeam) > 0 {
		fmt.Printf("\nActive Team:\n")
		for _, agent := range session.ActiveTeam {
			fmt.Printf("  - %s\n", agent)
		}
	}

	if len(session.Messages) > 0 {
		fmt.Printf("\nMessages (%d):\n", len(session.Messages))
		for i, msg := range session.Messages {
			if i >= 5 {
				fmt.Printf("  ... and %d more\n", len(session.Messages)-5)
				break
			}
			fmt.Printf("  [%s] %s: %s\n", msg.Role, msg.Agent, truncate(msg.Content, 80))
		}
	}

	return nil
}

func deleteSession(cmd *cobra.Command, args []string) error {
	sessionID := args[0]

	// Load configuration
	cfg, err := config.LoadClientConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create client
	c := client.NewClient(cfg.Client.ServerURL)

	// Delete session
	ctx := context.Background()
	if err := c.DeleteSession(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	fmt.Printf("✅ Session %s deleted\n", sessionID)

	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
