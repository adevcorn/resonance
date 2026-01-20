---
name: git-workflow
description: Git workflow patterns, commit conventions, branch naming, and PR creation best practices
category: workflow
capabilities: []
---

# Git Workflow Guide

This skill provides standardized git workflow patterns for the Resonance multi-agent development system. Follow these conventions to ensure consistent branch management, commit messages, and pull request creation.

## When to Use This Skill

Use this skill when you need guidance on:
- Creating branches
- Writing commit messages
- Preparing pull requests
- Following git best practices
- Understanding the project's git workflow

## Branch Naming Conventions

Use semantic branch names that clearly indicate the purpose:

### Patterns

- `feature/<short-description>` - New features or capabilities
- `fix/<bug-description>` - Bug fixes
- `refactor/<scope>` - Code refactoring without functionality changes
- `docs/<topic>` - Documentation updates
- `test/<scope>` - Test additions or improvements
- `chore/<task>` - Maintenance tasks

### Examples

```bash
feature/agent-skills-system
fix/coordinator-state-persistence
refactor/tool-registry
docs/api-documentation
test/skill-loader
chore/update-dependencies
```

## Commit Message Conventions

### Structure

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Message Types

| Type | Purpose | Example |
|------|---------|---------|
| `feat` | New feature | `feat(skills): add activate_skill tool` |
| `fix` | Bug fix | `fix(agent): resolve prompt injection issue` |
| `docs` | Documentation | `docs(readme): update installation steps` |
| `style` | Formatting, no code change | `style: apply gofmt` |
| `refactor` | Code change without feature/fix | `refactor(loader): simplify skill discovery` |
| `perf` | Performance improvement | `perf(cache): optimize skill lookup` |
| `test` | Adding/updating tests | `test(skill): add loader unit tests` |
| `build` | Build system changes | `build: update go modules` |
| `ci` | CI/CD changes | `ci: add skill tests to pipeline` |
| `chore` | Maintenance | `chore: update dependencies` |

### Guidelines

1. **Use imperative mood**: "Add feature" not "Added feature"
2. **Keep subject line ≤ 50 characters**
3. **Capitalize subject line**
4. **No period at end of subject**
5. **Separate subject from body with blank line**
6. **Wrap body at 72 characters**
7. **Explain what and why, not how**

### Good Examples

```
feat(skills): add AgentSkills.io support

Implement skill discovery, loading, and activation following
the AgentSkills.io open standard. Agents can now load procedural
knowledge on-demand for specific tasks.

- Add YAML frontmatter parser
- Implement hot-reload with fsnotify
- Create activate_skill tool
- Add comprehensive test coverage

Closes #123
```

```
fix(registry): handle missing skills directory gracefully

Server would crash if skills/ directory didn't exist. Now falls
back gracefully and continues without skills.
```

### Bad Examples

❌ `updated stuff`  
❌ `Fixed bug`  
❌ `WIP`  
❌ `asdf`  
❌ `Added new feature for handling skills and also refactored some code and updated tests`

## Pull Request Workflow

### Creating a PR

When your feature is ready, create a pull request:

```bash
# 1. Ensure you're on your feature branch
git checkout feature/your-feature

# 2. Pull latest from main and rebase
git fetch origin
git rebase origin/main

# 3. Run tests locally
go test ./...

# 4. Push your branch
git push -u origin feature/your-feature

# 5. Create PR using GitHub CLI
gh pr create --title "Add your feature" --body "Description of changes"
```

### PR Title Format

Use the same format as commit messages:

```
feat(scope): Add new capability
fix(module): Resolve specific issue
docs: Update getting started guide
```

### PR Description Template

```markdown
## Summary

Brief description of what this PR does and why.

## Changes

- Bullet point of key changes
- Another important change
- One more change

## Testing

How this was tested:
- Unit tests added/updated
- Manual testing performed
- Integration tests pass

## Checklist

- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] No breaking changes (or documented)
- [ ] Follows code style guidelines
```

### Before Merging

1. ✅ All CI checks pass
2. ✅ Code reviewed and approved
3. ✅ Tests added/updated
4. ✅ Documentation updated
5. ✅ No merge conflicts
6. ✅ Commits are clean (consider squashing)

## Git Best Practices

### Commit Frequency

