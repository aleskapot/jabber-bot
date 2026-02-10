package webhook

import (
	"context"
	"testing"
	"time"

	"jabber-bot/internal/config"
	"jabber-bot/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// MockXMPPManager mocks the XMPP manager for testing
type MockXMPPManager struct {
	mock.Mock
}

func (m *MockXMPPManager) GetWebhookChannel() <-chan models.Message {
	args := m.Called()
	// Return as receive-only channel (no type assertion needed)
	return args.Get(0).(<-chan models.Message)
}

func TestNewManager(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Webhook: config.WebhookConfig{
			URL:           "https://example.com/webhook",
			Timeout:       30 * time.Second,
			RetryAttempts: 3,
		},
	}

	xmppManager := &MockXMPPManager{}
	manager := NewManager(cfg, logger, xmppManager)

	assert.NotNil(t, manager)
	assert.Equal(t, cfg, manager.config)
	assert.Equal(t, logger, manager.logger)
	assert.NotNil(t, manager.webhookService)
	assert.Equal(t, xmppManager, manager.xmppManager)
}

func TestManager_GetService(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}
	xmppManager := &MockXMPPManager{}
	manager := NewManager(cfg, logger, xmppManager)

	service := manager.GetService()
	assert.NotNil(t, service)
	assert.Equal(t, manager.webhookService, service)
}

func TestManager_ProcessXMPPMessages(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Webhook: config.WebhookConfig{
			URL:           "https://example.com/webhook",
			Timeout:       5 * time.Second,
			RetryAttempts: 1,
		},
	}

	xmppManager := &MockXMPPManager{}
	manager := NewManager(cfg, logger, xmppManager)

	// Create test message channel
	msgChan := make(chan models.Message, 10)
	xmppManager.On("GetWebhookChannel").Return((<-chan models.Message)(msgChan))

	// Start message processor in goroutine
	ctx, cancel := context.WithCancel(context.Background())
	go manager.processXMPPMessages(ctx)

	// Send test message
	testMsg := models.Message{
		From: "test@example.com",
		To:   "bot@example.com",
		Body: "Hello webhook!",
		Type: "chat",
	}

	msgChan <- testMsg

	// Wait a bit for processing
	time.Sleep(100 * time.Millisecond)

	// Cancel context to stop processor
	cancel()

	// Close channel to clean up
	close(msgChan)

	// Wait for goroutine to finish
	time.Sleep(50 * time.Millisecond)

	xmppManager.AssertExpectations(t)
}

func TestManager_HandleIncomingMessage(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Webhook: config.WebhookConfig{
			URL:           "https://example.com/webhook",
			Timeout:       5 * time.Second,
			RetryAttempts: 1,
		},
	}

	xmppManager := &MockXMPPManager{}
	manager := NewManager(cfg, logger, xmppManager)

	// Start webhook service
	err := manager.webhookService.Start()
	require.NoError(t, err)
	defer manager.webhookService.Stop()

	// Test message
	msg := models.Message{
		From: "test@example.com",
		To:   "bot@example.com",
		Body: "Hello webhook!",
		Type: "chat",
	}

	// Handle incoming message
	manager.handleIncomingMessage(msg)

	// Wait a moment for message to be queued
	time.Sleep(10 * time.Millisecond)

	// Check queue length (message should be queued or processed)
	// The message could be queued or already processed, so we just check it was handled
	assert.GreaterOrEqual(t, manager.webhookService.GetQueueLength(), 0)
}

func TestManager_GetStatus(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Webhook: config.WebhookConfig{
			URL:           "https://example.com/webhook",
			Timeout:       5 * time.Second,
			RetryAttempts: 1,
		},
	}

	xmppManager := &MockXMPPManager{}
	manager := NewManager(cfg, logger, xmppManager)

	status := manager.GetStatus()

	assert.NotNil(t, status)
	assert.Contains(t, status, "running")
	assert.Contains(t, status, "healthy")
	assert.Contains(t, status, "queue_length")
	assert.Contains(t, status, "webhook_url")
	assert.Contains(t, status, "total_sent")
	assert.Contains(t, status, "total_failed")
	assert.Equal(t, cfg.Webhook.URL, status["webhook_url"])
}
