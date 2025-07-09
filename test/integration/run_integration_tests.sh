#!/bin/bash

# Integration Test Runner for AI Rebaser
# This script runs the integration tests with proper setup

set -e

echo "🧪 AI Rebaser Integration Test Runner"
echo "======================================"

# Check if OpenAI API key is set
if [ -z "$OPENAI_API_KEY" ]; then
    echo "❌ OPENAI_API_KEY environment variable is not set"
    echo "   Please set your OpenAI API key to run integration tests:"
    echo "   export OPENAI_API_KEY=your_api_key_here"
    echo ""
    echo "   Or skip AI tests (will only test Git operations):"
    echo "   export SKIP_AI_TESTS=true"
    echo ""
    exit 1
fi

# Check if git is installed
if ! command -v git &> /dev/null; then
    echo "❌ Git is not installed or not in PATH"
    exit 1
fi

# Check if go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed or not in PATH"
    exit 1
fi

echo "✅ OpenAI API key is set"
echo "✅ Git is available"
echo "✅ Go is available"
echo ""

# Set test timeout
export TEST_TIMEOUT=${TEST_TIMEOUT:-10m}

echo "🚀 Running integration tests..."
echo "   Timeout: $TEST_TIMEOUT"
echo "   API Key: ${OPENAI_API_KEY:0:10}..."
echo ""

# Run the integration tests
cd "$(dirname "$0")/../.."

# Run with verbose output and proper timeout
go test -v -timeout "$TEST_TIMEOUT" ./test/integration/... -run "TestRealWorldRebaseWorkflow"

echo ""
echo "🧪 Running error scenario tests..."
go test -v -timeout "$TEST_TIMEOUT" ./test/integration/... -run "TestErrorScenarios"

echo ""
echo "✅ All integration tests completed!"
echo ""
echo "💡 Tips:"
echo "   - Check the test output for AI-generated content"
echo "   - Verify that conflicts were resolved intelligently"
echo "   - Review the generated commit messages and PR descriptions"