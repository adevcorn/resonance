#!/bin/bash

# Test script to verify tool names are displayed correctly

echo "🧪 Testing tool name display..."
echo ""
echo "This test will:"
echo "1. Start the ensemble server in background"
echo "2. Run a simple task that uses server-side tools"
echo "3. Check if tool names are displayed"
echo ""

# Kill any existing server process
pkill -f ensemble-server 2>/dev/null

# Start server in background
echo "Starting ensemble-server..."
./bin/ensemble-server > /tmp/ensemble-server.log 2>&1 &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Check if server started
if ! ps -p $SERVER_PID > /dev/null; then
    echo "❌ Server failed to start. Check /tmp/ensemble-server.log"
    exit 1
fi

echo "✅ Server started (PID: $SERVER_PID)"
echo ""

# Run test task
echo "Running test task..."
echo "Expected: Tool names should be displayed (e.g., 'Assembling team (4 agents)...')"
echo "Actual output:"
echo "----------------------------------------"

./bin/ensemble run "Say hello" 2>&1 | tee /tmp/ensemble-client.log

echo "----------------------------------------"
echo ""

# Kill server
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

# Check logs for debug output
echo ""
echo "📋 Debug logs from client:"
echo "----------------------------------------"
grep "\[DEBUG\]" /tmp/ensemble-client.log || echo "No debug output found"
echo "----------------------------------------"

echo ""
echo "📋 Debug logs from server:"
echo "----------------------------------------"
grep "Sending server-side tool_call" /tmp/ensemble-server.log | tail -5 || echo "No debug output found"
echo "----------------------------------------"

echo ""
echo "✅ Test complete. Check the output above."
echo ""
echo "If you see '[DEBUG] Received tool_call: name=\"assemble_team\"', but"
echo "the display shows 'Using ...' instead of 'Assembling team (4 agents)...',"
echo "then the issue is in the formatToolAction function."
