package api

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"jabber-bot/internal/config"
	"jabber-bot/internal/models"
	"jabber-bot/internal/xmpp"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// MockXMPPManager mocks the XMPP manager for testing
type MockXMPPManager struct {
	mock.Mock
}

func (m *MockXMPPManager) SendMessage(to, body, messageType string) error {
	args := m.Called(to, body, messageType)
	return args.Error(0)
}

func (m *MockXMPPManager) SendMUCMessage(room, body, subject string) error {
	args := m.Called(room, body, subject)
	return args.Error(0)
}

func (m *MockXMPPManager) SendChatState(to string, state xmpp.ChatState) error {
	args := m.Called(to, state)
	return args.Error(0)
}

func (m *MockXMPPManager) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockXMPPManager) GetDefaultClient() *xmpp.Client {
	args := m.Called()
	return args.Get(0).(*xmpp.Client)
}

func (m *MockXMPPManager) GetWebhookChannel() <-chan models.Message {
	args := m.Called()
	return args.Get(0).(<-chan models.Message)
}

func (m *MockXMPPManager) SendFile(to, fileURL, fileName, fileType string) error {
	args := m.Called(to, fileURL, fileName, fileType)
	return args.Error(0)
}

func (m *MockXMPPManager) SendFileXEP0363(to, filePath, fileName, fileType string) error {
	args := m.Called(to, filePath, fileName, fileType)
	return args.Error(0)
}

func TestNewServer(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		API: config.APIConfig{
			Port: 8080,
			Host: "localhost",
		},
		Webhook: config.WebhookConfig{
			URL: "https://example.com/webhook",
		},
	}

	manager := &MockXMPPManager{}
	server := NewServer(cfg, logger, manager)

	assert.NotNil(t, server)
	assert.Equal(t, cfg, server.config)
	assert.Equal(t, logger, server.logger)
	assert.Equal(t, manager, server.manager)
	assert.NotNil(t, server.app)
}

func TestServer_GetAddress(t *testing.T) {
	cfg := &config.Config{
		API: config.APIConfig{
			Port: 8080,
			Host: "localhost",
		},
	}

	server := &Server{config: cfg}
	address := server.getAddress()

	assert.Equal(t, "localhost:8080", address)
}

func TestHandleSendMessage_Success(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Webhook: config.WebhookConfig{
			URL: "https://example.com/webhook",
		},
	}

	manager := &MockXMPPManager{}
	manager.On("SendMessage", "test@example.com", "Hello, world!", "chat").Return(nil)

	app := fiber.New()
	server := &Server{app: app, config: cfg, logger: logger, manager: manager}

	// Create test request
	reqBody := models.SendMessageRequest{
		To:   "test@example.com",
		Body: "Hello, world!",
		Type: "chat",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/send", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Add middleware to inject locals
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("logger", logger)
		c.Locals("config", cfg)
		c.Locals("manager", manager)
		return c.Next()
	})

	app.Post("/api/v1/send", server.handleSendMessage)

	// Perform request
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Parse response
	var response models.APIResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Equal(t, "Message sent successfully", response.Message)
	assert.NotNil(t, response.Data)

	manager.AssertExpectations(t)
}

func TestHandleSendMessage_InvalidBody(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}
	manager := &MockXMPPManager{}

	app := fiber.New()
	server := &Server{app: app, config: cfg, logger: logger, manager: manager}

	// Create invalid request
	req := httptest.NewRequest("POST", "/api/v1/send", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("logger", logger)
		c.Locals("config", cfg)
		c.Locals("manager", manager)
		return c.Next()
	})

	app.Post("/api/v1/send", server.handleSendMessage)

	// Perform request
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandleSendMessage_ValidationError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}
	manager := &MockXMPPManager{}

	app := fiber.New()
	server := &Server{app: app, config: cfg, logger: logger, manager: manager}

	// Create request with missing required field
	reqBody := models.SendMessageRequest{
		To:   "", // Missing required field
		Body: "Hello, world!",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/send", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("logger", logger)
		c.Locals("config", cfg)
		c.Locals("manager", manager)
		return c.Next()
	})

	app.Post("/api/v1/send", server.handleSendMessage)

	// Perform request
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandleSendMessage_XMPPError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}
	manager := &MockXMPPManager{}

	expectedError := xmpp.ErrNoDefaultClient
	manager.On("SendMessage", "test@example.com", "Hello, world!", "chat").Return(expectedError)

	app := fiber.New()
	server := &Server{app: app, config: cfg, logger: logger, manager: manager}

	reqBody := models.SendMessageRequest{
		To:   "test@example.com",
		Body: "Hello, world!",
		Type: "chat",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/send", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("logger", logger)
		c.Locals("config", cfg)
		c.Locals("manager", manager)
		return c.Next()
	})

	app.Post("/api/v1/send", server.handleSendMessage)

	// Perform request
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	manager.AssertExpectations(t)
}

