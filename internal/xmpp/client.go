package xmpp

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"jabber-bot/internal/config"
	"jabber-bot/internal/models"

	"go.uber.org/zap"
	"gosrc.io/xmpp"
	"gosrc.io/xmpp/stanza"
)

// Client represents XMPP client
type Client struct {
	config       *config.Config
	logger       *zap.Logger
	client       *xmpp.Client
	router       *xmpp.Router
	connected    bool
	messageChan  chan models.Message
	mu           sync.RWMutex
	cancelFunc   context.CancelFunc
	streamLogger *os.File
}

// NewClient creates new XMPP client
func NewClient(cfg *config.Config, logger *zap.Logger) *Client {
	return &Client{
		config:      cfg,
		logger:      logger,
		messageChan: make(chan models.Message, 100),
	}
}

// Connect establishes XMPP connection
func (c *Client) Connect(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	c.cancelFunc = cancel

	// Create temporary file for XMPP stream logging
	tempFile, err := os.CreateTemp("", "xmpp-stream-*.log")
	if err != nil {
		c.logger.Warn("Failed to create temp file for XMPP stream logging", zap.Error(err))
		tempFile = nil
	} else {
		c.logger.Info("Created XMPP stream log file", zap.String("file", tempFile.Name()))
		// Start goroutine to monitor and read from temp file
		go c.monitorXMPPStreamLogs(tempFile)
	}
	c.streamLogger = tempFile

	// Create XMPP client configuration
	clientConfig := xmpp.Config{
		TransportConfiguration: xmpp.TransportConfiguration{
			Address: c.config.XMPP.Server,
		},
		Jid:          c.config.XMPP.JID,
		Credential:   xmpp.Password(c.config.XMPP.Password),
		StreamLogger: tempFile,
	}

	// Create router
	c.router = xmpp.NewRouter()
	c.setupHandlers()

	// Create XMPP client
	client, err := xmpp.NewClient(&clientConfig, c.router, func(err error) {
		c.logger.Error("XMPP error", zap.Error(err))
	})
	if err != nil {
		return fmt.Errorf("failed to create XMPP client: %w", err)
	}
	c.client = client

	// Connect to XMPP server
	if err := c.client.Connect(); err != nil {
		return fmt.Errorf("failed to connect to XMPP server: %w", err)
	}

	c.setConnected(true)
	c.logger.Info("Successfully connected to XMPP server",
		zap.String("jid", c.config.XMPP.JID),
		zap.String("server", c.config.XMPP.Server),
	)

	// Start reconnection handler
	go c.handleReconnection(ctx)

	return nil
}

// Disconnect closes XMPP connection
func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cancelFunc != nil {
		c.cancelFunc()
	}

	c.setConnected(false)

	if c.client != nil {
		if err := c.client.Disconnect(); err != nil {
			c.logger.Error("Error during XMPP disconnect", zap.Error(err))
			return err
		}
	}

	// Clean up stream logger temp file if it exists
	if c.streamLogger != nil {
		//goland:noinspection GoUnhandledErrorResult
		c.streamLogger.Close()
		//goland:noinspection GoUnhandledErrorResult
		os.Remove(c.streamLogger.Name())
		c.streamLogger = nil
	}

	close(c.messageChan)
	c.logger.Info("XMPP client disconnected")
	return nil
}

