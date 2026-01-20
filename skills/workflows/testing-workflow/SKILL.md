---
name: testing-workflow
description: Comprehensive guide for writing and running tests in Go projects
category: workflow
capabilities: []
---

# Testing Workflow

A complete workflow for creating, running, and maintaining tests in Go projects.

## Overview

This skill teaches you how to write effective tests, run test suites, analyze coverage, and maintain test quality over time.

## Prerequisites

Before using this workflow, you should understand:
- How to read and write files (use `filesystem-operations` skill)
- How to execute commands (use `shell-execution` skill)
- Basic Go testing conventions

## Workflow Steps

### Step 1: Understand the Code to Test

First, read the code you need to test:

```json
{
  "action": "execute",
  "capability": "read_file",
  "parameters": {
    "path": "internal/server/mypackage/myfile.go"
  }
}
```

### Step 2: Check for Existing Tests

Look for existing test files:

```json
{
  "action": "execute",
  "capability": "list_directory",
  "parameters": {
    "path": "internal/server/mypackage",
    "recursive": false
  }
}
```

Test files in Go end with `_test.go`. If a test file exists, read it to understand existing test patterns:

```json
{
  "action": "execute",
  "capability": "read_file",
  "parameters": {
    "path": "internal/server/mypackage/myfile_test.go"
  }
}
```

### Step 3: Write Test File

Create or update the test file. Go test files follow these conventions:

**File naming**: `{filename}_test.go` (e.g., `registry.go` → `registry_test.go`)

**Package naming**: `package mypackage` or `package mypackage_test` for black-box testing

**Test function naming**: `func Test{FunctionName}_{Scenario}(t *testing.T)`

**Example test structure:**

```go
package mypackage

import (
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRegistry(t *testing.T) {
	reg := NewRegistry()
	assert.NotNil(t, reg)
	assert.NotNil(t, reg.items)
}

func TestRegister_Success(t *testing.T) {
	// Arrange
	reg := NewRegistry()
	item := &Item{ID: "test", Name: "Test Item"}
	
	// Act
	err := reg.Register(item)
	
	// Assert
	require.NoError(t, err)
	assert.True(t, reg.Has("test"))
}

func TestRegister_DuplicateError(t *testing.T) {
	// Arrange
	reg := NewRegistry()
	item := &Item{ID: "test", Name: "Test Item"}
	reg.Register(item)
	
	// Act
	err := reg.Register(item)
	
	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}
```

Write the test file:

```json
{
  "action": "execute",
  "capability": "write_file",
  "parameters": {
    "path": "internal/server/mypackage/myfile_test.go",
    "content": "package mypackage\n\nimport (\n\t\"testing\"\n\t...\n)\n\nfunc TestMyFunction(t *testing.T) {\n\t...\n}\n"
  }
}
```

### Step 4: Run Tests

Run the specific test file:

```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {
    "command": "go test ./internal/server/mypackage -v"
  }
}
```

Or run all tests in the project:

```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {
    "command": "go test ./... -v"
  }
}
```

### Step 5: Check Test Coverage

Run tests with coverage report:

```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {
    "command": "go test ./internal/server/mypackage -cover -coverprofile=coverage.out"
  }
}
```

View detailed coverage:

```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {
    "command": "go tool cover -func=coverage.out"
  }
}
```

### Step 6: Fix Failing Tests

If tests fail, read the error output and:

1. **Identify the failure**: Which test failed and why?
2. **Read the code**: Understand what the test expects
3. **Fix the issue**: Either fix the code or fix the test
4. **Re-run tests**: Verify the fix works

## Test Writing Best Practices

### Test Organization

Use **table-driven tests** for testing multiple scenarios:

```go
func TestCalculate(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
		wantErr  bool
	}{
		{"positive number", 5, 25, false},
		{"zero", 0, 0, false},
		{"negative number", -5, 25, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Calculate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
```

### Assertion Libraries

This project uses `testify`:

- `assert.Equal(t, expected, actual)` - Compare values (continues on failure)
- `require.Equal(t, expected, actual)` - Compare values (stops on failure)
- `assert.NoError(t, err)` - Check no error occurred
- `assert.Error(t, err)` - Check error occurred
- `assert.Contains(t, str, substring)` - Check substring exists
- `assert.True(t, condition)` / `assert.False(t, condition)` - Boolean checks
- `assert.Nil(t, value)` / `assert.NotNil(t, value)` - Nil checks

