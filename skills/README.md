# Skills Directory

This directory contains skills that agents can discover and use through the `active_tool`.

## Architecture Overview

The Resonance multi-agent system uses a **skills-first architecture**:

1. **Agents** don't have direct access to capabilities
2. **Agents** use `active_tool` to search for and discover skills
3. **Skills** contain instructions on how to use capabilities
4. **Capabilities** are the actual executable actions (hidden from agents)

## Skill Structure

Each skill is a directory containing:

```
skill-name/
├── SKILL.md         # Skill instructions (required)
├── scripts/         # Optional: Executable scripts
├── references/      # Optional: Reference materials
└── assets/          # Optional: Supporting files
```

### SKILL.md Format

Every skill must have a `SKILL.md` file with YAML frontmatter:

```markdown
---
name: skill-name
description: Brief description of what this skill does
category: capability|workflow|integration
capabilities:
  - capability_name_1
  - capability_name_2
---

# Skill Title

## Overview
Description of what this skill helps with.

## How to Use

### Step 1: Search for Skills
\`\`\`json
{"action": "search_skills", "query": "relevant search terms"}
\`\`\`

### Step 2: Load This Skill
\`\`\`json
{"action": "load_skill", "skill_name": "skill-name"}
\`\`\`

### Step 3: Execute Capability
\`\`\`json
{
  "action": "execute",
  "capability": "capability_name",
  "parameters": {
    "param1": "value1"
  }
}
\`\`\`

## Examples
[Concrete examples of using this skill]

## Parameters
[Detailed parameter documentation]
```

## Skill Categories

### 1. Capability Skills (`skills/capabilities/`)

These skills teach agents how to use **individual capabilities** (low-level actions).

**Examples:**
- `filesystem-operations` - How to read/write files
- `shell-execution` - How to run shell commands
- `web-access` - How to fetch web content

**Characteristics:**
- One skill per capability domain
- Focus on parameter formats and usage patterns
- Provide concrete examples

### 2. Workflow Skills (`skills/workflows/`)

These skills teach agents **multi-step workflows** that combine multiple capabilities.

**Examples:**
- `git-workflow` - How to use git for version control
- `testing-workflow` - How to write and run tests

**Characteristics:**
- Combine multiple capabilities
- Document best practices
- Provide step-by-step instructions

### 3. Integration Skills (`skills/integrations/`)

These skills teach agents how to work with **external systems and APIs**.

**Examples:**
- `github-api` - How to interact with GitHub
- `slack-integration` - How to send Slack messages

**Characteristics:**
- Focus on external service integration
- Document authentication patterns
- Provide API usage examples

## Agent Workflow

### 1. Discovery Phase

Agent needs to perform a task but doesn't know how:

```json
{
  "action": "search_skills",
  "query": "read files from filesystem",
  "max_results": 5
}
```

Response contains relevant skills ranked by relevance.

### 2. Learning Phase

Agent loads the most relevant skill to learn how to use it:

```json
{
  "action": "load_skill",
  "skill_name": "filesystem-operations"
}
```

Response contains full skill instructions, examples, and parameter schemas.

### 3. Execution Phase

Agent uses the learned knowledge to execute the capability:

```json
{
  "action": "execute",
  "capability": "read_file",
  "parameters": {
    "path": "/path/to/file.txt"
  }
}
```

## Creating New Skills

### Capability Skills

1. Create directory: `skills/capabilities/your-skill-name/`
2. Create `SKILL.md` with frontmatter listing the capabilities
3. Document each capability's parameters and return values
4. Provide 3-5 concrete examples
5. Test by having an agent discover and use it

### Workflow Skills

1. Create directory: `skills/workflows/your-skill-name/`
2. Create `SKILL.md` with empty `capabilities: []` list
3. Document the multi-step workflow
4. Reference other skills that provide required capabilities
5. Provide end-to-end examples

## Skill Search Algorithm

The skill registry searches based on:

1. **Skill name** (weight: 3.0)
2. **Description** (weight: 2.0)
3. **Capabilities** (weight: 2.5)
4. **Category** (weight: 1.5)

Higher weights mean stronger matches. Use descriptive names and keywords.

## Available Capabilities

Current system capabilities (updated automatically):

- `read_file` - Read file contents (server-side)
- `write_file` - Write file contents (client-side)
- `list_directory` - List directory contents (server-side)
- `execute_command` - Execute shell commands (client-side)
- `fetch_url` - Fetch web page content (server-side)
- `web_search` - Search the web (server-side)

### Client-Side vs Server-Side

- **Server-side**: Capability executes on the server (safe, read-only operations)
- **Client-side**: Capability execution is delegated to the client (write operations, commands)

This distinction is handled transparently by the capability registry.

## Best Practices

### For Skill Authors

1. **Be specific**: Use concrete examples, not abstract descriptions
2. **Show parameters**: Document all parameter fields with types and examples
3. **Explain errors**: Include common error scenarios and solutions
4. **Use keywords**: Include search-friendly terms in descriptions
5. **Keep focused**: One skill should teach one thing well

### For Capability Developers

1. **Register capabilities**: Add to capability registry in `cmd/ensemble-server/main.go`
2. **Document location**: Specify if client-side or server-side
3. **Validate inputs**: Capabilities should validate parameters
4. **Return JSON**: Always return `json.RawMessage`
5. **Handle errors**: Return descriptive error messages

## Debugging

### Skill Not Found

```bash
# Check if skill loaded
./bin/ensemble-server  # Look for "Loaded skill" log messages
```

### Low Search Relevance

- Add more keywords to skill description
- Add relevant capabilities to frontmatter
- Use common terminology in skill name

### Capability Not Found

```bash
# List registered capabilities
# (Check main.go capability registration section)
```

## Architecture Benefits

1. **Reduced Context**: Agents only load skills they need
2. **Dynamic Discovery**: New skills are available immediately
3. **Self-Documenting**: Skills contain their own usage instructions
4. **Flexible**: Add skills without recompiling
5. **Searchable**: Agents find skills using natural language

## Migration Notes

This architecture replaced the previous static tool system where:
- Agents had all tool schemas in their system prompts
- Tools were directly registered and called
- No discovery mechanism existed

The new system:
- Agents discover capabilities through skill search
- Skills provide just-in-time documentation
- Capability registry handles execution transparently

For more information, see `SKILLS_IMPLEMENTATION.md`.
