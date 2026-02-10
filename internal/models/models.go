package models

// Message represents an XMPP message
type Message struct {
	ID      string `json:"id"`
	From    string `json:"from"`
	To      string `json:"to"`
	Body    string `json:"body"`
	Type    string `json:"type"`
	Subject string `json:"subject"`
	Thread  string `json:"thread"`
	Stamp   string `json:"stamp"`
}

// SendMessageRequest represents API request to send a message
type SendMessageRequest struct {
	To   string `json:"to" validate:"required"`
	Body string `json:"body" validate:"required"`
	Type string `json:"type,omitempty"`
}

// SendMUCMessageRequest represents API request to send a message to MUC
type SendMUCMessageRequest struct {
	Room    string `json:"room" validate:"required"`
	Body    string `json:"body" validate:"required"`
	Subject string `json:"subject,omitempty"`
}

// WebhookPayload represents payload sent to webhook endpoint
type WebhookPayload struct {
	Message   Message `json:"message"`
	Timestamp string  `json:"timestamp"`
	Source    string  `json:"source"`
}

// StatusResponse represents API response with status information
type StatusResponse struct {
	XMPPConnected bool   `json:"xmpp_connected"`
	APIRunning    bool   `json:"api_running"`
	WebhookConfig string `json:"webhook_url"`
	Version       string `json:"version"`
}

// APIResponse represents standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    int    `json:"code"`
}
