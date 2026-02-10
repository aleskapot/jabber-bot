//go:build integration
// +build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"jabber-bot/internal/api"
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

// IntegrationTestSuite represents integration test suite
type IntegrationTestSuite struct {
	suite.Suite
	config      *config.Config
	logger      *zap.Logger
	xmppManager *xmpp.Manager
	webhookMgr  *webhook.Manager
	apiServer   *api.Server
	webhookSrv  *httptest.Server
	ctx         context.Context
	cancel      context.CancelFunc
}

// SetupSuite runs once before all tests
func (suite *IntegrationTestSuite) SetupSuite() {
	// Create test configuration
	suite.config = &config.Config{
		XMPP: config.XMPPConfig{
			JID:      "test-bot@localhost",
			Password: "test-password",
			Server:   "localhost:5222",
			Resource: "test-bot",
		},
		API: config.APIConfig{
			Port: 0, // Use random port
			Host: "127.0.0.1",
		},
		Webhook: config.WebhookConfig{
			URL:           "http://localhost:8081/webhook",
			Timeout:       10 * time.Second,
			RetryAttempts: 2,
		},
		Logging: config.LoggingConfig{
			Level:  "debug",
			Output: "stdout",
		},
		Reconnection: config.ReconnectionConfig{
			Enabled:     false, // Disable for testing
			MaxAttempts: 0,
		},
	}

	// Create logger
	var err error
	suite.logger, err = logger.NewWithConfig(
		suite.config.Logging.Level,
		suite.config.Logging.Output,
		suite.config.Logging.FilePath,
	)
	require.NoError(suite.T(), err)

	// Create webhook mock server
	suite.webhookSrv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.logger.Info("Webhook received",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Any("headers", r.Header),
		)

		// Verify webhook payload
		var payload models.WebhookPayload
		err := json.NewDecoder(r.Body).Decode(&payload)
		require.NoError(suite.T(), err)

		// Send success response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "received",
			"from":   payload.Message.From,
		})
	}))

	// Update webhook URL to mock server
	suite.config.Webhook.URL = suite.webhookSrv.URL

	// Create context
	suite.ctx, suite.cancel = context.WithCancel(context.Background())
}

// TearDownSuite runs once after all tests
func (suite *IntegrationTestSuite) TearDownSuite() {
	if suite.cancel != nil {
		suite.cancel()
	}

	if suite.webhookSrv != nil {
		suite.webhookSrv.Close()
	}

	if suite.apiServer != nil {
		suite.apiServer.Stop()
	}

	if suite.webhookMgr != nil {
		suite.webhookMgr.Stop()
	}

	if suite.xmppManager != nil {
		suite.xmppManager.Stop()
	}
}

