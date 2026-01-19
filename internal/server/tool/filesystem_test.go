package tool

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestReadFileTool(t *testing.T) {
	// Create a temp file for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "Hello, World!"

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create tool
	tool := NewReadFileTool(tmpDir)

	// Test reading the file
	input := ReadFileInput{
		Path: "test.txt",
	}
	inputData, _ := json.Marshal(input)

	output, err := tool.Execute(context.Background(), inputData)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	var result ReadFileOutput
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Failed to unmarshal output: %v", err)
	}

	if result.Content != testContent {
		t.Errorf("Expected content %q, got %q", testContent, result.Content)
	}
}

func TestListDirectoryTool(t *testing.T) {
	// Create a temp directory with some files
	tmpDir := t.TempDir()

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("test"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

	// Create tool
	tool := NewListDirectoryTool(tmpDir)

	// Test listing the directory
	input := ListDirectoryInput{
		Path: ".",
	}
	inputData, _ := json.Marshal(input)

	output, err := tool.Execute(context.Background(), inputData)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	var result ListDirectoryOutput
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Failed to unmarshal output: %v", err)
	}

	if len(result.Files) != 3 {
		t.Errorf("Expected 3 files, got %d", len(result.Files))
	}
}

func TestPathEscaping(t *testing.T) {
	tmpDir := t.TempDir()
	tool := NewReadFileTool(tmpDir)

	// Try to escape the base directory
	input := ReadFileInput{
		Path: "../../../etc/passwd",
	}
	inputData, _ := json.Marshal(input)

	_, err := tool.Execute(context.Background(), inputData)
	if err == nil {
		t.Error("Expected error for path escaping, got nil")
	}
	if err != nil && err.Error() != "access denied: path escapes project directory" {
		t.Logf("Got expected error: %v", err)
	}
}
