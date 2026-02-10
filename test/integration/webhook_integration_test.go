//go:build integration
// +build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"jabber-bot/internal/config"
	"jabber-bot/internal/models"
	"jabber-bot/internal/webhook"
	"jabber-bot/internal/xmpp"
	"jabber-bot/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

// WebhookIntegrationTestSuite tests webhook service integration
type WebhookIntegrationTestSuite struct {
	suite.Suite
	logger     *zap.Logger
	config     *config.Config
	webhookSrv *WebhookTestServer
	webhookMgr *webhook.Manager
	ctx        context.Context
	cancel     context.CancelFunc
}

// WebhookTestServer is a mock webhook server for testing
type WebhookTestServer struct {
	mux          *http.ServeMux
	server       *http.Server
	url          string
	receivedMsgs []WebhookMessage
	mu           sync.Mutex
}

type WebhookMessage struct {
	Headers map[string][]string
	Body    []byte
	Method  string
	Path    string
}

// NewWebhookTestServer creates a new test webhook server
func NewWebhookTestServer() *WebhookTestServer {
	mux := http.NewServeMux()
	wts := &WebhookTestServer{
		mux:          mux,
		receivedMsgs: make([]WebhookMessage, 0),
	}

	mux.HandleFunc("/webhook", wts.handleWebhook)
	mux.HandleFunc("/health", wts.handleHealth)

	server := &http.Server{
		Addr:    "127.0.0.1:0", // Random port
		Handler: mux,
	}

	wts.server = server
	return wts
}

// Start starts the webhook test server
func (wts *WebhookTestServer) Start() error {
	wts.server.ListenAndServe()
	return nil
}

// Stop stops the webhook test server
func (wts *WebhookTestServer) Stop() error {
	return wts.server.Shutdown(context.Background())
}

// GetURL returns the server URL
func (wts *WebhookTestServer) GetURL() string {
	return wts.url
}

// GetReceivedMessages returns all received webhook messages
func (wts *WebhookTestServer) GetReceivedMessages() []WebhookMessage {
	wts.mu.Lock()
	defer wts.mu.Unlock()
	return append([]WebhookMessage(nil), wts.receivedMsgs...)
}

// ClearMessages clears all received messages
func (wts *WebhookTestServer) ClearMessages() {
	wts.mu.Lock()
	defer wts.mu.Unlock()
	wts.receivedMsgs = make([]WebhookMessage, 0)
}

// handleWebhook handles webhook requests
func (wts *WebhookTestServer) handleWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	msg := WebhookMessage{
		Headers: r.Header,
		Body:    body,
		Method:  r.Method,
		Path:    r.URL.Path,
	}

	wts.mu.Lock()
	wts.receivedMsgs = append(wts.receivedMsgs, msg)
	wts.mu.Unlock()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "received"}`))
}

// handleHealth handles health check requests
func (wts *WebhookTestServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy"}`))
}

// SetupSuite runs once before all tests
func (suite *WebhookIntegrationTestSuite) SetupSuite() {
	var err error
	suite.logger, err = logger.New()
	require.NoError(suite.T(), err)

	// Create webhook test server
	suite.webhookSrv = NewWebhookTestServer()

	// Start server on random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(suite.T(), err)

	port := listener.Addr().(*net.TCPAddr).Port
	suite.webhookSrv.url = fmt.Sprintf("http://127.0.0.1:%d", port)

	go suite.webhookSrv.server.Serve(listener)

	// Create test configuration
	suite.config = &config.Config{
		Webhook: config.WebhookConfig{
			URL:           suite.webhookSrv.GetURL() + "/webhook",
			Timeout:       5 * time.Second,
			RetryAttempts: 3,
		},
		Logging: config.LoggingConfig{
			Level:  "debug",
			Output: "stdout",
		},
	}

	// Create context
	suite.ctx, suite.cancel = context.WithCancel(context.Background())

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
}

// TearDownSuite runs once after all tests
func (suite *WebhookIntegrationTestSuite) TearDownSuite() {
	if suite.cancel != nil {
		suite.cancel()
	}

	if suite.webhookMgr != nil {
		suite.webhookMgr.Stop()
	}

	if suite.webhookSrv != nil {
		suite.webhookSrv.Stop()
	}
}

// SetupTest runs before each test
func (suite *WebhookIntegrationTestSuite) SetupTest() {
	// Clear previous messages
	suite.webhookSrv.ClearMessages()
}

