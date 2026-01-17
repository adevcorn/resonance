package client

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/adevcorn/ensemble/internal/config"
)

// Checker checks permissions for tool execution
type Checker struct {
	config      *config.PermissionsSettings
	projectPath string
}

// NewChecker creates a new permission checker
func NewChecker(permissions *config.PermissionsSettings, projectPath string) *Checker {
	return &Checker{
		config:      permissions,
		projectPath: projectPath,
	}
}

// CheckFilePath checks if a file path is allowed
func (c *Checker) CheckFilePath(path string) error {
	// Resolve to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// If path is relative, resolve against project path
	if !filepath.IsAbs(path) {
		absPath = filepath.Join(c.projectPath, path)
	}

	// Check if path is within project directory
	relPath, err := filepath.Rel(c.projectPath, absPath)
	if err != nil || strings.HasPrefix(relPath, "..") {
		return fmt.Errorf("access denied: path outside project directory")
	}

	// Check if path is allowed
	if !c.IsPathAllowed(absPath) {
		return fmt.Errorf("access denied: path not allowed")
	}

	return nil
}

// CheckCommand checks if a command is allowed
func (c *Checker) CheckCommand(cmd string, args []string) error {
	// Check if command is allowed
	if !c.IsCommandAllowed(cmd) {
		return fmt.Errorf("access denied: command not allowed: %s", cmd)
	}

	// Check for dangerous patterns
	fullCmd := cmd + " " + strings.Join(args, " ")

	dangerousPatterns := []string{
		"rm -rf /",
		"rm -rf /*",
		":(){ :|:& };:", // Fork bomb
		"mkfs",
		"dd if=/dev/zero",
		"> /dev/sda",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(fullCmd, pattern) {
			return fmt.Errorf("access denied: dangerous command pattern detected")
		}
	}

	return nil
}

// IsPathAllowed checks if a path is allowed
func (c *Checker) IsPathAllowed(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	// Normalize path
	absPath = filepath.Clean(absPath)

	// Check denied paths first
	for _, denied := range c.config.File.DeniedPaths {
		deniedPath := c.resolvePath(denied)
		if c.matchesPath(absPath, deniedPath) {
			return false
		}
	}

	// Check allowed paths
	if len(c.config.File.AllowedPaths) == 0 {
		// If no allowed paths specified, allow project directory
		relPath, err := filepath.Rel(c.projectPath, absPath)
		return err == nil && !strings.HasPrefix(relPath, "..")
	}

	for _, allowed := range c.config.File.AllowedPaths {
		allowedPath := c.resolvePath(allowed)
		if c.matchesPath(absPath, allowedPath) {
			return true
		}
	}

	return false
}

// IsCommandAllowed checks if a command is allowed
func (c *Checker) IsCommandAllowed(cmd string) bool {
	// Extract base command (without path)
	baseCmd := filepath.Base(cmd)

	// Check denied commands first
	for _, denied := range c.config.Exec.DeniedCommands {
		if c.matchesCommand(baseCmd, denied) {
			return false
		}
	}

	// Check allowed commands
	if len(c.config.Exec.AllowedCommands) == 0 {
		// If no allowed commands specified, deny by default for security
		return false
	}

	for _, allowed := range c.config.Exec.AllowedCommands {
		if c.matchesCommand(baseCmd, allowed) {
			return true
		}
	}

	return false
}

// resolvePath resolves a path pattern relative to project
func (c *Checker) resolvePath(path string) string {
	if path == "." {
		return c.projectPath
	}

	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}

	return filepath.Clean(filepath.Join(c.projectPath, path))
}

// matchesPath checks if a path matches a pattern
func (c *Checker) matchesPath(path, pattern string) bool {
	// Clean both paths
	path = filepath.Clean(path)
	pattern = filepath.Clean(pattern)

	// Exact match
	if path == pattern {
		return true
	}

	// Check if path is under pattern directory
	relPath, err := filepath.Rel(pattern, path)
	if err != nil {
		return false
	}

	// Path is under pattern if relative path doesn't start with ..
	return !strings.HasPrefix(relPath, "..")
}

// matchesCommand checks if a command matches a pattern
func (c *Checker) matchesCommand(cmd, pattern string) bool {
	// Extract base commands
	baseCmd := filepath.Base(cmd)
	basePattern := filepath.Base(pattern)

	// Simple wildcard matching
	if basePattern == "*" {
		return true
	}

	// Exact match
	if baseCmd == basePattern {
		return true
	}

	// Check if pattern contains the command (for patterns like "rm -rf")
	if strings.Contains(pattern, " ") {
		return strings.HasPrefix(cmd+" ", pattern+" ")
	}

	return false
}
