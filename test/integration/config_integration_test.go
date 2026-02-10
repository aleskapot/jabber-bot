//go:build integration
// +build integration

package integration

import (
	"fmt"
	"os"
	"testing"
	"time"

	"jabber-bot/internal/config"
	"jabber-bot/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

// ConfigIntegrationTestSuite tests configuration loading in integration context
type ConfigIntegrationTestSuite struct {
	suite.Suite
	logger *zap.Logger
}

// SetupSuite runs once before all tests
func (suite *ConfigIntegrationTestSuite) SetupSuite() {
	var err error
	suite.logger, err = logger.New()
	require.NoError(suite.T(), err)
}

// TestRealConfigFile tests loading a real configuration file
func (suite *ConfigIntegrationTestSuite) TestRealConfigFile() {
	// Create a temporary config file for testing
	configContent := `
xmpp:
  jid: "integration-test@localhost"
  password: "integration-password"
  server: "localhost:5222"
  resource: "integration-bot"

api:
  port: 9090
  host: "127.0.0.1"

webhook:
  url: "https://webhook.example.com/integration"
  timeout: 45s
  retry_attempts: 4

logging:
  level: "warn"
  output: "stdout"
  file_path: "/tmp/integration-test.log"

reconnection:
  enabled: true
  max_attempts: 8
  backoff: "15s"
`

	// Write to temporary file
	tempDir := os.TempDir()
	tempFile := fmt.Sprintf("%s%sintegration-test-config.yaml", tempDir, string(os.PathSeparator))
	err := os.WriteFile(tempFile, []byte(configContent), 0644)
	require.NoError(suite.T(), err)
	defer os.Remove(tempFile)

	// Load configuration
	cfg, err := config.Load(tempFile)
	require.NoError(suite.T(), err)

	// Verify all values are loaded correctly
	assert.Equal(suite.T(), "integration-test@localhost", cfg.XMPP.JID)
	assert.Equal(suite.T(), "integration-password", cfg.XMPP.Password)
	assert.Equal(suite.T(), "localhost:5222", cfg.XMPP.Server)
	assert.Equal(suite.T(), "integration-bot", cfg.XMPP.Resource)

	assert.Equal(suite.T(), 9090, cfg.API.Port)
	assert.Equal(suite.T(), "127.0.0.1", cfg.API.Host)

	assert.Equal(suite.T(), "https://webhook.example.com/integration", cfg.Webhook.URL)
	assert.Equal(suite.T(), 45*time.Second, cfg.Webhook.Timeout)
	assert.Equal(suite.T(), 4, cfg.Webhook.RetryAttempts)

	assert.Equal(suite.T(), "warn", cfg.Logging.Level)
	assert.Equal(suite.T(), "stdout", cfg.Logging.Output)
	assert.Equal(suite.T(), "/tmp/integration-test.log", cfg.Logging.FilePath)

	assert.True(suite.T(), cfg.Reconnection.Enabled)
	assert.Equal(suite.T(), 8, cfg.Reconnection.MaxAttempts)
	assert.Equal(suite.T(), 15*time.Second, cfg.Reconnection.Backoff)
}

// TestConfigWithEnvVars tests configuration with environment variables
func (suite *ConfigIntegrationTestSuite) TestConfigWithEnvVars() {
	// Set environment variables
	os.Setenv("JABBER_BOT_XMPP_JID", "env-bot@localhost")
	os.Setenv("JABBER_BOT_API_PORT", "9999")
	os.Setenv("JABBER_BOT_WEBHOOK_RETRY_ATTEMPTS", "10")
	os.Setenv("JABBER_BOT_LOGGING_LEVEL", "error")
	os.Setenv("JABBER_BOT_RECONNECTION_ENABLED", "false")
	defer func() {
		os.Unsetenv("JABBER_BOT_XMPP_JID")
		os.Unsetenv("JABBER_BOT_API_PORT")
		os.Unsetenv("JABBER_BOT_WEBHOOK_RETRY_ATTEMPTS")
		os.Unsetenv("JABBER_BOT_LOGGING_LEVEL")
		os.Unsetenv("JABBER_BOT_RECONNECTION_ENABLED")
	}()

	// Create config file
	configContent := `
xmpp:
  jid: "file-bot@localhost"
  password: "password123"
  server: "localhost:5222"

api:
  port: 8080

webhook:
  retry_attempts: 3

logging:
  level: "info"

reconnection:
  enabled: true
`

	tempDir := os.TempDir()
	tempFile := fmt.Sprintf("%s%senv-test-config.yaml", tempDir, string(os.PathSeparator))
	err := os.WriteFile(tempFile, []byte(configContent), 0644)
	require.NoError(suite.T(), err)
	defer os.Remove(tempFile)

	// Load configuration
	cfg, err := config.Load(tempFile)
	require.NoError(suite.T(), err)

	// Environment variables should override file values
	assert.Equal(suite.T(), "env-bot@localhost", cfg.XMPP.JID)
	assert.Equal(suite.T(), 9999, cfg.API.Port)
	assert.Equal(suite.T(), 10, cfg.Webhook.RetryAttempts)
	assert.Equal(suite.T(), "error", cfg.Logging.Level)
	assert.False(suite.T(), cfg.Reconnection.Enabled)

	// Non-overridden values should remain
	assert.Equal(suite.T(), "password123", cfg.XMPP.Password)
	assert.Equal(suite.T(), "localhost:5222", cfg.XMPP.Server)
	assert.Equal(suite.T(), "0.0.0.0", cfg.API.Host) // default value
}

// TestConfigValidation tests configuration validation in real scenarios
func (suite *ConfigIntegrationTestSuite) TestConfigValidation() {
	tests := []struct {
		name        string
		config      string
		expectError bool
		description string
	}{
		{
			name: "minimal_valid_config",
			config: `
xmpp:
  jid: "test@localhost"
  password: "test"
  server: "localhost:5222"
`,
			expectError: false,
			description: "Minimal valid configuration should load successfully",
		},
		{
			name: "missing_xmpp_server",
			config: `
xmpp:
  jid: "test@localhost"
  password: "test"
`,
			expectError: false, // Config doesn't validate required fields, just loads with defaults
			description: "Missing XMPP server should load without error (no validation currently)",
		},
		{
			name: "invalid_yaml",
			config: `
xmpp:
  jid: "test@localhost"
  password: "test"
  server: "localhost:5222"
  invalid: [unclosed array
`,
			expectError: true,
			description: "Invalid YAML should cause parsing error",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tempDir := os.TempDir()
			tempFile := fmt.Sprintf("%s%sconfig-test-%s.yaml", tempDir, string(os.PathSeparator), tt.name)
			err := os.WriteFile(tempFile, []byte(tt.config), 0644)
			require.NoError(suite.T(), err)
			defer os.Remove(tempFile)

			cfg, err := config.Load(tempFile)

			if tt.expectError {
				suite.T().Logf("Expected error for: %s", tt.description)
				assert.Error(suite.T(), err)
			} else {
				suite.T().Logf("Should succeed for: %s", tt.description)
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), cfg)
			}
		})
	}
}

