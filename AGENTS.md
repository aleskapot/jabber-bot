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
│   └── config.yaml    # Default configuration (⚠️  contains example credentials - change for production!)
├── docs/               # Documentation
│   ├── API.md          # API documentation
│   ├── API_EXAMPLES.md # API usage examples
│   ├── DEVELOPMENT.md  # Development guide
│   ├── DEPLOYMENT.md   # Deployment instructions
│   ├── QUICK_START.md  # Quick start guide
│   ├── README.md       # Documentation index
│   ├── openapi.yaml    # OpenAPI specification (YAML)
│   └── openapi.json    # OpenAPI specification (JSON)
├── scripts/            # Build and deployment scripts
│   ├── build.bat       # Windows build script
│   ├── deploy.bat      # Windows deployment script
│   ├── deploy.sh       # Unix deployment script
│   ├── run-tests.bat   # Test runner (Windows)
│   ├── run-tests.sh    # Test runner (Unix)
│   ├── run-integration-tests.bat  # Integration tests (Windows)
│   └── run-integration-tests.sh   # Integration tests (Unix)
├── test/               # Test files
│   └── integration/    # Integration tests
├── docker/             # Docker configurations
│   ├── grafana/        # Grafana dashboards and datasources
│   ├── prometheus.yml  # Prometheus configuration
│   └── prosody/        # XMPP test server configuration
│       └── prosody.cfg.lua
├── .github/            # GitHub workflows
│   └── workflows/      # CI/CD pipelines
│       └── tests-codeql.yml  # CodeQL analysis and test workflow
├── bin/                # Build output directory
├── Makefile            # Build commands
├── go.mod              # Go module definition
├── go.sum              # Go module checksums
├── .air.toml           # Air live reload configuration
├── .env.example        # Environment variables template
├── docker-compose.yml  # Docker Compose configuration (single file for all environments)
├── Dockerfile          # Docker image definition (multi-stage build)
├── test-api.sh         # API testing script
```

## 🛠️ Technology Stack

### Core Technologies

**Go Version:** 1.26.0

**Main Dependencies:**
- `github.com/gofiber/fiber/v2 v2.52.11` - Web framework
- `github.com/spf13/viper v1.21.0` - Configuration management
- `go.uber.org/zap v1.27.1` - Structured logging
- `gosrc.io/xmpp v0.5.1` - XMPP client library
- `github.com/stretchr/testify v1.11.1` - Testing framework

**Development Tools:**
- `air` - Live reload development
- `golangci-lint` - Go linting
- `gosec` - Security scanning

### Architecture Patterns

- **Clean Architecture** - Separation of concerns with internal/pkg structure
- **Dependency Injection** - Config-based service initialization
- **Graceful Shutdown** - Context-based lifecycle management
- **Structured Logging** - Zap-based consistent logging
- **Configuration Management** - Viper with multiple sources (file, env, flags)

## 🚀 Quick Start Commands

### Essential Commands (Use These First)

**Build and Run:**
```bash
make run              # Build and run with default config
make build            # Build binary only
make build-all        # Build for all platforms
go run cmd/server/main.go -config configs/config.yaml  # Direct run
```

**Testing:**
```bash
make test             # Run unit tests
make test-coverage    # Run tests with coverage report
make test-integration # Run integration tests
make test-all         # Run all tests (unit + integration)
make quick-test       # Quick unit tests (no coverage)
make quick-integration # Quick integration tests
```

**Code Quality:**
```bash
make fmt              # Format Go code
make lint             # Run golangci-lint
make clean            # Clean build artifacts
make generate         # Generate code
```

**Dependencies:**
```bash
make deps             # Download dependencies
make install-tools    # Install development tools
```

**Docker:**
```bash
make docker-build     # Build Docker image
make docker-run       # Start Docker containers
make docker-stop      # Stop Docker containers
```

**Development:**
```bash
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
- Run: `make test-integration` or `make quick-integration` (requires `INTEGRATION_TESTS=1`)
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
make quick-test       # Quick unit tests (no coverage)
make quick-integration # Quick integration tests
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

**Dependency Commands:**
```bash
make deps             # Download dependencies
make install-tools    # Install development tools
```

