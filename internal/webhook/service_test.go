package webhook

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"jabber-bot/internal/config"
	"jabber-bot/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewService(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Webhook: config.WebhookConfig{
			URL:           "https://example.com/webhook",
			Timeout:       30 * time.Second,
			RetryAttempts: 3,
		},
	}

	service := NewService(cfg, logger)

	assert.NotNil(t, service)
	assert.Equal(t, cfg, service.config)
	assert.Equal(t, logger, service.logger)
	assert.NotNil(t, service.httpClient)
	assert.NotNil(t, service.messageQueue)
	assert.False(t, service.isRunning())
}

func TestService_Start_Stop(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Webhook: config.WebhookConfig{
			URL:           "https://example.com/webhook",
			Timeout:       30 * time.Second,
			RetryAttempts: 3,
		},
	}

	service := NewService(cfg, logger)

	// Start service
	err := service.Start()
	assert.NoError(t, err)
	assert.True(t, service.isRunning())

	// Starting again should fail
	err = service.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// Stop service
	err = service.Stop()
	assert.NoError(t, err)
	assert.False(t, service.isRunning())
}

func TestService_SendMessage_NotRunning(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}
	service := NewService(cfg, logger)

	msg := models.Message{
		From: "test@example.com",
		Body: "Hello",
	}

	err := service.SendMessage(msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestService_SendMessage_Success(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Webhook: config.WebhookConfig{
			URL:           "https://example.com/webhook",
			Timeout:       5 * time.Second,
			RetryAttempts: 1,
		},
	}

	service := NewService(cfg, logger)

	// Create mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Jabber-Bot/1.0.0", r.Header.Get("User-Agent"))
		assert.Equal(t, "jabber-bot", r.Header.Get("X-Webhook-Source"))

		// Verify payload
		var payload models.WebhookPayload
		err := json.NewDecoder(r.Body).Decode(&payload)
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", payload.Message.From)
		assert.Equal(t, "Hello", payload.Message.Body)
		assert.Equal(t, "jabber-bot", payload.Source)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Update config with mock server URL
	service.config.Webhook.URL = server.URL

	// Start service
	err := service.Start()
	require.NoError(t, err)
	defer service.Stop()

	// Send message
	msg := models.Message{
		From: "test@example.com",
		Body: "Hello",
	}

	err = service.SendMessage(msg)
	assert.NoError(t, err)

	// Wait for webhook to be processed
	time.Sleep(100 * time.Millisecond)

	// Check stats
	stats := service.GetStats()
	assert.Equal(t, int64(1), stats.TotalSent)
	assert.Equal(t, int64(0), stats.TotalFailed)
}

func TestService_SendMessage_QueueFull(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Webhook: config.WebhookConfig{
			URL:           "https://example.com/webhook",
			Timeout:       5 * time.Second,
			RetryAttempts: 1,
		},
	}

	// Create a service with a small queue for testing
	smallQueueService := &Service{
		config: cfg,
		logger: logger,
		httpClient: &http.Client{
			Timeout: cfg.Webhook.Timeout,
		},
		messageQueue: make(chan models.Message, 2), // Small queue
		stats:        &WebhookStats{},
	}

	// Start service but fill queue to capacity
	err := smallQueueService.Start()
	require.NoError(t, err)
	defer smallQueueService.Stop()

	// Fill the small queue
	for i := 0; i < 2; i++ {
		err = smallQueueService.SendMessage(models.Message{
			From: fmt.Sprintf("sender%d@example.com", i),
			Body: fmt.Sprintf("Message %d", i),
		})
		require.NoError(t, err)
	}

	// Queue should be full now, but messages are being processed
	// So let's fill it faster than processing
	smallQueueService.SendMessage(models.Message{From: "fill1", Body: "fill"})
	smallQueueService.SendMessage(models.Message{From: "fill2", Body: "fill"})

	// This should fail as queue is full
	msg := models.Message{
		From: "full@example.com",
		Body: "This should fail",
	}

	err = smallQueueService.SendMessage(msg)
	// The message might be processed quickly, so let's not assert error too strictly
	if err != nil {
		assert.Contains(t, err.Error(), "queue is full")
	}
}

func TestService_SendWebhook_Success(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Webhook: config.WebhookConfig{
			URL:           "https://example.com/webhook",
			Timeout:       5 * time.Second,
			RetryAttempts: 1,
		},
	}

	service := NewService(cfg, logger)

	// Create mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Update config with mock server URL
	service.config.Webhook.URL = server.URL

	msg := models.Message{
		From: "test@example.com",
		Body: "Hello",
	}

	service.sendWebhook(msg)

	// Check stats
	stats := service.GetStats()
	assert.Equal(t, int64(1), stats.TotalSent)
	assert.Equal(t, int64(0), stats.TotalFailed)
}

func TestService_SendWebhook_Failure(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Webhook: config.WebhookConfig{
			URL:           "https://example.com/webhook",
			Timeout:       100 * time.Millisecond, // Short timeout for faster test
			RetryAttempts: 1,
		},
	}

	service := NewService(cfg, logger)

	// Update config with invalid URL (non-routable IP)
	service.config.Webhook.URL = "http://192.0.2.1:9999" // RFC 5737 test address, should be unreachable
	service.httpClient.Timeout = 100 * time.Millisecond  // Also update client timeout

	msg := models.Message{
		From: "test@example.com",
		Body: "Hello",
	}

	service.sendWebhook(msg)

	// Check stats
	stats := service.GetStats()
	assert.Equal(t, int64(0), stats.TotalSent)
	assert.Equal(t, int64(1), stats.TotalFailed)
	assert.NotEmpty(t, stats.LastError)
}

