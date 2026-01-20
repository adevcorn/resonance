package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// ReadFileInput is the input for read_file tool
type ReadFileInput struct {
	Path string `json:"path"`
}

// ReadFileOutput is the output for read_file tool
type ReadFileOutput struct {
	Content string `json:"content"`
}

// WriteFileInput is the input for write_file tool
type WriteFileInput struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// WriteFileOutput is the output for write_file tool
type WriteFileOutput struct {
	Success bool `json:"success"`
}

// ListDirectoryInput is the input for list_directory tool
type ListDirectoryInput struct {
	Path      string `json:"path"`
	Recursive bool   `json:"recursive"`
}

// ListDirectoryOutput is the output for list_directory tool
type ListDirectoryOutput struct {
	Entries []DirEntry `json:"entries"`
}

// DirEntry represents a directory entry
type DirEntry struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
}

// ReadFile reads a file
func ReadFile(ctx context.Context, projectPath string, input ReadFileInput) (*ReadFileOutput, error) {
	// Resolve path relative to project
	path := resolvePath(projectPath, input.Path)

	// Read file
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return &ReadFileOutput{
		Content: string(content),
	}, nil
}

// WriteFile writes to a file
func WriteFile(ctx context.Context, projectPath string, input WriteFileInput) (*WriteFileOutput, error) {
	// Resolve path relative to project
	path := resolvePath(projectPath, input.Path)

	// Create parent directories if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, []byte(input.Content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return &WriteFileOutput{
		Success: true,
	}, nil
}

// ListDirectory lists directory contents
func ListDirectory(ctx context.Context, projectPath string, input ListDirectoryInput) (*ListDirectoryOutput, error) {
	// Resolve path relative to project
	path := resolvePath(projectPath, input.Path)

	var entries []DirEntry

	if input.Recursive {
		// Walk directory recursively
		err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip the root directory itself
			if p == path {
				return nil
			}

			// Get relative path
			relPath, err := filepath.Rel(projectPath, p)
			if err != nil {
				relPath = p
			}

			entries = append(entries, DirEntry{
				Name:  info.Name(),
				Path:  relPath,
				IsDir: info.IsDir(),
				Size:  info.Size(),
			})

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("failed to walk directory: %w", err)
		}
	} else {
		// List directory non-recursively
		dirEntries, err := os.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory: %w", err)
		}

		for _, entry := range dirEntries {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			entryPath := filepath.Join(path, entry.Name())
			relPath, err := filepath.Rel(projectPath, entryPath)
			if err != nil {
				relPath = entryPath
			}

			entries = append(entries, DirEntry{
				Name:  entry.Name(),
				Path:  relPath,
				IsDir: entry.IsDir(),
				Size:  info.Size(),
			})
		}
	}

	return &ListDirectoryOutput{
		Entries: entries,
	}, nil
}

// resolvePath resolves a path relative to project path
func resolvePath(projectPath, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(projectPath, path)
}
