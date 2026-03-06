package api

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"jabber-bot/internal/models"
	"jabber-bot/internal/xmpp"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// handleSendMessage handles POST /api/v1/send
func (s *Server) handleSendMessage(c *fiber.Ctx) error {
	logger := c.Locals("logger").(*zap.Logger)
	manager := c.Locals("manager").(XMPPManagerInterface)

	// Parse request body
	var req models.SendMessageRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Warn("Invalid request body",
			zap.Error(err),
			zap.String("request_id", c.GetRespHeader("X-Request-ID")),
		)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate request
	if err := s.validateSendMessageRequest(&req); err != nil {
		logger.Warn("Request validation failed",
			zap.Error(err),
			zap.String("request_id", c.GetRespHeader("X-Request-ID")),
		)
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	logger.Info("Sending message",
		zap.String("to", req.To),
		zap.String("type", req.Type),
		zap.Int("body_length", len(req.Body)),
		zap.String("request_id", c.GetRespHeader("X-Request-ID")),
	)

	// Send message via XMPP manager
	err := manager.SendMessage(req.To, req.Body, req.Type)
	if err != nil {
		logger.Error("Failed to send XMPP message",
			zap.Error(err),
			zap.String("to", req.To),
			zap.String("request_id", c.GetRespHeader("X-Request-ID")),
		)

		response := models.ErrorResponse{
			Success: false,
			Error:   "Failed to send message: " + err.Error(),
			Code:    fiber.StatusInternalServerError,
		}

		return c.Status(fiber.StatusInternalServerError).JSON(response)
	}

	// Success response
	response := models.APIResponse{
		Success: true,
		Message: "Message sent successfully",
		Data: map[string]interface{}{
			"to":          req.To,
			"type":        req.Type,
			"body_length": len(req.Body),
			"sent_at":     time.Now().UTC().Format(time.RFC3339),
			"request_id":  c.GetRespHeader("X-Request-ID"),
		},
	}

	return c.JSON(response)
}

// handleSendMUCMessage handles POST /api/v1/send-muc
func (s *Server) handleSendMUCMessage(c *fiber.Ctx) error {
	logger := c.Locals("logger").(*zap.Logger)
	manager := c.Locals("manager").(XMPPManagerInterface)

	// Parse request body
	var req models.SendMUCMessageRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Warn("Invalid request body",
			zap.Error(err),
			zap.String("request_id", c.GetRespHeader("X-Request-ID")),
		)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate request
	if err := s.validateSendMUCMessageRequest(&req); err != nil {
		logger.Warn("Request validation failed",
			zap.Error(err),
			zap.String("request_id", c.GetRespHeader("X-Request-ID")),
		)
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	logger.Info("Sending MUC message",
		zap.String("room", req.Room),
		zap.String("subject", req.Subject),
		zap.Int("body_length", len(req.Body)),
		zap.String("request_id", c.GetRespHeader("X-Request-ID")),
	)

	// Send MUC message via XMPP manager
	err := manager.SendMUCMessage(req.Room, req.Body, req.Subject)
	if err != nil {
		logger.Error("Failed to send MUC message",
			zap.Error(err),
			zap.String("room", req.Room),
			zap.String("request_id", c.GetRespHeader("X-Request-ID")),
		)

		response := models.ErrorResponse{
			Success: false,
			Error:   "Failed to send MUC message: " + err.Error(),
			Code:    fiber.StatusInternalServerError,
		}

		return c.Status(fiber.StatusInternalServerError).JSON(response)
	}

	// Success response
	response := models.APIResponse{
		Success: true,
		Message: "MUC message sent successfully",
		Data: map[string]interface{}{
			"room":        req.Room,
			"subject":     req.Subject,
			"body_length": len(req.Body),
			"sent_at":     time.Now().UTC().Format(time.RFC3339),
			"request_id":  c.GetRespHeader("X-Request-ID"),
		},
	}

	return c.JSON(response)
}

// handleSendChatState handles POST /api/v1/chat-state
func (s *Server) handleSendChatState(c *fiber.Ctx) error {
	//goland:noinspection DuplicatedCode
	logger := c.Locals("logger").(*zap.Logger)
	manager := c.Locals("manager").(XMPPManagerInterface)

	var req models.SendChatStateRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Warn("Invalid request body",
			zap.Error(err),
			zap.String("request_id", c.GetRespHeader("X-Request-ID")),
		)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if err := s.validateSendChatStateRequest(&req); err != nil {
		logger.Warn("Request validation failed",
			zap.Error(err),
			zap.String("request_id", c.GetRespHeader("X-Request-ID")),
		)
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	logger.Info("Sending chat state",
		zap.String("to", req.To),
		zap.String("state", req.State),
		zap.String("request_id", c.GetRespHeader("X-Request-ID")),
	)

	state := xmpp.ChatState(req.State)
	err := manager.SendChatState(req.To, state)
	if err != nil {
		logger.Error("Failed to send chat state",
			zap.Error(err),
			zap.String("to", req.To),
			zap.String("request_id", c.GetRespHeader("X-Request-ID")),
		)

		response := models.ErrorResponse{
			Success: false,
			Error:   "Failed to send chat state: " + err.Error(),
			Code:    fiber.StatusInternalServerError,
		}

		return c.Status(fiber.StatusInternalServerError).JSON(response)
	}

	response := models.APIResponse{
		Success: true,
		Message: "Chat state sent successfully",
		Data: map[string]interface{}{
			"to":         req.To,
			"state":      req.State,
			"sent_at":    time.Now().UTC().Format(time.RFC3339),
			"request_id": c.GetRespHeader("X-Request-ID"),
		},
	}

	return c.JSON(response)
}

// handleStatus handles GET /api/v1/status
func (s *Server) handleStatus(c *fiber.Ctx) error {
	logger := c.Locals("logger").(*zap.Logger)
	manager := c.Locals("manager").(XMPPManagerInterface)

	logger.Debug("Status requested",
		zap.String("request_id", c.GetRespHeader("X-Request-ID")),
	)

	// Get connection status
	xmppConnected := manager.IsConnected()

	// Build status response
	response := models.StatusResponse{
		XMPPConnected: xmppConnected,
		APIRunning:    true,
		WebhookConfig: s.config.Webhook.URL,
		Version:       "1.0.0",
	}

	return c.JSON(response)
}

// handleWebhookStatus handles GET /api/v1/webhook/status
func (s *Server) handleWebhookStatus(c *fiber.Ctx) error {
	logger := c.Locals("logger").(*zap.Logger)

	logger.Debug("Webhook status requested",
		zap.String("request_id", c.GetRespHeader("X-Request-ID")),
	)

	// For now, return basic webhook status
	// In a real implementation, you would get this from the webhook manager
	webhookStatus := map[string]interface{}{
		"running":      false, // Would come from webhook manager
		"healthy":      false, // Would come from webhook manager
		"queue_length": 0,     // Would come from webhook manager
		"webhook_url":  s.config.Webhook.URL,
		"total_sent":   int64(0),
		"total_failed": int64(0),
	}

	return c.JSON(webhookStatus)
}

// handleHealth handles GET /api/v1/health
func (s *Server) handleHealth(c *fiber.Ctx) error {
	manager := c.Locals("manager").(XMPPManagerInterface)

	// Simple health check
	healthy := true

	// Check XMPP connection
	if !manager.IsConnected() {
		healthy = false
	}

	status := fiber.Map{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	if !healthy {
		status["status"] = "error"
		status["error"] = "XMPP connection lost"
		return c.Status(fiber.StatusServiceUnavailable).JSON(status)
	}

	return c.JSON(status)
}

// handleRoot handles GET /
func (s *Server) handleRoot(c *fiber.Ctx) error {
	info := fiber.Map{
		"name":        "Jabber Bot API",
		"version":     "1.0.0",
		"description": "XMPP Jabber bot with RESTful API",
		"endpoints": map[string]string{
			"send":         "/api/v1/send - Send XMPP message",
			"send_muc":     "/api/v1/send-muc - Send MUC message",
			"send_file":    "/api/v1/send-file - Send file via XMPP",
			"status":       "/api/v1/status - Get bot status",
			"health":       "/api/v1/health - Health check",
			"webhook":      "/api/v1/webhook/status - Get webhook status",
			"docs":         "/docs - API documentation",
			"openapi":      "/openapi.yaml - OpenAPI specification (YAML)",
			"openapi_json": "/openapi.json - OpenAPI specification (JSON)",
		},
	}

	return c.JSON(info)
}

// handleDocs handles GET /docs
func (s *Server) handleDocs(c *fiber.Ctx) error {
	docs := `# Jabber Bot API Documentation

## Overview
This API provides endpoints to send XMPP messages and receive webhook notifications.

## Base URL
http://localhost:8080/api/v1

## OpenAPI Specification
- **YAML**: /openapi.yaml
- **JSON**: /openapi.json
- **Swagger UI**: Use any OpenAPI/Swagger viewer with the above URLs

## Endpoints

### Send Message
**POST /api/v1/send**
Send a message to an XMPP user.

**Request Body:**
{
  "to": "user@example.com",
  "body": "Hello, world!",
  "type": "chat" // optional, defaults to "chat"
}

**Response:**
{
  "success": true,
  "message": "Message sent successfully",
  "data": {
    "to": "user@example.com",
    "type": "chat",
    "body_length": 13,
    "sent_at": "2023-12-01T12:00:00Z",
    "request_id": "abc123"
  }
}

### Send MUC Message
**POST /api/v1/send-muc**
Send a message to a Multi-User Chat room.

**Request Body:**
{
  "room": "room@conference.example.com",
  "body": "Hello, room!",
  "subject": "Room Topic" // optional
}

**Response:**
{
  "success": true,
  "message": "MUC message sent successfully",
  "data": {
    "room": "room@conference.example.com",
    "subject": "Room Topic",
    "body_length": 14,
    "sent_at": "2023-12-01T12:00:00Z",
    "request_id": "abc123"
  }
}

### Send File
**POST /api/v1/send-file**
Upload and send a file to a user via XMPP using XEP-0066 (Out-of-Band Data).

**Request (multipart/form-data):**
- to (required): Recipient JID (e.g., user@example.com)
- description (optional): Description of the file
- file (required): The file to upload

**Response:**
{
  "success": true,
  "message": "File sent successfully",
  "data": {
    "to": "user@example.com",
    "description": "Project documentation",
    "file": {
      "name": "document.pdf",
      "size": 1048576,
      "type": "application/pdf",
      "url": "http://localhost:8080/files/document_1700000000000.pdf",
      "uploaded_at": "2023-12-01T12:00:00Z"
    },
    "sent_at": "2023-12-01T12:00:00Z",
    "request_id": "abc123"
  }
}

**Configuration:**
- file_transfer.max_size: Maximum file size (default: 10 MB)
- file_transfer.storage_path: Directory for temporary file storage
- file_transfer.base_url: Base URL for public file access (optional)

**Supported Formats:** Any file type. MIME type is auto-detected from file extension.

### Get Status
**GET /api/v1/status**
Get the current status of the bot.

**Response:**
{
  "xmpp_connected": true,
  "api_running": true,
  "webhook_url": "https://example.com/webhook",
  "version": "1.0.0"
}

### Webhook Status
**GET /api/v1/webhook/status**
Get webhook service status and statistics.

**Response:**
{
  "running": true,
  "healthy": true,
  "queue_length": 0,
  "webhook_url": "https://example.com/webhook",
  "total_sent": 150,
  "total_failed": 2,
  "last_sent": "2023-12-01T11:45:00Z",
  "last_failure": "2023-12-01T10:30:00Z",
  "last_error": "Connection timeout"
}

### Health Check
**GET /api/v1/health**
Simple health check endpoint.

**Response:**
{
  "status": "ok",
  "timestamp": "2023-12-01T12:00:00Z"
}

## Error Responses
All error responses follow this format:
{
  "success": false,
  "error": "Error message",
  "code": 400
}

## Headers
- X-Request-ID: Unique request identifier for tracking
- Content-Type: application/json for all API responses

## Authentication
Currently, the API does not require authentication. In production, consider implementing:
- API keys (X-API-Key header)
- OAuth2 tokens
- Rate limiting

## Webhook Integration
The bot can forward incoming XMPP messages to configured webhook URLs. See the OpenAPI specification for the webhook payload format.

## Rate Limiting
Not implemented in the current version. Consider implementing rate limiting for production use.

## Testing
Use the provided OpenAPI specification files to test the API with tools like:
- Swagger UI
- Postman (import OpenAPI spec)
- curl examples from the specification
`

	c.Type("text/plain; charset=utf-8")
	return c.SendString(docs)
}

// handleOpenAPIYAML handles GET /openapi.yaml
func (s *Server) handleOpenAPIYAML(c *fiber.Ctx) error {
	c.Type("application/x-yaml")
	return c.SendFile("./docs/openapi.yaml")
}

// handleOpenAPIJSON handles GET /openapi.json
func (s *Server) handleOpenAPIJSON(c *fiber.Ctx) error {
	c.Type("application/json")
	return c.SendFile("./docs/openapi.json")
}

// validateSendMessageRequest validates send message request
func (s *Server) validateSendMessageRequest(req *models.SendMessageRequest) error {
	if strings.TrimSpace(req.To) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "to field is required")
	}

	if strings.TrimSpace(req.Body) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "body field is required")
	}

	if len(req.Body) > 10000 {
		return fiber.NewError(fiber.StatusBadRequest, "body field too long (max 10000 characters)")
	}

	// Basic JID validation
	if !strings.Contains(req.To, "@") {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JID format")
	}

	return nil
}