### Platform-Specific Scripts

**Windows (.bat):**
- `scripts/build.bat` - Build for Windows
- `scripts/run-tests.bat` - Run tests
- `scripts/run-integration-tests.bat` - Run integration tests
- `scripts/deploy.bat` - Deploy

**Unix (.sh):**
- `scripts/run-tests.sh` - Run tests
- `scripts/run-integration-tests.sh` - Run integration tests
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
  api_key: "your-secret-api-key-change-in-production"
  auth_enabled: true

webhook:
  url: "https://example.com/webhook/somepath"
  timeout: 30s
  retry_attempts: 3
  test_mode_suffix: "-test"  # Suffix for webhook URLs when [test] prefix is detected

logging:
  level: "debug"  # debug, info, warn, error
  output: "stdout"  # stdout, stderr, file
  file_path: ""  # for output=file
  
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
- `JABBER_BOT_API_KEY`
- `JABBER_BOT_WEBHOOK_URL`
- `JABBER_BOT_LOG_LEVEL`
- `JABBER_BOT_LOG_OUTPUT`

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
- `GET /docs/openapi.yaml` - OpenAPI spec (YAML)
- `GET /docs/openapi.json` - OpenAPI spec (JSON)

### Adding New Endpoints

1. **Add handler function** in `internal/api/handlers.go`
2. **Register route** in `internal/api/server.go` `setupRoutes()`
3. **Add tests** in `internal/api/handlers_test.go`
4. **Update OpenAPI spec** in `docs/openapi.yaml` and `docs/openapi.json`

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

### Test XMPP Server (Prosody)

**Configuration:** `docker/prosody/prosody.cfg.lua`

The project includes a Prosody XMPP server configuration for testing:
- Allows registration for development (`allow_registration = true`)
- In-memory storage for easy cleanup
- Pre-configured admin user: `admin@localhost`
- Standard XMPP modules enabled

**Running Prosody:**
```bash
# Prosody can be started separately or via docker-compose
# Check docker-compose.yml for integration with the bot

# Manual Prosody setup (if needed)
docker run -p 5222:5222 -p 5280:5280 \
  -v $(pwd)/docker/prosody:/etc/prosody \
  prosody/prosody
```

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
- **Test Mode Detection** - Automatically adds suffix to webhook URLs when messages contain [test] prefix

### Webhook Configuration

**Settings:**
- URL endpoint
- Timeout
- Retry attempts
- Success/failure handling

## 🐳 Docker Development

### Docker Compose Services

**Configuration:** `docker-compose.yml` (single file for all environments)

The Docker Compose file includes:
- Application container with health checks
- Environment variable configuration
- Volume mounting support (commented out by default)
- Automatic restart policy

**Note:** The deployment scripts (`scripts/deploy.sh`, `scripts/deploy.bat`) reference `docker-compose.dev.yml` and `docker-compose.prod.yml` files, but these are not yet implemented. Currently, the single `docker-compose.yml` file is used for all environments with environment variables controlling configuration.

**Build Arguments:**
- `VERSION` - Application version (default: 1.0.0)
- `BUILD_TIME` - Build timestamp
- `GIT_COMMIT` - Git commit hash

**Environment Variables:**
All configuration can be passed via environment variables (see `.env.example`).

**Health Check:**
- Endpoint: `http://localhost:8080/api/v1/health`
- Interval: 30s
- Timeout: 10s
- Retries: 3
- Start period: 40s

### Build and Run Commands

```bash
# Build Docker image
make docker-build     # or: docker build -t jabber-bot .

# Start containers (if using docker-compose with additional services)
make docker-run       # or: docker-compose up -d

# Stop containers
make docker-stop      # or: docker-compose down

# Run with custom environment
docker run -p 8080:8080 \
  -e JABBER_BOT_XMPP_JID="bot@example.com" \
  -e JABBER_BOT_XMPP_PASSWORD="secret" \
  -e JABBER_BOT_WEBHOOK_URL="https://your-webhook.com" \
  jabber-bot
```

**Note:** The project uses a single `docker-compose.yml` file. For different environments, use environment variables or separate override files.

