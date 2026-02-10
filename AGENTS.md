# AGENTS.md - Jabber Bot Development Guide

This document provides comprehensive information for AI agents and developers working with the Jabber Bot codebase.

## 📁 Project Structure

```
jabber-bot/
├── cmd/server/           # Application entry points
│   └── main.go         # Main application entry point
├── internal/            # Private application code
│   ├── api/           # REST API server and handlers
│   │   ├── handlers.go     # HTTP request handlers
│   │   ├── server.go       # API server setup
│   │   ├── handlers_test.go
│   │   └── server_test.go
│   ├── config/        # Configuration management
│   │   ├── config.go       # Configuration structures and loading
│   │   └── config_test.go
│   ├── models/        # Data models and structures
│   │   └── models.go       # Request/response models
│   ├── webhook/       # Webhook system
│   │   ├── manager.go      # Webhook manager
│   │   ├── service.go      # Webhook service implementation
│   │   ├── manager_test.go
│   │   └── service_test.go
│   └── xmpp/          # XMPP client implementation
│       ├── client.go       # XMPP client
│       ├── manager.go      # XMPP manager
│       ├── client_test.go
│       └── manager_test.go
├── pkg/                # Public library code
│   └── logger/        # Logging utilities
│       └── logger.go       # Logger setup and configuration
├── configs/            # Configuration files
│   └── config.yaml    # Default configuration
├── docs/               # Documentation
│   ├── API.md          # API documentation
│   ├── DEVELOPMENT.md  # Development guide
│   ├── DEPLOYMENT.md   # Deployment instructions
│   └── README.md       # Documentation index
├── scripts/            # Build and deployment scripts
│   ├── build.bat       # Windows build script
│   ├── deploy.sh       # Deployment script
│   ├── run-tests.bat   # Test runner (Windows)
│   └── run-tests.sh    # Test runner (Unix)
├── test/               # Test files
│   └── integration/    # Integration tests
├── docker/             # Docker configurations
├── .github/            # GitHub workflows
│   └── workflows/      # CI/CD pipelines
├── bin/                # Build output directory
├── openapi.yaml        # OpenAPI specification (YAML)
├── openapi.json        # OpenAPI specification (JSON)
├── Makefile            # Build commands
├── go.mod              # Go module definition
└── .air.toml           # Air live reload configuration
```

## 🚀 Quick Start Commands

### Essential Commands (Use These First)

**Build and Run:**
```bash
make run              # Build and run with default config
go run cmd/server/main.go -config configs/config.yaml  # Direct run
```

**Testing:**
```bash
make test             # Run unit tests
make test-integration # Run integration tests
make test-all         # Run all tests (unit + integration)
make quick-test       # Quick unit tests (no coverage)
```

**Code Quality:**
```bash
make fmt              # Format Go code
make lint             # Run golangci-lint
make clean            # Clean build artifacts
```

**Development:**
```bash
make deps             # Download dependencies
air                   # Live reload development (requires air)
```

## 🔧 Development Workflow

### 1. Setting Up Development Environment

**Configuration:**
- Main config: `configs/config.yaml`
- Environment variables: Override config values
- Environment file: `.env.example` (copy to `.env`)

### 2. Running the Application

**For development with live reload:**
```bash
# Install air first
go install github.com/cosmtrek/air@latest
air  # Uses .air.toml configuration
```

**For production builds:**
```bash
make build            # Build for current platform
make build-all        # Build for all platforms
```

**Docker development:**
```bash
make docker-build     # Build Docker image
make docker-run       # Start with docker-compose
make docker-stop      # Stop containers
```

### 3. Testing Strategy

**Unit Tests:**
- Located: `*_test.go` files alongside source files
- Run: `make test` or `make quick-test`
- Coverage: `make test-coverage`

**Integration Tests:**
- Located: `test/integration/`
- Run: `make test-integration` (requires `INTEGRATION_TESTS=1`)
- Tag: Use `-tags=integration` build tag

**Test Requirements:**
- XMPP test server (configurable)
- Mock webhook endpoints
- Test configuration files

## 🏗️ Build System

### Makefile Commands

**Core Commands:**
```bash
make build            # Build binary to bin/jabber-bot
make build-all        # Cross-platform builds
make run              # Build and run
make clean            # Clean artifacts
```

**Testing Commands:**
```bash
make test             # Unit tests
make test-coverage    # Tests with HTML coverage report
make test-integration # Integration tests
make test-all         # All test suites
```

