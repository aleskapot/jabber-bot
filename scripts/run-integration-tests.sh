#!/bin/bash

# Integration Test Runner for Jabber Bot

set -e

echo "🧪 Starting Integration Tests for Jabber Bot"
echo "=========================================="

# Check if required tools are available
command -v go >/dev/null 2>&1 || { echo "❌ Go is required but not installed."; exit 1; }

# Set environment variables for integration tests
export INTEGRATION_TESTS=1
export JABBER_BOT_LOG_LEVEL=debug

# Create test directories
mkdir -p /tmp/jabber-bot-tests
mkdir -p test/reports

echo "📋 Running Integration Tests..."

# Run integration tests with coverage
echo ""
echo "🔧 Configuration Integration Tests..."
go test -v -tags=integration -timeout 30s ./test/integration/config_integration_test.go -coverprofile=test/reports/config-integration-coverage.out

echo ""
echo "🔌 Webhook Integration Tests..."
go test -v -tags=integration -timeout 60s ./test/integration/webhook_integration_test.go -coverprofile=test/reports/webhook-integration-coverage.out

echo ""
echo "🌐 Full Integration Tests..."
go test -v -tags=integration -timeout 60s ./test/integration/integration_test.go -coverprofile=test/reports/full-integration-coverage.out

# Combine coverage reports
echo ""
echo "📊 Combining coverage reports..."
go tool cover -merge=test/reports/*-coverage.out -o test/reports/integration-coverage.out

# Generate HTML coverage report
echo ""
echo "📈 Generating HTML coverage report..."
go tool cover -html=test/reports/integration-coverage.out -o test/reports/integration-coverage.html

# Display coverage summary
echo ""
echo "📈 Integration Test Coverage Summary:"
go tool cover -func=test/reports/integration-coverage.out

# Run performance tests if requested
if [ "$1" = "--performance" ]; then
    echo ""
    echo "⚡ Running Performance Tests..."
    go test -v -tags=integration -bench=. ./test/integration/... -benchmem
fi

# Cleanup test artifacts
echo ""
echo "🧹 Cleaning up test artifacts..."
rm -rf /tmp/jabber-bot-tests

echo ""
echo "✅ Integration Tests Completed Successfully!"
echo "📊 Coverage Report: test/reports/integration-coverage.html"
echo "📋 Detailed Reports: test/reports/"