- **Commit often**: Make small, logical commits
- **One concern per commit**: Each commit should address one thing
- **Test before committing**: Ensure tests pass

### Before Pushing

```bash
# Review what you're about to push
git log origin/main..HEAD

# Check diff
git diff origin/main...HEAD

# Run tests
go test ./...

# Run linter
golangci-lint run
```

### Cleaning Up History

If you have many small commits, consider squashing before creating PR:

```bash
# Interactive rebase to squash commits
git rebase -i origin/main

# In the editor, change 'pick' to 'squash' for commits to combine
```

### Handling Conflicts

```bash
# Pull latest from main
git fetch origin

# Rebase onto main
git rebase origin/main

# Resolve conflicts in files
# After resolving each file:
git add <file>

# Continue rebase
git rebase --continue

# If things go wrong:
git rebase --abort
```

## Common Tasks

### Starting a New Feature

```bash
# Ensure main is up to date
git checkout main
git pull origin main

# Create feature branch
git checkout -b feature/my-feature

# Make changes and commit
git add .
git commit -m "feat: initial implementation"

# Push when ready
git push -u origin feature/my-feature
```

### Updating Your Branch

```bash
# When main has moved forward
git fetch origin
git rebase origin/main

# If you've already pushed, force push (with lease for safety)
git push --force-with-lease
```

### Fixing a Bug

```bash
# Create fix branch
git checkout -b fix/issue-description

# Make fix and commit
git add <fixed-files>
git commit -m "fix: resolve issue with X

Detailed explanation of what was broken and how it's fixed.

Fixes #issue-number"

# Push and create PR
git push -u origin fix/issue-description
```

## Git Commands Reference

### Essential Commands

```bash
# Status
git status

# Add files
git add <file>
git add .                    # Add all changes
git add -p                   # Add interactively

# Commit
git commit -m "message"
git commit --amend           # Modify last commit

# Push
git push
git push -u origin <branch>  # Set upstream
git push --force-with-lease  # Safe force push

# Pull/Fetch
git pull
git fetch origin
git pull --rebase            # Rebase instead of merge

# Branching
git branch                   # List branches
git checkout -b <branch>     # Create and switch
git branch -d <branch>       # Delete local branch

# Rebasing
git rebase main
git rebase -i main           # Interactive rebase

# Stashing
git stash                    # Stash changes
git stash pop                # Apply and remove stash
git stash list               # List stashes

# Viewing History
git log
git log --oneline
git log --graph --all        # Visual branch history
```

## Troubleshooting

### Accidentally Committed to Wrong Branch

```bash
# Move commit to correct branch
git log  # Note the commit SHA
git checkout correct-branch
git cherry-pick <commit-sha>
git checkout wrong-branch
git reset --hard HEAD~1  # Remove commit
```

### Need to Undo Last Commit

```bash
# Keep changes in working directory
git reset --soft HEAD~1

# Discard changes
git reset --hard HEAD~1
```

### Want to Change Commit Message

```bash
# Last commit only
git commit --amend -m "new message"

# Older commits
git rebase -i HEAD~3  # For last 3 commits
# Change 'pick' to 'reword' for commits to change
```

## Integration with Resonance Workflow

### Development State

When in DEVELOPMENT state:
- Create feature branch
- Make frequent, atomic commits
- Follow commit message conventions
- Push regularly to backup work

### Review State

Before handing off to REVIEW:
- Rebase on latest main
- Squash WIP commits if needed
- Ensure all tests pass
- Create draft PR if not already done

### CI/CD State

When moving to CI_CD:
- Mark PR as ready for review
- Ensure branch is up to date
- All commits follow conventions
- PR description is complete

### Validation State

After code is validated:
- Get final approval
- Ensure clean commit history
- Ready to merge

### Complete State

After merge:
- Delete feature branch (local and remote)
- Pull latest main
- Start next feature from clean main

## References

- [Conventional Commits](https://www.conventionalcommits.org/)
- [Git Best Practices](https://git-scm.com/book/en/v2)
- [Semantic Commit Messages](https://gist.github.com/joshbuchea/6f47e86d2510bce28f8e7f42ae84c716)

## Helper Scripts

This skill includes helper scripts in the `scripts/` directory:

- `create-pr.sh` - Automated PR creation with validation
