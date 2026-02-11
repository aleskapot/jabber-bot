package webhook

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTestModeUtils_isTestMessage(t *testing.T) {
	tests := []struct {
		name string
		body string
		want bool
	}{
		{
			name: "valid test message with prefix",
			body: "[test] hello world",
			want: true,
		},
		{
			name: "valid test message with leading spaces",
			body: "   [test] hello world",
			want: true,
		},
		{
			name: "test message with only prefix",
			body: "[test]",
			want: true,
		},
		{
			name: "normal message",
			body: "hello world",
			want: false,
		},
		{
			name: "message with test in middle",
			body: "hello [test] world",
			want: false,
		},
		{
			name: "message with test prefix without brackets",
			body: "test hello world",
			want: false,
		},
		{
			name: "empty message",
			body: "",
			want: false,
		},
		{
			name: "case sensitive test",
			body: "[TEST] hello world",
			want: false,
		},
	}

	utils := NewTestModeUtils("")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.isTestMessage(tt.body)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTestModeUtils_removeTestPrefix(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "remove test prefix with spaces",
			body: "[test] hello world",
			want: "hello world",
		},
		{
			name: "remove test prefix with multiple spaces",
			body: "[test]    hello world",
			want: "hello world",
		},
		{
			name: "remove test prefix with leading spaces",
			body: "   [test] hello world",
			want: "hello world",
		},
		{
			name: "test prefix only",
			body: "[test]",
			want: "",
		},
		{
			name: "test prefix with only spaces",
			body: "[test]     ",
			want: "",
		},
		{
			name: "normal message unchanged",
			body: "hello world",
			want: "hello world",
		},
		{
			name: "test in middle unchanged",
			body: "hello [test] world",
			want: "hello [test] world",
		},
	}

	utils := NewTestModeUtils("")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.removeTestPrefix(tt.body)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTestModeUtils_getTestWebhookURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		suffix  string
		want    string
	}{
		{
			name:    "simple webhook URL",
			baseURL: "https://example.com/webhook",
			suffix:  "-test",
			want:    "https://example.com/webhook-test",
		},
		{
			name:    "webhook URL with path",
			baseURL: "https://example.com/webhook/path",
			suffix:  "-test",
			want:    "https://example.com/webhook-test/path",
		},
		{
			name:    "webhook URL with query params",
			baseURL: "https://example.com/webhook?param=value",
			suffix:  "-test",
			want:    "https://example.com/webhook-test?param=value",
		},
		{
			name:    "webhook URL with port and path",
			baseURL: "https://example.com:8080/webhook/api",
			suffix:  "-test",
			want:    "https://example.com:8080/webhook-test/api",
		},
		{
			name:    "already test webhook URL unchanged",
			baseURL: "https://example.com/webhook-test",
			suffix:  "-test",
			want:    "https://example.com/webhook-test",
		},
		{
			name:    "non-webhook URL unchanged",
			baseURL: "https://example.com/api/endpoint",
			suffix:  "-test",
			want:    "https://example.com/api/endpoint",
		},
		{
			name:    "custom suffix",
			baseURL: "https://example.com/webhook",
			suffix:  "-staging",
			want:    "https://example.com/webhook-staging",
		},
		{
			name:    "empty URL",
			baseURL: "",
			suffix:  "-test",
			want:    "",
		},
		{
			name:    "complex URL with multiple path segments",
			baseURL: "https://n8n.example.com/webhook/production/v1/messages",
			suffix:  "-test",
			want:    "https://n8n.example.com/webhook-test/production/v1/messages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils := NewTestModeUtils(tt.suffix)
			got := utils.getTestWebhookURL(tt.baseURL)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTestModeUtils_ProcessTestMessage(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		webhookURL string
		suffix     string
		wantBody   string
		wantURL    string
		wantTest   bool
	}{
		{
			name:       "test message processing",
			body:       "[test] hello world",
			webhookURL: "https://example.com/webhook",
			suffix:     "-test",
			wantBody:   "hello world",
			wantURL:    "https://example.com/webhook-test",
			wantTest:   true,
		},
		{
			name:       "normal message unchanged",
			body:       "hello world",
			webhookURL: "https://example.com/webhook",
			suffix:     "-test",
			wantBody:   "hello world",
			wantURL:    "https://example.com/webhook",
			wantTest:   false,
		},
		{
			name:       "test message with non-webhook URL",
			body:       "[test] hello",
			webhookURL: "https://example.com/api",
			suffix:     "-test",
			wantBody:   "hello",
			wantURL:    "https://example.com/api",
			wantTest:   true,
		},
		{
			name:       "test message with already test webhook",
			body:       "[test] debug",
			webhookURL: "https://example.com/webhook-test",
			suffix:     "-test",
			wantBody:   "debug",
			wantURL:    "https://example.com/webhook-test",
			wantTest:   true,
		},
		{
			name:       "test message with complex URL",
			body:       "[test] testing",
			webhookURL: "https://n8n.example.com/webhook/production/v1",
			suffix:     "-test",
			wantBody:   "testing",
			wantURL:    "https://n8n.example.com/webhook-test/production/v1",
			wantTest:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils := NewTestModeUtils(tt.suffix)
			gotBody, gotURL, gotTest := utils.ProcessTestMessage(tt.body, tt.webhookURL)

			assert.Equal(t, tt.wantBody, gotBody)
			assert.Equal(t, tt.wantURL, gotURL)
			assert.Equal(t, tt.wantTest, gotTest)
		})
	}
}

func TestNewTestModeUtils(t *testing.T) {
	t.Run("with custom suffix", func(t *testing.T) {
		utils := NewTestModeUtils("-staging")
		assert.Equal(t, "-staging", utils.testModeSuffix)
	})

	t.Run("with empty suffix uses default", func(t *testing.T) {
		utils := NewTestModeUtils("")
		assert.Equal(t, "-test", utils.testModeSuffix)
	})
}
