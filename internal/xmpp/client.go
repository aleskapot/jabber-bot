package xmpp

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"jabber-bot/internal/config"
	"jabber-bot/internal/metrics"
	"jabber-bot/internal/models"

	"go.uber.org/zap"
	"gosrc.io/xmpp"
	"gosrc.io/xmpp/stanza"
)

var (
	// ErrNotConnected is returned when a send is attempted without an active connection.
	ErrNotConnected = errors.New("XMPP client is not connected")
	// ErrConnectionLost is returned when a send fails because the connection was lost.
	ErrConnectionLost = errors.New("XMPP connection lost")
)

type UploadSlot struct {
	PutURL string
	GetURL string
}

// Client represents XMPP client
type Client struct {
	config       *config.Config
	logger       *zap.Logger
	client       *xmpp.Client
	connected    int32
	messageChan  chan models.Message
	mu           sync.RWMutex
	cancelFunc   context.CancelFunc
	streamLogger *os.File

	// Stream logger lifecycle
	streamLoggerDone chan struct{}

	// Track the actual connection state reported by the XMPP library
	// This is updated by the EventHandler when the library reports state changes
	libraryConnected int32

	// Signal for immediate reconnection attempts
	reconnectCh chan struct{}

	// Set when Disconnect is called so in-flight reconnects do not resurrect
	// the connection after shutdown.
	stopped int32

	// Ensures the incoming message channel is only closed once
	closeMessageOnce sync.Once
}

// NewClient creates new XMPP client
func NewClient(cfg *config.Config, logger *zap.Logger) *Client {
	return &Client{
		config:      cfg,
		logger:      logger,
		messageChan: make(chan models.Message, 100),
		reconnectCh: make(chan struct{}, 1),
	}
}

// Connect establishes XMPP connection
func (c *Client) Connect(ctx context.Context) error {
	atomic.StoreInt32(&c.stopped, 0)
	ctx, cancel := context.WithCancel(ctx)
	c.mu.Lock()
	c.cancelFunc = cancel
	c.mu.Unlock()

	tempFile := c.startStreamLogger()

	// Create XMPP client
	client, err := c.createXMPPClient(tempFile)
	if err != nil {
		c.stopStreamLogger()
		return err
	}

	c.mu.Lock()
	c.client = client
	c.mu.Unlock()

	// Connect to XMPP server
	if err := client.Connect(); err != nil {
		c.mu.Lock()
		if c.client == client {
			c.client = nil
		}
		c.mu.Unlock()
		go func() {
			_ = client.Disconnect()
		}()
		c.stopStreamLogger()
		return fmt.Errorf("failed to connect to XMPP server: %w", err)
	}

	c.mu.Lock()
	if atomic.LoadInt32(&c.stopped) == 1 || c.client != client {
		c.mu.Unlock()
		go func() {
			_ = client.Disconnect()
		}()
		c.stopStreamLogger()
		return nil
	}
	c.setConnected(true)
	atomic.StoreInt32(&c.libraryConnected, 1)
	c.mu.Unlock()

	c.logger.Info("Successfully connected to XMPP server",
		zap.String("jid", c.config.XMPP.JID),
		zap.String("server", c.config.XMPP.Server),
	)

	// Start reconnection handler
	go c.handleReconnection(ctx)

	return nil
}

