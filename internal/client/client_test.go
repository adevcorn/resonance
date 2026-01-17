package client

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/adevcorn/ensemble/internal/config"
	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectDetection(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ensemble-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a go.mod file to mark it as a project
	goModPath := tmpDir + "/go.mod"
	err = os.WriteFile(goModPath, []byte("module test\n"), 0644)
	require.NoError(t, err)

	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// Detect project
	project, err := DetectProject()
	require.NoError(t, err)
	assert.Contains(t, project.Path(), tmpDir)
}

func TestProjectInfo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ensemble-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a Go project
	goModPath := tmpDir + "/go.mod"
	goModContent := `module github.com/test/project

go 1.21
`
	err = os.WriteFile(goModPath, []byte(goModContent), 0644)
	require.NoError(t, err)

	project, err := NewProject(tmpDir)
	require.NoError(t, err)

	ctx := context.Background()
	info, err := project.GetInfo(ctx)
	require.NoError(t, err)

	assert.Equal(t, tmpDir, info.Path)
	assert.Equal(t, "go", info.Language)
}

func TestPermissionChecker(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ensemble-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	permissions := &config.PermissionsSettings{
		File: config.FilePermissions{
			AllowedPaths: []string{"."},
			DeniedPaths:  []string{".git", ".env"},
		},
		Exec: config.ExecPermissions{
			AllowedCommands: []string{"git", "go", "npm"},
			DeniedCommands:  []string{"rm -rf", "sudo"},
		},
	}

	checker := NewChecker(permissions, tmpDir)

	// Test allowed file path
	err = checker.CheckFilePath("src/main.go")
	assert.NoError(t, err)

	// Test denied file path
	err = checker.CheckFilePath(".git/config")
	assert.Error(t, err)

	// Test allowed command
	err = checker.CheckCommand("git", []string{"status"})
	assert.NoError(t, err)

	// Test denied command
	err = checker.CheckCommand("sudo", []string{"rm", "-rf", "/"})
	assert.Error(t, err)

	// Test unlisted command (should be denied by default)
	err = checker.CheckCommand("unknown-cmd", []string{})
	assert.Error(t, err)
}

func TestPermissionCheckerPathTraversal(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ensemble-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	permissions := &config.PermissionsSettings{
		File: config.FilePermissions{
			AllowedPaths: []string{"."},
		},
	}

	checker := NewChecker(permissions, tmpDir)

	// Test path traversal attempt
	err = checker.CheckFilePath("../../../etc/passwd")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "outside project directory")
}

func TestExecutor(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ensemble-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	permissions := &config.PermissionsSettings{
		File: config.FilePermissions{
			AllowedPaths: []string{"."},
		},
		Exec: config.ExecPermissions{
			AllowedCommands: []string{"echo"},
		},
	}

	checker := NewChecker(permissions, tmpDir)
	executor := NewExecutor(tmpDir, checker)

	ctx := context.Background()

	// Test write_file
	t.Run("write_file", func(t *testing.T) {
		call := createToolCall("write_file", map[string]interface{}{
			"path":    "test.txt",
			"content": "Hello, World!",
		})

		result, err := executor.Execute(ctx, call)
		require.NoError(t, err)
		assert.Empty(t, result.Error)
	})

	// Test read_file
	t.Run("read_file", func(t *testing.T) {
		call := createToolCall("read_file", map[string]interface{}{
			"path": "test.txt",
		})

		result, err := executor.Execute(ctx, call)
		require.NoError(t, err)
		assert.Empty(t, result.Error)
		assert.Contains(t, string(result.Result), "Hello, World!")
	})

	// Test list_directory
	t.Run("list_directory", func(t *testing.T) {
		call := createToolCall("list_directory", map[string]interface{}{
			"path":      ".",
			"recursive": false,
		})

		result, err := executor.Execute(ctx, call)
		require.NoError(t, err)
		assert.Empty(t, result.Error)
	})
}

func createToolCall(name string, args map[string]interface{}) protocol.ToolCall {
	argsBytes, _ := json.Marshal(args)
	return protocol.ToolCall{
		ID:        "test-call-1",
		ToolName:  name,
		Arguments: argsBytes,
	}
}
