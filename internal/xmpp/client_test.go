package xmpp

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"jabber-bot/internal/config"
	"jabber-bot/internal/models"

	gosrc "gosrc.io/xmpp"

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

func newTestXMPPClient(t *testing.T) *gosrc.Client {
	t.Helper()

	client, err := gosrc.NewClient(
		&gosrc.Config{
			Jid:        "test@example.com",
			Credential: gosrc.Password("password"),
		},
		gosrc.NewRouter(),
		func(error) {},
	)
	require.NoError(t, err)
	return client
}

func TestClient_IsConnected(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}
	client := NewClient(cfg, logger)

	// Initially not connected
	assert.False(t, client.IsConnected())

	// Simulate connection
	xmppClient := newTestXMPPClient(t)
	client.mu.Lock()
	client.client = xmppClient
	client.mu.Unlock()
	client.setConnected(true)
	atomic.StoreInt32(&client.libraryConnected, 1)
	assert.True(t, client.IsConnected())

	// Simulate disconnection
	client.setConnected(false)
	assert.False(t, client.IsConnected())
}

func TestClient_IsConnected_NoClientPointer(t *testing.T) {
	logger := zaptest.NewLogger(t)
	client := NewClient(&config.Config{}, logger)

	// Flags alone must not report a usable connection without a client.
	client.setConnected(true)
	atomic.StoreInt32(&client.libraryConnected, 1)

	assert.False(t, client.IsConnected())
}

func TestClient_HandleClientError_IgnoresStaleClient(t *testing.T) {
	logger := zaptest.NewLogger(t)
	client := NewClient(&config.Config{}, logger)
	currentClient := newTestXMPPClient(t)
	staleClient := newTestXMPPClient(t)

	client.mu.Lock()
	client.client = currentClient
	client.mu.Unlock()
	client.setConnected(true)
	atomic.StoreInt32(&client.libraryConnected, 1)

	client.handleClientError(staleClient, fmt.Errorf("write: broken pipe"))

	assert.True(t, client.IsConnected())
	select {
	case <-client.reconnectCh:
		t.Fatal("stale client error must not trigger reconnect")
	default:
	}
}

func TestClient_HandleClientError_MarksDisconnected(t *testing.T) {
	logger := zaptest.NewLogger(t)
	client := NewClient(&config.Config{}, logger)
	xmppClient := newTestXMPPClient(t)

	client.mu.Lock()
	client.client = xmppClient
	client.mu.Unlock()
	client.setConnected(true)
	atomic.StoreInt32(&client.libraryConnected, 1)

	client.handleClientError(xmppClient, fmt.Errorf("write: broken pipe"))

	assert.False(t, client.IsConnected())
	select {
	case <-client.reconnectCh:
	default:
		t.Fatal("expected reconnect signal")
	}
}

func TestClient_SendMessage_ConnectionLost(t *testing.T) {
	logger := zaptest.NewLogger(t)
	client := NewClient(&config.Config{}, logger)
	xmppClient := newTestXMPPClient(t)

	client.mu.Lock()
	client.client = xmppClient
	client.mu.Unlock()
	client.setConnected(true)
	atomic.StoreInt32(&client.libraryConnected, 1)

	err := client.SendMessage("test@example.com", "Hello", "chat")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrConnectionLost)
	assert.False(t, client.IsConnected())

	select {
	case <-client.reconnectCh:
	default:
		t.Fatal("expected reconnect signal")
	}
}

func TestIsConnectionError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "broken pipe", err: fmt.Errorf("write: broken pipe"), want: true},
		{name: "connection reset", err: fmt.Errorf("read: connection reset by peer"), want: true},
		{name: "closed", err: fmt.Errorf("use of closed network connection"), want: true},
		{name: "not connected", err: fmt.Errorf("client is not connected"), want: true},
		{name: "validation", err: fmt.Errorf("invalid JID format"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isConnectionError(tt.err))
		})
	}
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

func TestClient_ReconnectionDisabled(t *testing.T) {
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

	// handleReconnection should return immediately when disabled
	done := make(chan struct{})
	go func() {
		client.handleReconnection(context.Background())
		close(done)
	}()

	select {
	case <-done:
		// Expected - returned immediately
	case <-time.After(100 * time.Millisecond):
		t.Fatal("handleReconnection did not return immediately when reconnection is disabled")
	}
}

func TestClient_ReconnectionEnabled(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		XMPP: config.XMPPConfig{
			JID:      "test@example.com",
			Password: "password",
			Server:   "localhost:5222",
			Resource: "test",
		},
		Reconnection: config.ReconnectionConfig{
			Enabled:     true,
			MaxAttempts: 1,
			Backoff:     10 * time.Millisecond,
		},
	}

	client := NewClient(cfg, logger)

	// Verify reconnection config is applied
	assert.True(t, client.config.Reconnection.Enabled)
	assert.Equal(t, 1, client.config.Reconnection.MaxAttempts)
	assert.Equal(t, 10*time.Millisecond, client.config.Reconnection.Backoff)
}

func TestClient_DefaultConfig_HasReconnectionEnabled(t *testing.T) {
	cfg := &config.Config{}

	// Zero-value config has reconnection disabled
	assert.False(t, cfg.Reconnection.Enabled, "zero-value config has reconnection disabled")
}

func TestClient_StreamLoggerLifecycle(t *testing.T) {
	client := NewClient(&config.Config{}, zaptest.NewLogger(t))

	logFile := client.startStreamLogger()
	require.NotNil(t, logFile)
	name := logFile.Name()

	_, err := os.Stat(name)
	require.NoError(t, err)

	client.stopStreamLogger()

	_, err = os.Stat(name)
	assert.True(t, os.IsNotExist(err), "stream log file should be removed")
}

func TestClient_ReconnectBackoff_Bounded(t *testing.T) {
	cfg := &config.Config{
		Reconnection: config.ReconnectionConfig{
			Backoff:    100 * time.Millisecond,
			MaxBackoff: time.Second,
		},
	}
	client := NewClient(cfg, zaptest.NewLogger(t))

	for attempt := 1; attempt <= 20; attempt++ {
		delay := client.reconnectBackoff(attempt)
		assert.Greater(t, delay, time.Duration(0))
		assert.LessOrEqual(t, delay, time.Second+time.Second/5,
			"backoff must stay within max plus jitter")
	}
}