// TestConfigProductionScenario tests production-like configuration
func (suite *ConfigIntegrationTestSuite) TestConfigProductionScenario() {
	configContent := `
xmpp:
  jid: "production-bot@company.com"
  password: "${JABBER_BOT_PASSWORD}"
  server: "xmpp.company.com:5222"
  resource: "production-service"

api:
  port: 80
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
  backoff: "30s"
`

	tempDir := os.TempDir()
	tempFile := fmt.Sprintf("%s%sproduction-config.yaml", tempDir, string(os.PathSeparator))
	err := os.WriteFile(tempFile, []byte(configContent), 0644)
	require.NoError(suite.T(), err)
	defer os.Remove(tempFile)

	// Set password environment variable
	os.Setenv("JABBER_BOT_PASSWORD", "super-secret-password")
	defer os.Unsetenv("JABBER_BOT_PASSWORD")

	cfg, err := config.Load(tempFile)
	require.NoError(suite.T(), err)

	// Verify production-like values
	assert.Equal(suite.T(), "production-bot@company.com", cfg.XMPP.JID)
	assert.Equal(suite.T(), "${JABBER_BOT_PASSWORD}", cfg.XMPP.Password) // Note: env substitution would need additional implementation
	assert.Equal(suite.T(), "xmpp.company.com:5222", cfg.XMPP.Server)
	assert.Equal(suite.T(), "production-service", cfg.XMPP.Resource)

	assert.Equal(suite.T(), 80, cfg.API.Port)
	assert.Equal(suite.T(), "0.0.0.0", cfg.API.Host)

	assert.Equal(suite.T(), "https://api.company.com/webhooks/jabber", cfg.Webhook.URL)
	assert.Equal(suite.T(), 30*time.Second, cfg.Webhook.Timeout)
	assert.Equal(suite.T(), 5, cfg.Webhook.RetryAttempts)

	assert.Equal(suite.T(), "info", cfg.Logging.Level)
	assert.Equal(suite.T(), "file", cfg.Logging.Output)
	assert.Equal(suite.T(), "/var/log/jabber-bot/production.log", cfg.Logging.FilePath)

	assert.True(suite.T(), cfg.Reconnection.Enabled)
	assert.Equal(suite.T(), 10, cfg.Reconnection.MaxAttempts)
	assert.Equal(suite.T(), 30*time.Second, cfg.Reconnection.Backoff)
}