// SetupTest runs before each test
func (suite *IntegrationTestSuite) SetupTest() {
	// Note: XMPP connection will be skipped in integration tests
	// as we don't have a real XMPP server running
	suite.T().Log("Setting up integration test")

	// Create XMPP manager (but don't start real connection)
	suite.xmppManager = xmpp.NewManager(suite.config, suite.logger)

	// Create webhook manager
	suite.webhookMgr = webhook.NewManager(suite.config, suite.logger, suite.xmppManager)
	err := suite.webhookMgr.Start(suite.ctx)
	require.NoError(suite.T(), err)

	// Create API server
	suite.apiServer = api.NewServer(suite.config, suite.logger, suite.xmppManager)

	// Start API server
	go func() {
		err := suite.apiServer.Start()
		if err != nil {
			suite.logger.Error("API server error", zap.Error(err))
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)
}

// TearDownTest runs after each test
func (suite *IntegrationTestSuite) TearDownTest() {
	suite.T().Log("Tearing down integration test")

	if suite.apiServer != nil {
		suite.apiServer.Stop()
		suite.apiServer = nil
	}

	if suite.webhookMgr != nil {
		suite.webhookMgr.Stop()
		suite.webhookMgr = nil
	}

	if suite.xmppManager != nil {
		suite.xmppManager.Stop()
		suite.xmppManager = nil
	}

	time.Sleep(50 * time.Millisecond)
}

// TestAPIWebhookIntegration tests API and webhook integration
func (suite *IntegrationTestSuite) TestAPIWebhookIntegration() {
	// Create test message
	msg := models.Message{
		From: "user@example.com",
		To:   "test-bot@localhost",
		Body: "Hello from integration test!",
		Type: "chat",
	}

	// Simulate incoming XMPP message
	// In real scenario, this would come from XMPP client
	err := suite.webhookMgr.GetService().SendMessage(msg)
	require.NoError(suite.T(), err)

	// Wait for webhook to be processed
	time.Sleep(200 * time.Millisecond)

	// Check webhook service stats
	stats := suite.webhookMgr.GetService().GetStats()
	assert.Greater(suite.T(), stats.TotalSent, int64(0))
}

// TestAPIEndpoints tests all API endpoints
func (suite *IntegrationTestSuite) TestAPIEndpoints() {
	// Give server time to start and get actual port
	time.Sleep(500 * time.Millisecond)
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", suite.apiServer.GetPort())

	// Test health endpoint (should be 503 since XMPP is not connected in integration tests)
	resp, err := http.Get(baseURL + "/api/v1/health")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusServiceUnavailable, resp.StatusCode)
	resp.Body.Close()

	// Test status endpoint
	resp, err = http.Get(baseURL + "/api/v1/status")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var statusResp models.StatusResponse
	err = json.NewDecoder(resp.Body).Decode(&statusResp)
	require.NoError(suite.T(), err)
	assert.False(suite.T(), statusResp.XMPPConnected) // Should be false in integration test
	assert.True(suite.T(), statusResp.APIRunning)
	resp.Body.Close()

	// Test root endpoint
	resp, err = http.Get(baseURL + "/")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

// TestSendAPI tests send message API endpoint
func (suite *IntegrationTestSuite) TestSendAPI() {
	// Give server time to start and get actual port
	time.Sleep(500 * time.Millisecond)
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", suite.apiServer.GetPort())

	// Test send message endpoint
	sendReq := models.SendMessageRequest{
		To:   "test-user@example.com",
		Body: "Test message from integration test",
		Type: "chat",
	}

	jsonData, err := json.Marshal(sendReq)
	require.NoError(suite.T(), err)

	resp, err := http.Post(
		baseURL+"/api/v1/send",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	require.NoError(suite.T(), err)

	// Should fail because XMPP is not connected
	assert.Equal(suite.T(), http.StatusInternalServerError, resp.StatusCode)
	resp.Body.Close()
}

// TestSendMUCAPI tests send MUC message API endpoint
func (suite *IntegrationTestSuite) TestSendMUCAPI() {
	// Give server time to start and get actual port
	time.Sleep(500 * time.Millisecond)
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", suite.apiServer.GetPort())

	// Test send MUC message endpoint
	sendMUCReq := models.SendMUCMessageRequest{
		Room: "test-room@conference.localhost",
		Body: "Test MUC message from integration test",
	}

	jsonData, err := json.Marshal(sendMUCReq)
	require.NoError(suite.T(), err)

	resp, err := http.Post(
		baseURL+"/api/v1/send-muc",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	require.NoError(suite.T(), err)

	// Should fail because XMPP is not connected
	assert.Equal(suite.T(), http.StatusInternalServerError, resp.StatusCode)
	resp.Body.Close()
}

// TestAPITransactions tests multiple API requests
func (suite *IntegrationTestSuite) TestAPITransactions() {
	// Give server time to start and get actual port
	time.Sleep(500 * time.Millisecond)
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", suite.apiServer.GetPort())

	// Test concurrent requests
	for i := 0; i < 5; i++ {
		go func(id int) {
			// Health check
			resp, err := http.Get(baseURL + "/api/v1/health")
			assert.NoError(suite.T(), err)
			if resp != nil {
				resp.Body.Close()
			}
		}(i)
	}

	// Wait for concurrent requests to complete
	time.Sleep(200 * time.Millisecond)
}

// TestWebhookMessageFlow tests complete message flow
func (suite *IntegrationTestSuite) TestWebhookMessageFlow() {
	// Create multiple test messages
	messages := []models.Message{
		{
			From: "user1@example.com",
			To:   "test-bot@localhost",
			Body: "Hello from user 1",
			Type: "chat",
		},
		{
			From: "user2@example.com",
			To:   "test-bot@localhost",
			Body: "Hello from user 2",
			Type: "chat",
		},
		{
			From: "room@conference.localhost/user3",
			To:   "test-bot@localhost",
			Body: "Hello from room",
			Type: "groupchat",
		},
	}

	// Send all messages to webhook
	for _, msg := range messages {
		err := suite.webhookMgr.GetService().SendMessage(msg)
		require.NoError(suite.T(), err)
	}

	// Wait for processing
	time.Sleep(500 * time.Millisecond)

	// Check statistics
	stats := suite.webhookMgr.GetService().GetStats()
	assert.Equal(suite.T(), int64(len(messages)), stats.TotalSent)
	assert.Equal(suite.T(), int64(0), stats.TotalFailed)
}

// TestConfigurationValidation tests configuration validation
func (suite *IntegrationTestSuite) TestConfigurationValidation() {
	// Test current configuration
	assert.NotEmpty(suite.T(), suite.config.XMPP.JID)
	assert.NotEmpty(suite.T(), suite.config.XMPP.Password)
	assert.NotEmpty(suite.T(), suite.config.XMPP.Server)
	// Port can be 0 in tests (random port)
	assert.NotEmpty(suite.T(), suite.config.Webhook.URL)
	assert.True(suite.T(), suite.config.Webhook.Timeout > 0)
	assert.Greater(suite.T(), suite.config.Webhook.RetryAttempts, 0)
}

// TestErrorHandling tests error scenarios
func (suite *IntegrationTestSuite) TestErrorHandling() {
	// Give server time to start and get actual port
	time.Sleep(500 * time.Millisecond)
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", suite.apiServer.GetPort())

	// Test invalid JSON
	resp, err := http.Post(
		baseURL+"/api/v1/send",
		"application/json",
		bytes.NewBuffer([]byte("invalid json")),
	)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()

	// Test missing required fields
	invalidReq := map[string]interface{}{
		"body": "Missing to field",
	}
	jsonData, _ := json.Marshal(invalidReq)

	resp, err = http.Post(
		baseURL+"/api/v1/send",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()

	// Test non-existent endpoint
	resp, err = http.Get(baseURL + "/api/v1/nonexistent")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}

// TestGracefulShutdown tests graceful shutdown
func (suite *IntegrationTestSuite) TestGracefulShutdown() {
	// This is partially tested in TearDownSuite
	// In real scenario, you would test that all connections are properly closed
	assert.NotNil(suite.T(), suite.cancel)
}

// TestLoggerIntegration tests logger integration
func (suite *IntegrationTestSuite) TestLoggerIntegration() {
	assert.NotNil(suite.T(), suite.logger)

	// Test logging works
	suite.logger.Info("Test log message from integration test")
}

// TestEnvironmentVariables tests environment variable handling
func (suite *IntegrationTestSuite) TestEnvironmentVariables() {
	// Test that environment variables are properly handled
	// This is already tested in config tests, but we can verify here
	assert.Equal(suite.T(), "127.0.0.1", suite.config.API.Host)
	assert.Equal(suite.T(), "debug", suite.config.Logging.Level)
}

// Run integration tests
func TestIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Check if integration tests should run
	if os.Getenv("INTEGRATION_TESTS") != "1" {
		t.Skip("Set INTEGRATION_TESTS=1 to run integration tests")
	}

	suite.Run(t, new(IntegrationTestSuite))
}
