#!/bin/bash
# Helper script for creating pull requests with consistent format

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo -e "${RED}Error: GitHub CLI (gh) is not installed${NC}"
    echo "Install from: https://cli.github.com/"
    exit 1
fi

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo -e "${RED}Error: Not in a git repository${NC}"
    exit 1
fi

# Get current branch name
CURRENT_BRANCH=$(git branch --show-current)

if [ "$CURRENT_BRANCH" = "main" ] || [ "$CURRENT_BRANCH" = "master" ]; then
    echo -e "${RED}Error: Cannot create PR from main/master branch${NC}"
    echo "Create a feature branch first: git checkout -b feature/your-feature"
    exit 1
fi

# Check if there are uncommitted changes
if ! git diff-index --quiet HEAD --; then
    echo -e "${YELLOW}Warning: You have uncommitted changes${NC}"
    echo "Commit or stash them before creating PR"
    exit 1
fi

# Push current branch to remote
echo -e "${GREEN}Pushing branch to remote...${NC}"
git push -u origin "$CURRENT_BRANCH"

# Extract type and scope from branch name
if [[ $CURRENT_BRANCH =~ ^(feature|fix|refactor|docs|test)/(.+)$ ]]; then
    TYPE="${BASH_REMATCH[1]}"
    SCOPE="${BASH_REMATCH[2]}"
    
    # Map branch type to commit type
    case $TYPE in
        feature) COMMIT_TYPE="feat" ;;
        *) COMMIT_TYPE="$TYPE" ;;
    esac
else
    echo -e "${YELLOW}Branch name doesn't follow convention${NC}"
    COMMIT_TYPE="feat"
    SCOPE="general"
fi

# Get list of commits on this branch (compared to main)
COMMITS=$(git log --oneline origin/main..HEAD)
NUM_COMMITS=$(echo "$COMMITS" | wc -l | xargs)

echo ""
echo -e "${GREEN}Branch: $CURRENT_BRANCH${NC}"
echo -e "${GREEN}Commits on this branch: $NUM_COMMITS${NC}"
echo ""
echo "$COMMITS"
echo ""

# Prompt for PR title
read -p "Enter PR title (or press Enter to auto-generate): " PR_TITLE

if [ -z "$PR_TITLE" ]; then
    # Auto-generate title from scope
    PR_TITLE="$COMMIT_TYPE($SCOPE): ${SCOPE//-/ }"
fi

# Prompt for PR description
echo ""
echo "Enter PR description (Ctrl+D when done):"
echo ""
PR_BODY=$(cat)

if [ -z "$PR_BODY" ]; then
    # Default template
    PR_BODY="## Summary
- 

## Testing
- 

## Related Issues
Closes #"
fi

# Create PR
echo ""
echo -e "${GREEN}Creating pull request...${NC}"

PR_URL=$(gh pr create \
    --title "$PR_TITLE" \
    --body "$PR_BODY" \
    --base main)

echo ""
echo -e "${GREEN}✓ Pull request created: $PR_URL${NC}"
echo ""

# Ask if user wants to open in browser
read -p "Open PR in browser? (y/N): " OPEN_BROWSER

if [ "$OPEN_BROWSER" = "y" ] || [ "$OPEN_BROWSER" = "Y" ]; then
    gh pr view --web
fi
