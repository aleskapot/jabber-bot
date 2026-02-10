package api

import (
	"errors"
	"fmt"
	"jabber-bot/internal/config"
	"jabber-bot/internal/models"
	"jabber-bot/internal/xmpp"
	"net"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"go.uber.org/zap"
)

// XMPPManagerInterface defines the interface for XMPP manager operations
type XMPPManagerInterface interface {
	SendMessage(to, body, messageType string) error
	SendMUCMessage(room, body, subject string) error
	IsConnected() bool
	GetDefaultClient() *xmpp.Client
	GetWebhookChannel() <-chan models.Message
}

// Server represents the API server
type Server struct {
	app        *fiber.App
	config     *config.Config
	logger     *zap.Logger
	manager    XMPPManagerInterface
	actualPort int
}

// NewServer creates new API server
func NewServer(cfg *config.Config, logger *zap.Logger, manager XMPPManagerInterface) *Server {
	app := fiber.New(fiber.Config{
		ErrorHandler: errorHandler,
	})

	server := &Server{
		app:     app,
		config:  cfg,
		logger:  logger,
		manager: manager,
	}

	server.setupMiddleware()
	server.setupRoutes()

	return server
}

// setupMiddleware configures Fiber middleware
func (s *Server) setupMiddleware() {
	// Add Request ID
	s.app.Use(requestid.New())

	// Recovery middleware
	s.app.Use(recover.New())

	// Request logger
	s.app.Use(logger.New(logger.Config{
		Format:     "${time} | ${status} | ${latency} | ${method} | ${path} | ${ip} | ${error}\n",
		TimeFormat: "2006-01-02 15:04:05",
	}))

	// Custom middleware to inject logger and config
	s.app.Use(func(c *fiber.Ctx) error {
		c.Locals("logger", s.logger)
		c.Locals("config", s.config)
		c.Locals("manager", s.manager)
		return c.Next()
	})
}

// setupRoutes configures API routes
func (s *Server) setupRoutes() {
	api := s.app.Group("/api/v1")

	// Health endpoint (public - no auth required)
	api.Get("/health", s.handleHealth)

	// Apply authentication middleware to protected endpoints
	if s.config.API.Enabled {
		api.Use(s.AuthMiddleware())
	}

	// Message endpoints (protected)
	api.Post("/send", s.handleSendMessage)
	api.Post("/send-muc", s.handleSendMUCMessage)

	// Status endpoints (protected)
	api.Get("/status", s.handleStatus)
	api.Get("/webhook/status", s.handleWebhookStatus)

	// Documentation (public)
	s.app.Get("/", s.handleRoot)
	s.app.Get("/docs", s.handleDocs)
	s.app.Get("/openapi.yaml", s.handleOpenAPIYAML)
	s.app.Get("/openapi.json", s.handleOpenAPIJSON)
}

// Start starts the API server
func (s *Server) Start() error {
	// Use net.Listen to get actual port
	listener, err := net.Listen("tcp", s.getAddress())
	if err != nil {
		return err
	}

	// Get actual port
	if s.config.API.Port == 0 {
		s.actualPort = listener.Addr().(*net.TCPAddr).Port
	} else {
		s.actualPort = s.config.API.Port
	}

	s.logger.Info("Starting API server",
		zap.Int("port", s.actualPort),
	)

	// Start server with listener
	return s.app.Listener(listener)
}

// Stop stops the API server
func (s *Server) Stop() error {
	s.logger.Info("Stopping API server")
	return s.app.Shutdown()
}

// getAddress returns the server address
func (s *Server) getAddress() string {
	return s.config.API.Host + ":" + fmt.Sprintf("%d", s.config.API.Port)
}

// GetPort returns the actual port the server is listening on
func (s *Server) GetPort() int {
	if s.actualPort != 0 {
		return s.actualPort
	}
	return s.config.API.Port
}

// errorHandler custom error handler for Fiber
func errorHandler(c *fiber.Ctx, err error) error {
	log := c.Locals("logger").(*zap.Logger)

	// Default error response
	response := models.ErrorResponse{
		Success: false,
		Error:   "Internal server error",
		Code:    fiber.StatusInternalServerError,
	}

	// Type assertion to get Fiber error (handles wrapped errors)
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		response.Code = fiberErr.Code
		response.Error = fiberErr.Message
	} else {
		// Log unexpected errors
		log.Error("API error",
			zap.Error(err),
			zap.String("path", c.Path()),
			zap.String("method", c.Method()),
		)
	}

	// Check for validation errors
	if err.Error() == "validation failed" {
		response.Code = fiber.StatusBadRequest
		response.Error = "Request validation failed"
	}

	// Return JSON response
	return c.Status(response.Code).JSON(response)
}
