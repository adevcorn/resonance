#!/bin/bash

# Ensemble Quick Start Script
# This script helps you get Ensemble running quickly

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Ensemble Quick Start${NC}"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed. Please install Go 1.22 or later.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Go found: $(go version)${NC}"

# Check for API keys
if [ -z "$ANTHROPIC_API_KEY" ] && [ -z "$OPENAI_API_KEY" ]; then
    echo -e "${YELLOW}⚠️  No API key found!${NC}"
    echo ""
    echo "Please set one of the following:"
    echo "  export ANTHROPIC_API_KEY='sk-ant-...'"
    echo "  export OPENAI_API_KEY='sk-...'"
    echo ""
    read -p "Press Enter after setting your API key, or Ctrl+C to exit..."
fi

if [ -n "$ANTHROPIC_API_KEY" ]; then
    echo -e "${GREEN}✅ Anthropic API key found${NC}"
elif [ -n "$OPENAI_API_KEY" ]; then
    echo -e "${GREEN}✅ OpenAI API key found${NC}"
fi

# Build binaries
echo ""
echo -e "${BLUE}📦 Building binaries...${NC}"
make build

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Build failed!${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Build complete!${NC}"
echo ""

# Ask what to do
echo "What would you like to do?"
echo ""
echo "1) Start the server"
echo "2) Run a demo task (requires server running in another terminal)"
echo "3) Exit"
echo ""
read -p "Enter choice [1-3]: " choice

case $choice in
    1)
        echo ""
        echo -e "${BLUE}🌐 Starting Ensemble server...${NC}"
        echo ""
        echo "The server will start on http://localhost:8080"
        echo "Press Ctrl+C to stop the server"
        echo ""
        ./bin/ensemble-server
        ;;
    2)
        echo ""
        echo -e "${BLUE}🎯 Running demo task...${NC}"
        echo ""
        echo "Make sure the server is running in another terminal!"
        echo ""
        read -p "Press Enter to continue..."
        ./bin/ensemble run "analyze this Go project and tell me what it does"
        ;;
    3)
        echo ""
        echo "To start later:"
        echo "  Server: ./bin/ensemble-server"
        echo "  Client: ./bin/ensemble run 'your task'"
        echo ""
        echo "See QUICKSTART.md for more details."
        ;;
    *)
        echo "Invalid choice"
        exit 1
        ;;
esac