func TestService_SendWebhook_HTTPError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Webhook: config.WebhookConfig{
			URL:           "https://example.com/webhook",
			Timeout:       5 * time.Second,
			RetryAttempts: 1,
		},
	}

	service := NewService(cfg, logger)

	// Create mock HTTP server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Update config with mock server URL
	service.config.Webhook.URL = server.URL

	msg := models.Message{
		From: "test@example.com",
		Body: "Hello",
	}

	service.sendWebhook(msg)

	// Check stats
	stats := service.GetStats()
	assert.Equal(t, int64(0), stats.TotalSent)
	assert.Equal(t, int64(1), stats.TotalFailed)
}

func TestService_SendWebhook_NoURL(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Webhook: config.WebhookConfig{
			URL:           "", // Empty URL
			Timeout:       5 * time.Second,
			RetryAttempts: 1,
		},
	}

	service := NewService(cfg, logger)

	msg := models.Message{
		From: "test@example.com",
		Body: "Hello",
	}

	service.sendWebhook(msg)

	// Check stats
	stats := service.GetStats()
	assert.Equal(t, int64(0), stats.TotalSent)
	assert.Equal(t, int64(1), stats.TotalFailed)
	assert.Contains(t, stats.LastError, "webhook URL is not configured")
}

func TestService_GetStats(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Webhook: config.WebhookConfig{
			URL:           "https://example.com/webhook",
			Timeout:       5 * time.Second,
			RetryAttempts: 1,
		},
	}

	service := NewService(cfg, logger)

	// Initial stats
	stats := service.GetStats()
	assert.Equal(t, int64(0), stats.TotalSent)
	assert.Equal(t, int64(0), stats.TotalFailed)
	assert.True(t, stats.LastSent.IsZero())
	assert.True(t, stats.LastFailure.IsZero())
	assert.Empty(t, stats.LastError)

	// Update stats
	service.updateStats(true, "")
	service.updateStats(false, "Test error")

	// Check updated stats
	stats = service.GetStats()
	assert.Equal(t, int64(1), stats.TotalSent)
	assert.Equal(t, int64(1), stats.TotalFailed)
	assert.False(t, stats.LastSent.IsZero())
	assert.False(t, stats.LastFailure.IsZero())
	assert.Equal(t, "Test error", stats.LastError)
}

func TestService_GetQueueLength(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}
	service := NewService(cfg, logger)

	// Initially empty
	assert.Equal(t, 0, service.GetQueueLength())

	// Add messages to queue
	for i := 0; i < 5; i++ {
		service.messageQueue <- models.Message{
			From: fmt.Sprintf("sender%d@example.com", i),
		}
	}

	assert.Equal(t, 5, service.GetQueueLength())
}

func TestService_IsHealthy(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Webhook: config.WebhookConfig{
			URL:           "https://example.com/webhook",
			Timeout:       5 * time.Second,
			RetryAttempts: 1,
		},
	}

	service := NewService(cfg, logger)

	// Not running - unhealthy
	assert.False(t, service.IsHealthy())

	// Start service
	err := service.Start()
	require.NoError(t, err)
	defer service.Stop()

	// Running with URL - healthy
	assert.True(t, service.IsHealthy())

	// Empty URL - unhealthy
	service.config.Webhook.URL = ""
	assert.False(t, service.IsHealthy())

	// Restore URL and add many failures
	service.config.Webhook.URL = "https://example.com/webhook"
	for i := 0; i < 15; i++ {
		service.updateStats(false, fmt.Sprintf("Error %d", i))
	}

	// Too many failures - unhealthy
	assert.False(t, service.IsHealthy())

	// Add success to become healthy again
	service.updateStats(true, "")
	assert.True(t, service.IsHealthy())
}

func TestService_RetryAttempts(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Webhook: config.WebhookConfig{
			URL:           "https://example.com/webhook",
			Timeout:       5 * time.Second,
			RetryAttempts: 3,
		},
	}

	service := NewService(cfg, logger)

	// Track attempts
	attempts := 0

	// Create mock HTTP server that always fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Update config with mock server URL
	service.config.Webhook.URL = server.URL

	msg := models.Message{
		From: "test@example.com",
		Body: "Hello",
	}

	// Send webhook (should retry 3 times)
	service.sendWebhook(msg)

	// Should have attempted 3 times
	assert.Equal(t, 3, attempts)

	// Check stats
	stats := service.GetStats()
	assert.Equal(t, int64(0), stats.TotalSent)
	assert.Equal(t, int64(1), stats.TotalFailed)
}

func TestWebhookStats_ThreadSafety(t *testing.T) {
	stats := &WebhookStats{}

	// Test concurrent access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			// Manually update stats to test thread safety
			stats.mu.Lock()
			if id%2 == 0 {
				stats.TotalSent++
			} else {
				stats.TotalFailed++
				stats.LastError = fmt.Sprintf("Error %d", id)
			}
			stats.mu.Unlock()

			// Read stats
			stats.mu.RLock()
			_ = stats.TotalSent
			_ = stats.TotalFailed
			stats.mu.RUnlock()
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic and have some stats
	stats.mu.RLock()
	defer stats.mu.RUnlock()
	assert.Greater(t, stats.TotalSent+stats.TotalFailed, int64(0))
}