// TestWebhookServiceStartStop tests webhook service lifecycle
func (suite *WebhookIntegrationTestSuite) TestWebhookServiceStartStop() {
	// Create webhook manager with minimal XMPP config
	xmppConfig := &config.Config{
		XMPP: config.XMPPConfig{
			JID:      "test@localhost",
			Password: "test",
			Server:   "localhost:5222",
		},
		Webhook: suite.config.Webhook,
		Logging: suite.config.Logging,
	}
	xmppManager := xmpp.NewManager(xmppConfig, suite.logger)
	suite.webhookMgr = webhook.NewManager(suite.config, suite.logger, xmppManager)

	// Start webhook manager
	err := suite.webhookMgr.Start(suite.ctx)
	require.NoError(suite.T(), err)

	// Check status
	status := suite.webhookMgr.GetStatus()
	assert.True(suite.T(), status["running"].(bool))

	// Stop webhook manager
	err = suite.webhookMgr.Stop()
	require.NoError(suite.T(), err)

	// Check status after stop
	status = suite.webhookMgr.GetStatus()
	assert.False(suite.T(), status["running"].(bool))
}

// TestWebhookMessageDelivery tests message delivery to webhook endpoint
func (suite *WebhookIntegrationTestSuite) TestWebhookMessageDelivery() {
	// Create webhook manager with minimal XMPP config
	xmppConfig := &config.Config{
		XMPP: config.XMPPConfig{
			JID:      "test@localhost",
			Password: "test",
			Server:   "localhost:5222",
		},
		Webhook: suite.config.Webhook,
		Logging: suite.config.Logging,
	}
	xmppManager := xmpp.NewManager(xmppConfig, suite.logger)
	suite.webhookMgr = webhook.NewManager(suite.config, suite.logger, xmppManager)

	// Start webhook manager
	err := suite.webhookMgr.Start(suite.ctx)
	require.NoError(suite.T(), err)
	defer suite.webhookMgr.Stop()

	// Send test message directly to webhook service
	msg := models.Message{
		From: "user@example.com",
		To:   "bot@localhost",
		Body: "Hello from webhook integration test!",
		Type: "chat",
	}

	err = suite.webhookMgr.GetService().SendMessage(msg)
	require.NoError(suite.T(), err)

	// Wait for webhook to be processed
	time.Sleep(200 * time.Millisecond)

	// Check that webhook server received the message
	receivedMsgs := suite.webhookSrv.GetReceivedMessages()
	assert.Len(suite.T(), receivedMsgs, 1)

	// Verify webhook request
	receivedMsg := receivedMsgs[0]
	assert.Equal(suite.T(), "POST", receivedMsg.Method)
	assert.Equal(suite.T(), "/webhook", receivedMsg.Path)
	assert.Equal(suite.T(), "application/json", receivedMsg.Headers["Content-Type"][0])

	// Verify webhook payload
	var payload models.WebhookPayload
	err = json.Unmarshal(receivedMsg.Body, &payload)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), msg.From, payload.Message.From)
	assert.Equal(suite.T(), msg.To, payload.Message.To)
	assert.Equal(suite.T(), msg.Body, payload.Message.Body)
	assert.Equal(suite.T(), msg.Type, payload.Message.Type)
	assert.Equal(suite.T(), "jabber-bot", payload.Source)
	assert.NotEmpty(suite.T(), payload.Timestamp)

	// Check webhook service statistics
	stats := suite.webhookMgr.GetService().GetStats()
	assert.Equal(suite.T(), int64(1), stats.TotalSent)
	assert.Equal(suite.T(), int64(0), stats.TotalFailed)
	assert.False(suite.T(), stats.LastSent.IsZero())
}

