#!/bin/bash
# Start the Gemini CLI bridge for OAuth authentication

set -e

BRIDGE_DIR="internal/server/provider/gemini/bridge"

echo "🚀 Starting Gemini CLI Bridge..."
echo ""

# Check if Gemini CLI is installed
if ! command -v gemini &> /dev/null; then
    echo "❌ Error: Gemini CLI not found"
    echo ""
    echo "Install it with:"
    echo "  npm install -g @google/gemini-cli"
    echo ""
    exit 1
fi

echo "✓ Gemini CLI found: $(which gemini)"

# Check if bridge dependencies are installed
if [ ! -d "$BRIDGE_DIR/node_modules" ]; then
    echo ""
    echo "📦 Installing bridge dependencies..."
    cd "$BRIDGE_DIR"
    npm install
    cd - > /dev/null
fi

echo "✓ Bridge dependencies installed"
echo ""
echo "Starting bridge on http://localhost:3001..."
echo "Press Ctrl+C to stop"
echo ""

# Start the bridge
cd "$BRIDGE_DIR"
npm start
