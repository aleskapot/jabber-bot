# Development Guide

This guide covers development setup, coding standards, and contribution guidelines for Jabber Bot.

## Table of Contents
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Debugging](#debugging)
- [Architecture](#architecture)
- [Adding Features](#adding-features)
- [Performance](#performance)
- [Contributing](#contributing)

## Development Setup

### Prerequisites

- **Go**: 1.21 or later
- **Docker**: 20.10+ (for integration tests)
- **Git**: Recent version
- **Make**: For build automation

### Quick Setup

```bash
# Clone repository
git clone https://github.com/your-org/jabber-bot.git
cd jabber-bot

# Install dependencies
make deps

# Install development tools
make install-tools

# Run tests
make test

# Build application
make build

# Run development environment
make run
```

### IDE Setup

#### VS Code

Install these extensions:
- Go (golang.go)
- Docker (ms-azuretools.vscode-docker)
- Makefile Tools (ms-vscode.makefile-tools)

Configure workspace settings (`.vscode/settings.json`):

```json
{
  "go.useLanguageServer": true,
  "go.formatTool": "goimports",
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "go.testFlags": ["-v"],
  "go.coverOnSave": true,
  "files.exclude": {
    "**/bin": true,
    "**/coverage.out": true,
    "**/coverage.html": true
  }
}
```

#### GoLand/IntelliJ

- Import project as Go module
- Enable goimports on save
- Configure golangci-lint
- Set up test configuration

### Environment Configuration

Create development configuration:

```bash
# Copy development config
cp configs/config.yaml configs/dev-config.yaml

# Edit for your development setup
vim configs/dev-config.yaml
```

Example development config:

```yaml
# Development XMPP settings (use test server)
xmpp:
  jid: "dev-bot@localhost"
  password: "dev-password"
  server: "localhost:5222"
  resource: "dev-bot"

# Development API
api:
  port: 8081  # Different port
  host: "127.0.0.1"

# Local webhook (for testing)
webhook:
  url: "http://localhost:3000/webhook"
  timeout: 5s
  retry_attempts: 1

# Debug logging
logging:
  level: "debug"
  output: "stdout"

# Fast reconnection for dev
reconnection:
  enabled: true
  max_attempts: 3
  backoff: "2s"
```

## Project Structure

```
jabber-bot/
├── cmd/server/          # Application entry point
├── internal/            # Internal packages
│   ├── api/           # REST API handlers
│   ├── config/        # Configuration management
│   ├── models/        # Data models
│   ├── webhook/       # Webhook service
│   └── xmpp/          # XMPP client
├── pkg/                # Public packages
│   └── logger/        # Logging utilities
├── configs/            # Configuration files
├── docs/               # Documentation
├── scripts/            # Build and deploy scripts
├── test/               # Test files
│   └── integration/   # Integration tests
├── docker/             # Docker configs
├── .env.example        # Environment template
├── Makefile            # Build automation
├── go.mod              # Go module definition
├── go.sum              # Dependency checksums
└── README.md           # Project documentation
```

### Package Responsibilities

- **cmd/**: Application entry points
- **internal/**: Private application code
- **pkg/**: Reusable public packages
- **configs/**: Configuration files
- **docs/**: Project documentation
- **scripts/**: Automation scripts
- **test/**: Test utilities and integration tests

## Coding Standards

### Go Guidelines

Follow [Effective Go](https://golang.org/doc/effective_go.html) and the Go Code Review Comments:

```go
// Package declaration
package xmpp

// Import grouping
import (
    "context"
    "fmt"
    "time"

    "jabber-bot/internal/config"
    "github.com/fluux/go-xmpp"
)

// Constants
const (
    DefaultTimeout = 30 * time.Second
    MaxRetries     = 3
)

// Variables
var (
    ErrConnectionLost = fmt.Errorf("connection lost")
)

// Struct with comments
type Client struct {
    config *config.Config
    logger *zap.Logger
    client *xmpp.Client
}

// Method with context
func (c *Client) SendMessage(ctx context.Context, to, body string) error {
    // Implementation
}

// Interface
type MessageHandler interface {
    HandleMessage(msg Message) error
}
```

### Naming Conventions

- **Packages**: lowercase, single word when possible
- **Constants**: `UPPER_SNAKE_CASE` or `camelCase` for exported
- **Variables**: `camelCase`, descriptive names
- **Functions**: `PascalCase` for exported, `camelCase` for private
- **Interfaces**: `InterfaceName` or `Verb-er` pattern
- **Files**: `snake_case.go`

### Error Handling

```go
// Good: Use fmt.Errorf for context
func (c *Client) Connect() error {
    if c.config == nil {
        return fmt.Errorf("client config is required")
    }
    
    conn, err := xmpp.NewClient(c.config)
    if err != nil {
        return fmt.Errorf("failed to create XMPP client: %w", err)
    }
    
    c.client = conn
    return nil
}

// Good: Define error variables
var (
    ErrNotConnected = errors.New("client not connected")
    ErrInvalidJID   = errors.New("invalid JID format")
)

// Good: Use custom error types
type XMPPError struct {
    Code    string
    Message string
}

func (e *XMPPError) Error() string {
    return e.Message
}
```

### Logging

```go
import "go.uber.org/zap"

// Structured logging
logger.Info("User connected",
    zap.String("user_id", userID),
    zap.String("jid", jid),
    zap.Duration("connection_time", duration),
)

// Debug for development
logger.Debug("Processing message",
    zap.String("from", msg.From),
    zap.String("to", msg.To),
    zap.Int("body_length", len(msg.Body)),
)

// Error with context
logger.Error("Failed to send message",
    zap.Error(err),
    zap.String("to", recipient),
    zap.String("message_id", msgID),
)
```

## Testing

### Unit Tests

```go
package xmpp

import (
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "go.uber.org/zap/zaptest"
)

func TestClient_SendMessage(t *testing.T) {
    // Setup
    logger := zaptest.NewLogger(t)
    cfg := &config.Config{...}
    client := NewClient(cfg, logger)
    
    // Test
    err := client.SendMessage("test@example.com", "Hello", "chat")
    
    // Assertions
    require.NoError(t, err)
    assert.Equal(t, "chat", msg.Type)
}

func TestClient_SendMessage_NotConnected(t *testing.T) {
    // Test error case
    logger := zaptest.NewLogger(t)
    cfg := &config.Config{...}
    client := NewClient(cfg, logger)
    
    err := client.SendMessage("test@example.com", "Hello", "chat")
    
    // Should fail when not connected
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "not connected")
}
```

### Table-Driven Tests

```go
func TestValidateJID(t *testing.T) {
    tests := []struct {
        name    string
        jid     string
        wantErr bool
    }{
        {
            name:    "valid JID",
            jid:     "user@example.com",
            wantErr: false,
        },
        {
            name:    "missing domain",
            jid:     "user",
            wantErr: true,
        },
        {
            name:    "empty string",
            jid:     "",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateJID(tt.jid)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Mocking

```go
//go:generate mockery --name XMPPClient --output ./mocks
type XMPPClient interface {
    Send(msg Message) error
    Connect() error
    Disconnect() error
}

func TestService_ProcessMessage(t *testing.T) {
    // Create mock
    mockClient := &mocks.XMPPClient{}
    mockClient.On("Send", mock.Anything).Return(nil)
    
    // Test
    service := NewService(mockClient)
    err := service.ProcessMessage(msg)
    
    // Verify
    require.NoError(t, err)
    mockClient.AssertExpectations(t)
}
```

### Integration Tests

```go
//go:build integration
// +build integration

package integration_test

import (
    "testing"
    "time"
)

func TestEndToEnd(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    // Setup real components
    cfg := loadTestConfig()
    client := NewClient(cfg)
    
    // Connect
    err := client.Connect()
    require.NoError(t, err)
    defer client.Disconnect()
    
    // Test real flow
    err = client.SendMessage("test@example.com", "Hello", "chat")
    require.NoError(t, err)
    
    // Wait for processing
    time.Sleep(100 * time.Millisecond)
    
    // Verify results
    // ...
}
```

## Debugging

### Development Tools

```bash
# Install delve debugger
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug with delve
dlv debug ./cmd/server

# Or use VS Code debugging (F5)
```

### Logging

Enable debug logging:

```bash
# Environment variable
export JABBER_BOT_LOG_LEVEL=debug

# Or in config
logging:
  level: "debug"
  output: "stdout"
```

### Performance Profiling

```go
import (
    _ "net/http/pprof"
    "net/http"
)

// Add to main
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()

// Access profiles
# curl http://localhost:6060/debug/pprof/profile > cpu.prof
# go tool pprof cpu.prof
```

### Memory Analysis

```bash
# Build with memory profiling
go build -o jabber-bot ./cmd/server

# Run with memory profiling
JABBER_BOT_LOG_LEVEL=debug ./jabber-bot &
PID=$!

# Capture memory profile
curl http://localhost:6060/debug/pprof/heap > heap.prof

# Analyze
go tool pprof heap.prof
```

## Architecture

### Component Diagram

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

### Data Flow

```
External Client → API Server → XMPP Manager → XMPP Client → XMPP Server
                                                    ↓
Message Received ← XMPP Client ← XMPP Manager ← Webhook Service ← Webhook Endpoint
```

### Key Interfaces

```go
type XMPPClient interface {
    Connect(ctx context.Context) error
    Disconnect() error
    SendMessage(to, body, type string) error
    IsConnected() bool
    GetMessageChannel() <-chan Message
}

type WebhookService interface {
    SendMessage(msg Message) error
    GetStats() WebhookStats
    IsHealthy() bool
}

type ConfigManager interface {
    Load(path string) (*Config, error)
    Watch(callback func(*Config)) error
}
```

## Adding Features

### New API Endpoint

1. **Define request/response models:**
```go
// internal/models/models.go
type NewFeatureRequest struct {
    Param1 string `json:"param1"`
    Param2 int    `json:"param2"`
}

type NewFeatureResponse struct {
    Result string `json:"result"`
}
```

2. **Add handler:**
```go
// internal/api/handlers.go
func (s *Server) handleNewFeature(c *fiber.Ctx) error {
    var req NewFeatureRequest
    if err := c.BodyParser(&req); err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "Invalid request")
    }
    
    // Implement logic
    result := processNewFeature(req)
    
    response := NewFeatureResponse{Result: result}
    return c.JSON(response)
}
```

3. **Add route:**
```go
// internal/api/server.go
func (s *Server) setupRoutes() {
    api := s.app.Group("/api/v1")
    api.Post("/new-feature", s.handleNewFeature)
}
```

4. **Add tests:**
```go
// internal/api/handlers_test.go
func TestHandleNewFeature(t *testing.T) {
    // Test cases...
}
```

### New XMPP Feature

1. **Extend XMPP client interface:**
```go
type XMPPClient interface {
    // Existing methods...
    SendPresence(to, status string) error  // New method
}
```

2. **Implement in client:**
```go
func (c *Client) SendPresence(to, status string) error {
    if !c.isConnected() {
        return ErrNotConnected
    }
    
    presence := xmpp.Presence{To: to, Status: status}
    return c.client.Send(presence)
}
```

3. **Add API endpoint:**
```go
func (s *Server) handleSendPresence(c *fiber.Ctx) error {
    // Parse request
    // Call XMPP client
    // Return response
}
```

### Configuration Options

1. **Add to config struct:**
```go
type Config struct {
    // Existing fields...
    NewFeature NewFeatureConfig `mapstructure:"new_feature"`
}

type NewFeatureConfig struct {
    Enabled bool   `mapstructure:"enabled"`
    Option  string `mapstructure:"option"`
}
```

2. **Add default values:**
```go
func Load(configPath string) (*Config, error) {
    // ...
    if cfg.NewFeature.Option == "" {
        cfg.NewFeature.Option = "default"
    }
    // ...
}
```

3. **Add to config file template.**

## Performance

### Guidelines

- Use connection pooling for HTTP clients
- Implement proper timeout handling
- Use buffered channels for message passing
- Avoid memory leaks with proper cleanup
- Use sync.Pool for frequently allocated objects

### Benchmarks

```go
func BenchmarkSendMessage(b *testing.B) {
    client := setupTestClient()
    defer client.Disconnect()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        client.SendMessage("test@example.com", "benchmark message", "chat")
    }
}
```

### Memory Optimization

```go
// Use sync.Pool for message objects
var messagePool = sync.Pool{
    New: func() interface{} {
        return &Message{}
    },
}

func (c *Client) processMessage(raw string) {
    msg := messagePool.Get().(*Message)
    defer messagePool.Put(msg)
    
    // Reset and use message
    *msg = Message{}
    parseMessage(msg, raw)
    handleMessage(msg)
}
```

## Contributing

### Git Workflow

```bash
# Create feature branch
git checkout -b feature/new-feature

# Make changes
# Add tests
# Update documentation

# Run tests
make test
make test-integration
make lint

# Commit changes
git add .
git commit -m "feat: add new feature description"

# Push and create PR
git push origin feature/new-feature
```

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add new API endpoint for sending presence
fix: handle empty message bodies correctly
docs: update deployment guide
test: add integration tests for webhook service
refactor: simplify XMPP connection handling
```

### Pull Request Template

```markdown
## Description
Brief description of changes.

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing performed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests added/updated
```

### Code Review

- Focus on correctness, readability, and performance
- Ensure proper error handling
- Verify tests coverage
- Check for security issues
- Validate API compatibility

## Getting Help

- Check existing issues and documentation
- Join development discussions
- Ask questions in appropriate channels
- Follow contribution guidelines

For detailed questions, open an issue with the `question` label.