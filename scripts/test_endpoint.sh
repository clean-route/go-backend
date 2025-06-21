#!/bin/bash

# Simple test script for the clean-route API endpoint
# Usage: ./test_endpoint.sh

echo "=== Testing Clean Route API ==="
echo ""

# Check if server is running
echo "Checking if server is running..."
if curl -s "http://localhost:9000/health" > /dev/null; then
    echo "✓ Server is running"
else
    echo "✗ Server is not running. Please start the server first."
    exit 1
fi

echo ""

# Test the route endpoint
echo "Testing route endpoint with test.json..."
echo ""

curl -X POST \
  -H "Content-Type: application/json" \
  -d @test.json \
  http://localhost:9000/route

echo ""
echo "=== Test Complete ===" 