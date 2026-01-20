package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adevcorn/ensemble/internal/protocol"
)

// ListDirectoryInput is the input for listing a directory
type ListDirectoryInput struct {
	Path string `json:"path"`
}

// FileInfo represents information about a file or directory
type FileInfo struct {
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
}

// ListDirectoryOutput is the result of listing a directory
type ListDirectoryOutput struct {
	Path  string     `json:"path"`
	Files []FileInfo `json:"files"`
}

// ListDirectoryTool allows agents to list directory contents
type ListDirectoryTool struct {
	baseDir string
}

// NewListDirectoryTool creates a new list directory tool
func NewListDirectoryTool(baseDir string) *ListDirectoryTool {
	if baseDir == "" {
		baseDir = "."
	}
	return &ListDirectoryTool{
		baseDir: baseDir,
	}
}

// Name returns the tool name
func (l *ListDirectoryTool) Name() string {
	return "list_directory"
}

// Description returns the tool description
func (l *ListDirectoryTool) Description() string {
	return "List the contents of a directory. Provide the directory path relative to the project root."
}

// Parameters returns the JSON Schema for list_directory parameters
func (l *ListDirectoryTool) Parameters() json.RawMessage {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the directory to list (relative to project root, use '.' for project root)",
			},
		},
		"required": []string{"path"},
	}

	data, _ := json.Marshal(schema)
	return data
}

// Execute lists the directory
func (l *ListDirectoryTool) Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var dirInput ListDirectoryInput
	if err := json.Unmarshal(input, &dirInput); err != nil {
		return nil, fmt.Errorf("invalid list_directory input: %w", err)
	}

	if dirInput.Path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	// Resolve path relative to base directory
	fullPath := filepath.Join(l.baseDir, dirInput.Path)

	// Security: ensure path doesn't escape base directory
	absBase, err := filepath.Abs(l.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve base directory: %w", err)
	}

	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve directory path: %w", err)
	}

	relPath, err := filepath.Rel(absBase, absPath)
	if err != nil || (len(relPath) >= 2 && relPath[:2] == "..") {
		return nil, fmt.Errorf("access denied: path escapes project directory")
	}

	// Read directory
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	// Convert to FileInfo list
	files := make([]FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue // Skip entries we can't stat
		}

		files = append(files, FileInfo{
			Name:  entry.Name(),
			IsDir: entry.IsDir(),
			Size:  info.Size(),
		})
	}

	output := ListDirectoryOutput{
		Path:  dirInput.Path,
		Files: files,
	}

	data, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}

	return data, nil
}

// ExecutionLocation returns where this tool executes
func (l *ListDirectoryTool) ExecutionLocation() protocol.ExecutionLocation {
	return protocol.ExecutionLocationServer
}
