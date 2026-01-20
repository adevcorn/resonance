---
name: code-review-workflow
description: Systematic process for conducting thorough and constructive code reviews
category: workflow
capabilities: []
---

# Code Review Workflow

A comprehensive workflow for reviewing code changes systematically and providing actionable feedback.

## Overview

This skill guides you through conducting effective code reviews that:
- Catch bugs and design issues early
- Improve code quality and maintainability
- Share knowledge across the team
- Ensure consistency with project standards

## Prerequisites

You should understand how to:
- Read files and explore directories (use `filesystem-operations` skill)
- View git changes (use `git-workflow` skill)
- Run tests (use `testing-workflow` skill)

## Code Review Process

### Step 1: Understand the Context

**Read the change description:**

```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {
    "command": "git log -1 --pretty=format:'%B'"
  }
}
```

**View the list of changed files:**

```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {
    "command": "git diff --name-status HEAD~1"
  }
}
```

**See the summary of changes:**

```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {
    "command": "git diff --stat HEAD~1"
  }
}
```

### Step 2: Review the Changes

**View the diff:**

```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {
    "command": "git diff HEAD~1"
  }
}
```

**Read specific changed files:**

```json
{
  "action": "execute",
  "capability": "read_file",
  "parameters": {
    "path": "internal/server/tool/active_tool.go"
  }
}
```

**Check surrounding context:**

For understanding the bigger picture, read related files:

```json
{
  "action": "execute",
  "capability": "read_file",
  "parameters": {
    "path": "internal/server/tool/registry.go"
  }
}
```

### Step 3: Check Code Quality

**Run all tests:**

```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {
    "command": "go test ./... -v"
  }
}
```

**Check test coverage:**

```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {
    "command": "go test ./... -cover"
  }
}
```

**Run linters and static analysis:**

```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {
    "command": "go vet ./..."
  }
}
```

**Build the project:**

```json
{
  "action": "execute",
  "capability": "execute_command",
  "parameters": {
    "command": "go build ./..."
  }
}
```

## Review Checklist

### Correctness

- [ ] Does the code do what it's supposed to do?
- [ ] Are there any logic errors or edge cases missed?
- [ ] Are error conditions handled properly?
- [ ] Are nil checks in place where needed?
- [ ] Are there any potential panics?
- [ ] Are there race conditions in concurrent code?

### Design

- [ ] Is the approach appropriate for the problem?
- [ ] Does it follow project architecture patterns?
- [ ] Are abstractions at the right level?
- [ ] Is the code modular and reusable?
- [ ] Are dependencies managed correctly?
- [ ] Does it violate SOLID principles?

### Testing

- [ ] Are there tests for new functionality?
- [ ] Do tests cover edge cases?
- [ ] Are tests clear and maintainable?
- [ ] Is test coverage adequate?
- [ ] Do all tests pass?
- [ ] Are there integration tests if needed?

### Readability

- [ ] Are names clear and descriptive?
- [ ] Is the code self-documenting?
- [ ] Are complex parts commented?
- [ ] Is formatting consistent?
- [ ] Is the code easy to understand?
- [ ] Are there any confusing constructs?

### Performance

- [ ] Are there obvious performance issues?
- [ ] Are there unnecessary allocations?
- [ ] Are expensive operations cached if appropriate?
- [ ] Are database queries optimized?
- [ ] Is there potential for memory leaks?

### Security

- [ ] Is input validated?
- [ ] Are there SQL injection risks?
- [ ] Is authentication/authorization correct?
- [ ] Are secrets handled securely?
- [ ] Are there XSS or CSRF vulnerabilities?
- [ ] Is data sanitized before output?

### Maintainability

- [ ] Will this be easy to change later?
- [ ] Are magic numbers avoided?
- [ ] Are there hardcoded values that should be configurable?
- [ ] Is error handling consistent?
- [ ] Is logging appropriate?
- [ ] Is the code DRY (not repetitive)?

## Common Code Smells

### Go-Specific Issues

**1. Error handling**

❌ **Bad:**
```go
result, _ := doSomething()
```

✅ **Good:**
```go
result, err := doSomething()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}
```

**2. Nil pointer checks**

❌ **Bad:**
```go
func process(data *Data) {
    value := data.Field // Panic if data is nil
}
```

✅ **Good:**
```go
func process(data *Data) error {
    if data == nil {
        return fmt.Errorf("data cannot be nil")
    }
    value := data.Field // Safe
}
```

**3. Goroutine leaks**

❌ **Bad:**
```go
go func() {
    // Never returns, leaks goroutine
    for {
        select {
        case msg := <-ch:
            process(msg)
        }
    }
}()
```

✅ **Good:**
```go
go func() {
    for {
        select {
        case msg := <-ch:
            process(msg)
        case <-ctx.Done():
            return // Clean exit
        }
    }
}()
```

**4. Mutex usage**

❌ **Bad:**
```go
mu.Lock()
if err != nil {
    return err // Lock not released!
}
mu.Unlock()
```

✅ **Good:**
```go
mu.Lock()
defer mu.Unlock()
if err != nil {
    return err // Lock released by defer
}
```

**5. Context usage**

❌ **Bad:**
```go
func doWork() {
    // No way to cancel long-running operation
    time.Sleep(10 * time.Minute)
}
```

