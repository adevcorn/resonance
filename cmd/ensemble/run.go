package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/adevcorn/ensemble/internal/client"
	"github.com/adevcorn/ensemble/internal/config"
	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [task]",
	Short: "Run a task with Ensemble agents",
	Long: `Run a task with Ensemble's multi-agent system.
The agents will collaborate to accomplish the task.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runTask,
}

func runTask(cmd *cobra.Command, args []string) error {
	task := args[0]

	// Load configuration
	cfg, err := config.LoadClientConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Detect or load project
	var project *client.Project
	if cfg.Project.AutoDetect {
		project, err = client.DetectProject()
		if err != nil {
			return fmt.Errorf("failed to detect project: %w", err)
		}
		fmt.Printf("📁 Project detected: %s\n", project.Path())
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		project, err = client.NewProject(cwd)
		if err != nil {
			return fmt.Errorf("failed to load project: %w", err)
		}
	}

	// Create client
	c := client.NewClient(cfg.Client.ServerURL)

	// Create session
	ctx := context.Background()
	session, err := c.CreateSession(ctx, project.Path())
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	fmt.Printf("🔗 Session created: %s\n", session.ID)

	// Gather project info
	projectInfo, err := project.GetInfo(ctx)
	if err != nil {
		fmt.Printf("⚠️  Warning: failed to gather project info: %v\n", err)
	}

	// Connect WebSocket
	ws, err := c.ConnectWebSocket(ctx, session.ID)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}
	defer ws.Close()

	fmt.Printf("🚀 Starting task: %s\n\n", task)

	// Create executor
	checker := client.NewChecker(&cfg.Permissions, project.Path())
	executor := client.NewExecutor(project.Path(), checker)

	// Set up callbacks
	ws.SetCallbacks(
		// onMessage
		func(msg protocol.Message) error {
			fmt.Printf("\n🤖 %s:\n%s\n", msg.Agent, msg.Content)
			return nil
		},
		// onToolCall
		func(call protocol.ToolCall) (protocol.ToolResult, error) {
			fmt.Printf("\n🔧 Tool call: %s\n", call.ToolName)
			result, err := executor.Execute(ctx, call)
			if err != nil {
				fmt.Printf("❌ Tool execution failed: %v\n", err)
			}
			return result, nil
		},
		// onComplete
		func(summary string, artifacts []string) error {
			fmt.Printf("\n✅ Task completed!\n\n")
			fmt.Printf("Summary:\n%s\n", summary)
			if len(artifacts) > 0 {
				fmt.Printf("\nArtifacts:\n")
				for _, artifact := range artifacts {
					fmt.Printf("  - %s\n", artifact)
				}
			}
			return nil
		},
		// onError
		func(err error) error {
			fmt.Printf("\n❌ Error: %v\n", err)
			return err
		},
	)

	// Start task
	if err := ws.Start(task, projectInfo); err != nil {
		return fmt.Errorf("failed to start task: %w", err)
	}

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Create context for listening
	listenCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Listen for messages in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- ws.Listen(listenCtx)
	}()

	// Wait for completion or interruption
	select {
	case <-sigCh:
		fmt.Printf("\n\n⚠️  Interrupted, cancelling task...\n")
		cancel()
		ws.Cancel()
		return nil
	case err := <-errCh:
		if err != nil && err != context.Canceled {
			return fmt.Errorf("WebSocket error: %w", err)
		}
	}

	return nil
}