// TestWebhookRetryLogic tests webhook retry on failure
func (suite *WebhookIntegrationTestSuite) TestWebhookRetryLogic() {
	// Create config that points to non-existent webhook URL
	configWithInvalidURL := *suite.config
	configWithInvalidURL.Webhook.URL = "http://127.0.0.1:9999/webhook"
	configWithInvalidURL.Webhook.Timeout = 1 * time.Second
	configWithInvalidURL.Webhook.RetryAttempts = 2

	// Add XMPP config for manager creation
	configWithInvalidURL.XMPP = config.XMPPConfig{
		JID:      "test@localhost",
		Password: "test",
		Server:   "localhost:5222",
	}

	xmppManager := xmpp.NewManager(&configWithInvalidURL, suite.logger)
	suite.webhookMgr = webhook.NewManager(&configWithInvalidURL, suite.logger, xmppManager)

	// Start webhook manager
	err := suite.webhookMgr.Start(suite.ctx)
	require.NoError(suite.T(), err)
	defer suite.webhookMgr.Stop()

	// Send test message
	msg := models.Message{
		From: "user@example.com",
		Body: "This should fail and retry",
	}

	err = suite.webhookMgr.GetService().SendMessage(msg)
	require.NoError(suite.T(), err)

	// Wait for retries to complete
	time.Sleep(3 * time.Second)

	// Check webhook service statistics
	stats := suite.webhookMgr.GetService().GetStats()
	assert.Equal(suite.T(), int64(0), stats.TotalSent)
	assert.Equal(suite.T(), int64(1), stats.TotalFailed)
	assert.False(suite.T(), stats.LastFailure.IsZero())
	// Check for connection error (different on Windows vs Unix)
	assert.True(suite.T(),
		strings.Contains(stats.LastError, "connection refused") ||
			strings.Contains(stats.LastError, "No connection could be made") ||
			strings.Contains(stats.LastError, "actively refused"))
}

// TestWebhookConcurrency tests concurrent message delivery
func (suite *WebhookIntegrationTestSuite) TestWebhookConcurrency() {
	// Create webhook manager with minimal XMPP config
	xmppConfig := &config.Config{
		XMPP: config.XMPPConfig{
			JID:      "test@localhost",
			Password: "test",
			Server:   "localhost:5222",
		},
		Webhook: suite.config.Webhook,
		Logging: suite.config.Logging,
	}
	xmppManager := xmpp.NewManager(xmppConfig, suite.logger)
	suite.webhookMgr = webhook.NewManager(suite.config, suite.logger, xmppManager)

	// Start webhook manager
	err := suite.webhookMgr.Start(suite.ctx)
	require.NoError(suite.T(), err)
	defer suite.webhookMgr.Stop()

	// Send multiple messages concurrently
	numMessages := 10
	for i := 0; i < numMessages; i++ {
		go func(id int) {
			msg := models.Message{
				From: fmt.Sprintf("user%d@example.com", id),
				Body: fmt.Sprintf("Message %d", id),
			}

			err := suite.webhookMgr.GetService().SendMessage(msg)
			assert.NoError(suite.T(), err)
		}(i)
	}

	// Wait for all messages to be processed
	time.Sleep(1 * time.Second)

	// Check that webhook server received all messages
	receivedMsgs := suite.webhookSrv.GetReceivedMessages()
	assert.GreaterOrEqual(suite.T(), len(receivedMsgs), numMessages-2) // Allow some messages to be in queue

	// Check webhook service statistics
	stats := suite.webhookMgr.GetService().GetStats()
	assert.GreaterOrEqual(suite.T(), stats.TotalSent, int64(numMessages-2))
}

// TestWebhookQueueTests tests webhook queue behavior
func (suite *WebhookIntegrationTestSuite) TestWebhookQueueTests() {
	// Create webhook manager with minimal XMPP config
	xmppConfig := &config.Config{
		XMPP: config.XMPPConfig{
			JID:      "test@localhost",
			Password: "test",
			Server:   "localhost:5222",
		},
		Webhook: suite.config.Webhook,
		Logging: suite.config.Logging,
	}
	xmppManager := xmpp.NewManager(xmppConfig, suite.logger)
	suite.webhookMgr = webhook.NewManager(suite.config, suite.logger, xmppManager)

	// Start webhook manager
	err := suite.webhookMgr.Start(suite.ctx)
	require.NoError(suite.T(), err)
	defer suite.webhookMgr.Stop()

	// Fill queue quickly
	service := suite.webhookMgr.GetService()
	for i := 0; i < 500; i++ {
		msg := models.Message{
			From: fmt.Sprintf("user%d@example.com", i),
			Body: fmt.Sprintf("Message %d", i),
		}

		err := service.SendMessage(msg)
		if err != nil {
			suite.T().Logf("Message %d failed: %v", i, err)
			break
		}
	}

	// Wait for processing
	time.Sleep(2 * time.Second)

	// Check statistics
	stats := service.GetStats()
	suite.T().Logf("Total sent: %d, Total failed: %d, Queue length: %d",
		stats.TotalSent, stats.TotalFailed, service.GetQueueLength())

	assert.Greater(suite.T(), stats.TotalSent, int64(0))
}

// Run webhook integration tests
func TestWebhookIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	if os.Getenv("INTEGRATION_TESTS") != "1" {
		t.Skip("Set INTEGRATION_TESTS=1 to run integration tests")
	}

	suite.Run(t, new(WebhookIntegrationTestSuite))
}