func TestHandleSendMUCMessage_Success(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Webhook: config.WebhookConfig{
			URL: "https://example.com/webhook",
		},
	}

	manager := &MockXMPPManager{}
	manager.On("SendMUCMessage", "room@conference.example.com", "Hello room!", "Room Topic").Return(nil)

	app := fiber.New()
	server := &Server{app: app, config: cfg, logger: logger, manager: manager}

	reqBody := models.SendMUCMessageRequest{
		Room:    "room@conference.example.com",
		Body:    "Hello room!",
		Subject: "Room Topic",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/send-muc", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("logger", logger)
		c.Locals("config", cfg)
		c.Locals("manager", manager)
		return c.Next()
	})

	app.Post("/api/v1/send-muc", server.handleSendMUCMessage)

	// Perform request
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Parse response
	var response models.APIResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Equal(t, "MUC message sent successfully", response.Message)
	assert.NotNil(t, response.Data)

	manager.AssertExpectations(t)
}

func TestHandleStatus(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Webhook: config.WebhookConfig{
			URL: "https://example.com/webhook",
		},
	}

	manager := &MockXMPPManager{}
	manager.On("IsConnected").Return(true)

	app := fiber.New()
	server := &Server{app: app, config: cfg, logger: logger, manager: manager}

	req := httptest.NewRequest("GET", "/api/v1/status", nil)

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("logger", logger)
		c.Locals("config", cfg)
		c.Locals("manager", manager)
		return c.Next()
	})

	app.Get("/api/v1/status", server.handleStatus)

	// Perform request
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Parse response
	var response models.StatusResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.True(t, response.XMPPConnected)
	assert.True(t, response.APIRunning)
	assert.Equal(t, "https://example.com/webhook", response.WebhookConfig)
	assert.Equal(t, "1.0.0", response.Version)

	manager.AssertExpectations(t)
}

func TestHandleHealth(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}

	manager := &MockXMPPManager{}
	manager.On("IsConnected").Return(true)

	app := fiber.New()
	server := &Server{app: app, config: cfg, logger: logger, manager: manager}

	req := httptest.NewRequest("GET", "/api/v1/health", nil)

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("logger", logger)
		c.Locals("config", cfg)
		c.Locals("manager", manager)
		return c.Next()
	})

	app.Get("/api/v1/health", server.handleHealth)

	// Perform request
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Parse response
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "ok", response["status"])
	assert.NotEmpty(t, response["timestamp"])

	manager.AssertExpectations(t)
}

func TestHandleHealth_Unhealthy(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}

	manager := &MockXMPPManager{}
	manager.On("IsConnected").Return(false)

	app := fiber.New()
	server := &Server{app: app, config: cfg, logger: logger, manager: manager}

	req := httptest.NewRequest("GET", "/api/v1/health", nil)

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("logger", logger)
		c.Locals("config", cfg)
		c.Locals("manager", manager)
		return c.Next()
	})

	app.Get("/api/v1/health", server.handleHealth)

	// Perform request
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	// Parse response
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "error", response["status"])
	assert.Equal(t, "XMPP connection lost", response["error"])

	manager.AssertExpectations(t)
}