// SendMessage sends message to specified JID
func (c *Client) SendMessage(to, body, messageType string) error {
	if !c.isConnected() {
		return fmt.Errorf("XMPP client is not connected")
	}

	if messageType == "" {
		messageType = "chat"
	}

	msg := stanza.Message{
		Attrs: stanza.Attrs{
			To:   to,
			Type: stanza.StanzaType(messageType),
		},
		Body: body,
	}

	if err := c.client.Send(msg); err != nil {
		c.logger.Error("Failed to send XMPP message",
			zap.String("to", to),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send message: %w", err)
	}

	c.logger.Info("Message sent successfully",
		zap.String("to", to),
		zap.String("type", messageType),
		zap.Int("body_length", len(body)),
	)

	return nil
}

// SendMUCMessage sends message to Multi-User Chat room
func (c *Client) SendMUCMessage(room, body, subject string) error {
	if !c.isConnected() {
		return fmt.Errorf("XMPP client is not connected")
	}

	msg := stanza.Message{
		Attrs: stanza.Attrs{
			To:   room,
			Type: stanza.StanzaType("groupchat"),
		},
		Body: body,
	}

	if subject != "" {
		msg.Subject = subject
	}

	if err := c.client.Send(msg); err != nil {
		c.logger.Error("Failed to send MUC message",
			zap.String("room", room),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send MUC message: %w", err)
	}

	c.logger.Info("MUC message sent successfully",
		zap.String("room", room),
		zap.Int("body_length", len(body)),
	)

	return nil
}

// IsConnected returns connection status
func (c *Client) IsConnected() bool {
	return c.isConnected()
}

// GetMessageChannel returns channel for incoming messages
func (c *Client) GetMessageChannel() <-chan models.Message {
	return c.messageChan
}

// setupHandlers sets up XMPP message handlers
func (c *Client) setupHandlers() {
	c.router.HandleFunc("message", func(s xmpp.Sender, p stanza.Packet) {
		msg, ok := p.(stanza.Message)
		if !ok {
			return
		}

		// Skip empty messages or system messages
		if msg.Body == "" || msg.From == "" {
			return
		}

		// Convert to internal model
		message := models.Message{
			ID:      msg.Id,
			From:    msg.From,
			To:      msg.To,
			Body:    msg.Body,
			Type:    string(msg.Type),
			Subject: msg.Subject,
			Thread:  msg.Thread,
			Stamp:   "",
		}

		// Send to channel (non-blocking)
		select {
		case c.messageChan <- message:
			c.logger.Debug("Message received and queued",
				zap.String("from", msg.From),
				zap.String("to", msg.To),
				zap.String("type", string(msg.Type)),
			)
		default:
			c.logger.Warn("Message channel full, dropping message",
				zap.String("from", msg.From),
			)
		}
	})
}

// handleReconnection handles automatic reconnection
func (c *Client) handleReconnection(ctx context.Context) {
	if !c.config.Reconnection.Enabled {
		return
	}

	c.logger.Info("Reconnection enabled, starting reconnection monitor")

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Second):
			if !c.isConnected() && c.client != nil {
				c.logger.Warn("XMPP connection lost, attempting to reconnect")
				if err := c.reconnect(); err != nil {
					c.logger.Error("Reconnection failed", zap.Error(err))
				}
			}
		}
	}
}

// reconnect attempts to reconnect to XMPP server
func (c *Client) reconnect() error {
	for attempt := 1; attempt <= c.config.Reconnection.MaxAttempts; attempt++ {
		c.logger.Info("Reconnection attempt",
			zap.Int("attempt", attempt),
			zap.Int("max_attempts", c.config.Reconnection.MaxAttempts),
		)

		time.Sleep(c.config.Reconnection.Backoff)

		if err := c.client.Connect(); err != nil {
			c.logger.Error("Reconnection attempt failed",
				zap.Int("attempt", attempt),
				zap.Error(err),
			)
			continue
		}

		c.setConnected(true)
		c.logger.Info("Reconnection successful",
			zap.Int("attempt", attempt),
		)
		return nil
	}

	return fmt.Errorf("failed to reconnect after %d attempts", c.config.Reconnection.MaxAttempts)
}

// isConnected returns connection status (thread-safe)
func (c *Client) isConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// setConnected sets connection status (thread-safe)
func (c *Client) setConnected(connected bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connected = connected
}

// monitorXMPPStreamLogs monitors XMPP stream log file and pipes new content to zap logger
func (c *Client) monitorXMPPStreamLogs(tempFile *os.File) {
	//goland:noinspection GoUnhandledErrorResult
	defer tempFile.Close()

	// Get initial file size
	info, err := tempFile.Stat()
	if err != nil {
		c.logger.Error("Failed to stat temp file", zap.Error(err))
		return
	}

	lastPos := info.Size()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-time.After(100 * time.Millisecond):
			// Check file size
			info, err := tempFile.Stat()
			if err != nil {
				c.logger.Error("Failed to stat temp file", zap.Error(err))
				return
			}

			currentSize := info.Size()
			if currentSize > lastPos {
				// Read new content
				_, err = tempFile.Seek(lastPos, 0)
				if err != nil {
					c.logger.Error("Failed to seek in temp file", zap.Error(err))
					return
				}

				buf := make([]byte, currentSize-lastPos)
				_, err = tempFile.Read(buf)
				if err != nil {
					c.logger.Error("Failed to read from temp file", zap.Error(err))
					return
				}

				// Log the new content
				content := string(buf)
				lines := strings.Split(content, "\n")

				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line != "" {
						c.logger.Debug("XMPP stream", zap.String("data", line))
					}
				}

				lastPos = currentSize
			}
		}
	}
}