// TestConfigDevelopmentScenario tests development configuration
func (suite *ConfigIntegrationTestSuite) TestConfigDevelopmentScenario() {
	configContent := `
xmpp:
  jid: "dev-bot@localhost"
  password: "dev-password"
  server: "localhost:5222"
  resource: "dev-bot"

api:
  port: 8080
  host: "127.0.0.1"

webhook:
  url: "http://localhost:3000/webhook"
  timeout: 5s
  retry_attempts: 1

logging:
  level: "debug"
  output: "stdout"

reconnection:
  enabled: false
`

	tempDir := os.TempDir()
	tempFile := fmt.Sprintf("%s%sdev-config.yaml", tempDir, string(os.PathSeparator))
	err := os.WriteFile(tempFile, []byte(configContent), 0644)
	require.NoError(suite.T(), err)
	defer os.Remove(tempFile)

	cfg, err := config.Load(tempFile)
	require.NoError(suite.T(), err)

	// Verify development-like values
	assert.Equal(suite.T(), "dev-bot@localhost", cfg.XMPP.JID)
	assert.Equal(suite.T(), "dev-password", cfg.XMPP.Password)
	assert.Equal(suite.T(), "localhost:5222", cfg.XMPP.Server)
	assert.Equal(suite.T(), "dev-bot", cfg.XMPP.Resource)

	assert.Equal(suite.T(), 8080, cfg.API.Port)
	assert.Equal(suite.T(), "127.0.0.1", cfg.API.Host)

	assert.Equal(suite.T(), "http://localhost:3000/webhook", cfg.Webhook.URL)
	assert.Equal(suite.T(), 5*time.Second, cfg.Webhook.Timeout)
	assert.Equal(suite.T(), 1, cfg.Webhook.RetryAttempts)

	assert.Equal(suite.T(), "debug", cfg.Logging.Level)
	assert.Equal(suite.T(), "stdout", cfg.Logging.Output)

	assert.False(suite.T(), cfg.Reconnection.Enabled)
}

// Run config integration tests
func TestConfigIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	if os.Getenv("INTEGRATION_TESTS") != "1" {
		t.Skip("Set INTEGRATION_TESTS=1 to run integration tests")
	}

	suite.Run(t, new(ConfigIntegrationTestSuite))
}