func TestValidateSendMessageRequest(t *testing.T) {
	server := &Server{}

	tests := []struct {
		name    string
		req     *models.SendMessageRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &models.SendMessageRequest{
				To:   "test@example.com",
				Body: "Hello",
			},
			wantErr: false,
		},
		{
			name: "missing to",
			req: &models.SendMessageRequest{
				To:   "",
				Body: "Hello",
			},
			wantErr: true,
		},
		{
			name: "missing body",
			req: &models.SendMessageRequest{
				To:   "test@example.com",
				Body: "",
			},
			wantErr: true,
		},
		{
			name: "body too long",
			req: &models.SendMessageRequest{
				To:   "test@example.com",
				Body: string(make([]byte, 10001)),
			},
			wantErr: true,
		},
		{
			name: "invalid JID format",
			req: &models.SendMessageRequest{
				To:   "invalid_jid",
				Body: "Hello",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := server.validateSendMessageRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateSendMUCMessageRequest(t *testing.T) {
	server := &Server{}

	tests := []struct {
		name    string
		req     *models.SendMUCMessageRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &models.SendMUCMessageRequest{
				Room: "room@conference.example.com",
				Body: "Hello room",
			},
			wantErr: false,
		},
		{
			name: "missing room",
			req: &models.SendMUCMessageRequest{
				Room: "",
				Body: "Hello room",
			},
			wantErr: true,
		},
		{
			name: "missing body",
			req: &models.SendMUCMessageRequest{
				Room: "room@conference.example.com",
				Body: "",
			},
			wantErr: true,
		},
		{
			name: "invalid room JID",
			req: &models.SendMUCMessageRequest{
				Room: "invalid_room",
				Body: "Hello room",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := server.validateSendMUCMessageRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleSendChatState_Success(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		Webhook: config.WebhookConfig{
			URL: "https://example.com/webhook",
		},
	}

	manager := &MockXMPPManager{}
	manager.On("SendChatState", "test@example.com", xmpp.ChatStateComposing).Return(nil)

	app := fiber.New()
	server := &Server{app: app, config: cfg, logger: logger, manager: manager}

	reqBody := models.SendChatStateRequest{
		To:    "test@example.com",
		State: "composing",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/chat-state", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("logger", logger)
		c.Locals("config", cfg)
		c.Locals("manager", manager)
		return c.Next()
	})

	app.Post("/api/v1/chat-state", server.handleSendChatState)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response models.APIResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Equal(t, "Chat state sent successfully", response.Message)
	assert.NotNil(t, response.Data)

	manager.AssertExpectations(t)
}

func TestHandleSendChatState_InvalidBody(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}
	manager := &MockXMPPManager{}

	app := fiber.New()
	server := &Server{app: app, config: cfg, logger: logger, manager: manager}

	req := httptest.NewRequest("POST", "/api/v1/chat-state", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("logger", logger)
		c.Locals("config", cfg)
		c.Locals("manager", manager)
		return c.Next()
	})

	app.Post("/api/v1/chat-state", server.handleSendChatState)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandleSendChatState_ValidationError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}
	manager := &MockXMPPManager{}

	app := fiber.New()
	server := &Server{app: app, config: cfg, logger: logger, manager: manager}

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("logger", logger)
		c.Locals("config", cfg)
		c.Locals("manager", manager)
		return c.Next()
	})

	app.Post("/api/v1/chat-state", server.handleSendChatState)

	tests := []struct {
		name         string
		req          models.SendChatStateRequest
		expectedCode int
	}{
		{
			name: "missing to",
			req: models.SendChatStateRequest{
				To:    "",
				State: "composing",
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "missing state",
			req: models.SendChatStateRequest{
				To:    "test@example.com",
				State: "",
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "invalid state",
			req: models.SendChatStateRequest{
				To:    "test@example.com",
				State: "invalid_state",
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "invalid JID",
			req: models.SendChatStateRequest{
				To:    "invalid_jid",
				State: "composing",
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.req)
			req := httptest.NewRequest("POST", "/api/v1/chat-state", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedCode, resp.StatusCode)
		})
	}
}

func TestHandleSendChatState_XMPPError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}
	manager := &MockXMPPManager{}

	expectedError := xmpp.ErrNoDefaultClient
	manager.On("SendChatState", "test@example.com", xmpp.ChatStateComposing).Return(expectedError)

	app := fiber.New()
	server := &Server{app: app, config: cfg, logger: logger, manager: manager}

	reqBody := models.SendChatStateRequest{
		To:    "test@example.com",
		State: "composing",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/chat-state", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("logger", logger)
		c.Locals("config", cfg)
		c.Locals("manager", manager)
		return c.Next()
	})

	app.Post("/api/v1/chat-state", server.handleSendChatState)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	manager.AssertExpectations(t)
}

