#!/bin/bash
set -e

echo "🔨 Building Ensemble..."

# Create bin directory
mkdir -p bin

# Build server
echo "📦 Building server..."
go build -o bin/ensemble-server cmd/ensemble-server/main.go

# Build client
echo "📦 Building client..."
go build -o bin/ensemble cmd/ensemble/main.go

echo "✅ Build complete!"
echo ""
echo "Next steps:"
echo "1. Set your API key: export ANTHROPIC_API_KEY='your-key'"
echo "2. Start server: ./bin/ensemble-server"
echo "3. Run a task: ./bin/ensemble run 'your task'"
echo ""
echo "See QUICKSTART.md for detailed instructions."