✅ **Good:**
```go
func doWork(ctx context.Context) error {
    select {
    case <-time.After(10 * time.Minute):
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

### General Issues

**1. Deep nesting**

❌ **Bad:**
```go
if a {
    if b {
        if c {
            if d {
                // Too deeply nested
            }
        }
    }
}
```

✅ **Good (early returns):**
```go
if !a {
    return
}
if !b {
    return
}
if !c {
    return
}
if !d {
    return
}
// Main logic at top level
```

**2. Large functions**

Functions over 50 lines are candidates for refactoring. Look for:
- Logical sections that can be extracted
- Repeated code patterns
- Single Responsibility Principle violations

**3. God objects**

Structs with too many responsibilities should be split:

```go
// Before: Does too much
type Service struct {
    db       *sql.DB
    cache    *Cache
    logger   *Logger
    metrics  *Metrics
    // ... 20 more fields
}

// After: Separated concerns
type UserService struct {
    repo   *UserRepository
    events *EventPublisher
}
```

## Providing Feedback

### Feedback Principles

1. **Be constructive**: Suggest improvements, don't just criticize
2. **Be specific**: Point to exact lines and explain the issue
3. **Prioritize**: Distinguish between must-fix and nice-to-have
4. **Educate**: Explain why something is an issue
5. **Recognize good work**: Call out clever solutions

### Feedback Template

**Critical issues (must fix):**
```
❌ CRITICAL: Nil pointer dereference at line 42

The code assumes `data` is never nil, but it can be nil when called from X.

Suggested fix:
```go
if data == nil {
    return fmt.Errorf("data cannot be nil")
}
```

**Minor improvements (optional):**
```
💡 SUGGESTION: Consider using a switch statement (line 15-30)

The if-else chain could be more readable as a switch:
```go
switch status {
case StatusPending:
    // ...
case StatusComplete:
    // ...
}
```

**Positive feedback:**
```
✅ NICE: Great use of table-driven tests! This makes it easy to add new test cases.
```

### Priority Levels

**🔴 P0 - Blocker:**
- Security vulnerabilities
- Data loss risks
- Crashes or panics
- Breaking changes without migration

**🟡 P1 - Important:**
- Logic errors
- Missing error handling
- Performance issues
- Missing tests

**🟢 P2 - Nice to have:**
- Style inconsistencies
- Minor improvements
- Documentation updates
- Code organization

## Review Workflow Example

Complete example of reviewing a pull request:

```json
// 1. Get the commit message and understand the change
{"action": "execute", "capability": "execute_command", "parameters": {"command": "git log -1 --pretty=format:'%B'"}}

// 2. See what files changed
{"action": "execute", "capability": "execute_command", "parameters": {"command": "git diff --name-only HEAD~1"}}

// 3. View the diff
{"action": "execute", "capability": "execute_command", "parameters": {"command": "git diff HEAD~1"}}

// 4. Read the main changed file
{"action": "execute", "capability": "read_file", "parameters": {"path": "internal/server/tool/active_tool.go"}}

// 5. Check if tests were added
{"action": "execute", "capability": "read_file", "parameters": {"path": "internal/server/tool/active_tool_test.go"}}

// 6. Run the tests
{"action": "execute", "capability": "execute_command", "parameters": {"command": "go test ./internal/server/tool -v"}}

// 7. Check test coverage
{"action": "execute", "capability": "execute_command", "parameters": {"command": "go test ./internal/server/tool -cover"}}

// 8. Run static analysis
{"action": "execute", "capability": "execute_command", "parameters": {"command": "go vet ./..."}}

// 9. Build to ensure it compiles
{"action": "execute", "capability": "execute_command", "parameters": {"command": "go build ./..."}}

// 10. Read related files for context
{"action": "execute", "capability": "read_file", "parameters": {"path": "internal/server/capability/registry.go"}}
```

## Post-Review Actions

After completing the review:

1. **Summarize findings**: Create a clear summary of issues found
2. **Categorize by priority**: Group feedback by severity
3. **Suggest next steps**: Clear action items for the author
4. **Offer to help**: Be available for questions

### Example Review Summary

```markdown
## Code Review Summary

### Overview
Reviewed changes to active_tool implementation. Overall good work with some areas for improvement.

### Critical Issues (Must Fix)
1. ❌ Line 42: Nil pointer check needed for skillRegistry
2. ❌ Line 78: Error not propagated to caller

### Important Issues (Should Fix)
1. ⚠️  Missing test coverage for error paths
2. ⚠️  Line 120: Consider using context with timeout

### Suggestions (Nice to Have)
1. 💡 Line 55: Could simplify with early return
2. 💡 Consider adding examples in comments

### Positive Highlights
1. ✅ Excellent test organization with table-driven tests
2. ✅ Clear error messages with good context
3. ✅ Well-documented public API

### Next Steps
1. Fix critical issues (nil checks and error handling)
2. Add tests for error paths
3. Address important issues if time permits

Happy to discuss any of these points!
```

## Tips for Effective Reviews

1. **Review in small batches**: Don't try to review 1000+ lines at once
2. **Focus on important issues**: Don't nitpick style if there are bigger problems
3. **Run the code locally**: Understanding is better than guessing
4. **Check related files**: Context matters
5. **Look for patterns**: Repeated issues might indicate a systemic problem
6. **Trust but verify**: Tests passing doesn't mean the code is correct
7. **Be timely**: Review soon after changes are submitted
8. **Be thorough but kind**: Quality feedback delivered respectfully

## Summary

Effective code reviews:

1. **Understand** the change and its context
2. **Check** for correctness, design, tests, and quality
3. **Test** that changes work as intended
4. **Provide** specific, prioritized, constructive feedback
5. **Recognize** good work and help the team improve

Use this systematic workflow to conduct reviews that improve code quality while supporting team growth.
