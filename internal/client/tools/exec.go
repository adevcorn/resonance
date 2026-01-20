package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"
)

// ExecuteCommandInput is the input for execute_command tool
type ExecuteCommandInput struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Cwd     string   `json:"cwd,omitempty"`
}

// ExecuteCommandOutput is the output for execute_command tool
type ExecuteCommandOutput struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
}

// ExecuteCommand runs a shell command
func ExecuteCommand(ctx context.Context, projectPath string, input ExecuteCommandInput) (*ExecuteCommandOutput, error) {
	// Determine working directory
	cwd := projectPath
	if input.Cwd != "" {
		if filepath.IsAbs(input.Cwd) {
			cwd = input.Cwd
		} else {
			cwd = filepath.Join(projectPath, input.Cwd)
		}
	}

	// Create command with timeout context
	cmdCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Execute the command through shell to properly handle full command strings
	// like "git status" or "go build ./..."
	var cmd *exec.Cmd
	if len(input.Args) > 0 {
		// If Args are provided, use them directly (legacy format)
		cmd = exec.CommandContext(cmdCtx, input.Command, input.Args...)
	} else {
		// Otherwise, execute the full command string through a shell
		// This handles commands like "git status", "npm run build", etc.
		cmd = exec.CommandContext(cmdCtx, "sh", "-c", input.Command)
	}
	cmd.Dir = cwd

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run command
	err := cmd.Run()

	output := &ExecuteCommandOutput{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: 0,
	}

	// Get exit code
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			output.ExitCode = exitErr.ExitCode()
		} else {
			// Command failed to start or context deadline exceeded
			return nil, fmt.Errorf("command failed: %w", err)
		}
	}

	return output, nil
}
