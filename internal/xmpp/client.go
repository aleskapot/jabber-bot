package xmpp

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"jabber-bot/internal/config"
	"jabber-bot/internal/models"

	"go.uber.org/zap"
	"gosrc.io/xmpp"
	"gosrc.io/xmpp/stanza"
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
	router       *xmpp.Router
	connected    int32
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

	// XEP-0184: Request delivery receipt for chat messages (not groupchat)
	if messageType != "groupchat" {
		msg.Extensions = append(msg.Extensions, stanza.ReceiptRequest{})
	}

	// XEP-0085: Add active chat state notification for non-groupchat messages
	if messageType != "groupchat" && body != "" {
		msg.Extensions = append(msg.Extensions, stanza.StateActive{})
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

// SendChatState sends a chat state notification (XEP-0085) to a JID
func (c *Client) SendChatState(to string, state ChatState) error {
	if !c.isConnected() {
		return fmt.Errorf("XMPP client is not connected")
	}

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

	if err := c.client.Send(msg); err != nil {
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

	if err := c.client.Send(receipt); err != nil {
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
	if !c.isConnected() {
		return fmt.Errorf("XMPP client is not connected")
	}

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

	if err := c.client.Send(msg); err != nil {
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

	respChan, err := c.client.SendIQ(context.Background(), &iq)
	if err != nil {
		return "", fmt.Errorf("failed to send service discovery request: %w", err)
	}

	select {
	case resp, ok := <-respChan:
		if !ok {
			return "", fmt.Errorf("discovery response channel closed")
		}

		if resp.Attrs.Type == stanza.IQTypeError {
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
	if !c.isConnected() {
		return fmt.Errorf("XMPP client is not connected")
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

	if err := c.client.Send(msg); err != nil {
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

	respChan, err := c.client.SendIQ(context.Background(), &iq)
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

		if resp.Attrs.Type == stanza.IQTypeError {
			c.logger.Error("Upload slot request failed",
				zap.String("iq_id", iqID),
				zap.String("type", string(resp.Attrs.Type)),
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
	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

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

// GetMessageChannel returns channel for incoming messages
func (c *Client) GetMessageChannel() <-chan models.Message {
	return c.messageChan
}

// setupHandlers sets up XMPP message handlers
func (c *Client) setupHandlers() {
	// Message received handler
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
	c.router.HandleFunc("iq", func(s xmpp.Sender, p stanza.Packet) {
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
	return atomic.LoadInt32(&c.connected) == 1
}

// setConnected sets connection status (thread-safe)
func (c *Client) setConnected(connected bool) {
	if connected {
		atomic.StoreInt32(&c.connected, 1)
	} else {
		atomic.StoreInt32(&c.connected, 0)
	}
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
