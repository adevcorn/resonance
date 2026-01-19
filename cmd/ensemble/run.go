package main

import (
	"context"
	"encoding/json"
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
			// Check if this is a collaboration message
			if msgType, ok := msg.Metadata["type"].(string); ok && msgType == "collaboration" {
				// Display collaboration message with special formatting
				fmt.Printf("   💬 %s\n", msg.Content)
				return nil
			}

			// Regular agent message
			if msg.Content != "" {
				fmt.Printf("\n🤖 %s:\n%s\n", msg.Agent, msg.Content)
			}
			return nil
		},
		// onToolCall (client-side tools only)
		func(call protocol.ToolCall) (protocol.ToolResult, error) {
			// Show what tool is being called
			action := formatToolAction(call)
			fmt.Printf("   🔧 %s...", action)

			// Execute the tool
			result, err := executor.Execute(ctx, call)

			if err != nil {
				fmt.Printf(" ❌ Failed: %v\n", err)
			} else {
				// Show success with result summary
				summary := formatToolResult(call.ToolName, result)
				fmt.Printf(" %s\n", summary)
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

	// Set up server tool callbacks (for display only)
	ws.SetServerToolCallbacks(
		// onServerToolStart - show that server is executing a tool
		func(call protocol.ToolCall) error {
			action := formatToolAction(call)
			fmt.Printf("   🔧 %s...", action)
			return nil
		},
		// onServerToolEnd - show server tool result
		func(call protocol.ToolCall, result protocol.ToolResult) error {
			if result.Error != "" {
				fmt.Printf(" ❌ Failed: %s\n", result.Error)
			} else {
				summary := formatToolResult(call.ToolName, result)
				fmt.Printf(" %s\n", summary)
			}
			return nil
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

// formatToolAction creates a human-readable description of a tool call
func formatToolAction(tc protocol.ToolCall) string {
	// Parse arguments to extract useful info
	var args map[string]interface{}
	json.Unmarshal(tc.Arguments, &args)

	switch tc.ToolName {
	case "read_file":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("Reading %s", path)
		}
		return "Reading file"

	case "write_file":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("Writing %s", path)
		}
		return "Writing file"

	case "list_directory":
		if path, ok := args["path"].(string); ok {
			if path == "." {
				return "Listing project files"
			}
			return fmt.Sprintf("Listing %s", path)
		}
		return "Listing directory"

	case "execute_command":
		if cmd, ok := args["command"].(string); ok {
			return fmt.Sprintf("Running: %s", cmd)
		}
		return "Executing command"

	case "fetch_url":
		if url, ok := args["url"].(string); ok {
			return fmt.Sprintf("Fetching %s", url)
		}
		return "Fetching URL"

	case "web_search":
		if query, ok := args["query"].(string); ok {
			return fmt.Sprintf("Searching: %s", query)
		}
		return "Web search"

	case "collaborate":
		action, _ := args["action"].(string)
		toAgent, hasAgent := args["to_agent"].(string)

		switch action {
		case "broadcast":
			return "Broadcasting to team"
		case "direct":
			if hasAgent {
				return fmt.Sprintf("Messaging %s", toAgent)
			}
			return "Messaging agent"
		case "help":
			if hasAgent {
				return fmt.Sprintf("Requesting help from %s", toAgent)
			}
			return "Requesting help"
		case "complete":
			return "Signaling completion"
		default:
			return "Collaborating"
		}

	case "assemble_team":
		if agents, ok := args["agents"].([]interface{}); ok {
			return fmt.Sprintf("Assembling team (%d agents)", len(agents))
		}
		return "Assembling team"

	default:
		return fmt.Sprintf("Using %s", tc.ToolName)
	}
}

// formatToolResult creates a human-readable summary of a tool result
func formatToolResult(toolName string, result protocol.ToolResult) string {
	// Parse result to extract useful info
	var resultData map[string]interface{}
	json.Unmarshal(result.Result, &resultData)

	switch toolName {
	case "read_file":
		if content, ok := resultData["content"].(string); ok {
			size := len(content)
			if size > 1024 {
				return fmt.Sprintf("✅ Read %d KB", size/1024)
			}
			return fmt.Sprintf("✅ Read %d bytes", size)
		}
		return "✅ Done"

	case "write_file":
		if path, ok := resultData["path"].(string); ok {
			return fmt.Sprintf("✅ Wrote %s", path)
		}
		return "✅ Done"

	case "list_directory":
		if files, ok := resultData["files"].([]interface{}); ok {
			fileCount := 0
			dirCount := 0
			for _, f := range files {
				if fileMap, ok := f.(map[string]interface{}); ok {
					if isDir, ok := fileMap["is_dir"].(bool); ok && isDir {
						dirCount++
					} else {
						fileCount++
					}
				}
			}
			if dirCount > 0 {
				return fmt.Sprintf("✅ Found %d files, %d directories", fileCount, dirCount)
			}
			return fmt.Sprintf("✅ Found %d files", fileCount)
		}
		return "✅ Done"

	case "execute_command":
		if exitCode, ok := resultData["exit_code"].(float64); ok {
			if exitCode == 0 {
				return "✅ Success"
			}
			return fmt.Sprintf("⚠️  Exit code %d", int(exitCode))
		}
		return "✅ Done"

	case "fetch_url":
		if statusCode, ok := resultData["status_code"].(float64); ok {
			if statusCode >= 200 && statusCode < 300 {
				if bodyLen, ok := resultData["body"].(string); ok {
					return fmt.Sprintf("✅ %d OK (%d bytes)", int(statusCode), len(bodyLen))
				}
				return fmt.Sprintf("✅ %d OK", int(statusCode))
			}
			return fmt.Sprintf("⚠️  %d", int(statusCode))
		}
		return "✅ Done"

	case "web_search":
		if results, ok := resultData["results"].([]interface{}); ok {
			return fmt.Sprintf("✅ Found %d results", len(results))
		}
		return "✅ Done"

	case "collaborate":
		if delivered, ok := resultData["delivered"].(bool); ok && delivered {
			return "✅ Delivered"
		}
		return "✅ Done"

	case "assemble_team":
		if team, ok := resultData["team"].([]interface{}); ok {
			return fmt.Sprintf("✅ Assembled %d agents", len(team))
		}
		return "✅ Done"

	default:
		return "✅ Done"
	}
}
