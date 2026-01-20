package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadFile(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "ensemble-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test file
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "Hello, World!"
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	require.NoError(t, err)

	// Test reading file
	ctx := context.Background()
	input := ReadFileInput{
		Path: "test.txt",
	}

	output, err := ReadFile(ctx, tmpDir, input)
	require.NoError(t, err)
	assert.Equal(t, testContent, output.Content)
}

func TestReadFileNotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ensemble-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	input := ReadFileInput{
		Path: "nonexistent.txt",
	}

	_, err = ReadFile(ctx, tmpDir, input)
	assert.Error(t, err)
}

func TestWriteFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ensemble-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	input := WriteFileInput{
		Path:    "test.txt",
		Content: "Test content",
	}

	output, err := WriteFile(ctx, tmpDir, input)
	require.NoError(t, err)
	assert.True(t, output.Success)

	// Verify file was written
	content, err := os.ReadFile(filepath.Join(tmpDir, "test.txt"))
	require.NoError(t, err)
	assert.Equal(t, "Test content", string(content))
}

func TestWriteFileCreatesDirectories(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ensemble-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	input := WriteFileInput{
		Path:    "subdir/nested/test.txt",
		Content: "Nested content",
	}

	output, err := WriteFile(ctx, tmpDir, input)
	require.NoError(t, err)
	assert.True(t, output.Success)

	// Verify file exists
	filePath := filepath.Join(tmpDir, "subdir", "nested", "test.txt")
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, "Nested content", string(content))
}

func TestListDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ensemble-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test structure
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("content"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

	ctx := context.Background()
	input := ListDirectoryInput{
		Path:      ".",
		Recursive: false,
	}

	output, err := ListDirectory(ctx, tmpDir, input)
	require.NoError(t, err)
	assert.Len(t, output.Entries, 3)

	// Check entries
	names := make(map[string]bool)
	for _, entry := range output.Entries {
		names[entry.Name] = true
	}
	assert.True(t, names["file1.txt"])
	assert.True(t, names["file2.txt"])
	assert.True(t, names["subdir"])
}

func TestListDirectoryRecursive(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ensemble-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create nested structure
	os.Mkdir(filepath.Join(tmpDir, "dir1"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "dir1", "file.txt"), []byte("content"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "dir2"), 0755)

	ctx := context.Background()
	input := ListDirectoryInput{
		Path:      ".",
		Recursive: true,
	}

	output, err := ListDirectory(ctx, tmpDir, input)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(output.Entries), 3) // dir1, dir2, file.txt
}
