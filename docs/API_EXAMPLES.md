# API Examples

This document provides comprehensive examples of how to use the Jabber Bot API.

## Base URL
```
http://localhost:8080/api/v1
```

## Authentication
Currently, the API doesn't require authentication. This may change in future versions.

## Headers
```
Content-Type: application/json
X-Request-ID: <unique-identifier>  // Optional
```

## Error Responses
All error responses follow this format:
```json
{
  "success": false,
  "error": "Error message",
  "code": 400
}
```

## Endpoints

### 1. Send Message

**Endpoint:** `POST /send`

Send a message to an XMPP user.

#### Request
```json
{
  "to": "user@example.com",
  "body": "Hello, world!",
  "type": "chat"
}
```

#### Response (Success)
```json
{
  "success": true,
  "message": "Message sent successfully",
  "data": {
    "to": "user@example.com",
    "type": "chat",
    "body_length": 13,
    "sent_at": "2023-12-01T12:00:00Z",
    "request_id": "abc123"
  }
}
```

#### Examples

**Simple message:**
```bash
curl -X POST http://localhost:8080/api/v1/send \
  -H "Content-Type: application/json" \
  -d '{
    "to": "friend@example.com",
    "body": "Hi there! How are you?"
  }'
```

**Message with type:**
```bash
curl -X POST http://localhost:8080/api/v1/send \
  -H "Content-Type: application/json" \
  -d '{
    "to": "support@company.com",
    "body": "I need help with your service",
    "type": "chat"
  }'
```

### 2. Send MUC Message

**Endpoint:** `POST /send-muc`

Send a message to a Multi-User Chat room.

#### Request
```json
{
  "room": "room@conference.example.com",
  "body": "Hello, room!",
  "subject": "New Topic"
}
```

#### Response (Success)
```json
{
  "success": true,
  "message": "MUC message sent successfully",
  "data": {
    "room": "room@conference.example.com",
    "subject": "New Topic",
    "body_length": 14,
    "sent_at": "2023-12-01T12:00:00Z",
    "request_id": "def456"
  }
}
```

#### Examples

**Join and send message:**
```bash
curl -X POST http://localhost:8080/api/v1/send-muc \
  -H "Content-Type: application/json" \
  -d '{
    "room": "general@conference.example.com",
    "body": "Hello everyone! 👋"
  }'
```

**Change room subject:**
```bash
curl -X POST http://localhost:8080/api/v1/send-muc \
  -H "Content-Type: application/json" \
  -d '{
    "room": "general@conference.example.com",
    "body": "Changing room topic...",
    "subject": "Daily Standup"
  }'
```

### 3. Get Status

**Endpoint:** `GET /status`

Get the current status of the bot.

#### Response
```json
{
  "xmpp_connected": true,
  "api_running": true,
  "webhook_url": "https://example.com/webhook",
  "version": "1.0.0"
}
```

#### Example
```bash
curl -X GET http://localhost:8080/api/v1/status
```

### 4. Health Check

**Endpoint:** `GET /health`

Simple health check endpoint.

#### Response (Healthy)
```json
{
  "status": "ok",
  "timestamp": "2023-12-01T12:00:00Z"
}
```

#### Response (Unhealthy)
```json
{
  "status": "error",
  "timestamp": "2023-12-01T12:00:00Z",
  "error": "XMPP connection lost"
}
```

#### Example
```bash
curl -X GET http://localhost:8080/api/v1/health
```

### 5. Webhook Status

**Endpoint:** `GET /webhook/status`

Get webhook service status.

#### Response
```json
{
  "running": true,
  "healthy": true,
  "queue_length": 0,
  "webhook_url": "https://example.com/webhook",
  "total_sent": 42,
  "total_failed": 2,
  "last_sent": "2023-12-01T11:55:00Z",
  "last_failure": "2023-12-01T11:30:00Z",
  "last_error": "connection timeout"
}
```

#### Example
```bash
curl -X GET http://localhost:8080/api/v1/webhook/status
```

## Error Scenarios

### Validation Errors

**Missing required field:**
```bash
curl -X POST http://localhost:8080/api/v1/send \
  -H "Content-Type: application/json" \
  -d '{"body": "Missing to field"}'
```

Response:
```json
{
  "success": false,
  "error": "to field is required",
  "code": 400
}
```

**Invalid JID format:**
```bash
curl -X POST http://localhost:8080/api/v1/send \
  -H "Content-Type: application/json" \
  -d '{"to": "invalid-jid", "body": "test"}'
```

Response:
```json
{
  "success": false,
  "error": "invalid JID format",
  "code": 400
}
```

**Message too long:**
```bash
curl -X POST http://localhost:8080/api/v1/send \
  -H "Content-Type: application/json" \
  -d '{"to": "user@example.com", "body": "very long message..."}'
```

Response:
```json
{
  "success": false,
  "error": "body field too long (max 10000 characters)",
  "code": 400
}
```

### Service Errors

**XMPP not connected:**
```bash
curl -X POST http://localhost:8080/api/v1/send \
  -H "Content-Type: application/json" \
  -d '{"to": "user@example.com", "body": "test"}'
```

Response:
```json
{
  "success": false,
  "error": "Failed to send message: XMPP client is not connected",
  "code": 500
}
```

## Advanced Examples

### Bulk Messages

Send multiple messages using a script:

