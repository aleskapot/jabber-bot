package xmpp

import (
	"context"
	"fmt"
	"testing"
	"time"

	"jabber-bot/internal/config"
	"jabber-bot/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// MockXMPPClient mocks the XMPP client for testing
type MockXMPPClient struct {
	mock.Mock
}

func (m *MockXMPPClient) Connect() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockXMPPClient) Disconnect() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockXMPPClient) Send(msg interface{}) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *MockXMPPClient) Recv() (interface{}, error) {
	args := m.Called()
	return args.Get(0), args.Error(1)
}

func TestClient_NewClient(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		XMPP: config.XMPPConfig{
			JID:      "test@example.com",
			Password: "password",
			Server:   "example.com:5222",
			Resource: "test",
		},
	}

	client := NewClient(cfg, logger)

	assert.NotNil(t, client)
	assert.Equal(t, cfg, client.config)
	assert.Equal(t, logger, client.logger)
	assert.False(t, client.IsConnected())
}

func TestClient_SendMessage_NotConnected(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}
	client := NewClient(cfg, logger)

	err := client.SendMessage("test@example.com", "Hello", "chat")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestClient_SendMUCMessage_NotConnected(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}
	client := NewClient(cfg, logger)

	err := client.SendMUCMessage("room@conference.example.com", "Hello room", "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestClient_IsConnected(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}
	client := NewClient(cfg, logger)

	// Initially not connected
	assert.False(t, client.IsConnected())

	// Simulate connection
	client.setConnected(true)
	assert.True(t, client.IsConnected())

	// Simulate disconnection
	client.setConnected(false)
	assert.False(t, client.IsConnected())
}

func TestClient_GetMessageChannel(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}
	client := NewClient(cfg, logger)

	ch := client.GetMessageChannel()
	assert.NotNil(t, ch)

	// Channel should be initially empty
	select {
	case msg := <-ch:
		t.Fatalf("Expected empty channel, got message: %v", msg)
	default:
		// Expected
	}
}

func TestClient_SetConnected_ThreadSafety(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}
	client := NewClient(cfg, logger)

	// Test concurrent access
	done := make(chan bool, 10)

	// Start multiple goroutines that change connection status
	for i := 0; i < 10; i++ {
		go func() {
			client.setConnected(true)
			client.setConnected(false)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic
	assert.False(t, client.IsConnected())
}

func TestClient_MessageChannel_Capacity(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}
	client := NewClient(cfg, logger)

	// Send messages to channel up to capacity
	for i := 0; i < 100; i++ {
		select {
		case client.messageChan <- models.Message{
			From: fmt.Sprintf("sender%d@example.com", i),
			Body: fmt.Sprintf("Message %d", i),
		}:
			// Message sent successfully
		default:
			// Channel should be able to handle 100 messages
			t.Fatalf("Channel should handle at least 100 messages, failed at message %d", i)
		}
	}

	// Channel should be close to capacity
	assert.Equal(t, 100, len(client.messageChan))

	// Clear channel
	for i := 0; i < 100; i++ {
		<-client.messageChan
	}
}

func TestManager_NewManager(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Reconnection: config.ReconnectionConfig{
			Enabled:     true,
			MaxAttempts: 5,
			Backoff:     5 * time.Second,
		},
	}

	manager := NewManager(cfg, logger)

	assert.NotNil(t, manager)
	assert.Equal(t, cfg, manager.config)
	assert.Equal(t, logger, manager.logger)
	assert.Empty(t, manager.clients)
	assert.NotNil(t, manager.webhookChan)
}

func TestManager_GetDefaultClient_NoClients(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}
	manager := NewManager(cfg, logger)

	client := manager.GetDefaultClient()
	assert.Nil(t, client)
}

func TestManager_SendMessage_NoDefaultClient(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}
	manager := NewManager(cfg, logger)

	err := manager.SendMessage("test@example.com", "Hello", "chat")
	assert.Error(t, err)
	assert.Equal(t, ErrNoDefaultClient, err)
}

func TestManager_SendMUCMessage_NoDefaultClient(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}
	manager := NewManager(cfg, logger)

	err := manager.SendMUCMessage("room@conference.example.com", "Hello room", "")
	assert.Error(t, err)
	assert.Equal(t, ErrNoDefaultClient, err)
}

func TestManager_IsConnected_NoClients(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}
	manager := NewManager(cfg, logger)

	connected := manager.IsConnected()
	assert.False(t, connected)
}

func TestXMPPError(t *testing.T) {
	err := &XMPPError{
		Code:    "TEST_ERROR",
		Message: "Test error message",
	}

	assert.Equal(t, "Test error message", err.Error())
	assert.Equal(t, "TEST_ERROR", err.Code)
}

func TestXMPPLoggerAdapter_Printf(t *testing.T) {
	logger := zaptest.NewLogger(t)
	adapter := &xmppLoggerAdapter{logger: logger}

	// Should not panic
	adapter.Printf("Test message: %s", "test")
}

func TestXMPPLoggerAdapter_Println(t *testing.T) {
	logger := zaptest.NewLogger(t)
	adapter := &xmppLoggerAdapter{logger: logger}

	// Should not panic
	adapter.Println("Test message", "test")
}

// Integration test placeholder - would require actual XMPP server
func TestClient_Connect_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test would require a running XMPP server
	// For now, just verify the method exists
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		XMPP: config.XMPPConfig{
			JID:      "test@example.com",
			Password: "password",
			Server:   "localhost:5222",
			Resource: "test",
		},
		Reconnection: config.ReconnectionConfig{
			Enabled: false,
		},
	}

	client := NewClient(cfg, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Connect(ctx)

	// Expect error since no XMPP server is running
	assert.Error(t, err)

	// Verify error type
	require.Error(t, err)
}
