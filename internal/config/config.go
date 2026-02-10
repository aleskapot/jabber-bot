package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	XMPP         XMPPConfig         `mapstructure:"xmpp"`
	API          APIConfig          `mapstructure:"api"`
	Webhook      WebhookConfig      `mapstructure:"webhook"`
	Logging      LoggingConfig      `mapstructure:"logging"`
	Reconnection ReconnectionConfig `mapstructure:"reconnection"`
}

type XMPPConfig struct {
	JID      string `mapstructure:"jid"`
	Password string `mapstructure:"password"`
	Server   string `mapstructure:"server"`
	Resource string `mapstructure:"resource"`
}

type APIConfig struct {
	Port    int    `mapstructure:"port"`
	Host    string `mapstructure:"host"`
	APIKey  string `mapstructure:"api_key"`
	Enabled bool   `mapstructure:"auth_enabled"`
}

type WebhookConfig struct {
	URL           string        `mapstructure:"url"`
	Timeout       time.Duration `mapstructure:"timeout"`
	RetryAttempts int           `mapstructure:"retry_attempts"`
}

type LoggingConfig struct {
	Level    string `mapstructure:"level"`
	Output   string `mapstructure:"output"`
	FilePath string `mapstructure:"file_path"`
}

type ReconnectionConfig struct {
	Enabled     bool          `mapstructure:"enabled"`
	MaxAttempts int           `mapstructure:"max_attempts"`
	Backoff     time.Duration `mapstructure:"backoff"`
}

func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// Set environment variable prefix
	viper.SetEnvPrefix("JABBER_BOT")
	viper.AutomaticEnv()

	// Set environment variable key replacer for nested structs
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Set default values
	if config.API.Port == 0 {
		config.API.Port = 8080
	}
	if config.API.Host == "" {
		config.API.Host = "0.0.0.0"
	}

	// If API key is empty, disable authentication
	if config.API.APIKey == "" {
		config.API.APIKey = ""
		config.API.Enabled = false
	}
	if config.Webhook.Timeout == 0 {
		config.Webhook.Timeout = 30 * time.Second
	}
	if config.Webhook.RetryAttempts == 0 {
		config.Webhook.RetryAttempts = 3
	}
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	if config.Logging.Output == "" {
		config.Logging.Output = "stdout"
	}
	if config.Reconnection.MaxAttempts == 0 {
		config.Reconnection.MaxAttempts = 5
	}
	if config.Reconnection.Backoff == 0 {
		config.Reconnection.Backoff = 5 * time.Second
	}

	return &config, nil
}
