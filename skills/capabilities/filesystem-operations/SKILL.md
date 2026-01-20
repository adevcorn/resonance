---
name: filesystem-operations
description: Read, write, and navigate the project filesystem
category: capability
capabilities:
  - read_file
  - write_file
  - list_directory
---

# Filesystem Operations

Access and modify files in the project directory using active_tool.

## Reading Files

### When to Use
- Examine code before making modifications
- Understand configuration files
- Read documentation
- Review test files

### How to Execute

```json
{
  "action": "execute",
  "capability": "read_file",
  "parameters": {
    "path": "relative/path/to/file.go"
  }
}
```

### Returns

```json
{
  "content": "package main...",
  "path": "relative/path/to/file.go",
  "size": 1234
}
```

### Important Notes
- Paths are relative to project root
- Cannot escape project directory with `../`
- Returns error if file doesn't exist
- Content returned as string

### Examples

Read configuration file:
```json
{
  "action": "execute",
  "capability": "read_file",
  "parameters": {"path": "config.yaml"}
}
```

Read source code:
```json
{
  "action": "execute",
  "capability": "read_file",
  "parameters": {"path": "internal/server/agent/agent.go"}
}
```

## Writing Files

### When to Use
- Create new source files
- Modify existing code
- Update configuration
- Generate documentation

### How to Execute

```json
{
  "action": "execute",
  "capability": "write_file",
  "parameters": {
    "path": "path/to/file.go",
    "content": "package main\n\nfunc main() {}\n"
  }
}
```

### Important Notes
- **This is a client-side capability** - execution happens on the client
- Creates parent directories automatically if they don't exist
- Overwrites file if it already exists
- Preserves file permissions

### Best Practices
- **Always read_file first** to see current content before modifying
- Preserve existing formatting and style
- Test after writing (run builds/tests)
- Use meaningful commit messages after file changes

### Examples

Create a new Go file:
```json
{
  "action": "execute",
  "capability": "write_file",
  "parameters": {
    "path": "internal/newpackage/new.go",
    "content": "package newpackage\n\n// NewFunc does something\nfunc NewFunc() error {\n\treturn nil\n}\n"
  }
}
```

Update configuration:
```json
{
  "action": "execute",
  "capability": "write_file",
  "parameters": {
    "path": "config.yaml",
    "content": "server:\n  port: 8080\n  host: localhost\n"
  }
}
```

## Listing Directories

### When to Use
- Explore project structure
- Find files in a directory
- Understand codebase organization
- Discover available tests or packages

### How to Execute

```json
{
  "action": "execute",
  "capability": "list_directory",
  "parameters": {
    "path": "internal/server"
  }
}
```

Use `"."` for project root:
```json
{
  "action": "execute",
  "capability": "list_directory",
  "parameters": {
    "path": "."
  }
}
```

### Returns

```json
{
  "path": "internal/server",
  "files": [
    {"name": "agent", "is_dir": true, "size": 0},
    {"name": "tool", "is_dir": true, "size": 0},
    {"name": "main.go", "is_dir": false, "size": 5432}
  ]
}
```

### Examples

List all packages in internal/:
```json
{
  "action": "execute",
  "capability": "list_directory",
  "parameters": {"path": "internal"}
}
```

Explore test directory:
```json
{
  "action": "execute",
  "capability": "list_directory",
  "parameters": {"path": "internal/server/agent"}
}
```

## Complete Workflow Example

### Task: Add a new capability to the system

1. **Search for this skill**:
   ```json
   {"action": "search_skills", "query": "read write files"}
   ```

2. **Load this skill** to learn filesystem operations:
   ```json
   {"action": "load_skill", "skill_name": "filesystem-operations"}
   ```

3. **Explore the directory** to understand structure:
   ```json
   {"action": "execute", "capability": "list_directory", "parameters": {"path": "internal/server/capability"}}
   ```

4. **Read an existing file** to understand patterns:
   ```json
   {"action": "execute", "capability": "read_file", "parameters": {"path": "internal/server/capability/filesystem.go"}}
   ```

5. **Write the new capability**:
   ```json
   {"action": "execute", "capability": "write_file", "parameters": {"path": "internal/server/capability/new.go", "content": "..."}}
   ```

6. **Verify it was created**:
   ```json
   {"action": "execute", "capability": "list_directory", "parameters": {"path": "internal/server/capability"}}
   ```

## Security Considerations

- All file paths are restricted to the project directory
- Cannot access files outside the project with `../` or absolute paths
- Server validates all paths before operations
- Client-side tools (write_file) execute with client's permissions

## Troubleshooting

**Error: "access denied: path escapes project directory"**
- You're trying to access files outside the project
- Use paths relative to project root only
- Don't use `../` to go up directories outside project

**Error: "failed to read file: no such file or directory"**
- File doesn't exist at the specified path
- Use list_directory to see available files
- Check path spelling and capitalization

**Write not working**
- write_file is client-side, ensure client is processing it
- Check client logs for execution details
- Verify parent directories can be created
