#!/bin/bash

# Test script for Jabber Bot API OpenAPI endpoints

echo "=== Testing Jabber Bot API Endpoints ==="
echo

# Test root endpoint
echo "1. Testing root endpoint (/):"
curl -s http://localhost:8080/ | jq '.'
echo

# Test OpenAPI YAML endpoint
echo "2. Testing OpenAPI YAML endpoint (/openapi.yaml):"
curl -s -I http://localhost:8080/openapi.yaml
echo

# Test OpenAPI JSON endpoint
echo "3. Testing OpenAPI JSON endpoint (/openapi.json):"
curl -s http://localhost:8080/openapi.json | jq '.info.title'
echo

# Test status endpoint
echo "4. Testing status endpoint (/api/v1/status):"
curl -s http://localhost:8080/api/v1/status | jq '.'
echo

# Test health endpoint
echo "5. Testing health endpoint (/api/v1/health):"
curl -s http://localhost:8080/api/v1/health | jq '.'
echo

echo "=== API Test Complete ==="