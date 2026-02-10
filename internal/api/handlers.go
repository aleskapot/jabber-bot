package api

import (
	"strings"
	"time"

	"jabber-bot/internal/models"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// handleSendMessage handles POST /api/v1/send
func (s *Server) handleSendMessage(c *fiber.Ctx) error {
	//goland:noinspection DuplicatedCode
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
	return c.SendFile("./openapi.yaml")
}

// handleOpenAPIJSON handles GET /openapi.json
func (s *Server) handleOpenAPIJSON(c *fiber.Ctx) error {
	c.Type("application/json")
	return c.SendFile("./openapi.json")
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
