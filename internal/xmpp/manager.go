package xmpp

import (
	"context"
	"sync"
	"time"

	"jabber-bot/internal/config"
	"jabber-bot/internal/models"

	"go.uber.org/zap"
)

// Manager manages XMPP connections and message handling
type Manager struct {
	config      *config.Config
	logger      *zap.Logger
	clients     map[string]*Client
	mu          sync.RWMutex
	webhookChan chan models.Message
}

// NewManager creates new XMPP manager
func NewManager(cfg *config.Config, logger *zap.Logger) *Manager {
	return &Manager{
		config:      cfg,
		logger:      logger,
		clients:     make(map[string]*Client),
		webhookChan: make(chan models.Message, 1000),
	}
}

// Start starts XMPP manager
func (m *Manager) Start() error {
	m.logger.Info("Starting XMPP manager")

	// Create default client
	client := NewClient(m.config, m.logger)

	// Connect using background context
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		return err
	}

	m.mu.Lock()
	m.clients["default"] = client
	m.mu.Unlock()

	// Start webhook dispatcher
	go m.dispatchWebhooks()

	m.logger.Info("XMPP manager started successfully")
	return nil
}

// Stop stops XMPP manager and all connections
func (m *Manager) Stop() error {
	m.logger.Info("Stopping XMPP manager")

	m.mu.Lock()
	defer m.mu.Unlock()

	// Disconnect all clients
	for id, client := range m.clients {
		if err := client.Disconnect(); err != nil {
			m.logger.Error("Error disconnecting client",
				zap.String("client_id", id),
				zap.Error(err),
			)
		}
	}

	close(m.webhookChan)
	m.clients = make(map[string]*Client)

	m.logger.Info("XMPP manager stopped")
	return nil
}

// GetDefaultClient returns the default XMPP client
func (m *Manager) GetDefaultClient() *Client {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if client, exists := m.clients["default"]; exists {
		return client
	}
	return nil
}

// SendMessage sends message using default client
func (m *Manager) SendMessage(to, body, messageType string) error {
	client := m.GetDefaultClient()
	if client == nil {
		return ErrNoDefaultClient
	}

	return client.SendMessage(to, body, messageType)
}

// SendMUCMessage sends MUC message using default client
func (m *Manager) SendMUCMessage(room, body, subject string) error {
	client := m.GetDefaultClient()
	if client == nil {
		return ErrNoDefaultClient
	}

	return client.SendMUCMessage(room, body, subject)
}

// IsConnected checks if default client is connected
func (m *Manager) IsConnected() bool {
	client := m.GetDefaultClient()
	return client != nil && client.IsConnected()
}

// GetWebhookChannel returns channel for webhook messages
func (m *Manager) GetWebhookChannel() <-chan models.Message {
	return m.webhookChan
}

// dispatchWebhooks processes messages for webhook delivery
func (m *Manager) dispatchWebhooks() {
	m.logger.Info("Starting webhook dispatcher")
	defer m.logger.Info("Webhook dispatcher stopped")

	// Get message channels from all clients
	var messageChans []<-chan models.Message

	m.mu.RLock()
	for _, client := range m.clients {
		messageChans = append(messageChans, client.GetMessageChannel())
	}
	m.mu.RUnlock()

	if len(messageChans) == 0 {
		m.logger.Warn("No message channels available for webhook dispatch")
		return
	}

	// Use fan-in pattern to receive messages from all clients
	merged := m.mergeChannels(messageChans...)

	for {
		select {
		case msg, ok := <-merged:
			if !ok {
				return
			}

			// Forward message directly to webhook channel
			select {
			case m.webhookChan <- msg:
				m.logger.Debug("Message forwarded to webhook channel",
					zap.String("from", msg.From),
					zap.String("to", msg.To),
				)
			default:
				m.logger.Warn("Webhook channel full, dropping message",
					zap.String("from", msg.From),
				)
			}

		case <-time.After(30 * time.Second):
			// Periodically check if we need to update message channels
			// (in case clients are added/removed)
			m.mu.RLock()
			currentClients := len(m.clients)
			m.mu.RUnlock()

			if currentClients != len(messageChans) {
				m.logger.Info("Client count changed, updating message channels")
				// Restart dispatcher to update channels
				return
			}
		}
	}
}

// mergeChannels merges multiple channels into one using fan-in pattern
func (m *Manager) mergeChannels(channels ...<-chan models.Message) <-chan models.Message {
	output := make(chan models.Message)

	var wg sync.WaitGroup
	wg.Add(len(channels))

	for _, ch := range channels {
		go func(c <-chan models.Message) {
			defer wg.Done()
			for msg := range c {
				output <- msg
			}
		}(ch)
	}

	// Start goroutine to close output channel when all inputs are closed
	go func() {
		wg.Wait()
		close(output)
	}()

	return output
}

// Errors
var (
	ErrNoDefaultClient = &XMPPError{
		Code:    "NO_DEFAULT_CLIENT",
		Message: "No default XMPP client available",
	}
)

// XMPPError represents XMPP related errors
type XMPPError struct {
	Code    string
	Message string
}

func (e *XMPPError) Error() string {
	return e.Message
}
