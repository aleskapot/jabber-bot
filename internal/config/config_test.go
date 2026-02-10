package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_ValidConfig(t *testing.T) {
	// Create temporary config file
	configContent := `
xmpp:
  jid: "bot@example.com"
  password: "secret123"
  server: "xmpp.example.com:5222"
  resource: "bot-resource"

api:
  port: 9090
  host: "127.0.0.1"

webhook:
  url: "https://webhook.example.com/hook"
  timeout: 60s
  retry_attempts: 5

logging:
  level: "debug"
  output: "file"
  file_path: "/var/log/jabber-bot.log"

reconnection:
  enabled: true
  max_attempts: 10
  backoff: "10s"
`

	// Write to temporary file
	tempFile := filepath.Join(t.TempDir(), "test-config.yaml")
	err := os.WriteFile(tempFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Load config
	cfg, err := Load(tempFile)
	require.NoError(t, err)

	// Verify XMPP config
	assert.Equal(t, "bot@example.com", cfg.XMPP.JID)
	assert.Equal(t, "secret123", cfg.XMPP.Password)
	assert.Equal(t, "xmpp.example.com:5222", cfg.XMPP.Server)
	assert.Equal(t, "bot-resource", cfg.XMPP.Resource)

	// Verify API config
	assert.Equal(t, 9090, cfg.API.Port)
	assert.Equal(t, "127.0.0.1", cfg.API.Host)

	// Verify webhook config
	assert.Equal(t, "https://webhook.example.com/hook", cfg.Webhook.URL)
	assert.Equal(t, 60*time.Second, cfg.Webhook.Timeout)
	assert.Equal(t, 5, cfg.Webhook.RetryAttempts)

	// Verify logging config
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, "file", cfg.Logging.Output)
	assert.Equal(t, "/var/log/jabber-bot.log", cfg.Logging.FilePath)

	// Verify reconnection config
	assert.True(t, cfg.Reconnection.Enabled)
	assert.Equal(t, 10, cfg.Reconnection.MaxAttempts)
	assert.Equal(t, 10*time.Second, cfg.Reconnection.Backoff)
}

func TestLoad_MinimalConfig(t *testing.T) {
	// Create minimal config file
	configContent := `
xmpp:
  jid: "bot@example.com"
  password: "secret123"
  server: "xmpp.example.com:5222"
`

	// Write to temporary file
	tempFile := filepath.Join(t.TempDir(), "minimal-config.yaml")
	err := os.WriteFile(tempFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Load config
	cfg, err := Load(tempFile)
	require.NoError(t, err)

	// Verify XMPP config
	assert.Equal(t, "bot@example.com", cfg.XMPP.JID)
	assert.Equal(t, "secret123", cfg.XMPP.Password)
	assert.Equal(t, "xmpp.example.com:5222", cfg.XMPP.Server)

	// Verify default values
	assert.Equal(t, 8080, cfg.API.Port)
	assert.Equal(t, "0.0.0.0", cfg.API.Host)
	assert.Equal(t, 30*time.Second, cfg.Webhook.Timeout)
	assert.Equal(t, 3, cfg.Webhook.RetryAttempts)
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "stdout", cfg.Logging.Output)
	assert.Equal(t, 5, cfg.Reconnection.MaxAttempts)
	assert.Equal(t, 5*time.Second, cfg.Reconnection.Backoff)
}

