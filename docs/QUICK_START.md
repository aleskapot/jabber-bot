# Quick Start Guide

Get Jabber Bot up and running in minutes with this comprehensive quick start guide.

## Prerequisites

- **XMPP Server**: Access to an XMPP server
- **Webhook Endpoint**: URL to receive message notifications
- **Docker** (recommended) or Go 1.21+

## Option 1: Docker Quick Start (Recommended)

### 1. Clone Repository

```bash
git clone https://github.com/your-org/jabber-bot.git
cd jabber-bot
```

### 2. Configure

```bash
# Copy configuration template
cp .env.example .env

# Edit configuration
nano .env
```

Add your settings:
```bash
# XMPP Configuration
JABBER_BOT_XMPP_JID=your-bot@your-xmpp-server.com
JABBER_BOT_XMPP_PASSWORD=your-secure-password
JABBER_BOT_XMPP_SERVER=your-xmpp-server.com:5222

# Webhook Configuration
JABBER_BOT_WEBHOOK_URL=https://your-webhook-endpoint.com/receive

# API Configuration
JABBER_BOT_API_PORT=8080
```

### 3. Start Bot

```bash
# Development environment
./scripts/deploy.sh dev

# Or production environment
./scripts/deploy.sh prod
```

### 4. Verify

```bash
# Check bot status
curl http://localhost:8080/api/v1/status

# Health check
curl http://localhost:8080/api/v1/health
```

## Option 2: Manual Installation

### 1. Build from Source

```bash
# Clone repository
git clone https://github.com/your-org/jabber-bot.git
cd jabber-bot

# Install dependencies
go mod download

# Build application
go build -o bin/jabber-bot ./cmd/server
```

### 2. Configure

Create `configs/config.yaml`:

```yaml
xmpp:
  jid: "your-bot@your-xmpp-server.com"
  password: "your-secure-password"
  server: "your-xmpp-server.com:5222"
  resource: "bot"

api:
  port: 8080
  host: "0.0.0.0"

webhook:
  url: "https://your-webhook-endpoint.com/receive"
  timeout: 30s
  retry_attempts: 3

logging:
  level: "info"
  output: "stdout"

reconnection:
  enabled: true
  max_attempts: 5
  backoff: "5s"
```

### 3. Run

```bash
# Start bot
./bin/jabber-bot -config configs/config.yaml
```

## Option 3: Pre-built Binary

### 1. Download

```bash
# Download latest release
wget https://github.com/your-org/jabber-bot/releases/latest/download/jabber-bot-linux-amd64

# Make executable
chmod +x jabber-bot-linux-amd64
```

### 2. Configure

Create `.env` file:
```bash
JABBER_BOT_XMPP_JID=your-bot@your-xmpp-server.com
JABBER_BOT_XMPP_PASSWORD=your-secure-password
JABBER_BOT_XMPP_SERVER=your-xmpp-server.com:5222
JABBER_BOT_WEBHOOK_URL=https://your-webhook-endpoint.com/receive
```

### 3. Run

```bash
./jabber-bot-linux-amd64
```

## Testing Your Setup

### 1. Send Test Message

```bash
# Send message via API
curl -X POST http://localhost:8080/api/v1/send \
  -H "Content-Type: application/json" \
  -d '{
    "to": "your-friend@your-xmpp-server.com",
    "body": "Hello from Jabber Bot! 🤖"
  }'
```

### 2. Send to Group Chat

```bash
# Send to MUC room
curl -X POST http://localhost:8080/api/v1/send-muc \
  -H "Content-Type: application/json" \
  -d '{
    "room": "general@conference.your-xmpp-server.com",
    "body": "Hello from bot in group chat!"
  }'
```

### 3. Check Webhook

Your webhook endpoint should receive JSON like this:

```json
{
  "message": {
    "id": "msg-123",
    "from": "sender@your-xmpp-server.com/resource",
    "to": "your-bot@your-xmpp-server.com",
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

## Next Steps

### 1. Monitor Your Bot

```bash
# Check webhook status
curl http://localhost:8080/api/v1/webhook/status

# Enable monitoring stack
./scripts/deploy.sh monitoring
```

Access:
- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090

### 2. Deploy to Production

```bash
# Production deployment
./scripts/deploy.sh prod

# With monitoring
docker-compose -f docker-compose.prod.yml -f docker-compose.yml up -d
```

### 3. Integrate with Your Application

**Python Example:**
```python
import requests

def send_notification(user, message):
    response = requests.post(
        "http://localhost:8080/api/v1/send",
        json={
            "to": f"{user}@your-xmpp-server.com",
            "body": message
        }
    )
    return response.json()

# Use in your app
send_notification("admin", "System alert: CPU usage high!")
```

**Node.js Example:**
```javascript
const axios = require('axios');

async function sendAlert(message) {
    try {
        const response = await axios.post(
            'http://localhost:8080/api/v1/send',
            {
                to: 'alerts@your-xmpp-server.com',
                body: message
            }
        );
        return response.data;
    } catch (error) {
        console.error('Failed to send alert:', error);
    }
}

// Use in your app
sendAlert('Server backup completed successfully!');
```

## Troubleshooting

### Bot Won't Start

```bash
# Check configuration
./bin/jabber-bot -config configs/config.yaml -check

# Verify XMPP credentials
telnet your-xmpp-server.com 5222
```

### Messages Not Sending

```bash
# Check XMPP connection
curl http://localhost:8080/api/v1/status

# Check webhook status
curl http://localhost:8080/api/v1/webhook/status

# Enable debug logging
export JABBER_BOT_LOG_LEVEL=debug
```

### Webhook Not Receiving

```bash
# Test webhook endpoint
curl -X POST https://your-webhook-endpoint.com/test \
  -H "Content-Type: application/json" \
  -d '{"test": true}'

# Check firewall rules
curl -v http://localhost:8080/api/v1/health
```

## Common Issues

| Issue | Solution |
|-------|----------|
| **Authentication failed** | Verify XMPP JID and password |
| **Connection timeout** | Check XMPP server accessibility |
| **Webhook fails** | Verify webhook URL is reachable |
| **High memory usage** | Check message queue size |
| **API returns 500** | Check XMPP connection status |

## Need Help?

- 📖 [Documentation](docs/README.md)
- 🐛 [Issues](https://github.com/your-org/jabber-bot/issues)
- 💬 [Discussions](https://github.com/your-org/jabber-bot/discussions)
- 📧 [Support](mailto:support@your-org.com)

## Quick Reference

### Environment Variables

```bash
# Required
JABBER_BOT_XMPP_JID=bot@server.com
JABBER_BOT_XMPP_PASSWORD=password
JABBER_BOT_XMPP_SERVER=server.com:5222
JABBER_BOT_WEBHOOK_URL=https://webhook.example.com

# Optional
JABBER_BOT_LOG_LEVEL=info
JABBER_BOT_API_PORT=8080
JABBER_BOT_RECONNECTION_ENABLED=true
```

### API Endpoints

```bash
POST /api/v1/send          # Send message
POST /api/v1/send-muc      # Send to group chat
GET  /api/v1/status        # Bot status
GET  /api/v1/health        # Health check
GET  /api/v1/webhook/status # Webhook status
```

### Docker Commands

```bash
# Start development
./scripts/deploy.sh dev

# Start production
./scripts/deploy.sh prod

# Stop services
docker-compose down

# View logs
docker-compose logs -f jabber-bot
```

🎉 **Congratulations!** Your Jabber Bot is now running and ready to send/receive messages.