**Quality Commands:**
```bash
make fmt              # Format code
make lint             # Lint code
make generate         # Generate code
```

**Docker Commands:**
```bash
make docker-build     # Build image
make docker-run       # Run containers
make docker-stop      # Stop containers
```

### Platform-Specific Scripts

**Windows (.bat):**
- `scripts/build.bat` - Build for Windows
- `scripts/run-tests.bat` - Run tests
- `scripts/deploy.bat` - Deploy

**Unix (.sh):**
- `scripts/run-tests.sh` - Run tests
- `scripts/deploy.sh` - Deploy

## ⚙️ Configuration Management

### Configuration Files

**Primary Config:** `configs/config.yaml`
```yaml
xmpp:
  jid: "bot@jabber.skbis.ru"
  password: "password"
  server: "jabber.skbis.ru:5222"
  resource: "bot"

api:
  port: 8080
  host: "0.0.0.0"

webhook:
  url: "https://example.com/webhook"
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

**Environment Variables:**
- `JABBER_BOT_XMPP_JID`
- `JABBER_BOT_XMPP_PASSWORD`
- `JABBER_BOT_XMPP_SERVER`
- `JABBER_BOT_API_PORT`
- `JABBER_BOT_API_HOST`
- `JABBER_BOT_WEBHOOK_URL`

### Configuration Loading

Configuration is loaded in this priority order:
1. Command line flags
2. Environment variables
3. Configuration file (`configs/config.yaml`)
4. Default values

## 🔌 API Development

### API Endpoints Structure

**Base URL:** `http://localhost:8080/api/v1`

**Core Endpoints:**
- `POST /send` - Send XMPP message
- `POST /send-muc` - Send MUC message
- `GET /status` - Get bot status
- `GET /health` - Health check
- `GET /webhook/status` - Webhook status

**Documentation Endpoints:**
- `GET /` - API root info
- `GET /docs` - Plain text documentation
- `GET /openapi.yaml` - OpenAPI spec (YAML)
- `GET /openapi.json` - OpenAPI spec (JSON)

### Adding New Endpoints

1. **Add handler function** in `internal/api/handlers.go`
2. **Register route** in `internal/api/server.go` `setupRoutes()`
3. **Add tests** in `internal/api/handlers_test.go`
4. **Update OpenAPI spec** in `openapi.yaml` and `openapi.json`

### Request/Response Models

**Location:** `internal/models/models.go`

**Common Patterns:**
```go
type SendMessageRequest struct {
    To   string `json:"to" validate:"required"`
    Body string `json:"body" validate:"required"`
    Type string `json:"type,omitempty"`
}

type APIResponse struct {
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

## 📡 XMPP Integration

### XMPP Client Architecture

**Files:**
- `internal/xmpp/client.go` - Low-level XMPP client
- `internal/xmpp/manager.go` - High-level XMPP manager

**Key Features:**
- Automatic reconnection
- Message routing and processing
- MUC (group chat) support
- Presence management

### XMPP Configuration

**Required Settings:**
- JID (username@domain)
- Password
- Server address (domain:port)
- Resource (optional)

### Testing XMPP Features

**Test Setup:**
- Use `test/integration/` for integration tests
- Mock XMPP server for unit tests
- Test with real XMPP server for integration

## 🪝 Webhook System

### Webhook Architecture

**Files:**
- `internal/webhook/manager.go` - Webhook manager
- `internal/webhook/service.go` - Webhook service

**Features:**
- HTTP webhook delivery
- Retry mechanism
- Queue management
- Statistics tracking

### Webhook Configuration

**Settings:**
- URL endpoint
- Timeout
- Retry attempts
- Success/failure handling

## 🐳 Docker Development

### Docker Compose Services

**Development:** `docker-compose.dev.yml`
- Application container
- XMPP server (Prosody)
- Monitoring tools

**Production:** `docker-compose.prod.yml`
- Production-ready configuration
- SSL/TLS setup
- External service connections

### Build Commands

```bash
# Development
docker-compose -f docker-compose.dev.yml up -d

# Production
docker-compose -f docker-compose.prod.yml up -d

# Build and push
make docker-build
docker build -t jabber-bot .
```

## 🔒 Security Considerations

### Current Security Features
- Input validation
- Error handling
- Request ID tracking
- TLS support for XMPP

### Security Improvements Needed
- API authentication (API keys)
- Rate limiting
- HTTPS enforcement
- Input sanitization
- Audit logging

### Security Testing

**Commands:**
```bash
# Run security scan (in CI/CD)
gosec ./...

