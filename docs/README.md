# Documentation

Welcome to the Jabber Bot documentation hub.

## Table of Contents

### Getting Started
- [**Quick Start Guide**](QUICK_START.md) - Get running in minutes
- [API Examples](API_EXAMPLES.md) - Complete API usage examples
- [Deployment Guide](DEPLOYMENT.md) - Production deployment options

### Development
- [**Development Guide**](DEVELOPMENT.md) - Setup, coding standards, contribution
- [Architecture Overview](ARCHITECTURE.md) - System design and components
- [Testing Strategy](TESTING.md) - Testing approaches and guidelines

### Operations
- [Configuration Reference](CONFIGURATION.md) - All configuration options
- [Monitoring Guide](MONITORING.md) - Health checks, metrics, alerting
- [Troubleshooting Guide](TROUBLESHOOTING.md) - Common issues and solutions

### API Reference
- [REST API Documentation](API.md) - Complete API reference
- [Webhook Specification](WEBHOOK.md) - Incoming webhook format
- [Message Formats](MESSAGES.md) - Supported message types

### Deployment
- [Docker Deployment](DOCKER.md) - Container deployment guide
- [Kubernetes Deployment](KUBERNETES.md) - K8s manifests and guidance
- [Systemd Service](SYSTEMD.md) - Linux service setup

### Security
- [Security Best Practices](SECURITY.md) - Security considerations
- [Authentication Guide](AUTHENTICATION.md) - Securing the API
- [Network Security](NETWORK_SECURITY.md) - Firewall and network setup

### Examples
- [Integration Examples](INTEGRATION_EXAMPLES.md) - Real-world integrations
- [CI/CD Pipelines](CICD.md) - GitHub Actions, GitLab CI examples
- [Monitoring Setup](MONITORING_SETUP.md) - Prometheus/Grafana configuration

---

## Quick Navigation

### 🚀 I want to...
- **Start quickly** → [Quick Start Guide](QUICK_START.md)
- **Deploy to production** → [Deployment Guide](DEPLOYMENT.md)
- **Integrate with my app** → [API Examples](API_EXAMPLES.md)
- **Contribute to the project** → [Development Guide](DEVELOPMENT.md)
- **Troubleshoot issues** → [Troubleshooting Guide](TROUBLESHOOTING.md)

### 📚 Learn about...
- **System architecture** → [Architecture Overview](ARCHITECTURE.md)
- **Configuration options** → [Configuration Reference](CONFIGURATION.md)
- **API endpoints** → [REST API Documentation](API.md)
- **Message handling** → [Message Formats](MESSAGES.md)

### 🔧 Set up...
- **Development environment** → [Development Guide](DEVELOPMENT.md)
- **Monitoring stack** → [Monitoring Guide](MONITORING.md)
- **Docker deployment** → [Docker Deployment](DOCKER.md)
- **Kubernetes cluster** → [Kubernetes Deployment](KUBERNETES.md)

### 🛡️ Secure...
- **API endpoints** → [Authentication Guide](AUTHENTICATION.md)
- **Network traffic** → [Network Security](NETWORK_SECURITY.md)
- **Production deployment** → [Security Best Practices](SECURITY.md)

---

## Key Concepts

### XMPP Bot
A Go-based XMPP (Jabber) bot that provides a RESTful API for sending messages and webhook notifications for incoming messages.

### Core Components
- **XMPP Client**: Handles XMPP protocol communication
- **REST API**: HTTP endpoints for message sending
- **Webhook Service**: Forwards incoming messages to your endpoints
- **Configuration**: Flexible YAML and environment-based configuration
- **Logging**: Structured logging with Zap

### Architecture Pattern
```
External Clients → REST API → XMPP Client → XMPP Server
                                        ↓
Message Notifications ← Webhook Service ← XMPP Client
```

---

## Contributing to Documentation

We welcome contributions to improve the documentation!

### Adding New Documentation

1. Create markdown file in appropriate directory
2. Add to this README's table of contents
3. Include examples and code snippets
4. Test all commands and examples
5. Submit pull request with `docs` label

### Documentation Standards

- Use clear, concise language
- Include practical examples
- Add code blocks with syntax highlighting
- Provide troubleshooting sections
- Keep content up-to-date with code changes

### Review Process

- Technical accuracy review
- User experience validation
- Cross-platform compatibility check
- Link and reference verification

---

## Getting Help

### 📖 Documentation First
- Search existing documentation
- Check quick start and troubleshooting guides
- Review API examples and configuration reference

### 🐛 Report Issues
- Use GitHub Issues for bug reports
- Include logs and configuration
- Provide reproduction steps
- Specify environment details

### 💬 Community Support
- GitHub Discussions for questions
- Check existing threads before posting
- Share configuration examples (redact secrets)
- Help others when you can

### 📧 Direct Support
- Security vulnerabilities: security@your-org.com
- Enterprise support: enterprise@your-org.com
- General questions: support@your-org.com

---

## Documentation Structure

```
docs/
├── QUICK_START.md              # Getting started quickly
├── API_EXAMPLES.md           # API usage examples
├── DEPLOYMENT.md             # Deployment options
├── DEVELOPMENT.md            # Development guide
├── API.md                   # Complete API reference
├── CONFIGURATION.md          # All config options
├── SECURITY.md               # Security best practices
├── TROUBLESHOOTING.md       # Common issues
└── INTEGRATION_EXAMPLES.md   # Real-world examples
```

Each document follows a consistent structure:
- Overview/Purpose
- Prerequisites
- Step-by-step instructions
- Examples and code snippets
- Troubleshooting
- Related resources

---

## Version Compatibility

Documentation is maintained for:
- **Latest stable release** (recommended for production)
- **Development branch** (bleeding edge features)
- **Previous major versions** (legacy support)

Version indicators:
- 🆕 New features
- ⚠️ Breaking changes
- 🔄 Updated content
- 🐛 Bug fixes

---

This documentation is continuously improved based on user feedback and evolving project needs. Your contributions help make it better for everyone!