// validateSendMUCMessageRequest validates send MUC message request
func (s *Server) validateSendMUCMessageRequest(req *models.SendMUCMessageRequest) error {
	if strings.TrimSpace(req.Room) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "room field is required")
	}

	if strings.TrimSpace(req.Body) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "body field is required")
	}

	if len(req.Body) > 10000 {
		return fiber.NewError(fiber.StatusBadRequest, "body field too long (max 10000 characters)")
	}

	// Basic room JID validation
	if !strings.Contains(req.Room, "@") {
		return fiber.NewError(fiber.StatusBadRequest, "invalid room JID format")
	}

	return nil
}

var validChatStates = map[string]bool{
	"active":    true,
	"composing": true,
	"paused":    true,
	"inactive":  true,
	"gone":      true,
}

func (s *Server) validateSendChatStateRequest(req *models.SendChatStateRequest) error {
	if strings.TrimSpace(req.To) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "to field is required")
	}

	if !strings.Contains(req.To, "@") {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JID format")
	}

	if strings.TrimSpace(req.State) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "state field is required")
	}

	if !validChatStates[req.State] {
		return fiber.NewError(fiber.StatusBadRequest, "invalid state. Must be one of: active, composing, paused, inactive, gone")
	}

	return nil
}

// handleSendFile handles POST /api/v1/send-file
func (s *Server) handleSendFile(c *fiber.Ctx) error {
	logger := c.Locals("logger").(*zap.Logger)
	manager := c.Locals("manager").(XMPPManagerInterface)

	// Get fields
	to := c.FormValue("to")
	description := c.FormValue("description")
	file, err := c.FormFile("file")
	if err != nil {
		logger.Warn("File not provided",
			zap.Error(err),
			zap.String("request_id", c.GetRespHeader("X-Request-ID")),
		)
		return fiber.NewError(fiber.StatusBadRequest, "File is required")
	}

	// Validate recipient
	if strings.TrimSpace(to) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "to field is required")
	}

	if !strings.Contains(to, "@") {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JID format")
	}

	// Check file size limit
	if file.Size > s.config.FileTransfer.MaxSize {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("File too large. Maximum size is %d bytes", s.config.FileTransfer.MaxSize))
	}

	// Ensure storage directory exists
	storagePath := s.config.FileTransfer.StoragePath
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		logger.Error("Failed to create storage directory",
			zap.Error(err),
			zap.String("storage_path", storagePath),
		)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to prepare file storage")
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	baseName := strings.TrimSuffix(filepath.Base(file.Filename), ext)
	uniqueName := fmt.Sprintf("%s_%d%s", baseName, time.Now().UnixNano(), ext)
	destPath := filepath.Join(storagePath, uniqueName)

	// Save the file
	src, err := file.Open()
	if err != nil {
		logger.Error("Failed to open uploaded file",
			zap.Error(err),
			zap.String("request_id", c.GetRespHeader("X-Request-ID")),
		)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to open uploaded file")
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		logger.Error("Failed to create destination file",
			zap.Error(err),
			zap.String("path", destPath),
		)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save file")
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		logger.Error("Failed to save file",
			zap.Error(err),
			zap.String("path", destPath),
		)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save file")
	}

	// Determine file type (MIME)
	fileType := file.Header.Get("Content-Type")
	if fileType == "" {
		// Try to detect from extension
		fileType = detectMimeType(ext)
	}

	// Check if we should use XEP-0363 (HTTP File Upload)
	useXEP0363 := s.config.FileTransfer.UseXEP0363

	var fileURL string

	if useXEP0363 {
		logger.Info("Uploading file via XEP-0363 HTTP File Upload",
			zap.String("to", to),
			zap.String("filename", file.Filename),
			zap.Int64("size", file.Size),
			zap.String("type", fileType),
			zap.String("request_id", c.GetRespHeader("X-Request-ID")),
		)

		// Send file via XEP-0363 (HTTP upload through XMPP server)
		err = manager.SendFileXEP0363(to, destPath, file.Filename, fileType)
		if err != nil {
			logger.Error("Failed to send file via XEP-0363",
				zap.Error(err),
				zap.String("to", to),
				zap.String("file", destPath),
				zap.String("request_id", c.GetRespHeader("X-Request-ID")),
			)

			// Clean up file on failure
			os.Remove(destPath)

			response := models.ErrorResponse{
				Success: false,
				Error:   "Failed to upload file: " + err.Error(),
				Code:    fiber.StatusInternalServerError,
			}

			return c.Status(fiber.StatusInternalServerError).JSON(response)
		}

		// Delete the local file after successful upload
		if err := os.Remove(destPath); err != nil {
			logger.Warn("Failed to delete local file after upload",
				zap.Error(err),
				zap.String("path", destPath),
			)
		} else {
			logger.Info("Local file deleted after successful upload",
				zap.String("path", destPath),
			)
		}

		fileURL = fmt.Sprintf("via XEP-0363")
	} else {
		// Build file URL if BaseURL is configured (XEP-0066 OOB)
		if s.config.FileTransfer.BaseURL != "" {
			fileURL = fmt.Sprintf("%s/%s", strings.TrimRight(s.config.FileTransfer.BaseURL, "/"), uniqueName)
		}

		logger.Info("File uploaded and saved",
			zap.String("to", to),
			zap.String("filename", file.Filename),
			zap.String("saved_as", uniqueName),
			zap.Int64("size", file.Size),
			zap.String("type", fileType),
			zap.String("url", fileURL),
			zap.String("request_id", c.GetRespHeader("X-Request-ID")),
		)

		// Send file via XMPP manager (XEP-0066 OOB)
		err = manager.SendFile(to, fileURL, file.Filename, fileType)
		if err != nil {
			logger.Error("Failed to send file via XMPP",
				zap.Error(err),
				zap.String("to", to),
				zap.String("file", destPath),
				zap.String("request_id", c.GetRespHeader("X-Request-ID")),
			)

			// Clean up file on failure
			os.Remove(destPath)

			response := models.ErrorResponse{
				Success: false,
				Error:   "Failed to send file: " + err.Error(),
				Code:    fiber.StatusInternalServerError,
			}

			return c.Status(fiber.StatusInternalServerError).JSON(response)
		}
	}

	// Build response
	fileInfo := models.FileInfo{
		Name:       file.Filename,
		Size:       file.Size,
		Type:       fileType,
		Path:       destPath,
		URL:        fileURL,
		UploadedAt: time.Now().UTC().Format(time.RFC3339),
	}

	response := models.APIResponse{
		Success: true,
		Message: "File sent successfully",
		Data: map[string]interface{}{
			"to":            to,
			"description":   description,
			"file":          fileInfo,
			"method":        map[bool]string{true: "XEP-0363 HTTP Upload", false: "XEP-0066 OOB"}[useXEP0363],
			"local_deleted": useXEP0363,
			"sent_at":       time.Now().UTC().Format(time.RFC3339),
			"request_id":    c.GetRespHeader("X-Request-ID"),
		},
	}

	return c.JSON(response)
}

// detectMimeType detects MIME type from file extension
func detectMimeType(ext string) string {
	switch strings.ToLower(ext) {
	case ".txt":
		return "text/plain"
	case ".pdf":
		return "application/pdf"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".mp3":
		return "audio/mpeg"
	case ".mp4":
		return "video/mp4"
	case ".zip":
		return "application/zip"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	default:
		return "application/octet-stream"
	}
}
