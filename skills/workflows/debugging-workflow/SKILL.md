---
name: debugging-workflow
description: Systematic approach to debugging code issues and finding root causes
category: workflow
capabilities: []
---

# Debugging Workflow

A systematic workflow for identifying, diagnosing, and fixing bugs in Go applications.

## Overview

This skill provides a structured approach to debugging that helps you:
- Reproduce issues reliably
- Identify root causes efficiently
- Fix bugs without introducing new ones
- Prevent similar issues in the future

## Prerequisites

You should know how to:
- Read and write files (use `filesystem-operations` skill)
- Execute commands and tests (use `shell-execution` skill)
- Run tests (use `testing-workflow` skill)

## The Debugging Process

### Step 1: Reproduce the Issue

**Gather Information:**

Read error reports, logs, or issue descriptions:

```json
{
  "action": "execute",
  "capability": "read_file",
  "parameters": {
    "path": "logs/error.log"
  }
}
```

**Create a Minimal Reproduction:**

1. Identify the exact steps to trigger the bug
2. Create a test that reproduces the issue
3. Verify the test fails consistently

```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {
    "command": "go test ./internal/server/mypackage -run TestBuggyFunction -v"
  }
}
```

### Step 2: Locate the Problem

**Read the Stack Trace:**

Analyze error messages and stack traces to identify:
- Which function failed
- What the error message says
- Where in the code the error occurred

**Example stack trace analysis:**
```
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x1234567]

goroutine 1 [running]:
main.processData(0x0, 0x0)
    /path/to/file.go:42 +0x85
```

This tells us: nil pointer dereference at `file.go:42` in `processData` function.

**Navigate to the problematic code:**

```json
{
  "action": "execute",
  "capability": "read_file",
  "parameters": {
    "path": "path/to/file.go"
  }
}
```

### Step 3: Understand the Context

**Read surrounding code:**

```json
{
  "action": "execute",
  "capability": "read_file",
  "parameters": {
    "path": "internal/server/mypackage/buggy.go"
  }
}
```

**Check function callers:**

Use `grep` to find where the buggy function is called:

```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {
    "command": "grep -r \"BuggyFunction\" internal/"
  }
}
```

**Review recent changes:**

```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {
    "command": "git log -p --follow -- internal/server/mypackage/buggy.go"
  }
}
```

### Step 4: Form a Hypothesis

Based on the information gathered:

1. **What do you think is happening?**
   - "The function receives a nil pointer when X condition occurs"
   - "The variable is not initialized before use"
   - "There's a race condition between goroutines"

2. **Why might this be happening?**
   - "Function assumes input is never nil"
   - "Initialization happens conditionally"
   - "Shared state without synchronization"

3. **How can you test this hypothesis?**
   - Add logging
   - Add assertions
   - Write a focused test

### Step 5: Test Your Hypothesis

**Add Debugging Output:**

Temporarily add print statements or logging:

```go
func processData(data *Data) error {
    fmt.Printf("DEBUG: data is nil: %v\n", data == nil)
    if data == nil {
        return fmt.Errorf("data is nil")
    }
    // ... rest of function
}
```

**Run with added debugging:**

```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {
    "command": "go test ./internal/server/mypackage -run TestBuggyFunction -v"
  }
}
```

**Use Go's built-in debugging tools:**

```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {
    "command": "go test -v -run TestBuggyFunction ./... 2>&1 | grep -A 10 -B 10 ERROR"
  }
}
```

### Step 6: Fix the Bug

**Implement the fix:**

```json
{
  "action": "execute",
  "capability": "read_file",
  "parameters": {
    "path": "internal/server/mypackage/buggy.go"
  }
}
```

Read the file, identify the fix, then write the corrected version:

```json
{
  "action": "execute",
  "capability": "write_file",
  "parameters": {
    "path": "internal/server/mypackage/buggy.go",
    "content": "package mypackage\n\n// Fixed version...\n"
  }
}
```

### Step 7: Verify the Fix

**Run the failing test:**

```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {
    "command": "go test ./internal/server/mypackage -run TestBuggyFunction -v"
  }
}
```

**Run the full test suite:**

```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {
    "command": "go test ./... -v"
  }
}
```

**Clean up debug code:**

Remove any temporary logging or debug statements before committing.

### Step 8: Prevent Future Occurrences

**Add a regression test:**

Write a test that specifically covers this bug case:

```go
func TestBugFix_NilPointerInProcessData(t *testing.T) {
    // This test ensures we don't regress on the nil pointer bug
    err := processData(nil)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "data is nil")
}
```

**Document the fix:**

Add comments explaining the edge case:

```go
func processData(data *Data) error {
    // Guard against nil input - can happen when called from X
    if data == nil {
        return fmt.Errorf("data cannot be nil")
    }
    // ... rest of function
}
```

## Common Debugging Patterns

### Nil Pointer Dereferences

**Symptom:** `panic: runtime error: invalid memory address`

**Debug approach:**
1. Find which variable is nil
2. Trace back to where it should have been initialized
3. Add nil checks or ensure proper initialization

**Fix pattern:**
```go
// Before
func process(data *Data) {
    result := data.Field // PANIC if data is nil
}

// After
func process(data *Data) error {
    if data == nil {
        return fmt.Errorf("data cannot be nil")
    }
    result := data.Field // Safe
}
```

### Race Conditions

**Symptom:** Intermittent failures, different results in different runs

**Debug approach:**
1. Run tests with race detector:
   ```bash
   go test -race ./...
   ```
2. Identify unsynchronized shared state
3. Add proper synchronization