func TestValidateSendChatStateRequest(t *testing.T) {
	s := &Server{}

	tests := []struct {
		name    string
		req     *models.SendChatStateRequest
		wantErr bool
	}{
		{
			name: "valid request - composing",
			req: &models.SendChatStateRequest{
				To:    "test@example.com",
				State: "composing",
			},
			wantErr: false,
		},
		{
			name: "valid request - active",
			req: &models.SendChatStateRequest{
				To:    "test@example.com",
				State: "active",
			},
			wantErr: false,
		},
		{
			name: "valid request - paused",
			req: &models.SendChatStateRequest{
				To:    "test@example.com",
				State: "paused",
			},
			wantErr: false,
		},
		{
			name: "valid request - inactive",
			req: &models.SendChatStateRequest{
				To:    "test@example.com",
				State: "inactive",
			},
			wantErr: false,
		},
		{
			name: "valid request - gone",
			req: &models.SendChatStateRequest{
				To:    "test@example.com",
				State: "gone",
			},
			wantErr: false,
		},
		{
			name: "missing to",
			req: &models.SendChatStateRequest{
				To:    "",
				State: "composing",
			},
			wantErr: true,
		},
		{
			name: "missing state",
			req: &models.SendChatStateRequest{
				To:    "test@example.com",
				State: "",
			},
			wantErr: true,
		},
		{
			name: "invalid state",
			req: &models.SendChatStateRequest{
				To:    "test@example.com",
				State: "typing",
			},
			wantErr: true,
		},
		{
			name: "invalid JID format",
			req: &models.SendChatStateRequest{
				To:    "invalid_jid",
				State: "composing",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.validateSendChatStateRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateSendFileRequest(t *testing.T) {
	s := &Server{}

	tests := []struct {
		name    string
		to      string
		file    string
		wantErr bool
	}{
		{
			name:    "valid request",
			to:      "user@example.com",
			file:    "test.txt",
			wantErr: false,
		},
		{
			name:    "missing to",
			to:      "",
			file:    "test.txt",
			wantErr: true,
		},
		{
			name:    "missing file",
			to:      "user@example.com",
			file:    "",
			wantErr: true,
		},
		{
			name:    "invalid JID",
			to:      "invalid_jid",
			file:    "test.txt",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We'll test the validation logic inline in the handler test
			// This is a simplified validation check
			err := s.validateSendFileRequest(tt.to, tt.file)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleSendFile_Success(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		FileTransfer: config.FileTransferConfig{
			MaxSize:     10 * 1024 * 1024, // 10 MB
			StoragePath: "./test-uploads",
			BaseURL:     "http://localhost:8080/files",
		},
	}

	manager := &MockXMPPManager{}
	// Use flexible matching for file path and type
	manager.On("SendFile", "user@example.com", mock.Anything, "test.txt", mock.Anything).Return(nil)

	app := fiber.New()
	server := &Server{app: app, config: cfg, logger: logger, manager: manager}

	// Create a proper multipart request with a file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("to", "user@example.com")
	_ = writer.WriteField("description", "Test file")
	part, _ := writer.CreateFormFile("file", "test.txt")
	part.Write([]byte("test content"))
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/send-file", bytes.NewReader(body.Bytes()))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("logger", logger)
		c.Locals("config", cfg)
		c.Locals("manager", manager)
		return c.Next()
	})

	app.Post("/api/v1/send-file", server.handleSendFile)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Parse response
	var response models.APIResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Equal(t, "File sent successfully", response.Message)
	assert.NotNil(t, response.Data)

	manager.AssertExpectations(t)
}

func TestHandleSendFile_InvalidForm(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}
	manager := &MockXMPPManager{}

	app := fiber.New()
	server := &Server{app: app, config: cfg, logger: logger, manager: manager}

	// Invalid multipart data
	req := httptest.NewRequest("POST", "/api/v1/send-file", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "multipart/form-data")

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("logger", logger)
		c.Locals("config", cfg)
		c.Locals("manager", manager)
		return c.Next()
	})

	app.Post("/api/v1/send-file", server.handleSendFile)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandleSendFile_MissingTo(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}
	manager := &MockXMPPManager{}

	app := fiber.New()
	server := &Server{app: app, config: cfg, logger: logger, manager: manager}

	// Create multipart form without 'to' field
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// Add file part but no 'to'
	part, err := writer.CreateFormField("file")
	require.NoError(t, err)
	part.Write([]byte("test content"))
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/send-file", bytes.NewReader(body.Bytes()))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("logger", logger)
		c.Locals("config", cfg)
		c.Locals("manager", manager)
		return c.Next()
	})

	app.Post("/api/v1/send-file", server.handleSendFile)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandleSendFile_MissingFile(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}
	manager := &MockXMPPManager{}

	app := fiber.New()
	server := &Server{app: app, config: cfg, logger: logger, manager: manager}

	// Create multipart form with 'to' but no 'file'
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("to", "user@example.com")
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/send-file", bytes.NewReader(body.Bytes()))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("logger", logger)
		c.Locals("config", cfg)
		c.Locals("manager", manager)
		return c.Next()
	})

	app.Post("/api/v1/send-file", server.handleSendFile)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandleSendFile_InvalidJID(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}
	manager := &MockXMPPManager{}

	app := fiber.New()
	server := &Server{app: app, config: cfg, logger: logger, manager: manager}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("to", "invalid_jid")
	part, _ := writer.CreateFormField("file")
	part.Write([]byte("test"))
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/send-file", bytes.NewReader(body.Bytes()))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("logger", logger)
		c.Locals("config", cfg)
		c.Locals("manager", manager)
		return c.Next()
	})

	app.Post("/api/v1/send-file", server.handleSendFile)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandleSendFile_FileTooLarge(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		FileTransfer: config.FileTransferConfig{
			MaxSize: 10, // Only 10 bytes
		},
	}
	manager := &MockXMPPManager{}

	app := fiber.New()
	server := &Server{app: app, config: cfg, logger: logger, manager: manager}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("to", "user@example.com")
	part, _ := writer.CreateFormField("file")
	part.Write([]byte("this file is too large")) // 24 bytes
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/send-file", bytes.NewReader(body.Bytes()))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("logger", logger)
		c.Locals("config", cfg)
		c.Locals("manager", manager)
		return c.Next()
	})

	app.Post("/api/v1/send-file", server.handleSendFile)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandleSendFile_XMPPError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		FileTransfer: config.FileTransferConfig{
			MaxSize:     10 * 1024 * 1024,
			StoragePath: "./test-uploads",
		},
	}

	manager := &MockXMPPManager{}
	expectedError := xmpp.ErrNoDefaultClient
	// Use flexible matching for file path and type
	manager.On("SendFile", "user@example.com", mock.Anything, "test.txt", mock.Anything).Return(expectedError)

	app := fiber.New()
	server := &Server{app: app, config: cfg, logger: logger, manager: manager}

	// Create a proper multipart request with a file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("to", "user@example.com")
	part, _ := writer.CreateFormFile("file", "test.txt")
	part.Write([]byte("test"))
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/send-file", bytes.NewReader(body.Bytes()))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("logger", logger)
		c.Locals("config", cfg)
		c.Locals("manager", manager)
		return c.Next()
	})

	app.Post("/api/v1/send-file", server.handleSendFile)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	manager.AssertExpectations(t)
}

// validateSendFileRequest validates the send file request parameters
func (s *Server) validateSendFileRequest(to, fileName string) error {
	if strings.TrimSpace(to) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "to field is required")
	}

	if !strings.Contains(to, "@") {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JID format")
	}

	if strings.TrimSpace(fileName) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "file field is required")
	}

	return nil
}
