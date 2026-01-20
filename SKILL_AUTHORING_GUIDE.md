# Skill Authoring Guide

A comprehensive guide for creating new skills in the Resonance multi-agent system.

## What is a Skill?

A **skill** is a self-contained markdown document that teaches agents how to perform tasks using the system's capabilities. Skills are discovered dynamically through search and loaded just-in-time when needed.

## Skill Types

### 1. Capability Skills (`skills/capabilities/`)

**Purpose**: Teach agents how to use individual low-level capabilities.

**Characteristics:**
- Document one capability domain (e.g., filesystem, shell, web)
- Focus on parameter formats and usage patterns
- Provide concrete examples for each capability
- List all related capabilities in frontmatter

**Example**: `filesystem-operations` teaches `read_file`, `write_file`, `list_directory`

### 2. Workflow Skills (`skills/workflows/`)

**Purpose**: Teach agents multi-step processes combining multiple capabilities.

**Characteristics:**
- Empty capabilities list in frontmatter (or reference other skills' capabilities)
- Step-by-step procedures
- Best practices and common patterns
- Link to capability skills for low-level operations

**Example**: `testing-workflow` teaches how to write tests, run test suites, analyze coverage

### 3. Integration Skills (`skills/integrations/`)

**Purpose**: Teach agents how to work with external systems and APIs.

**Characteristics:**
- API authentication and usage patterns
- Service-specific conventions
- Common operations and workflows
- Error handling for external services

**Example**: Future skills like `github-api`, `slack-integration`

## Skill Structure

### Directory Layout

```
skills/
├── capabilities/
│   └── my-capability/
│       ├── SKILL.md           # Required: Main skill content
│       ├── scripts/           # Optional: Helper scripts
│       ├── references/        # Optional: Reference materials
│       └── assets/            # Optional: Supporting files
├── workflows/
│   └── my-workflow/
│       └── SKILL.md
└── integrations/
    └── my-integration/
        └── SKILL.md
```

### SKILL.md Format

Every skill **must** have a `SKILL.md` file with this structure:

```markdown
---
name: skill-name
description: One-sentence description of what this skill teaches
category: capability|workflow|integration
capabilities:
  - capability_name_1
  - capability_name_2
---

# Skill Title

## Overview
[High-level description of what this skill teaches]

## Prerequisites
[What agents need to know before using this skill]

## [Main Content Sections]
[Step-by-step instructions, examples, best practices]

## Summary
[Key takeaways]
```

## YAML Frontmatter Fields

### Required Fields

**`name`** (string, required)
- Unique identifier for the skill
- Use lowercase with hyphens: `my-skill-name`
- Should match the directory name
- Used in `load_skill` action

**`description`** (string, required)
- One-sentence summary (< 100 characters)
- Used in search results
- Should include key searchable terms
- Example: "Read and write files on the filesystem"

**`category`** (string, required)
- Must be one of: `capability`, `workflow`, `integration`
- Affects search ranking and organization
- Helps agents understand skill type

**`capabilities`** (array of strings, required)
- List of capability names this skill teaches
- For capability skills: list all capabilities covered
- For workflow skills: empty array `[]` or list referenced capabilities
- Used for skill search and discovery

### Optional Fields

**`metadata`** (object, optional)
- Custom key-value pairs for additional information
- Examples: `author`, `version`, `updated`, `tags`

**Example frontmatter:**

```yaml
---
name: filesystem-operations
description: Read and write files, list directories on the filesystem
category: capability
capabilities:
  - read_file
  - write_file
  - list_directory
metadata:
  author: ensemble
  version: 1.0.0
  updated: 2025-01-20
---
```

## Content Guidelines

### 1. Structure

**Use clear hierarchy:**
```markdown
# Main Title (H1 - once per skill)

## Overview (H2)
[Introduction and purpose]

## Prerequisites (H2)
[What to know before starting]

## Main Sections (H2)
[Core content]

### Subsections (H3)
[Detailed topics]

#### Details (H4)
[Fine points]
```

**Order sections logically:**
1. Overview (what and why)
2. Prerequisites (what you need to know)
3. How to use (step-by-step)
4. Examples (concrete demonstrations)
5. Best practices (do's and don'ts)
6. Troubleshooting (common issues)
7. Summary (key takeaways)

### 2. Examples

**Always include JSON examples** showing active_tool usage:

```markdown
## How to Read a File

Use the `read_file` capability:

\`\`\`json
{
  "action": "execute",
  "capability": "read_file",
  "parameters": {
    "path": "README.md"
  }
}
\`\`\`
```

**Show the full 3-step workflow** where appropriate:

```markdown
## Complete Workflow

### Step 1: Search for skills
\`\`\`json
{
  "action": "search_skills",
  "query": "read files",
  "max_results": 3
}
\`\`\`

### Step 2: Load this skill
\`\`\`json
{
  "action": "load_skill",
  "skill_name": "filesystem-operations"
}
\`\`\`

### Step 3: Execute capability
\`\`\`json
{
  "action": "execute",
  "capability": "read_file",
  "parameters": {
    "path": "go.mod"
  }
}
\`\`\`
```

### 3. Parameter Documentation

**Document all parameters** for each capability:

```markdown
## read_file Parameters

| Parameter | Type   | Required | Description                    |
|-----------|--------|----------|--------------------------------|
| path      | string | Yes      | File path to read (absolute or relative) |
| encoding  | string | No       | File encoding (default: utf-8) |

**Example:**
\`\`\`json
{
  "action": "execute",
  "capability": "read_file",
  "parameters": {
    "path": "src/main.go"
  }
}
\`\`\`
```

### 4. Error Handling

**Document common errors** and how to handle them:

```markdown
## Common Errors

### File Not Found
If the file doesn't exist, you'll receive:
\`\`\`json
{
  "error": "file not found: /path/to/file"
}
\`\`\`

**Solution**: Verify the path exists using `list_directory` first.

### Permission Denied
If you don't have read access:
\`\`\`json
{
  "error": "permission denied: /path/to/file"
}
\`\`\`

**Solution**: Check file permissions or use a different file.
```

### 5. Best Practices

**Include do's and don'ts:**

```markdown
## Best Practices

✅ **Do:**
- Check if files exist before reading
- Use relative paths when possible
- Handle errors appropriately
- Validate file paths

❌ **Don't:**
- Hardcode absolute paths
- Ignore error messages
- Read extremely large files without consideration
- Use `write_file` without reading first (for updates)
```

### 6. Cross-References

**Link to related skills:**

```markdown
## Prerequisites

Before using this skill, you should understand:
- How to search for skills (see main `active_tool` documentation)
- Basic file path concepts

## Related Skills

- `git-workflow` - For reading files from version control
- `testing-workflow` - For reading and writing test files
```

## Writing Style

### Voice and Tone

- **Direct and instructional**: "Use this capability to..." not "You might want to..."
- **Active voice**: "Execute the command" not "The command should be executed"
- **Present tense**: "The function returns" not "The function will return"
- **Concise**: Get to the point quickly

### Formatting

**Use formatting for clarity:**
- **Bold** for emphasis and important terms
- `Code formatting` for capabilities, parameters, values
- > Blockquotes for notes and warnings
- - Bullet lists for options and features
- 1. Numbered lists for sequential steps

**Code blocks** should specify the language:
````markdown
```json
{"action": "execute"}
```

```go
func example() {}
```

```bash
go test ./...
```
````

### Search Optimization

**Include searchable keywords:**
- Use terms agents will likely search for
- Include synonyms and related concepts
- Put important keywords in the description
- Use keywords naturally in headers and content

**Example:**
```markdown
---
description: Read and write files, list directories on the filesystem
---

# Filesystem Operations

Learn how to read files, write files, list directories, and manage
file content using the filesystem capabilities.
```

This makes the skill discoverable via searches for "read", "write", "files", "directories", "filesystem", etc.

## Testing Your Skill

### 1. Verify YAML Frontmatter

```bash
# Check that YAML is valid
head -20 skills/workflows/my-skill/SKILL.md | grep -A 10 "^---"
```

### 2. Test Skill Loading

```bash
# Start server and check logs for skill loading
./bin/ensemble-server

# Look for:
# INF Loaded skill skill=my-skill-name
```

### 3. Test Skill Search

Start the server and use an agent to search for your skill:

```json
{
  "action": "search_skills",
  "query": "keywords from your skill description"
}
```

Verify your skill appears in the results with a good relevance score.

### 4. Test Skill Loading

```json
{
  "action": "load_skill",
  "skill_name": "my-skill-name"
}
```

Verify the full content is returned correctly.

### 5. Test with Real Agent

Run a task that would benefit from your skill and see if agents discover and use it.

## Checklist for New Skills

Before submitting a new skill, verify:

**Frontmatter:**
- [ ] `name` matches directory name
- [ ] `description` is clear and under 100 chars
- [ ] `category` is set correctly
- [ ] `capabilities` lists all relevant capabilities

**Content:**
- [ ] Has clear Overview section
- [ ] Lists Prerequisites if any
- [ ] Includes step-by-step instructions
- [ ] Has JSON examples for all capabilities
- [ ] Documents all parameters
- [ ] Includes error handling guidance
- [ ] Has best practices section
- [ ] Ends with clear Summary

**Quality:**
- [ ] All examples are valid JSON
- [ ] All capability names are correct
- [ ] All links and references work
- [ ] Markdown renders correctly
- [ ] No typos or grammatical errors
- [ ] Examples match the project's actual capabilities

**Testing:**
- [ ] Skill loads without errors
- [ ] Appears in search results for relevant queries
- [ ] Content is clear and actionable
- [ ] Examples can be executed successfully

## Example: Creating a New Capability Skill

Let's create a skill for a hypothetical `database-operations` capability:

### Step 1: Create Directory

```bash
mkdir -p skills/capabilities/database-operations
```

### Step 2: Create SKILL.md

```markdown
---
name: database-operations
description: Query and update database records using SQL
category: capability
capabilities:
  - query_database
  - execute_sql
  - get_schema
---

# Database Operations

Learn how to interact with databases: query records, execute SQL statements, and inspect database schemas.

## Overview

This skill teaches you how to use database capabilities to:
- Query data using SQL SELECT statements
- Execute INSERT, UPDATE, DELETE operations
- Get database schema information

## Prerequisites

You should understand:
- Basic SQL syntax
- How to use active_tool (search, load, execute)
- Database connection concepts

## Capabilities

### query_database

Execute SELECT queries and retrieve results.

**Parameters:**

| Parameter | Type   | Required | Description |
|-----------|--------|----------|-------------|
| query     | string | Yes      | SQL SELECT statement |
| database  | string | No       | Database name (default: main) |

**Example:**

\`\`\`json
{
  "action": "execute",
  "capability": "query_database",
  "parameters": {
    "query": "SELECT * FROM users WHERE active = true",
    "database": "production"
  }
}
\`\`\`

**Response:**

\`\`\`json
{
  "rows": [
    {"id": 1, "name": "Alice", "active": true},
    {"id": 2, "name": "Bob", "active": true}
  ],
  "count": 2
}
\`\`\`

### execute_sql

Execute INSERT, UPDATE, DELETE, or DDL statements.

**Parameters:**

| Parameter | Type   | Required | Description |
|-----------|--------|----------|-------------|
| sql       | string | Yes      | SQL statement to execute |
| database  | string | No       | Database name (default: main) |

**Example:**

\`\`\`json
{
  "action": "execute",
  "capability": "execute_sql",
  "parameters": {
    "sql": "UPDATE users SET active = false WHERE id = 5"
  }
}
\`\`\`

**Response:**

\`\`\`json
{
  "rows_affected": 1,
  "success": true
}
\`\`\`

### get_schema

Get schema information for tables.

**Parameters:**

| Parameter | Type   | Required | Description |
|-----------|--------|----------|-------------|
| table     | string | No       | Table name (omit for all tables) |
| database  | string | No       | Database name (default: main) |

**Example:**

\`\`\`json
{
  "action": "execute",
  "capability": "get_schema",
  "parameters": {
    "table": "users"
  }
}
\`\`\`

## Best Practices

✅ **Do:**
- Use parameterized queries when possible
- Validate input data before INSERT/UPDATE
- Check schema before writing queries
- Use transactions for multiple related changes
- Handle connection errors gracefully

❌ **Don't:**
- Use string concatenation for user input (SQL injection risk)
- Execute destructive operations without confirmation
- Ignore error messages
- Make schema assumptions without verification

## Common Errors

### Connection Failed

\`\`\`json
{"error": "database connection failed: timeout"}
\`\`\`

**Solution**: Check database is running and accessible.

### Invalid SQL Syntax

\`\`\`json
{"error": "SQL syntax error near 'FORM'"}
\`\`\`

**Solution**: Review SQL syntax, it should be FROM not FORM.

## Summary

Use database capabilities to:
1. **Query data**: Use `query_database` for SELECT statements
2. **Modify data**: Use `execute_sql` for INSERT/UPDATE/DELETE
3. **Inspect schema**: Use `get_schema` to understand table structure

Always validate queries and handle errors appropriately.
```

### Step 3: Test

```bash
# Start server
./bin/ensemble-server

# Verify in logs:
# INF Loaded skill skill=database-operations
```

## Tips for Great Skills

1. **Start with examples**: Show before explaining
2. **Be comprehensive**: Cover all common use cases
3. **Anticipate problems**: Document errors and solutions
4. **Stay focused**: One skill, one purpose
5. **Update regularly**: Keep skills current as capabilities change
6. **Get feedback**: Test with real agents and iterate

## Summary

Good skills are:
- **Discoverable**: Clear name and description with keywords
- **Complete**: Cover all aspects of the topic
- **Practical**: Concrete examples and workflows
- **Clear**: Easy to understand and follow
- **Tested**: Verified to work correctly

Use this guide to create skills that help agents learn and work effectively!
