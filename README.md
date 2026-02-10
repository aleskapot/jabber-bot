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