**Fix pattern:**
```go
// Before (race condition)
type Counter struct {
    count int
}

func (c *Counter) Increment() {
    c.count++ // Unsafe with concurrent access
}

// After (thread-safe)
type Counter struct {
    mu    sync.Mutex
    count int
}

func (c *Counter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++ // Safe
}
```

### Logic Errors

**Symptom:** Wrong output, incorrect behavior

**Debug approach:**
1. Add print statements at key points
2. Compare expected vs actual values
3. Trace logic flow step by step

**Debugging technique:**
```go
func calculate(x, y int) int {
    fmt.Printf("DEBUG: calculate(%d, %d)\n", x, y)
    
    result := x + y * 2
    fmt.Printf("DEBUG: result before adjustment: %d\n", result)
    
    if x > 10 {
        result = result / 2
        fmt.Printf("DEBUG: result after division: %d\n", result)
    }
    
    fmt.Printf("DEBUG: final result: %d\n", result)
    return result
}
```

### Off-by-One Errors

**Symptom:** Index out of range, incorrect loop iterations

**Debug approach:**
1. Check loop conditions carefully
2. Verify boundary values
3. Test with edge cases (empty, single item, max size)

**Common mistakes:**
```go
// Wrong: Skips last element
for i := 0; i < len(items)-1; i++ {
    process(items[i])
}

// Correct: Processes all elements
for i := 0; i < len(items); i++ {
    process(items[i])
}

// Or use range
for _, item := range items {
    process(item)
}
```

### Error Handling Issues

**Symptom:** Errors silently ignored, unexpected behavior

**Debug approach:**
1. Search for unchecked errors
2. Verify error propagation
3. Ensure errors are logged

**Pattern to find unchecked errors:**
```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {
    "command": "go vet ./..."
  }
}
```

**Fix pattern:**
```go
// Before (error ignored)
result, _ := doSomething()

// After (error handled)
result, err := doSomething()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}
```

## Debugging Commands Reference

### Compile with Debug Info
```bash
go build -gcflags="all=-N -l" ./cmd/myapp
```

### Run Tests with Verbose Output
```bash
go test -v ./internal/server/mypackage
```

### Run Single Failing Test
```bash
go test -v -run TestSpecificFunction ./internal/server/mypackage
```

### Run with Race Detector
```bash
go test -race ./...
```

### Get Stack Traces on Panic
```bash
GOTRACEBACK=all go test ./...
```

### Profile CPU Usage
```bash
go test -cpuprofile cpu.prof ./internal/server/mypackage
go tool pprof cpu.prof
```

### Profile Memory Usage
```bash
go test -memprofile mem.prof ./internal/server/mypackage
go tool pprof mem.prof
```

### Check for Vet Issues
```bash
go vet ./...
```

### Static Analysis
```bash
staticcheck ./...
```

## Debugging Checklist

When stuck on a bug, systematically check:

- [ ] Can you reproduce it reliably?
- [ ] Do you understand the error message?
- [ ] Have you read the relevant code?
- [ ] Have you checked recent changes (git log)?
- [ ] Have you added logging/debugging output?
- [ ] Have you tested your hypothesis?
- [ ] Have you checked for nil pointers?
- [ ] Have you run with -race detector?
- [ ] Have you verified inputs and outputs?
- [ ] Have you tested edge cases?
- [ ] Have you checked error handling?
- [ ] Have you read the test output carefully?

## Complete Debugging Example

Here's a full debugging session:

```json
// 1. Reproduce: Run failing test
{"action": "execute", "capability": "execute_command", "parameters": {"command": "go test ./internal/server/tool -run TestActiveTool_Execute -v"}}

// 2. Read the error output (from previous command)
// Error: "panic: runtime error: nil pointer dereference"

// 3. Read the source file
{"action": "execute", "capability": "read_file", "parameters": {"path": "internal/server/tool/active_tool.go"}}

// 4. Search for recent changes
{"action": "execute", "capability": "execute_command", "parameters": {"command": "git log -p --follow -n 5 -- internal/server/tool/active_tool.go"}}

// 5. Read the test to understand what it's testing
{"action": "execute", "capability": "read_file", "parameters": {"path": "internal/server/tool/active_tool_test.go"}}

// 6. Found the issue: skillRegistry is nil when active_tool is created
// 7. Fix: Update the code to check for nil

{"action": "execute", "capability": "write_file", "parameters": {"path": "internal/server/tool/active_tool.go", "content": "... fixed code ..."}}

// 8. Verify the fix
{"action": "execute", "capability": "execute_command", "parameters": {"command": "go test ./internal/server/tool -run TestActiveTool_Execute -v"}}

// 9. Run full test suite to ensure no regressions
{"action": "execute", "capability": "execute_command", "parameters": {"command": "go test ./... -v"}}
```

## Tips for Effective Debugging

1. **Read error messages carefully** - They often tell you exactly what's wrong
2. **Reproduce first** - If you can't reproduce it, you can't fix it
3. **Change one thing at a time** - Multiple changes make it hard to know what fixed it
4. **Use version control** - Commit before debugging so you can revert
5. **Take breaks** - Fresh eyes often spot issues immediately
6. **Ask for help** - Collaborate with other agents if stuck
7. **Document your findings** - Help future debuggers (including yourself)

## Summary

Effective debugging is systematic:

1. **Reproduce** the issue reliably
2. **Locate** the problematic code
3. **Understand** what's happening
4. **Hypothesize** the root cause
5. **Test** your hypothesis
6. **Fix** the bug
7. **Verify** the fix works
8. **Prevent** future occurrences with tests

Use this workflow to debug efficiently and learn from each bug you encounter.