// createXMPPClient creates a new XMPP client instance.
// Handlers are bound to the concrete client instance so that events from stale
// clients (after a reconnect) are ignored.
func (c *Client) createXMPPClient(tempFile *os.File) (*xmpp.Client, error) {
	// Create XMPP client configuration
	clientConfig := xmpp.Config{
		TransportConfiguration: xmpp.TransportConfiguration{
			Address: c.config.XMPP.Server,
		},
		Jid:               c.config.XMPP.JID,
		Credential:        xmpp.Password(c.config.XMPP.Password),
		StreamLogger:      tempFile,
		KeepaliveInterval: c.keepaliveInterval(),
		ConnectTimeout:    c.connectTimeoutSeconds(),
	}

	// Create router
	router := xmpp.NewRouter()
	c.setupHandlers(router)

	var client *xmpp.Client
	var err error
	client, err = xmpp.NewClient(&clientConfig, router, func(err error) {
		c.handleClientError(client, err)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create XMPP client: %w", err)
	}

	return client, nil
}

// Disconnect closes XMPP connection
func (c *Client) Disconnect() error {
	atomic.StoreInt32(&c.stopped, 1)

	c.mu.Lock()
	cancel := c.cancelFunc
	c.cancelFunc = nil
	c.mu.Unlock()
	if cancel != nil {
		cancel()
	}

	c.setConnected(false)
	atomic.StoreInt32(&c.libraryConnected, 0)

	c.mu.Lock()
	client := c.client
	c.client = nil
	c.mu.Unlock()

	var disconnectErr error
	if client != nil {
		if err := client.Disconnect(); err != nil {
			c.logger.Error("Error during XMPP disconnect", zap.Error(err))
			disconnectErr = err
		}
	}

	// Clean up stream logger temp file if it exists
	c.stopStreamLogger()

	c.closeMessageOnce.Do(func() {
		close(c.messageChan)
	})
	c.logger.Info("XMPP client disconnected")
	return disconnectErr
}

// SendMessage sends message to specified JID
func (c *Client) SendMessage(to, body, messageType string) error {
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

	// XEP-0184: Request delivery receipt for chat messages (not groupchat)
	if messageType != "groupchat" {
		msg.Extensions = append(msg.Extensions, stanza.ReceiptRequest{})
	}

	// XEP-0085: Add active chat state notification for non-groupchat messages
	if messageType != "groupchat" && body != "" {
		msg.Extensions = append(msg.Extensions, stanza.StateActive{})
	}

	if err := c.sendPacket(msg); err != nil {
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

	if err := c.sendPacket(msg); err != nil {
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

// SendChatState sends a chat state notification (XEP-0085) to a JID
func (c *Client) SendChatState(to string, state ChatState) error {
	var chatState stanza.MsgExtension
	switch state {
	case ChatStateActive:
		chatState = stanza.StateActive{}
	case ChatStateComposing:
		chatState = stanza.StateComposing{}
	case ChatStatePaused:
		chatState = stanza.StatePaused{}
	case ChatStateInactive:
		chatState = stanza.StateInactive{}
	case ChatStateGone:
		chatState = stanza.StateGone{}
	default:
		return fmt.Errorf("invalid chat state: %s", state)
	}

	msg := stanza.Message{
		Attrs: stanza.Attrs{
			To:   to,
			Type: stanza.StanzaType("chat"),
		},
		Body: "",
		Extensions: []stanza.MsgExtension{
			chatState,
		},
	}

	if err := c.sendPacket(msg); err != nil {
		c.logger.Error("Failed to send chat state notification",
			zap.String("to", to),
			zap.String("state", string(state)),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send chat state: %w", err)
	}

	c.logger.Info("Chat state notification sent",
		zap.String("to", to),
		zap.String("state", string(state)),
	)

	return nil
}

// SendDeliveryReceipt sends a delivery receipt (XEP-0184)
func (c *Client) SendDeliveryReceipt(to, messageID string) error {
	if to == "" || messageID == "" {
		return fmt.Errorf("recipient and message ID are required")
	}

	receipt := stanza.Message{
		Attrs: stanza.Attrs{
			To:   to,
			Type: stanza.StanzaType("chat"),
		},
		Extensions: []stanza.MsgExtension{
			stanza.ReceiptReceived{ID: messageID},
		},
	}

	if err := c.sendPacket(receipt); err != nil {
		c.logger.Error("Failed to send delivery receipt",
			zap.String("to", to),
			zap.String("message_id", messageID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send delivery receipt: %w", err)
	}

	c.logger.Info("Delivery receipt sent",
		zap.String("to", to),
		zap.String("message_id", messageID),
	)

	return nil
}

// SendFile sends a file to specified JID using OOB (Out-of-Band) URI
func (c *Client) SendFile(to, fileURL, fileName, fileType string) error {
	// Create message body with file info
	body := fmt.Sprintf("File: %s\nType: %s\nURL: %s", fileName, fileType, fileURL)

	msg := stanza.Message{
		Attrs: stanza.Attrs{
			To:   to,
			Type: stanza.StanzaType("chat"),
		},
		Body: body,
	}

	// Add OOB extension with HTTP URL (XEP-0066 Out-of-Band Data)
	// This allows the recipient to download the file from the provided URL
	oob := stanza.OOB{
		URL:  fileURL,
		Desc: fileName,
	}
	msg.Extensions = append(msg.Extensions, oob)

	// Request delivery receipt
	msg.Extensions = append(msg.Extensions, stanza.ReceiptRequest{})

	// Add active chat state
	msg.Extensions = append(msg.Extensions, stanza.StateActive{})

	if err := c.sendPacket(msg); err != nil {
		c.logger.Error("Failed to send file",
			zap.String("to", to),
			zap.String("file", fileName),
			zap.String("file_url", fileURL),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send file: %w", err)
	}

	c.logger.Info("File sent successfully",
		zap.String("to", to),
		zap.String("file", fileName),
		zap.String("file_type", fileType),
		zap.String("file_url", fileURL),
	)

	return nil
}

func (c *Client) discoverUploadService(serverDomain string) (string, error) {
	iqID := fmt.Sprintf("disco-%d", time.Now().UnixNano())

	iq := stanza.IQ{
		Attrs: stanza.Attrs{
			Id:   iqID,
			Type: stanza.IQTypeGet,
			To:   serverDomain,
		},
		Payload: &stanza.DiscoItems{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	respChan, err := c.sendIQ(ctx, &iq)
	if err != nil {
		return "", fmt.Errorf("failed to send service discovery request: %w", err)
	}

	select {
	case resp, ok := <-respChan:
		if !ok {
			return "", fmt.Errorf("discovery response channel closed")
		}

		if resp.Type == stanza.IQTypeError {
			return "", fmt.Errorf("service discovery failed")
		}

		items, ok := resp.Payload.(*stanza.DiscoItems)
		if !ok {
			return "", fmt.Errorf("invalid discovery response")
		}

		for _, item := range items.Items {
			if strings.Contains(item.JID, "upload") {
				return item.JID, nil
			}
		}

		return "", fmt.Errorf("no upload service found on domain %s", serverDomain)

	case <-time.After(10 * time.Second):
		return "", fmt.Errorf("service discovery timed out")
	}
}

func (c *Client) SendFileXEP0363(to, filePath, fileName, fileType string) error {
	if err := c.checkConnected(); err != nil {
		return err
	}

	domain := strings.Split(c.config.XMPP.JID, "@")
	if len(domain) < 2 {
		return fmt.Errorf("invalid JID format")
	}
	serverDomain := domain[1]

	uploadService, err := c.discoverUploadService(serverDomain)
	if err != nil {
		return fmt.Errorf("failed to discover upload service: %w", err)
	}

	c.logger.Info("Discovered upload service",
		zap.String("upload_service", uploadService),
	)

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	size := fileInfo.Size()
	maxSize := c.config.FileTransfer.MaxSize
	if maxSize > 0 && size > maxSize {
		return fmt.Errorf("file size %d exceeds maximum allowed %d", size, maxSize)
	}

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	slot, err := c.requestUploadSlot(uploadService, fileName, size, fileType)
	if err != nil {
		return fmt.Errorf("failed to request upload slot: %w", err)
	}

	timeout := c.config.FileTransfer.Timeout
	if timeout == 0 {
		timeout = 60 * time.Second
	}

	if err := c.uploadFileToURL(slot.PutURL, fileData, fileType, timeout); err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	if err := c.sendFileWithXEP0447(to, slot.GetURL, fileName, fileType, size, fileData); err != nil {
		return fmt.Errorf("failed to send file message: %w", err)
	}

	c.logger.Info("File uploaded and sent via XEP-0363",
		zap.String("to", to),
		zap.String("file", fileName),
		zap.Int64("size", size),
		zap.String("get_url", slot.GetURL),
	)

	return nil
}

func (c *Client) sendFileWithXEP0447(to, fileURL, fileName, fileType string, size int64, fileData []byte) error {
	hash := sha256.Sum256(fileData)
	hashBase64 := base64.StdEncoding.EncodeToString(hash[:])

	msg := stanza.Message{
		Attrs: stanza.Attrs{
			To:   to,
			From: c.config.XMPP.JID,
			Type: stanza.StanzaType("chat"),
		},
		Body: fileURL,
	}

	fileSharing := FileSharing{
		Disposition: "inline",
		File: &FileMetadata{
			MediaType: fileType,
			Name:      fileName,
			Size:      size,
			Hashes: []FileHash{
				{
					Algo:  "sha-256",
					Value: hashBase64,
				},
			},
			Desc: fileName,
		},
		Sources: &FileSources{
			URLSources: []URLDataSource{
				{
					Target: fileURL,
				},
			},
		},
	}

	msg.Extensions = append(msg.Extensions, fileSharing)

	oob := stanza.OOB{
		URL:  fileURL,
		Desc: fileName,
	}
	msg.Extensions = append(msg.Extensions, oob)

	msg.Extensions = append(msg.Extensions, stanza.ReceiptRequest{})
	msg.Extensions = append(msg.Extensions, stanza.StateActive{})

	if err := c.sendPacket(msg); err != nil {
		c.logger.Error("Failed to send file via XEP-0447",
			zap.String("to", to),
			zap.String("file", fileName),
			zap.String("file_url", fileURL),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send file: %w", err)
	}

	c.logger.Info("File sent via XEP-0447",
		zap.String("to", to),
		zap.String("file", fileName),
		zap.String("url", fileURL),
	)

	return nil
}

func (c *Client) requestUploadSlot(uploadService, filename string, size int64, contentType string) (*UploadSlot, error) {
	domain := strings.Split(c.config.XMPP.JID, "@")
	if len(domain) < 2 {
		return nil, fmt.Errorf("invalid JID format")
	}
	serverDomain := domain[1]

	if !strings.Contains(uploadService, ".") {
		uploadService = "upload." + serverDomain
	}

	iqID := fmt.Sprintf("upload-%d", time.Now().UnixNano())

	iq := stanza.IQ{
		Attrs: stanza.Attrs{
			Id:   iqID,
			Type: stanza.IQTypeGet,
			To:   uploadService,
			From: c.config.XMPP.JID,
		},
		Payload: UploadRequest{
			Filename:    filename,
			Size:        size,
			ContentType: contentType,
			XMLName: xml.Name{
				Space: nsHTTPUpload,
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	respChan, err := c.sendIQ(ctx, &iq)
	if err != nil {
		return nil, fmt.Errorf("failed to send upload slot request: %w", err)
	}

	c.logger.Debug("Sent upload slot request",
		zap.String("iq_id", iqID),
		zap.String("upload_service", uploadService),
		zap.String("filename", filename),
		zap.Int64("size", size),
	)

	select {
	case resp, ok := <-respChan:
		if !ok {
			return nil, fmt.Errorf("response channel closed")
		}

		if resp.Type == stanza.IQTypeError {
			c.logger.Error("Upload slot request failed",
				zap.String("iq_id", iqID),
				zap.String("type", string(resp.Type)),
			)
			return nil, fmt.Errorf("slot request failed")
		}

		slot := c.parseSlotResponse(resp)
		if slot != nil {
			c.logger.Info("Received upload slot",
				zap.String("put_url", slot.PutURL),
				zap.String("get_url", slot.GetURL),
			)
			return slot, nil
		}

		return nil, fmt.Errorf("failed to parse slot response")

	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("timeout waiting for slot response")
	}
}

func (c *Client) parseSlotResponse(resp stanza.IQ) *UploadSlot {
	if resp.Payload == nil {
		c.logger.Debug("No payload in IQ response")
		return nil
	}

	slotResp, ok := resp.Payload.(*UploadSlotResponse)
	if !ok {
		c.logger.Debug("Payload is not *UploadSlotResponse", zap.String("type", fmt.Sprintf("%T", resp.Payload)))
		return nil
	}

	if slotResp.Put.URL == "" || slotResp.Get.URL == "" {
		c.logger.Debug("Missing URLs in slot response")
		return nil
	}

	return &UploadSlot{
		PutURL: slotResp.Put.URL,
		GetURL: slotResp.Get.URL,
	}
}

func (c *Client) uploadFileToURL(putURL string, fileData []byte, contentType string, timeout time.Duration) error {
	req, err := http.NewRequest(http.MethodPut, putURL, bytes.NewReader(fileData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(fileData)))

	if c.config.XMPP.Password != "" {
		auth := base64.StdEncoding.EncodeToString([]byte(c.config.XMPP.JID + ":" + c.config.XMPP.Password))
		req.Header.Set("Authorization", "Basic "+auth)
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// IsConnected returns connection status
func (c *Client) IsConnected() bool {
	return c.isConnected()
}

// IsLibraryConnected returns the connection state as reported by the XMPP library
func (c *Client) IsLibraryConnected() bool {
	return atomic.LoadInt32(&c.libraryConnected) == 1
}

// GetMessageChannel returns channel for incoming messages
func (c *Client) GetMessageChannel() <-chan models.Message {
	return c.messageChan
}

// setupHandlers sets up XMPP message handlers on the given router
func (c *Client) setupHandlers(router *xmpp.Router) {
	// Message received handler
	router.HandleFunc("message", func(s xmpp.Sender, p stanza.Packet) {
		msg, ok := p.(stanza.Message)
		if !ok {
			return
		}

		// Skip empty messages or system messages
		if msg.Body == "" || msg.From == "" {
			return
		}

		// Convert to internal model
		receiptRequested := false
		for _, ext := range msg.Extensions {
			switch ext.(type) {
			case stanza.ReceiptRequest:
				receiptRequested = true
			case *stanza.ReceiptRequest:
				receiptRequested = true
			}
			if receiptRequested {
				break
			}
		}

		message := models.Message{
			ID:               msg.Id,
			From:             msg.From,
			To:               msg.To,
			Body:             msg.Body,
			Type:             string(msg.Type),
			Subject:          msg.Subject,
			Thread:           msg.Thread,
			Stamp:            "",
			ReceiptRequested: receiptRequested,
		}

		// Send to channel (non-blocking)
		select {
		case c.messageChan <- message:
			c.logger.Debug("Message received and queued",
				zap.String("from", msg.From),
				zap.String("to", msg.To),
				zap.String("type", string(msg.Type)),
				zap.Bool("receipt_requested", receiptRequested),
			)
		default:
			c.logger.Warn("Message channel full, dropping message",
				zap.String("from", msg.From),
			)
		}
	})

	// Information query received handler
	router.HandleFunc("iq", func(s xmpp.Sender, p stanza.Packet) {
		iq, ok := p.(*stanza.IQ)
		if !ok {
			return
		}

		if iq.Type != stanza.IQTypeGet {
			return
		}

		if iq.Payload == nil {
			return
		}

		switch iq.Payload.(type) {
		case *stanza.Version:
			c.logger.Debug("Received version query from server",
				zap.String("from", iq.From),
				zap.String("id", iq.Id),
			)

			versionResp := stanza.IQ{
				Attrs: stanza.Attrs{
					Id:   iq.Id,
					Type: stanza.IQTypeResult,
					To:   iq.From,
					From: iq.To,
				},
				Payload: &stanza.Version{},
			}
			versionResp.Payload.(*stanza.Version).SetInfo("jabber-bot", "1.0.0", "Linux")

			if err := s.Send(&versionResp); err != nil {
				c.logger.Error("Failed to send version response",
					zap.Error(err),
				)
			}
		}
	})
}

// handleReconnection handles automatic reconnection
func (c *Client) handleReconnection(ctx context.Context) {
	if !c.config.Reconnection.Enabled {
		return
	}

	interval := c.watchdogInterval()
	c.logger.Info("Reconnection enabled, starting reconnection monitor",
		zap.Duration("watchdog_interval", interval),
	)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.reconnectCh:
			if !c.isConnected() {
				if err := c.reconnect(ctx); err != nil {
					c.logger.Error("Reconnection failed", zap.Error(err))
				}
			}
		case <-ticker.C:
			if !c.isConnected() {
				c.logger.Warn("XMPP connection lost, attempting to reconnect")
				if err := c.reconnect(ctx); err != nil {
					c.logger.Error("Reconnection failed", zap.Error(err))
				}
			}
		}
	}
}

// reconnect performs a full reconnect to the XMPP server until it succeeds or
// the context is cancelled.
func (c *Client) reconnect(ctx context.Context) error {
	for attempt := 1; ; attempt++ {
		if atomic.LoadInt32(&c.stopped) == 1 {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if maxAttempts := c.config.Reconnection.MaxAttempts; maxAttempts > 0 && attempt > maxAttempts {
			return fmt.Errorf("reconnection failed after %d attempts", maxAttempts)
		}

		if attempt > 1 {
			delay := c.reconnectBackoff(attempt - 1)
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}

		c.logger.Info("Reconnection attempt",
			zap.Int("attempt", attempt),
			zap.Int("max_attempts", c.config.Reconnection.MaxAttempts),
		)

		// Detach the old client so that sends fail fast as not connected while
		// we establish a new session.
		c.mu.Lock()
		oldClient := c.client
		c.client = nil
		c.mu.Unlock()
		c.setConnected(false)
		atomic.StoreInt32(&c.libraryConnected, 0)

		if oldClient != nil {
			go func(client *xmpp.Client) {
				if err := client.Disconnect(); err != nil {
					c.logger.Debug("Error disconnecting old client during reconnect", zap.Error(err))
				}
			}(oldClient)
		}

		c.stopStreamLogger()
		tempFile := c.startStreamLogger()

		newClient, err := c.createXMPPClient(tempFile)
		if err != nil {
			c.logger.Error("Failed to create new XMPP client",
				zap.Int("attempt", attempt),
				zap.Error(err),
			)
			c.stopStreamLogger()
			continue
		}

		if err := newClient.Connect(); err != nil {
			c.logger.Error("Reconnection attempt failed",
				zap.Int("attempt", attempt),
				zap.Error(err),
			)
			go func() {
				_ = newClient.Disconnect()
			}()
			c.stopStreamLogger()
			continue
		}

		c.mu.Lock()
		if atomic.LoadInt32(&c.stopped) == 1 {
			c.mu.Unlock()
			go func() {
				_ = newClient.Disconnect()
			}()
			c.stopStreamLogger()
			return nil
		}
		c.client = newClient
		c.setConnected(true)
		atomic.StoreInt32(&c.libraryConnected, 1)
		c.mu.Unlock()

		metrics.RecordXMPPReconnect()
		c.logger.Info("Reconnection successful",
			zap.Int("attempt", attempt),
		)
		return nil
	}
}

// reconnectBackoff returns the delay before the given reconnect attempt.
func (c *Client) reconnectBackoff(attempt int) time.Duration {
	base := 5 * time.Second
	maxBackoff := 60 * time.Second
	if c.config != nil {
		if c.config.Reconnection.Backoff > 0 {
			base = c.config.Reconnection.Backoff
		}
		if c.config.Reconnection.MaxBackoff > 0 {
			maxBackoff = c.config.Reconnection.MaxBackoff
		}
	}
	if maxBackoff < base {
		maxBackoff = base
	}

	delay := base
	for i := 1; i < attempt && delay < maxBackoff; i++ {
		delay *= 2
	}
	if delay > maxBackoff {
		delay = maxBackoff
	}

	// Add +/-20% jitter to avoid a reconnect storm.
	if half := delay / 5; half > 0 {
		delay += time.Duration(rand.Int63n(int64(2*half)+1)) - half
	}
	if delay < time.Second {
		delay = time.Second
	}
	return delay
}

// handleClientError handles errors reported by the XMPP library transport.
func (c *Client) handleClientError(client *xmpp.Client, err error) {
	if err == nil || !c.isCurrentClient(client) {
		return
	}

	c.logger.Error("XMPP error", zap.Error(err))
	c.markDisconnected("client error")
}

// markDisconnected clears the connection state and wakes up the reconnect loop.
func (c *Client) markDisconnected(reason string) {
	wasConnected := atomic.LoadInt32(&c.connected) == 1
	c.setConnected(false)
	atomic.StoreInt32(&c.libraryConnected, 0)

	if wasConnected {
		metrics.RecordXMPPDisconnect()
		c.logger.Warn("XMPP connection lost", zap.String("reason", reason))
	}

	select {
	case c.reconnectCh <- struct{}{}:
	default:
	}
}

// isCurrentClient reports whether the given client is the currently active one.
func (c *Client) isCurrentClient(client *xmpp.Client) bool {
	if client == nil {
		return false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.client == client
}

// currentClient returns the currently active XMPP client.
func (c *Client) currentClient() *xmpp.Client {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.client
}

// checkConnected returns ErrNotConnected when there is no usable connection.
func (c *Client) checkConnected() error {
	if !c.isConnected() {
		return ErrNotConnected
	}
	return nil
}

// connectedClient returns the active client or ErrNotConnected.
func (c *Client) connectedClient() (*xmpp.Client, error) {
	if !c.isConnected() {
		return nil, ErrNotConnected
	}
	client := c.currentClient()
	if client == nil {
		return nil, ErrNotConnected
	}
	return client, nil
}

// sendPacket sends a stanza using the current connection. It never retries:
// a lost connection is reported to the caller so it can decide when to retry.
func (c *Client) sendPacket(packet stanza.Packet) error {
	client, err := c.connectedClient()
	if err != nil {
		return err
	}

	if err := client.Send(packet); err != nil {
		if isConnectionError(err) {
			c.logger.Error("XMPP send failed due to lost connection", zap.Error(err))
			c.markDisconnected("send failure")
			return fmt.Errorf("%w: %v", ErrConnectionLost, err)
		}
		return err
	}
	return nil
}

// sendIQ sends an IQ stanza and returns the result channel.
func (c *Client) sendIQ(ctx context.Context, iq *stanza.IQ) (chan stanza.IQ, error) {
	client, err := c.connectedClient()
	if err != nil {
		return nil, err
	}

	respChan, err := client.SendIQ(ctx, iq)
	if err != nil {
		if isConnectionError(err) {
			c.logger.Error("XMPP IQ send failed due to lost connection", zap.Error(err))
			c.markDisconnected("send failure")
			return nil, fmt.Errorf("%w: %v", ErrConnectionLost, err)
		}
		return nil, err
	}
	return respChan, nil
}

// isConnectionError reports whether an error indicates a broken transport.
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}

	msg := strings.ToLower(err.Error())
	for _, marker := range []string{
		"broken pipe",
		"connection reset",
		"connection refused",
		"connection aborted",
		"connection closed",
		"forcibly closed",
		"network is unreachable",
		"use of closed network connection",
		"no readwriter",
		"not connected",
		"i/o timeout",
	} {
		if strings.Contains(msg, marker) {
			return true
		}
	}
	return false
}

// isConnected returns connection status (thread-safe)
func (c *Client) isConnected() bool {
	if atomic.LoadInt32(&c.connected) != 1 {
		return false
	}
	if atomic.LoadInt32(&c.libraryConnected) != 1 {
		return false
	}
	return c.currentClient() != nil
}

// setConnected sets connection status (thread-safe)
func (c *Client) setConnected(connected bool) {
	if connected {
		atomic.StoreInt32(&c.connected, 1)
	} else {
		atomic.StoreInt32(&c.connected, 0)
	}
}

func (c *Client) keepaliveInterval() time.Duration {
	if c.config != nil && c.config.Reconnection.KeepaliveInterval > 0 {
		return c.config.Reconnection.KeepaliveInterval
	}
	return 30 * time.Second
}

func (c *Client) connectTimeoutSeconds() int {
	timeout := 10 * time.Second
	if c.config != nil && c.config.Reconnection.ConnectTimeout > 0 {
		timeout = c.config.Reconnection.ConnectTimeout
	}
	seconds := int(timeout / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	return seconds
}

func (c *Client) watchdogInterval() time.Duration {
	if c.config != nil && c.config.Reconnection.WatchdogInterval > 0 {
		return c.config.Reconnection.WatchdogInterval
	}
	return 10 * time.Second
}

// startStreamLogger creates a temporary stream log file and starts its monitor.
func (c *Client) startStreamLogger() *os.File {
	tempFile, err := os.CreateTemp("", "xmpp-stream-*.log")
	if err != nil {
		c.logger.Warn("Failed to create temp file for XMPP stream logging", zap.Error(err))
		c.mu.Lock()
		c.streamLogger = nil
		c.streamLoggerDone = nil
		c.mu.Unlock()
		return nil
	}

	done := make(chan struct{})
	c.logger.Info("Created XMPP stream log file", zap.String("file", tempFile.Name()))

	c.mu.Lock()
	c.streamLogger = tempFile
	c.streamLoggerDone = done
	c.mu.Unlock()

	go c.monitorXMPPStreamLogs(tempFile, done)
	return tempFile
}

// stopStreamLogger stops the monitor and removes the temporary stream log file.
func (c *Client) stopStreamLogger() {
	c.mu.Lock()
	tempFile := c.streamLogger
	done := c.streamLoggerDone
	c.streamLogger = nil
	c.streamLoggerDone = nil
	c.mu.Unlock()

	if done != nil {
		close(done)
	}
	if tempFile != nil {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
	}
}

// monitorXMPPStreamLogs monitors XMPP stream log file and pipes new content to zap logger
func (c *Client) monitorXMPPStreamLogs(tempFile *os.File, done <-chan struct{}) {
	defer func() {
		_ = tempFile.Close()
	}()

	// Get initial file size
	info, err := tempFile.Stat()
	if err != nil {
		return
	}

	lastPos := info.Size()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			// Check file size
			info, err := tempFile.Stat()
			if err != nil {
				return
			}

			currentSize := info.Size()
			if currentSize <= lastPos {
				continue
			}

			// Read new content
			if _, err = tempFile.Seek(lastPos, 0); err != nil {
				return
			}

			buf := make([]byte, currentSize-lastPos)
			if _, err = tempFile.Read(buf); err != nil {
				return
			}

			// Log the new content
			content := string(buf)
			lines := strings.SplitSeq(content, "\n")

			for line := range lines {
				line = strings.TrimSpace(line)
				if line != "" {
					c.logger.Debug("XMPP stream", zap.String("data", line))
				}
			}

			lastPos = currentSize
		}
	}
}
