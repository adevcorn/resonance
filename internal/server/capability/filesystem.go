package capability

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

// ReadFileCapability allows reading file contents
type ReadFileCapability struct {
	baseDir string
}

// NewReadFileCapability creates a new read file capability
func NewReadFileCapability(baseDir string) *ReadFileCapability {
	if baseDir == "" {
		baseDir = "."
	}
	return &ReadFileCapability{
		baseDir: baseDir,
	}
}

// Name returns the capability name
func (r *ReadFileCapability) Name() string {
	return "read_file"
}

// Execute reads the file
func (r *ReadFileCapability) Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
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
	if err != nil || (len(relPath) >= 2 && relPath[:2] == "..") {
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

// ExecutionLocation returns where this capability executes
func (r *ReadFileCapability) ExecutionLocation() protocol.ExecutionLocation {
	return protocol.ExecutionLocationServer
}

// WriteFileCapability is a client-side capability for writing files
type WriteFileCapability struct{}

// NewWriteFileCapability creates a new write file capability
func NewWriteFileCapability() *WriteFileCapability {
	return &WriteFileCapability{}
}

// Name returns the capability name
func (w *WriteFileCapability) Name() string {
	return "write_file"
}

// Execute should never be called on server - this is a client-side capability
func (w *WriteFileCapability) Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	// This should never be called because ExecutionLocation is Client
	return nil, fmt.Errorf("write_file is a client-side capability and should not be executed on server")
}

// ExecutionLocation returns where this capability executes
func (w *WriteFileCapability) ExecutionLocation() protocol.ExecutionLocation {
	return protocol.ExecutionLocationClient
}

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

// ListDirectoryCapability allows listing directory contents
type ListDirectoryCapability struct {
	baseDir string
}

// NewListDirectoryCapability creates a new list directory capability
func NewListDirectoryCapability(baseDir string) *ListDirectoryCapability {
	if baseDir == "" {
		baseDir = "."
	}
	return &ListDirectoryCapability{
		baseDir: baseDir,
	}
}

// Name returns the capability name
func (l *ListDirectoryCapability) Name() string {
	return "list_directory"
}

// Execute lists the directory
func (l *ListDirectoryCapability) Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
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

// ExecutionLocation returns where this capability executes
func (l *ListDirectoryCapability) ExecutionLocation() protocol.ExecutionLocation {
	return protocol.ExecutionLocationServer
}
