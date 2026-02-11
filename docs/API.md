# Jabber Bot API - OpenAPI Specification

This document describes the OpenAPI specification for the Jabber Bot API.

## Quick Start

### Accessing the API Specification

The API provides two formats of the OpenAPI specification:

- **YAML Format**: `http://localhost:8080/openapi.yaml`
- **JSON Format**: `http://localhost:8080/openapi.json`

### Interactive Documentation

You can use the specification files with various OpenAPI tools:

1. **Swagger UI**: Upload the YAML or JSON file to [Swagger Editor](https://editor.swagger.io/)
2. **Postman**: Import the OpenAPI specification
3. **Redoc**: Use with Redoc for beautiful API documentation
4. **Insomnia**: Import the specification for API testing

## API Endpoints Overview

### Base Information
- **Base URL**: `http://localhost:8080/api/v1`
- **Content-Type**: `application/json`
- **Authentication**: Currently not implemented (open API)

### Core Endpoints

#### Message Operations
- `POST /api/v1/send` - Send XMPP message to user
- `POST /api/v1/send-muc` - Send message to Multi-User Chat room

#### Status & Health
- `GET /api/v1/status` - Get comprehensive bot status
- `GET /api/v1/health` - Simple health check
- `GET /api/v1/webhook/status` - Webhook service status

#### Documentation
- `GET /` - API root information
- `GET /docs` - Plain text documentation
- `GET /openapi.yaml` - OpenAPI specification (YAML)
- `GET /openapi.json` - OpenAPI specification (JSON)

## Request Examples

### Send Message
```bash
curl -X POST http://localhost:8080/api/v1/send \
  -H "Content-Type: application/json" \
  -d '{
    "to": "user@example.com",
    "body": "Hello, world!",
    "type": "chat"
  }'
```

### Send MUC Message
```bash
curl -X POST http://localhost:8080/api/v1/send-muc \
  -H "Content-Type: application/json" \
  -d '{
    "room": "room@conference.example.com",
    "body": "Hello, room!",
    "subject": "Room Topic"
  }'
```

### Get Status
```bash
curl http://localhost:8080/api/v1/status
```

## Response Formats

### Success Response
```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": { ... }
}
```

### Error Response
```json
{
  "success": false,
  "error": "Error description",
  "code": 400
}
```

## Data Models

### SendMessageRequest
- `to` (string, required): JID of the recipient
- `body` (string, required): Message content (max 10,000 chars)
- `type` (string, optional): Message type (chat, groupchat, headline, normal)

### SendMUCMessageRequest
- `room` (string, required): JID of the MUC room
- `body` (string, required): Message content (max 10,000 chars)
- `subject` (string, optional): Room subject/topic

### StatusResponse
- `xmpp_connected` (boolean): XMPP connection status
- `api_running` (boolean): API server status
- `webhook_url` (string): Configured webhook URL
- `version` (string): Bot version

## Webhook Integration

The bot can forward incoming XMPP messages to configured webhook URLs:

### Webhook Payload Format
```json
{
  "message": {
    "id": "msg123",
    "from": "sender@example.com",
    "to": "bot@example.com",
    "body": "Hello bot",
    "type": "chat",
    "subject": "",
    "thread": "",
    "stamp": "2023-12-01T12:00:00Z"
  },
  "timestamp": "2023-12-01T12:00:00Z",
  "source": "jabber-bot"
}
```

### n8n Test Mode Support

The bot supports automatic test mode detection for n8n webhook integrations:

#### Test Mode Detection Rules
- **Message Detection**: Messages starting with `[test]` prefix are treated as test messages
- **URL Modification**: If webhook URL contains `webhook` but not `webhook-test`, the suffix `-test` is added
- **Message Processing**: The `[test]` prefix and following spaces are removed from the message body

#### Examples

**Normal Message**:
```
Input: "Hello world"
Webhook URL: https://example.com/webhook
Result: Sent to https://example.com/webhook with body "Hello world"
```

**Test Message**:
```
Input: "[test] Debug message"
Webhook URL: https://example.com/webhook
Result: Sent to https://example.com/webhook-test with body "Debug message"
```

**Test Message with Path**:
```
Input: "[test] Testing workflow"
Webhook URL: https://n8n.example.com/webhook/production/v1
Result: Sent to https://n8n.example.com/webhook-test/production/v1 with body "Testing workflow"
```

**Test Message with Non-Webhook URL**:
```
Input: "[test] Test message"
Webhook URL: https://example.com/api/endpoint
Result: Sent to https://example.com/api/endpoint with body "Test message" (URL unchanged)
```

#### Configuration
```yaml
webhook:
  url: "https://example.com/webhook"
  test_mode_suffix: "-test"  # Customizable suffix for test mode
  timeout: 30s
  retry_attempts: 3
```

#### Headers Added in Test Mode
- `Webhook-Test-Mode: "true"` - Indicates test mode processing
- Standard headers remain unchanged

#### Environment Variables
- `JABBER_BOT_WEBHOOK_TEST_MODE_SUFFIX`: Override default test suffix

## Error Codes

- `400` - Bad Request (validation errors, invalid JSON)
- `500` - Internal Server Error (XMPP errors, unexpected failures)
- `503` - Service Unavailable (XMPP connection lost)

## Headers

- `X-Request-ID`: Unique request identifier for tracking
- `Content-Type`: application/json for all API responses

## Configuration

The API can be configured via:
- Configuration file (YAML)
- Environment variables:
  - `JABBER_BOT_API_PORT`: API server port (default: 8080)
  - `JABBER_BOT_API_HOST`: API server host (default: 0.0.0.0)

## Testing with OpenAPI

### Using curl
```bash
# Get OpenAPI spec
curl http://localhost:8080/openapi.json

# Validate API responses against spec
# (Use OpenAPI validation tools)
```

### Using Postman
1. Download `openapi.json` or `openapi.yaml`
2. In Postman: Import → Raw Text → paste the specification
3. Postman will auto-generate all API endpoints

### Using Swagger UI
1. Go to [Swagger Editor](https://editor.swagger.io/)
2. File → Import File → upload `openapi.yaml`
3. Interactive API documentation available

## Development

### Generating Client SDKs
You can generate client SDKs using the OpenAPI specification:

```bash
# Using OpenAPI Generator
docker run --rm -v "${PWD}:/local" openapitools/openapi-generator-cli generate \
  -i /local/openapi.yaml \
  -g go \
  -o /local/client/go

# Generate for other languages
# -g python, javascript, java, php, etc.
```

### API Versioning
- Current version: v1
- Version included in all API paths: `/api/v1/`
- OpenAPI spec version: 1.0.0

## Security Considerations

### Current Limitations
- No authentication implemented
- No rate limiting
- No HTTPS enforcement

### Recommended Improvements
- Implement API key authentication (`X-API-Key` header)
- Add rate limiting middleware
- Enforce HTTPS in production
- Add request/response validation
- Implement audit logging

## Monitoring

### Health Checks
- `/api/v1/health`: Basic health check
- Returns HTTP 200 when healthy, 503 when XMPP connection lost

### Logging
All requests are logged with format:
```
{timestamp} | {status} | {latency} | {method} | {path} | {ip} | {error}
```

### Metrics
The API includes basic webhook metrics:
- Total sent messages
- Total failed deliveries
- Queue length
- Last activity timestamps

## Support

For issues and questions:
- Check the OpenAPI specification for detailed endpoint information
- Review the API logs for debugging
- Use the `/docs` endpoint for comprehensive documentation
- Examine webhook status for integration issues