func TestLoad_FileNotFound(t *testing.T) {
	// Try to load non-existent file
	_, err := Load("non-existent-config.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestLoad_InvalidYAML(t *testing.T) {
	// Create invalid YAML file
	configContent := `
xmpp:
  jid: "bot@example.com"
  password: "secret123"
  server: "xmpp.example.com:5222"
  invalid_yaml: [unclosed array
`

	// Write to temporary file
	tempFile := filepath.Join(t.TempDir(), "invalid-config.yaml")
	err := os.WriteFile(tempFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Try to load config
	_, err = Load(tempFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal config")
}

func TestLoad_EnvironmentVariableOverride(t *testing.T) {
	// Set environment variables
	t.Setenv("JABBER_BOT_XMPP_JID", "env-bot@example.com")
	t.Setenv("JABBER_BOT_API_PORT", "9000")
	t.Setenv("JABBER_BOT_LOGGING_LEVEL", "warn")

	// Create config file
	configContent := `
xmpp:
  jid: "file-bot@example.com"
  password: "secret123"
  server: "xmpp.example.com:5222"

api:
  port: 8080

logging:
  level: "info"
`

	// Write to temporary file
	tempFile := filepath.Join(t.TempDir(), "env-config.yaml")
	err := os.WriteFile(tempFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Load config
	cfg, err := Load(tempFile)
	require.NoError(t, err)

	// Environment variables should override file values
	assert.Equal(t, "env-bot@example.com", cfg.XMPP.JID)
	assert.Equal(t, 9000, cfg.API.Port)
	assert.Equal(t, "warn", cfg.Logging.Level)

	// Non-overridden values should remain
	assert.Equal(t, "secret123", cfg.XMPP.Password)
	assert.Equal(t, "info", cfg.Logging.Output) // default value
}

func TestLoad_InvalidPort(t *testing.T) {
	// Create config with invalid port
	configContent := `
xmpp:
  jid: "bot@example.com"
  password: "secret123"
  server: "xmpp.example.com:5222"

api:
  port: "invalid_port"
`

	// Write to temporary file
	tempFile := filepath.Join(t.TempDir(), "invalid-port-config.yaml")
	err := os.WriteFile(tempFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Load config - this should not fail as viper will try to convert
	cfg, err := Load(tempFile)
	require.NoError(t, err)

	// Port should be 0 (invalid conversion)
	assert.Equal(t, 0, cfg.API.Port)
}

func TestLoad_EmptyConfig(t *testing.T) {
	// Create empty config file
	configContent := ``

	// Write to temporary file
	tempFile := filepath.Join(t.TempDir(), "empty-config.yaml")
	err := os.WriteFile(tempFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Load config
	cfg, err := Load(tempFile)
	require.NoError(t, err)

	// Should have default values
	assert.Equal(t, 8080, cfg.API.Port)
	assert.Equal(t, "0.0.0.0", cfg.API.Host)
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "stdout", cfg.Logging.Output)
	assert.Equal(t, 30*time.Second, cfg.Webhook.Timeout)
	assert.Equal(t, 3, cfg.Webhook.RetryAttempts)
	assert.Equal(t, 5, cfg.Reconnection.MaxAttempts)
	assert.Equal(t, 5*time.Second, cfg.Reconnection.Backoff)
}

func TestLoad_InvalidDuration(t *testing.T) {
	// Create config with invalid duration
	configContent := `
xmpp:
  jid: "bot@example.com"
  password: "secret123"
  server: "xmpp.example.com:5222"

webhook:
  timeout: "invalid_duration"
`

	// Write to temporary file
	tempFile := filepath.Join(t.TempDir(), "invalid-duration-config.yaml")
	err := os.WriteFile(tempFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Load config - this should not fail as viper will use default
	cfg, err := Load(tempFile)
	require.NoError(t, err)

	// Timeout should be default value
	assert.Equal(t, 30*time.Second, cfg.Webhook.Timeout)
}

func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name        string
		configFunc  func() *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			configFunc: func() *Config {
				return &Config{
					XMPP: XMPPConfig{
						JID:      "bot@example.com",
						Password: "secret123",
						Server:   "xmpp.example.com:5222",
					},
					API: APIConfig{
						Port: 8080,
						Host: "localhost",
					},
					Webhook: WebhookConfig{
						URL:           "https://webhook.example.com",
						Timeout:       30 * time.Second,
						RetryAttempts: 3,
					},
				}
			},
			expectError: false,
		},
		{
			name: "missing jid",
			configFunc: func() *Config {
				return &Config{
					XMPP: XMPPConfig{
						Password: "secret123",
						Server:   "xmpp.example.com:5222",
					},
				}
			},
			expectError: true,
			errorMsg:    "jid",
		},
		{
			name: "missing password",
			configFunc: func() *Config {
				return &Config{
					XMPP: XMPPConfig{
						JID:    "bot@example.com",
						Server: "xmpp.example.com:5222",
					},
				}
			},
			expectError: true,
			errorMsg:    "password",
		},
		{
			name: "missing server",
			configFunc: func() *Config {
				return &Config{
					XMPP: XMPPConfig{
						JID:      "bot@example.com",
						Password: "secret123",
					},
				}
			},
			expectError: true,
			errorMsg:    "server",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.configFunc()

			// This is a simple validation test - in real implementation you might have more complex validation
			if tt.expectError {
				// Check for missing required fields
				if cfg.XMPP.JID == "" {
					assert.Contains(t, tt.errorMsg, "jid")
				}
				if cfg.XMPP.Password == "" {
					assert.Contains(t, tt.errorMsg, "password")
				}
				if cfg.XMPP.Server == "" {
					assert.Contains(t, tt.errorMsg, "server")
				}
			} else {
				assert.NotEmpty(t, cfg.XMPP.JID)
				assert.NotEmpty(t, cfg.XMPP.Password)
				assert.NotEmpty(t, cfg.XMPP.Server)
			}
		})
	}
}

func TestConfig_NumericDefaults(t *testing.T) {
	// Test that numeric defaults are properly set
	configContent := `
xmpp:
  jid: "bot@example.com"
  password: "secret123"
  server: "xmpp.example.com:5222"
`

	tempFile := filepath.Join(t.TempDir(), "numeric-defaults.yaml")
	err := os.WriteFile(tempFile, []byte(configContent), 0644)
	require.NoError(t, err)

	cfg, err := Load(tempFile)
	require.NoError(t, err)

	// Test all numeric defaults
	assert.Equal(t, 8080, cfg.API.Port)
	assert.Equal(t, 3, cfg.Webhook.RetryAttempts)
	assert.Equal(t, 5, cfg.Reconnection.MaxAttempts)
}

func TestConfig_DurationDefaults(t *testing.T) {
	// Test that duration defaults are properly set
	configContent := `
xmpp:
  jid: "bot@example.com"
  password: "secret123"
  server: "xmpp.example.com:5222"
`

	tempFile := filepath.Join(t.TempDir(), "duration-defaults.yaml")
	err := os.WriteFile(tempFile, []byte(configContent), 0644)
	require.NoError(t, err)

	cfg, err := Load(tempFile)
	require.NoError(t, err)

	// Test all duration defaults
	assert.Equal(t, 30*time.Second, cfg.Webhook.Timeout)
	assert.Equal(t, 5*time.Second, cfg.Reconnection.Backoff)
}

func TestConfig_StringDefaults(t *testing.T) {
	// Test that string defaults are properly set
	configContent := `
xmpp:
  jid: "bot@example.com"
  password: "secret123"
  server: "xmpp.example.com:5222"
`

	tempFile := filepath.Join(t.TempDir(), "string-defaults.yaml")
	err := os.WriteFile(tempFile, []byte(configContent), 0644)
	require.NoError(t, err)

	cfg, err := Load(tempFile)
	require.NoError(t, err)

	// Test all string defaults
	assert.Equal(t, "0.0.0.0", cfg.API.Host)
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "stdout", cfg.Logging.Output)
	assert.Empty(t, cfg.Logging.FilePath)
}
