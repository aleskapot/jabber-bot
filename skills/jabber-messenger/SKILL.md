---
name: jabber-messenger
description: Use when you need to send XMPP/Jabber messages via the Jabber Bot API. Provides functions to send messages, MUC messages, files, and check status.
version: 1.0.0
author: Hermes Agent
license: MIT
metadata:
  hermes:
    tags: [messaging, jabber, xmpp, api]
    related_skills: [http-request, api-integration]
---

# Jabber Messenger Skill

Send XMPP/Jabber messages through the Jabber Bot API using HTTP requests.

## Overview

This skill provides a simple interface to interact with a Jabber Bot API service that exposes REST endpoints for sending XMPP messages, MUC (Multi-User Chat) messages, file transfers, chat state notifications, and status checks. The skill handles authentication via API-Key header, formats requests according to the API specification, and standardizes responses for easy use in Hermes Agent workflows.

The Jabber Bot API OpenAPI specification can be found at: http://app1.mos.skbis.ru:8080/docs/openapi.json

## When to Use

- Sending one-to-one XMPP messages to users
- Sending messages to MUC (group chat) rooms
- Sending files via XMPP using XEP-0363 (HTTP File Upload)
- Sending chat state notifications (typing indicators, etc.)
- Checking the health and status of the Jabber bot service
- Integrating Jabber notifications into automated workflows

## Skill Functions

### `send_jabber_message(to, body, type="chat", api_key=None, base_url=None)`
Send a standard XMPP message.

Parameters:
- `to` (str): Recipient JID (e.g., "user@example.com")
- `body` (str): Message content (max 10,000 characters)
- `type` (str): Message type - "chat", "groupchat", "headline", or "normal" (default: "chat")
- `api_key` (str): API key for authentication (optional, can be set via JABBER_API_KEY env var)
- `base_url` (str): Base URL of the Jabber Bot API (optional, defaults to "http://localhost:8080")

Returns:
- dict: Response with keys `success` (bool), `message` (str), and `data` (dict) containing sent message details

### `send_muc_message(room, body, subject="", api_key=None, base_url=None)`
Send a message to a Multi-User Chat room.

Parameters:
- `room` (str): Room JID (e.g., "room@conference.example.com")
- `body` (str): Message content (max 10,000 characters)
- `subject` (str): Optional room subject/topic (max 200 characters)
- `api_key` (str): API key for authentication
- `base_url` (str): Base URL of the Jabber Bot API

Returns:
- dict: Response with `success`, `message`, and `data` containing sent MUC message details

### `send_file(to, file_path, description="", api_key=None, base_url=None)`
Send a file via XMPP using XEP-0363 (HTTP File Upload).

Parameters:
- `to` (str): Recipient JID
- `file_path` (str): Local path to the file to send
- `description` (str): Optional description of the file
- `api_key` (str): API key for authentication
- `base_url` (str): Base URL of the Jabber Bot API

Returns:
- dict: Response with `success`, `message`, and `data` containing file upload and send details

### `send_chat_state(to, state, api_key=None, base_url=None)`
Send a chat state notification (typing indicator, etc.).

Parameters:
- `to` (str): Recipient JID
- `state` (str): Chat state - "active", "composing", "paused", "inactive", or "gone"
- `api_key` (str): API key for authentication
- `base_url` (str): Base URL of the Jabber Bot API

Returns:
- dict: Response with `success`, `message`, and `data` containing chat state details

### `get_status(api_key=None, base_url=None)`
Get the Jabber bot status (XMPP connection, API status, webhook config).

Parameters:
- `api_key` (str): API key for authentication
- `base_url` (str): Base URL of the Jabber Bot API

Returns:
- dict: Response with `success`, `message`, and `data` containing status information

### `health_check(base_url=None)`
Perform a health check on the Jabber bot service.

Parameters:
- `base_url` (str): Base URL of the Jabber Bot API

Returns:
- dict: Response with `success`, `message`, and `data` containing health status

## Authentication

The Jabber Bot API uses API-Key authentication via the `API-Key` header. The skill supports two methods:

1. **Parameter**: Pass `api_key` directly to any function
2. **Environment Variable**: Set `JABBER_API_KEY` environment variable

If both are provided, the parameter takes precedence.

Example:
```bash
export JABBER_API_KEY="your-api-key-here"
```

## Error Handling

The skill standardizes error responses to match the successful format where possible. All functions return a dictionary with:
- `success`: Boolean indicating if the request succeeded
- `message`: Human-readable status message
- `data`: Additional data (on success) or error details (on failure)

Common HTTP errors are caught and returned as:
```json
{
  "success": false,
  "message": "Error description",
  "data": {
    "status_code": 401,
    "error": "Unauthorized - valid API key required"
  }
}
```

## Usage Examples

### Send a simple message
```python
result = send_jabber_message(
    to="friend@example.com",
    body="Hello from Hermes Agent! nya~",
    api_key="your-api-key"
)
if result["success"]:
    print("Message sent successfully!")
else:
    print(f"Failed to send: {result['message']}")
```

### Send a file
```python
result = send_file(
    to="friend@example.com",
    file_path="/path/to/document.pdf",
    description="Project documentation",
    api_key="your-api-key"
)
```

### Check bot status
```python
status = get_status(api_key="your-api-key")
if status["success"] and status["data"]["xmpp_connected"]:
    print("Jabber bot is connected and ready!")
```

## Common Pitfalls

1. **Incorrect JID format**: Ensure JIDs are properly formatted as `username@domain.com` or `username@domain.com/resource`. The API validates JID format and will return 400 errors for invalid formats.

2. **API key security**: Avoid hardcoding API keys in scripts. Use environment variables or secure secret management.

3. **Network connectivity**: The skill does not implement automatic retries for transient network failures. Consider wrapping calls in retry logic if needed for unstable connections.

4. **File size limits**: The send-file function respects the server's file transfer configuration (default 10 MB). Larger files will be rejected by the API.

5. **Base URL configuration**: The API specification shows example base URL `http://app1.mos.skbis.ru:8080` (production)

## Implementation Details

The skill uses the `terminal` tool to make HTTP requests via `curl` for maximum compatibility. Each function:
1. Constructs the appropriate URL and headers
2. Formats the request body (JSON or multipart/form-data for file uploads)
3. Executes the curl command with appropriate timeout
4. Parses the JSON response
5. Returns standardized result format

Timeouts are set to 30 seconds for all requests to balance responsiveness with network variability.

---