```bash
#!/bin/bash
API_URL="http://localhost:8080/api/v1/send"
USERS=("user1@example.com" "user2@example.com" "user3@example.com")
MESSAGE="Hello from bot!"

for user in "${USERS[@]}"; do
  curl -X POST "$API_URL" \
    -H "Content-Type: application/json" \
    -d "{\"to\": \"$user\", \"body\": \"$MESSAGE\"}"
  echo "Sent to $user"
  sleep 1
done
```

### Webhook Payload Example

When a message is received, the webhook receives this payload:

```json
{
  "message": {
    "id": "msg-123",
    "from": "sender@example.com/resource",
    "to": "bot@example.com",
    "body": "Hello bot!",
    "type": "chat",
    "subject": "",
    "thread": "",
    "stamp": "2023-12-01T12:00:00Z"
  },
  "timestamp": "2023-12-01T12:00:01Z",
  "source": "jabber-bot"
}
```

### Integration with Monitoring

Monitor API health using a cron job:

```bash
#!/bin/bash
HEALTH_URL="http://localhost:8080/api/v1/health"
RESPONSE=$(curl -s "$HEALTH_URL")

if echo "$RESPONSE" | grep -q '"status":"ok"'; then
    echo "$(date): Bot is healthy"
else
    echo "$(date): Bot is unhealthy! Response: $RESPONSE"
    # Send alert
    curl -X POST "https://alerts.example.com/notify" \
      -d "Jabber bot is down!"
fi
```

## Language Examples

### Python

```python
import requests
import json

BASE_URL = "http://localhost:8080/api/v1"

def send_message(to, body, message_type="chat"):
    """Send a message via the API"""
    url = f"{BASE_URL}/send"
    data = {
        "to": to,
        "body": body,
        "type": message_type
    }
    
    response = requests.post(url, json=data)
    return response.json()

def send_muc_message(room, body, subject=""):
    """Send a message to a MUC room"""
    url = f"{BASE_URL}/send-muc"
    data = {
        "room": room,
        "body": body,
        "subject": subject
    }
    
    response = requests.post(url, json=data)
    return response.json()

def get_status():
    """Get bot status"""
    url = f"{BASE_URL}/status"
    response = requests.get(url)
    return response.json()

# Usage
if __name__ == "__main__":
    # Send message
    result = send_message("friend@example.com", "Hello from Python!")
    print(result)
    
    # Send to MUC
    result = send_muc_message("general@conference.example.com", "Python bot here!")
    print(result)
    
    # Check status
    status = get_status()
    print(f"XMPP Connected: {status['xmpp_connected']}")
```

### Node.js

```javascript
const axios = require('axios');

const BASE_URL = 'http://localhost:8080/api/v1';

async function sendMessage(to, body, type = 'chat') {
  try {
    const response = await axios.post(`${BASE_URL}/send`, {
      to,
      body,
      type
    });
    return response.data;
  } catch (error) {
    console.error('Error sending message:', error.response?.data || error.message);
    throw error;
  }
}

async function sendMUCMessage(room, body, subject = '') {
  try {
    const response = await axios.post(`${BASE_URL}/send-muc`, {
      room,
      body,
      subject
    });
    return response.data;
  } catch (error) {
    console.error('Error sending MUC message:', error.response?.data || error.message);
    throw error;
  }
}

async function getStatus() {
  try {
    const response = await axios.get(`${BASE_URL}/status`);
    return response.data;
  } catch (error) {
    console.error('Error getting status:', error.response?.data || error.message);
    throw error;
  }
}

// Usage
(async () => {
  try {
    // Send message
    const result = await sendMessage('friend@example.com', 'Hello from Node.js!');
    console.log(result);
    
    // Send to MUC
    const mucResult = await sendMUCMessage('general@conference.example.com', 'Node.js bot here!');
    console.log(mucResult);
    
    // Check status
    const status = await getStatus();
    console.log(`XMPP Connected: ${status.xmpp_connected}`);
  } catch (error) {
    console.error('Operation failed:', error);
  }
})();
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

const BaseURL = "http://localhost:8080/api/v1"

type SendMessageRequest struct {
    To   string `json:"to"`
    Body string `json:"body"`
    Type string `json:"type,omitempty"`
}

type SendMUCMessageRequest struct {
    Room    string `json:"room"`
    Body    string `json:"body"`
    Subject string `json:"subject,omitempty"`
}

func sendMessage(to, body, messageType string) ([]byte, error) {
    req := SendMessageRequest{
        To:   to,
        Body: body,
        Type: messageType,
    }
    
    jsonData, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }
    
    resp, err := http.Post(BaseURL+"/send", "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    return json.NewDecoder(resp.Body).Decode(&struct{}{})
}

func sendMUCMessage(room, body, subject string) ([]byte, error) {
    req := SendMUCMessageRequest{
        Room:    room,
        Body:    body,
        Subject: subject,
    }
    
    jsonData, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }
    
    resp, err := http.Post(BaseURL+"/send-muc", "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    return json.NewDecoder(resp.Body).Decode(&struct{}{})
}

func main() {
    // Send message
    result, err := sendMessage("friend@example.com", "Hello from Go!", "chat")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    fmt.Printf("Result: %s\n", result)
    
    // Send to MUC
    mucResult, err := sendMUCMessage("general@conference.example.com", "Go bot here!", "")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    fmt.Printf("MUC Result: %s\n", mucResult)
}
```