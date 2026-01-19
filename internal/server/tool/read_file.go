package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adevcorn/ensemble/internal/protocol"
)

// ReadFileInput is the input for reading a file
type ReadFileInput struct {
	Path string `json:"path"`
}

// ReadFileOutput is the result of reading a file
type ReadFileOutput struct {
	Content string `json:"content"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
}

// ReadFileTool allows agents to read file contents
type ReadFileTool struct {
	baseDir string
}

// NewReadFileTool creates a new read file tool
func NewReadFileTool(baseDir string) *ReadFileTool {
	if baseDir == "" {
		baseDir = "."
	}
	return &ReadFileTool{
		baseDir: baseDir,
	}
}

// Name returns the tool name
func (r *ReadFileTool) Name() string {
	return "read_file"
}

// Description returns the tool description
func (r *ReadFileTool) Description() string {
	return "Read the contents of a file. Provide the file path relative to the project root."
}

// Parameters returns the JSON Schema for read_file parameters
func (r *ReadFileTool) Parameters() json.RawMessage {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to read (relative to project root)",
			},
		},
		"required": []string{"path"},
	}

	data, _ := json.Marshal(schema)
	return data
}

// Execute reads the file
func (r *ReadFileTool) Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var fileInput ReadFileInput
	if err := json.Unmarshal(input, &fileInput); err != nil {
		return nil, fmt.Errorf("invalid read_file input: %w", err)
	}

	if fileInput.Path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	// Resolve path relative to base directory
	fullPath := filepath.Join(r.baseDir, fileInput.Path)

	// Security: ensure path doesn't escape base directory
	absBase, err := filepath.Abs(r.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve base directory: %w", err)
	}

	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve file path: %w", err)
	}

	relPath, err := filepath.Rel(absBase, absPath)
	if err != nil || relPath[:2] == ".." {
		return nil, fmt.Errorf("access denied: path escapes project directory")
	}

	// Read file
	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Get file info for size
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	output := ReadFileOutput{
		Content: string(content),
		Path:    fileInput.Path,
		Size:    info.Size(),
	}

	data, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}

	return data, nil
}

// ExecutionLocation returns where this tool executes
func (r *ReadFileTool) ExecutionLocation() protocol.ExecutionLocation {
	return protocol.ExecutionLocationServer
}