### Docker Image Details

**Multi-stage build:**
1. **Builder stage** - Uses `golang:1.26-alpine` to compile the application
2. **Production stage** - Uses `alpine:latest` with minimal runtime dependencies

**Security:**
- Runs as non-root user (UID 1001)
- Includes ca-certificates and tzdata
- CGO_ENABLED=0 for static binary

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

### Metrics and Dashboards

**Prometheus:** Configuration in `docker/prometheus.yml`
- Scrapes metrics from the application
- Pre-configured for Docker environment

**Grafana:** Dashboards in `docker/grafana/`
- Pre-built dashboards for monitoring
- Includes datasource configuration for Prometheus

**Setup:**
```bash
# Start monitoring stack with docker-compose
docker-compose up -d

# Access Grafana at http://localhost:3000
# Access Prometheus at http://localhost:9090
```

## 🔄 CI/CD Pipeline

### GitHub Actions Workflow

**File:** `.github/workflows/tests-codeql.yml`

**Pipeline Stages:**
1. **CodeQL Analysis** - Static code analysis for security vulnerabilities
2. **Test** - Unit tests with coverage, upload to Codecov
3. **Integration Test** - Integration test suite with XMPP server
4. **Security Scan** - Gosec security scanner (in progress)

**Workflow Triggers:**
- Push to `main` or `develop` branches
- Pull requests to `main` branch
- Published releases

**Build Environment:**
- Runner: `ubuntu-latest`
- Go version: `1.26`

**Test Matrix:**
- Unit tests: `make test`
- Integration tests: `INTEGRATION_TESTS=1 make test-integration`
- Coverage: `make test-coverage` with Codecov upload

**Security Analysis:**
- CodeQL for Go language
- Automatic build detection
- Security alerts to GitHub Security tab

### Artifacts

The CI pipeline produces:
- Test results and coverage reports
- CodeQL analysis results
- Security scanning reports

### Local Testing Before Push

```bash
# Run all tests locally
make test-all

# Check code formatting
make fmt

# Run linter
make lint

# Verify build
make build
```

## 🛠️ Development Tools

### Required Tools

**Core:**
- Go 1.26+
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
# Build Docker image
make docker-build     # or: docker build -t jabber-bot .

# Run with environment configuration
docker run -p 8080:8080 \
  -e JABBER_BOT_XMPP_JID="bot@example.com" \
  -e JABBER_BOT_XMPP_PASSWORD="secret" \
  -e JABBER_BOT_WEBHOOK_URL="https://your-webhook.com" \
  jabber-bot

# Or use docker-compose
docker-compose up -d
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
- Use environment variables for production secrets
- Never commit real credentials to version control
- Use `.env.example` as template for required variables
- Consider using secret management tools (HashiCorp Vault, AWS Secrets Manager, etc.)

**Health Checks:**
- Docker health check configured in docker-compose.yml
- Health endpoint: `GET /api/v1/health`
- Use in orchestration platforms (Kubernetes, ECS, etc.)

**Monitoring:**
- Application logs via configured output (stdout/file)
- Metrics endpoint available for Prometheus (see docker/prometheus.yml)
- Grafana dashboards available in docker/grafana/

## 📚 Additional Resources

### Documentation

- `docs/API.md` - API documentation
- `docs/API_EXAMPLES.md` - API usage examples
- `docs/DEVELOPMENT.md` - Development guide
- `docs/DEPLOYMENT.md` - Deployment instructions
- `docs/QUICK_START.md` - Quick start guide
- `docs/README.md` - Documentation index
- `docs/openapi.yaml` - OpenAPI specification (YAML)
- `docs/openapi.json` - OpenAPI specification (JSON)

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
- `docs/openapi.yaml` - API specification (YAML)
- `docs/openapi.json` - API specification (JSON)

**Debugging Tips:**
- Use `make test-coverage` for detailed test reports with HTML output
- Check logs in `stdout` or configured file
- Use `air` for live development
- Use `make test-integration` for end-to-end validation
- Use `make quick-test` for fast unit testing during development

This guide should help any AI agent or developer quickly understand and work with the Jabber Bot codebase effectively.