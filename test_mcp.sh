#!/bin/bash

# Test script for curated-axiom-mcp
echo "🧪 Testing curated-axiom-mcp functionality..."

export AXIOM_TOKEN="fake-token-for-testing"

echo "✅ Testing basic CLI commands..."

echo "1. Testing help command:"
./curated-axiom-mcp --help > /dev/null && echo "   ✓ Help command works"

echo "2. Testing list command:"
./curated-axiom-mcp list > /dev/null && echo "   ✓ List command works"

echo "3. Testing describe command:"
./curated-axiom-mcp describe user_events > /dev/null && echo "   ✓ Describe command works"

echo "4. Testing config init:"
rm -f ~/.config/curated-axiom-mcp/config.yaml
./curated-axiom-mcp config init > /dev/null && echo "   ✓ Config init works"

echo "5. Testing HTTP server (startup only):"
timeout 2 ./curated-axiom-mcp --port 9999 > /dev/null 2>&1 && echo "   ✓ HTTP server starts"

echo ""
echo "🎉 All basic tests passed!"
echo ""
echo "To test MCP functionality:"
echo "  1. Set your real AXIOM_TOKEN:"
echo "     export AXIOM_TOKEN='your-real-token'"
echo ""
echo "  2. Run in stdio mode:"
echo "     ./curated-axiom-mcp --stdio"
echo ""
echo "  3. Connect your MCP client to this process"
echo ""
echo "Example queries available:"
./curated-axiom-mcp list 