# Check for vulnerabilities
go list -json -m all | nancy sleuth
```

## 📊 Monitoring and Logging

### Logging Configuration

**Logger Setup:** `pkg/logger/logger.go`

**Log Levels:**
- `debug` - Detailed debugging information (includes XMPP stream logs)
- `info` - General information (default)
- `warn` - Warning messages
- `error` - Error messages only

**Log Outputs:**
- `stdout` - Standard output (default)
- `stderr` - Standard error
- `file` - Log to file (specify `file_path`)

**XMPP Stream Logging:**
- XMPP stream data (SEND/RECV packets) is logged at debug level
- Set `logging.level: "debug"` in config to see XMPP communication
- Stream logs appear with field name "XMPP stream" and show raw XML packets
- Useful for debugging XMPP connections, authentication, and message flow

### Monitoring Endpoints

**Health Check:** `GET /api/v1/health`
**Status:** `GET /api/v1/status`
**Webhook Status:** `GET /api/v1/webhook/status`

## 🔄 CI/CD Pipeline

### GitHub Actions Workflow

**File:** `.github/workflows/ci-cd.yml`

**Pipeline Stages:**
1. **Test** - Unit tests, coverage, upload to Codecov
2. **Integration Test** - Integration test suite
3. **Security Scan** - Gosec security scanner
4. **Build** - Cross-platform builds
5. **Docker Build** - Multi-platform Docker images
6. **Deploy** - Production deployment (main branch)
7. **Release** - GitHub release creation

### Build Matrix

**Platforms:**
- Linux (amd64, arm64)
- Windows (amd64)
- macOS (amd64, arm64)

**Outputs:**
- Binary artifacts
- Docker images
- Coverage reports

## 🛠️ Development Tools

### Required Tools

**Core:**
- Go 1.25+
- Git

**Development:**
- `air` - Live reload
- `golangci-lint` - Linting
- `gosec` - Security scanning

**Optional:**
- Docker & Docker Compose
- XMPP client for testing

### IDE Configuration

**Recommended Extensions:**
- Go extension for VS Code
- Docker extension
- YAML extension

**Settings:**
- Go format on save
- Linting integration
- Debugging configuration

## 📝 Code Style and Conventions

### Go Conventions

**Package Structure:**
- Use `internal/` for private packages
- Use `pkg/` for public packages
- Follow standard Go project layout

**Naming:**
- PascalCase for exported
- camelCase for unexported
- Use descriptive names

**Error Handling:**
- Always handle errors
- Use wrapped errors with context
- Log errors appropriately

### Testing Conventions

**Test Files:**
- Name: `*_test.go`
- Location: Same package as source
- Use table-driven tests

## 🚀 Deployment

### Production Deployment

**Docker:**
```bash
# Build production image
docker build -f Dockerfile.prod -t jabber-bot:prod .

# Run with production config
docker run -p 8080:8080 -v ./configs:/app/configs jabber-bot:prod
```

**Binary:**
```bash
# Build for target platform
GOOS=linux GOARCH=amd64 make build

# Deploy binary
scp bin/jabber-bot user@server:/opt/jabber-bot/
```

### Configuration Management

**Environment:**
- Use environment-specific configs
- Secure secrets management
- Configuration validation

**Health Checks:**
- Implement readiness probes
- Liveness checks
- Monitoring integration

## 📚 Additional Resources

### Documentation

- `docs/API.md` - API documentation
- `docs/DEVELOPMENT.md` - Development guide
- `docs/DEPLOYMENT.md` - Deployment instructions
- `openapi.yaml` - OpenAPI specification

### External Documentation

- [Go Documentation](https://golang.org/doc/)
- [Fiber Framework](https://docs.gofiber.io/)
- [XMPP Protocol](https://xmpp.org/)
- [Viper Configuration](https://github.com/spf13/viper)

### Community

- GitHub Issues - Report bugs and request features
- GitHub Discussions - Community support
- Go Community - Go language support

---

## 🎯 Quick Reference

**Key Files:**
- `configs/config.yaml` - Main configuration
- `cmd/server/main.go` - Application entry
- `internal/api/handlers.go` - API handlers
- `Makefile` - Build commands
- `openapi.yaml` - API specification

**Debugging Tips:**
- Use `make test-coverage` for detailed test reports
- Check logs in `stdout` or configured file
- Use `air` for live development
- Use integration tests for end-to-end validation

This guide should help any AI agent or developer quickly understand and work with the Jabber Bot codebase effectively.