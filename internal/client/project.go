package client

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/adevcorn/ensemble/internal/protocol"
)

// Project represents the user's project context
type Project struct {
	path string
}

// NewProject creates a new project from a path
func NewProject(path string) (*Project, error) {
	// Validate path exists
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("project path does not exist: %w", err)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	return &Project{
		path: absPath,
	}, nil
}

// DetectProject detects the project root from the current directory
func DetectProject() (*Project, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	// Look for project markers
	markers := []string{
		".git",
		"go.mod",
		"package.json",
		"pyproject.toml",
		"requirements.txt",
		"Cargo.toml",
		"pom.xml",
		"build.gradle",
	}

	dir := cwd
	for {
		// Check for any marker
		for _, marker := range markers {
			markerPath := filepath.Join(dir, marker)
			if _, err := os.Stat(markerPath); err == nil {
				return NewProject(dir)
			}
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding markers, use current directory
			return NewProject(cwd)
		}
		dir = parent
	}
}

// GetInfo gathers project information
func (p *Project) GetInfo(ctx context.Context) (*protocol.ProjectInfo, error) {
	info := &protocol.ProjectInfo{
		Path:     p.path,
		Metadata: make(map[string]string),
	}

	// Get git information
	if gitBranch, err := p.getGitBranch(ctx); err == nil {
		info.GitBranch = gitBranch
	}

	if gitRemote, err := p.getGitRemote(ctx); err == nil {
		info.GitRemote = gitRemote
	}

	// Detect project type
	if lang, framework := p.detectProjectType(); lang != "" {
		info.Language = lang
		info.Framework = framework
	}

	return info, nil
}

// Path returns the project path
func (p *Project) Path() string {
	return p.path
}

// getGitBranch gets the current git branch
func (p *Project) getGitBranch(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = p.path

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// getGitRemote gets the git remote URL
func (p *Project) getGitRemote(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "remote", "get-url", "origin")
	cmd.Dir = p.path

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// detectProjectType detects the project language and framework
func (p *Project) detectProjectType() (language string, framework string) {
	// Go
	if _, err := os.Stat(filepath.Join(p.path, "go.mod")); err == nil {
		language = "go"
		// Try to detect framework
		if p.hasFile("go.mod", "github.com/gin-gonic/gin") {
			framework = "gin"
		} else if p.hasFile("go.mod", "github.com/gorilla/mux") {
			framework = "gorilla"
		}
		return
	}

	// JavaScript/TypeScript
	if _, err := os.Stat(filepath.Join(p.path, "package.json")); err == nil {
		language = "javascript"

		// Check for TypeScript
		if _, err := os.Stat(filepath.Join(p.path, "tsconfig.json")); err == nil {
			language = "typescript"
		}

		// Try to detect framework
		if p.hasFile("package.json", "\"react\"") {
			framework = "react"
		} else if p.hasFile("package.json", "\"next\"") {
			framework = "nextjs"
		} else if p.hasFile("package.json", "\"express\"") {
			framework = "express"
		} else if p.hasFile("package.json", "\"vue\"") {
			framework = "vue"
		}
		return
	}

	// Python
	if _, err := os.Stat(filepath.Join(p.path, "pyproject.toml")); err == nil {
		language = "python"
		if p.hasFile("pyproject.toml", "django") {
			framework = "django"
		} else if p.hasFile("pyproject.toml", "flask") {
			framework = "flask"
		} else if p.hasFile("pyproject.toml", "fastapi") {
			framework = "fastapi"
		}
		return
	}

	if _, err := os.Stat(filepath.Join(p.path, "requirements.txt")); err == nil {
		language = "python"
		if p.hasFile("requirements.txt", "Django") {
			framework = "django"
		} else if p.hasFile("requirements.txt", "Flask") {
			framework = "flask"
		} else if p.hasFile("requirements.txt", "fastapi") {
			framework = "fastapi"
		}
		return
	}

	// Rust
	if _, err := os.Stat(filepath.Join(p.path, "Cargo.toml")); err == nil {
		language = "rust"
		if p.hasFile("Cargo.toml", "actix-web") {
			framework = "actix"
		} else if p.hasFile("Cargo.toml", "rocket") {
			framework = "rocket"
		}
		return
	}

	// Java
	if _, err := os.Stat(filepath.Join(p.path, "pom.xml")); err == nil {
		language = "java"
		if p.hasFile("pom.xml", "spring-boot") {
			framework = "spring-boot"
		}
		return
	}

	if _, err := os.Stat(filepath.Join(p.path, "build.gradle")); err == nil {
		language = "java"
		if p.hasFile("build.gradle", "spring-boot") {
			framework = "spring-boot"
		}
		return
	}

	return "", ""
}

// hasFile checks if a file exists and contains a string
func (p *Project) hasFile(filename, content string) bool {
	data, err := os.ReadFile(filepath.Join(p.path, filename))
	if err != nil {
		return false
	}
	return strings.Contains(string(data), content)
}
