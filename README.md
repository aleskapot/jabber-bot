# Jabber Bot with REST API

🤖 A production-ready XMPP (Jabber) bot with RESTful API for sending messages and webhook notifications for incoming messages.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-ready-blue.svg)](docker/)
[![Tests](https://img.shields.io/badge/Tests-passing-brightgreen.svg)](#testing)

## ✨ Features

- 🚀 **RESTful API** - Send XMPP messages via HTTP endpoints
- 🏠 **MUC Support** - Send messages to group chats
- 🔔 **Webhook Notifications** - Forward incoming messages to your endpoints
- 🔄 **Auto-Reconnection** - Automatic reconnection with configurable backoff
- 📊 **Monitoring Ready** - Health checks, metrics, and observability
- 🔧 **Flexible Configuration** - YAML files and environment variables
- 📝 **Structured Logging** - Zap-based logging with multiple output options
- 🐳 **Docker Support** - Multi-environment Docker deployment
- 🧪 **Comprehensive Testing** - Unit and integration tests with high coverage
- 🛡️ **Production Ready** - Security best practices and performance optimization

## 🚀 Quick Start

### Option 1: Docker (Recommended)

```bash
# Clone and configure
git clone https://github.com/your-org/jabber-bot.git
cd jabber-bot
cp .env.example .env
# Edit .env with your XMPP credentials

# Start development environment
./scripts/deploy.sh dev
```

### Option 2: Binary

```bash
# Download and run
wget https://github.com/your-org/jabber-bot/releases/latest/download/jabber-bot-linux-amd64
chmod +x jabber-bot-linux-amd64
./jabber-bot-linux-amd64
```

### Option 3: From Source

```bash
# Build and run
git clone https://github.com/your-org/jabber-bot.git
cd jabber-bot
make build
make run
```

## 📖 Documentation

- [📚 **Quick Start**](docs/QUICK_START.md) - Get running in minutes
- [🔧 **API Examples**](docs/API_EXAMPLES.md) - Complete API usage examples
- [🚀 **Deployment Guide**](docs/DEPLOYMENT.md) - Production deployment options
- [👨‍💻 **Development Guide**](docs/DEVELOPMENT.md) - Setup and contribution guide
- [📖 **Full Documentation**](docs/README.md) - Complete documentation hub

## 🏗️ Architecture

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   XMPP      │    │   Webhook   │    │   Config    │
│   Client    │◄──►│   Service   │◄──►│  Manager    │
└─────────────┘    └─────────────┘    └─────────────┘
       ▲                    ▲                    ▲
       │                    │                    │
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│    API      │◄──►│   Manager   │◄──►│   Logger    │
│   Server    │    │             │    │             │
└─────────────┘    └─────────────┘    └─────────────┘
       ▲
       │
┌─────────────┐
│  External   │
│   Clients   │
└─────────────┘
```

## 🔌 API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/send` | Send message to XMPP user |
| `POST` | `/api/v1/send-muc` | Send message to group chat |
| `GET` | `/api/v1/status` | Get bot status and statistics |
| `GET` | `/api/v1/health` | Simple health check |
| `GET` | `/api/v1/webhook/status` | Webhook service status |

### Quick API Example

```bash
# Send a message
curl -X POST http://localhost:8080/api/v1/send \
  -H "Content-Type: application/json" \
  -d '{
    "to": "friend@example.com",
    "body": "Hello from Jabber Bot! 🤖"
  }'
```

## 🐳 Docker Deployment

### Development Environment

```bash
# Start with live reload and monitoring
./scripts/deploy.sh dev
```

### Production Environment

```bash
# Production with monitoring stack
./scripts/deploy.sh prod
```

### Management Commands

```bash
# Show status
./scripts/deploy.sh status

# View logs
./scripts/deploy.sh logs

# Restart services
./scripts/deploy.sh restart

# Stop and cleanup
./scripts/deploy.sh clean
```

## 🧪 Testing

```bash
# All tests with coverage
make test-all

# Only unit tests
make test

# Only integration tests
make test-integration

# Windows
scripts\run-tests.bat
```

## 📊 Project Structure

```
jabber-bot/
├── cmd/server/          # Application entry point
├── internal/            # Private application packages
│   ├── api/           # REST API handlers
│   ├── config/        # Configuration management
│   ├── models/        # Data models
│   ├── webhook/       # Webhook service
│   └── xmpp/          # XMPP client
├── pkg/logger/         # Public logging utilities
├── configs/            # Configuration files
├── docs/               # Complete documentation
├── scripts/            # Build and deployment scripts
├── test/               # Integration tests
├── docker/             # Docker configurations
└── Makefile            # Build automation
```

## ⚙️ Configuration

Basic configuration via `.env`:

```bash
# XMPP Settings (Required)
JABBER_BOT_XMPP_JID=your-bot@your-xmpp-server.com
JABBER_BOT_XMPP_PASSWORD=your-secure-password
JABBER_BOT_XMPP_SERVER=your-xmpp-server.com:5222

# Webhook Settings (Required)
JABBER_BOT_WEBHOOK_URL=https://your-webhook-endpoint.com/receive

# API Settings
JABBER_BOT_API_PORT=8080
JABBER_BOT_LOG_LEVEL=info
```

Advanced configuration via `configs/config.yaml`:

```yaml
xmpp:
  jid: "bot@company.com"
  password: "${JABBER_BOT_PASSWORD}"
  server: "xmpp.company.com:5222"
  resource: "production-bot"

api:
  port: 8080
  host: "0.0.0.0"

webhook:
  url: "https://api.company.com/webhooks/jabber"
  timeout: 30s
  retry_attempts: 5

logging:
  level: "info"
  output: "file"
  file_path: "/var/log/jabber-bot/production.log"

reconnection:
  enabled: true
  max_attempts: 10
  backoff: "5s"
```

## 🏃‍♂️ Build & Run

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run with default config
make run

# Run with custom config
./bin/jabber-bot -config configs/config.yaml
```

## 📈 Monitoring

When using the Docker monitoring stack:

- **Grafana Dashboard**: http://localhost:3000 (admin/admin)
- **Prometheus Metrics**: http://localhost:9090
- **Health Checks**: http://localhost:8080/api/v1/health

### Health Check Response

```json
{
  "status": "ok",
  "timestamp": "2023-12-01T12:00:00Z"
}
```

### Status Response

```json
{
  "xmpp_connected": true,
  "api_running": true,
  "webhook_url": "https://example.com/webhook",
  "version": "1.0.0"
}
```

## 🔗 n8n Integration

[Jabber Bot](https://github.com/your-org/jabber-bot) integrates seamlessly with [n8n](https://n8n.io/) for powerful workflow automation. Use n8n to create complex XMPP-based workflows, notifications, and automated responses.

### Quick Setup

1. **Install n8n** (if not already installed):
   ```bash
   # Docker (recommended)
   docker run -it --rm \
     --name n8n \
     -p 5678:5678 \
     n8nio/n8n
   
   # Or npm
   npm install n8n -g
   n8n start
   ```

2. **Configure Jabber Bot** with webhook URL pointing to n8n:
   ```yaml
   webhook:
     url: "http://n8n:5678/webhook/jabber-incoming"
   ```

3. **Start n8n**: Open http://localhost:5678 in your browser

### Basic n8n Workflow Examples

#### 1. XMPP Message Forwarding

Create a simple workflow that forwards incoming XMPP messages to Slack:

```javascript
// n8n Webhook Node (POST /webhook/jabber-incoming)
// Trigger: HTTP Request Node (POST /api/v1/send)

// Incoming webhook data structure:
{
  "from": "user@example.com",
  "body": "Hello!",
  "timestamp": "2023-12-01T12:00:00Z"
}

// Slack notification:
{
  "channel": "#xmpp-alerts",
  "text": "New XMPP message from {{$json.from}}: {{$json.body}}"
}
```

#### 2. Automated Response Workflow

Create an automated response system based on message content:

**Flow**: XMPP Message → Content Check → Conditional Response → Send Reply

```javascript
// Function Node for content analysis
const messages = [
  {
    condition: msg => msg.body.toLowerCase().includes('hello'),
    response: "Hello! I'm an automated bot. How can I help you?"
  },
  {
    condition: msg => msg.body.toLowerCase().includes('status'),
    response: "System status: All operational ✅"
  },
  {
    condition: msg => msg.body.toLowerCase().includes('help'),
    response: "Available commands: hello, status, help"
  }
];

const matchedMessage = messages.find(msg => msg.condition($input.first()));
const response = matchedMessage ? matchedMessage.response : "I didn't understand that. Try 'help'.";

return [{ json: { to: $input.first().from, body: response } }];
```

#### 3. Multi-Channel Broadcasting

Broadcast XMPP messages to multiple platforms:

```javascript
// After XMPP webhook triggers:
const message = $input.first().body;
const sender = $input.first().from;

// Send to Discord
await axios.post('https://discord.com/api/webhooks/YOUR_WEBHOOK', {
  content: `📧 XMPP from ${sender}: ${message}`
});

// Send to Telegram
await axios.post(`https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/sendMessage`, {
  chat_id: TELEGRAM_CHAT_ID,
  text: `📧 XMPP from ${sender}: ${message}`
});

// Send email (optional)
await axios.post('https://api.sendgrid.com/v3/mail/send', {
  personalizations: [{
    to: [{ email: 'admin@example.com' }],
    subject: 'New XMPP Message'
  }],
  from: { email: 'bot@example.com' },
  content: [{
    type: 'text/plain',
    value: `From: ${sender}\nMessage: ${message}`
  }]
});
```

### n8n Node Configuration

#### HTTP Request Node (Send XMPP Message)

```json
{
  "method": "POST",
  "url": "http://jabber-bot:8080/api/v1/send",
  "headers": {
    "Content-Type": "application/json",
    "API-Key": "YOUR_JABBER_BOT_API_KEY"
  },
  "body": {
    "to": "={{$json.to}}",
    "body": "={{$json.body}}",
    "type": "chat"
  }
}
```

#### Webhook Node (Receive XMPP Messages)

```json
{
  "path": "jabber-incoming",
  "httpMethod": "POST",
  "responseMode": "responseNode",
  "options": {
    "rawBody": true
  }
}
```

### Advanced Workflows

#### 1. Message Routing by Content Type

```javascript
// Function Node - Message Router
const message = $input.first().body.toLowerCase();
const sender = $input.first().from;

let route;

if (message.includes('urgent') || message.includes('emergency')) {
  route = 'urgent';
} else if (message.includes('meeting') || message.includes('schedule')) {
  route = 'scheduling';
} else if (message.includes('invoice') || message.includes('payment')) {
  route = 'billing';
} else {
  route = 'general';
}

return [{
  json: {
    from: sender,
    body: $input.first().body,
    route: route,
    timestamp: new Date().toISOString()
  }
}];
```

#### 2. Database Integration

Store and retrieve conversation history:

```javascript
// Save to database (SQLite, PostgreSQL, etc.)
const dbMessage = {
  id: crypto.randomUUID(),
  from: $input.first().from,
  body: $input.first().body,
  timestamp: new Date().toISOString(),
  processed: false
};

// Insert into database
await db.insert('xmpp_messages', dbMessage);

return [{ json: { message: "Saved to database", id: dbMessage.id } }];
```

#### 3. AI-Powered Responses

Integrate with OpenAI/Claude for intelligent responses:

```javascript
// OpenAI Integration Node
const openai = require('openai');
const client = new openai.OpenAI({ apiKey: YOUR_OPENAI_KEY });

const response = await client.chat.completions.create({
  model: "gpt-3.5-turbo",
  messages: [
    {
      role: "system",
      content: "You are a helpful XMPP bot assistant. Be concise and friendly."
    },
    {
      role: "user",
      content: $input.first().body
    }
  ],
  max_tokens: 150
});

const reply = response.choices[0].message.content;

return [{ 
  json: { 
    to: $input.first().from, 
    body: reply 
  } 
}];
```

### n8n Cron Jobs for XMPP

#### Scheduled Announcements

```javascript
// Cron Node: "0 9 * * 1-5" (Every weekday at 9 AM)
const announcement = `☀️ Good morning! Today's reminders:
• Standup meeting at 10 AM
• Code review deadline: 3 PM
• Deploy window: 6-8 PM`;

return [{
  json: {
    to: "team-chat@conference.example.com",
    body: announcement
  }
}];
```

#### Health Check Notifications

```javascript
// Check Jabber Bot health
const healthCheck = await axios.get('http://jabber-bot:8080/api/v1/health');

if (healthCheck.data.status !== 'ok') {
  // Send alert
  await axios.post('http://jabber-bot:8080/api/v1/send', {
    to: 'admin@example.com',
    body: '⚠️ Jabber Bot health check failed! Please check the service.'
  });
}
```

### Environment Variables for n8n

```bash
# n8n Configuration
N8N_BASIC_AUTH_ACTIVE=true
N8N_BASIC_AUTH_USER=admin
N8N_BASIC_AUTH_PASSWORD=your-secure-password

# XMPP Bot Integration
JABBER_BOT_API_URL=http://jabber-bot:8080
JABBER_BOT_API_KEY=your-api-key

# External Services
TELEGRAM_BOT_TOKEN=your-telegram-bot-token
TELEGRAM_CHAT_ID=your-chat-id
DISCORD_WEBHOOK_URL=your-discord-webhook
OPENAI_API_KEY=your-openai-key
SENDGRID_API_KEY=your-sendgrid-key
```

### Best Practices

1. **Error Handling**: Always include error handling in n8n workflows
2. **Rate Limiting**: Implement delays between API calls to avoid overwhelming services
3. **Security**: Store sensitive keys in n8n credentials, not in workflow JSON
4. **Logging**: Use n8n's built-in execution logs to debug workflows
5. **Testing**: Test workflows with test webhooks before connecting to production XMPP

### Troubleshooting

**Common Issues:**
- **Webhook not triggered**: Check n8n URL and Jabber Bot webhook configuration
- **API Authentication**: Verify API key in HTTP Request nodes
- **Message format**: Ensure JSON payload matches Jabber Bot API format
- **Network connectivity**: Ensure n8n can reach Jabber Bot API endpoints

**Debug Tips:**
- Use n8n's "Execute Workflow" feature with test data
- Check Jabber Bot logs for webhook delivery status
- Monitor n8n execution logs for errors
- Test API endpoints with curl before creating workflows

---

## 🌐 Language Examples

### Python

```python
import requests

def send_message(to, body):
    response = requests.post(
        "http://localhost:8080/api/v1/send",
        json={"to": to, "body": body}
    )
    return response.json()

send_message("friend@example.com", "Hello from Python!")
```

### Node.js

```javascript
const axios = require('axios');

await axios.post('http://localhost:8080/api/v1/send', {
    to: 'friend@example.com',
    body: 'Hello from Node.js!'
});
```

### Go

```go
resp, err := http.Post(
    "http://localhost:8080/api/v1/send",
    "application/json",
    bytes.NewBuffer([]byte(`{
        "to": "friend@example.com",
        "body": "Hello from Go!"
    }`))
```

## 🛡️ Security

- 🔒 Non-root Docker containers
- 🚫 No hardcoded secrets
- 🔐 TLS support for XMPP connections
- 🛡️ Input validation and sanitization
- 📊 Structured audit logging
- 🚦 Rate limiting readiness (configuration available)

## 🤝 Contributing

We welcome contributions! Please see our [Development Guide](docs/DEVELOPMENT.md) for:

- Development setup
- Coding standards
- Testing requirements
- Pull request process

### Quick Development Setup

```bash
# Clone repository
git clone https://github.com/your-org/jabber-bot.git
cd jabber-bot

# Install dependencies
make deps
make install-tools

# Run tests
make test-all

# Start development
make run-dev
```

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙋‍♂️ Support

- 📖 [Documentation](docs/README.md)
- 🐛 [Issues](https://github.com/your-org/jabber-bot/issues)
- 💬 [Discussions](https://github.com/your-org/jabber-bot/discussions)
- 📧 [Email](mailto:support@your-org.com)

---

<div align="center">
  <strong>🤖 Built with ❤️ for the XMPP community</strong>
</div>