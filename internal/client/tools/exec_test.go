package tools

import (
	"context"
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteCommand(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ensemble-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()

	// Use platform-appropriate commands
	var cmd string
	var args []string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "echo", "Hello"}
	} else {
		cmd = "echo"
		args = []string{"Hello"}
	}

	input := ExecuteCommandInput{
		Command: cmd,
		Args:    args,
	}

	output, err := ExecuteCommand(ctx, tmpDir, input)
	require.NoError(t, err)
	assert.Equal(t, 0, output.ExitCode)
	assert.Contains(t, output.Stdout, "Hello")
}

func TestExecuteCommandWithCwd(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ensemble-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()

	var cmd string
	var args []string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "cd"}
	} else {
		cmd = "pwd"
		args = []string{}
	}

	input := ExecuteCommandInput{
		Command: cmd,
		Args:    args,
		Cwd:     tmpDir,
	}

	output, err := ExecuteCommand(ctx, tmpDir, input)
	require.NoError(t, err)
	assert.Equal(t, 0, output.ExitCode)
	assert.Contains(t, output.Stdout, tmpDir)
}

func TestExecuteCommandNonZeroExit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ensemble-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()

	var cmd string
	var args []string
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "exit", "1"}
	} else {
		cmd = "sh"
		args = []string{"-c", "exit 1"}
	}

	input := ExecuteCommandInput{
		Command: cmd,
		Args:    args,
	}

	output, err := ExecuteCommand(ctx, tmpDir, input)
	require.NoError(t, err) // Command executed, but with non-zero exit
	assert.Equal(t, 1, output.ExitCode)
}

func TestExecuteCommandNotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ensemble-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	input := ExecuteCommandInput{
		Command: "nonexistent-command-xyz",
		Args:    []string{},
	}

	_, err = ExecuteCommand(ctx, tmpDir, input)
	assert.Error(t, err)
}
