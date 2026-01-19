package tools

import (
	"bytes"
	"context"
	"fmt"
	"os"
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
	fmt.Fprintf(os.Stderr, "[EXEC] Command: %s, Args: %v, Cwd: %s\n", input.Command, input.Args, input.Cwd)

	// Determine working directory
	cwd := projectPath
	if input.Cwd != "" {
		if filepath.IsAbs(input.Cwd) {
			cwd = input.Cwd
		} else {
			cwd = filepath.Join(projectPath, input.Cwd)
		}
	}

	fmt.Fprintf(os.Stderr, "[EXEC] Working directory: %s\n", cwd)

	// Create command with timeout context
	cmdCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Execute the command through shell to properly handle full command strings
	// like "git status" or "go build ./..."
	var cmd *exec.Cmd
	if len(input.Args) > 0 {
		// If Args are provided, use them directly (legacy format)
		fmt.Fprintf(os.Stderr, "[EXEC] Using args format: %s %v\n", input.Command, input.Args)
		cmd = exec.CommandContext(cmdCtx, input.Command, input.Args...)
	} else {
		// Otherwise, execute the full command string through a shell
		// This handles commands like "git status", "npm run build", etc.
		fmt.Fprintf(os.Stderr, "[EXEC] Using shell format: sh -c '%s'\n", input.Command)
		cmd = exec.CommandContext(cmdCtx, "sh", "-c", input.Command)
	}
	cmd.Dir = cwd

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run command
	fmt.Fprintf(os.Stderr, "[EXEC] Running command...\n")
	err := cmd.Run()
	fmt.Fprintf(os.Stderr, "[EXEC] Command complete. Stdout len: %d, Stderr len: %d, Error: %v\n",
		stdout.Len(), stderr.Len(), err)

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

	fmt.Fprintf(os.Stderr, "[EXEC] Returning output: stdout=%d bytes, stderr=%d bytes, exitcode=%d\n",
		len(output.Stdout), len(output.Stderr), output.ExitCode)

	return output, nil
}