### What to Test

✅ **Do test:**
- Public functions and methods
- Error conditions and edge cases
- Boundary values (0, -1, max values)
- Concurrent access (if applicable)
- Integration between components

❌ **Don't test:**
- Private implementation details (unless critical)
- Third-party library code
- Trivial getters/setters without logic

### Test Naming Conventions

Use descriptive test names that explain:
1. What is being tested
2. Under what conditions
3. What the expected outcome is

**Format**: `Test{Function}_{Condition}_{ExpectedResult}`

Examples:
- `TestRegister_ValidInput_Success`
- `TestRegister_DuplicateID_ReturnsError`
- `TestGet_NonexistentItem_ReturnsNotFoundError`

## Common Test Patterns

### Setup and Teardown

```go
func TestMyFeature(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	defer db.Close() // Cleanup
	
	// Test code here
}
```

### Testing with Context

```go
func TestWithTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	result, err := MyFunction(ctx)
	require.NoError(t, err)
}
```

### Mock Dependencies

```go
type mockService struct {
	called bool
	result string
}

func (m *mockService) DoSomething() string {
	m.called = true
	return m.result
}

func TestWithMock(t *testing.T) {
	mock := &mockService{result: "test"}
	
	service := NewService(mock)
	result := service.Process()
	
	assert.True(t, mock.called)
	assert.Equal(t, "test", result)
}
```

## Running Specific Tests

### Run single test function:
```bash
go test ./internal/server/mypackage -run TestMyFunction
```

### Run tests matching pattern:
```bash
go test ./internal/server/mypackage -run "TestRegister.*"
```

### Run with verbose output:
```bash
go test ./internal/server/mypackage -v
```

### Run with race detector:
```bash
go test ./internal/server/mypackage -race
```

## Coverage Goals

- **Critical code**: 80%+ coverage
- **Business logic**: 70%+ coverage
- **Utilities**: 60%+ coverage
- **Overall project**: 70%+ coverage

Coverage is a guide, not a goal. Focus on testing important paths and edge cases.

## Troubleshooting

### Tests Pass Locally but Fail in CI

- Check for timing issues (use proper waits, not sleeps)
- Verify test isolation (tests shouldn't depend on order)
- Check for hardcoded paths or environment dependencies

### Flaky Tests

- Identify root cause (timing, concurrency, external dependency)
- Add proper synchronization
- Use deterministic test data
- Consider mocking external dependencies

### Slow Tests

- Profile with `go test -cpuprofile`
- Use parallel tests: `t.Parallel()`
- Mock slow external services
- Consider integration vs unit test separation

## Complete Example

Here's a complete workflow for adding tests to a new feature:

```json
// 1. Read the code to understand what to test
{"action": "execute", "capability": "read_file", "parameters": {"path": "internal/server/tool/active_tool.go"}}

// 2. Check if test file exists
{"action": "execute", "capability": "list_directory", "parameters": {"path": "internal/server/tool"}}

// 3. Read existing test patterns
{"action": "execute", "capability": "read_file", "parameters": {"path": "internal/server/tool/registry_test.go"}}

// 4. Write new test file
{"action": "execute", "capability": "write_file", "parameters": {"path": "internal/server/tool/active_tool_test.go", "content": "..."}}

// 5. Run tests
{"action": "execute", "capability": "execute_command", "parameters": {"command": "go test ./internal/server/tool -v"}}

// 6. Check coverage
{"action": "execute", "capability": "execute_command", "parameters": {"command": "go test ./internal/server/tool -cover"}}

// 7. If tests fail, read output, fix, and re-run
{"action": "execute", "capability": "execute_command", "parameters": {"command": "go test ./internal/server/tool -v -run TestActiveTool"}}
```

## Summary

Good tests:
1. **Are focused**: Test one thing at a time
2. **Are isolated**: Don't depend on other tests
3. **Are fast**: Run quickly to encourage frequent use
4. **Are clear**: Easy to understand what's being tested
5. **Are reliable**: Same input → same output

Use this workflow to maintain high code quality and catch bugs early.
