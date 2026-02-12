package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"jabber-bot/internal/config"
	"jabber-bot/internal/models"

	"go.uber.org/zap"
)

// Service represents webhook service for sending notifications
type Service struct {
	config       *config.Config
	logger       *zap.Logger
	httpClient   *http.Client
	messageQueue chan models.Message
	mu           sync.RWMutex
	running      bool
	cancelFunc   context.CancelFunc
	stats        *Stats
	testMode     *TestModeUtils
}

// Stats contains webhook statistics
type Stats struct {
	TotalSent   int64     `json:"total_sent"`
	TotalFailed int64     `json:"total_failed"`
	LastSent    time.Time `json:"last_sent"`
	LastFailure time.Time `json:"last_failure"`
	LastError   string    `json:"last_error"`
	mu          sync.RWMutex
}

// NewService creates new webhook service
func NewService(cfg *config.Config, logger *zap.Logger) *Service {
	return &Service{
		config: cfg,
		logger: logger,
		httpClient: &http.Client{
			Timeout: cfg.Webhook.Timeout,
		},
		messageQueue: make(chan models.Message, 1000),
		stats:        &Stats{},
		testMode:     NewTestModeUtils(cfg.Webhook.TestModeSuffix),
	}
}

// Start starts webhook service
func (s *Service) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("webhook service is already running")
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFunc = cancel

	// Start webhook processor
	go s.processWebhooks(ctx)

	s.running = true
	s.logger.Info("Webhook service started",
		zap.String("url", s.config.Webhook.URL),
		zap.Duration("timeout", s.config.Webhook.Timeout),
		zap.Int("retry_attempts", s.config.Webhook.RetryAttempts),
	)

	return nil
}

// Stop stops webhook service
func (s *Service) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	if s.cancelFunc != nil {
		s.cancelFunc()
	}

	close(s.messageQueue)
	s.running = false

	s.logger.Info("Webhook service stopped")
	return nil
}

// SendMessage sends message to webhook endpoint
func (s *Service) SendMessage(msg models.Message) error {
	if !s.isRunning() {
		return fmt.Errorf("webhook service is not running")
	}

	select {
	case s.messageQueue <- msg:
		s.logger.Debug("Message queued for webhook",
			zap.String("from", msg.From),
			zap.String("to", msg.To),
		)
		return nil
	default:
		s.logger.Warn("Webhook queue full, dropping message",
			zap.String("from", msg.From),
			zap.Int("queue_length", len(s.messageQueue)),
		)
		return fmt.Errorf("webhook queue is full")
	}
}

// GetStats returns webhook statistics
func (s *Service) GetStats() Stats {
	s.stats.mu.RLock()
	defer s.stats.mu.RUnlock()

	return Stats{
		TotalSent:   s.stats.TotalSent,
		TotalFailed: s.stats.TotalFailed,
		LastSent:    s.stats.LastSent,
		LastFailure: s.stats.LastFailure,
		LastError:   s.stats.LastError,
	}
}

// isRunning checks if service is running (thread-safe)
func (s *Service) isRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// processWebhooks processes messages from queue and sends webhooks
func (s *Service) processWebhooks(ctx context.Context) {
	s.logger.Info("Starting webhook processor")
	defer s.logger.Info("Webhook processor stopped")

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-s.messageQueue:
			if !ok {
				return
			}
			s.sendWebhook(msg)
		}
	}
}

// sendWebhook sends webhook notification with retry logic
func (s *Service) sendWebhook(msg models.Message) {
	// Create webhook payload
	payload := models.WebhookPayload{
		Message:   msg,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Source:    "jabber-bot",
	}

	// Send with retries
	var lastErr error
	for attempt := 1; attempt <= s.config.Webhook.RetryAttempts; attempt++ {
		webhookURL := s.config.Webhook.URL
		// Check if test mode is detected and update URL
		if _, testURL, isTestMode := s.testMode.ProcessTestMessage(payload.Message.Body, s.config.Webhook.URL); isTestMode {
			webhookURL = testURL
		}

		err := s.sendWebhookAttempt(payload)
		if err == nil {
			// Success
			s.updateStats(true, "")
			s.logger.Info("Webhook sent successfully",
				zap.Int("attempt", attempt),
				zap.String("from", msg.From),
				zap.String("to", msg.To),
				zap.String("url", webhookURL),
			)
			return
		}

		lastErr = err
		s.logger.Warn("Webhook attempt failed",
			zap.Int("attempt", attempt),
			zap.Int("max_attempts", s.config.Webhook.RetryAttempts),
			zap.Error(err),
			zap.String("from", msg.From),
			zap.String("url", webhookURL),
		)

		// Don't wait after last attempt
		if attempt < s.config.Webhook.RetryAttempts {
			// Exponential backoff
			backoff := time.Duration(attempt*attempt) * time.Second
			time.Sleep(backoff)
		}
	}

	// All attempts failed
	var errorMsg string
	if lastErr != nil {
		errorMsg = lastErr.Error()
	} else {
		errorMsg = "unknown error"
	}
	// Determine final webhook URL for logging
	webhookURL := s.config.Webhook.URL
	if _, testURL, isTestMode := s.testMode.ProcessTestMessage(payload.Message.Body, s.config.Webhook.URL); isTestMode {
		webhookURL = testURL
	}

	s.updateStats(false, errorMsg)
	s.logger.Error("Webhook failed after all attempts",
		zap.Int("attempts", s.config.Webhook.RetryAttempts),
		zap.Error(lastErr),
		zap.String("from", msg.From),
		zap.String("to", msg.To),
		zap.String("url", webhookURL),
	)
}

// sendWebhookAttempt sends single webhook attempt
func (s *Service) sendWebhookAttempt(payload models.WebhookPayload) error {
	if s.config.Webhook.URL == "" {
		return fmt.Errorf("webhook URL is not configured")
	}

	// Process message for test mode
	processedBody, webhookURL, isTestMode := s.testMode.ProcessTestMessage(payload.Message.Body, s.config.Webhook.URL)

	// Update message body if test mode is detected
	if isTestMode {
		payload.Message.Body = processedBody
		s.logger.Debug("Test mode detected, using modified webhook URL",
			zap.String("original_url", s.config.Webhook.URL),
			zap.String("test_url", webhookURL),
			zap.String("original_body", payload.Message.Body),
		)
	}

	// Marshal payload
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Jabber-Bot/1.0.0")
	req.Header.Set("X-Webhook-Source", "jabber-bot")
	req.Header.Set("X-Webhook-Timestamp", payload.Timestamp)

	// Add test mode header for debugging
	if isTestMode {
		req.Header.Set("Webhook-Test-Mode", "true")
	}

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// updateStats updates webhook statistics
func (s *Service) updateStats(success bool, errorMsg string) {
	s.stats.mu.Lock()
	defer s.stats.mu.Unlock()

	if success {
		s.stats.TotalSent++
		s.stats.LastSent = time.Now().UTC()
	} else {
		s.stats.TotalFailed++
		s.stats.LastFailure = time.Now().UTC()
		s.stats.LastError = errorMsg
	}
}

// GetQueueLength returns current queue length
func (s *Service) GetQueueLength() int {
	return len(s.messageQueue)
}

// IsHealthy checks webhook service health
func (s *Service) IsHealthy() bool {
	if !s.isRunning() {
		return false
	}

	if s.config.Webhook.URL == "" {
		return false
	}

	stats := s.GetStats()

	// Check if we have too many recent failures
	if stats.TotalFailed > 10 && stats.LastSent.Before(stats.LastFailure) {
		// More than 10 failures and last operation was failure
		return false
	}

	return true
}
