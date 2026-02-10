# Integration Tests

This directory contains integration tests for the Jabber Bot project.

## Overview

Integration tests verify that different components work together correctly. They test:
- API endpoints
- Webhook functionality
- Configuration loading
- Component integration

## Running Tests

### Prerequisites

```bash
# Set environment variable to enable integration tests
export INTEGRATION_TESTS=1
```

### Run All Integration Tests

```bash
# Using the provided script
./scripts/run-integration-tests.sh

# Or manually
go test -v -tags=integration ./test/integration/...
```

### Run Specific Test Suites

```bash
# Configuration integration tests
go test -v -tags=integration ./test/integration/config_integration_test.go

# Webhook integration tests
go test -v -tags=integration ./test/integration/webhook_integration_test.go

# Full integration tests
go test -v -tags=integration ./test/integration/integration_test.go
```

### With Coverage

```bash
go test -v -tags=integration -coverprofile=coverage.out ./test/integration/...
go tool cover -html=coverage.out -o coverage.html
```

## Test Structure

### IntegrationTestSuite
Tests the complete application flow:
- API server startup/shutdown
- Webhook service integration
- Message flow from XMPP to webhook
- Error handling scenarios

### ConfigIntegrationTestSuite
Tests configuration loading in realistic scenarios:
- Real configuration files
- Environment variable overrides
- Production vs development configurations
- Validation and error handling

### WebhookIntegrationTestSuite
Tests webhook service with real HTTP servers:
- Message delivery
- Retry logic
- Concurrent requests
- Queue behavior

## Test Data

Integration tests use:
- Mock XMPP connections (since real XMPP server is not always available)
- Real HTTP servers for webhooks
- Temporary configuration files
- Environment variable testing

## CI/CD Integration

In CI/CD pipelines:

```yaml
- name: Run Integration Tests
  run: |
    export INTEGRATION_TESTS=1
    ./scripts/run-integration-tests.sh
```

## Debugging

Enable debug logging:

```bash
export JABBER_BOT_LOG_LEVEL=debug
go test -v -tags=integration ./test/integration/...
```

## Performance Testing

Run with performance benchmarks:

```bash
./scripts/run-integration-tests.sh --performance
```