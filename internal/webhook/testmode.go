package webhook

import (
	"net/url"
	"strings"
)

// TestModeUtils provides utilities for handling n8n test mode
type TestModeUtils struct {
	testModeSuffix string
}

// NewTestModeUtils creates new test mode utilities
func NewTestModeUtils(suffix string) *TestModeUtils {
	if suffix == "" {
		suffix = "-test"
	}
	return &TestModeUtils{
		testModeSuffix: suffix,
	}
}

// isTestMessage checks if message body starts with [test] prefix
func (t *TestModeUtils) isTestMessage(body string) bool {
	return strings.HasPrefix(strings.TrimSpace(body), "[test]")
}

// removeTestPrefix removes [test] prefix and following spaces from message body
func (t *TestModeUtils) removeTestPrefix(body string) string {
	trimmed := strings.TrimSpace(body)
	if strings.HasPrefix(trimmed, "[test]") {
		// Remove [test] and trim leading spaces
		withoutPrefix := strings.TrimPrefix(trimmed, "[test]")
		return strings.TrimSpace(withoutPrefix)
	}
	return body
}

// getTestWebhookURL modifies webhook URL for test mode
// Rules:
// - If URL contains "webhook" but not "webhook-test", add "-test" suffix
// - Otherwise return original URL
func (t *TestModeUtils) getTestWebhookURL(baseURL string) string {
	if baseURL == "" {
		return baseURL
	}

	// Parse URL to safely modify path
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		// If URL parsing fails, do simple string replacement
		return t.simpleURLModification(baseURL)
	}

	// Check if URL path contains "webhook" but not "webhook-test"
	path := parsedURL.Path
	if strings.Contains(path, "webhook") && !strings.Contains(path, "webhook-test") {
		// Replace "webhook" with "webhook-test"
		newPath := strings.Replace(path, "webhook", "webhook"+t.testModeSuffix, 1)
		parsedURL.Path = newPath
		return parsedURL.String()
	}

	return baseURL
}

// simpleURLModification provides fallback for URL modification
func (t *TestModeUtils) simpleURLModification(baseURL string) string {
	if strings.Contains(baseURL, "webhook") && !strings.Contains(baseURL, "webhook-test") {
		return strings.Replace(baseURL, "webhook", "webhook"+t.testModeSuffix, 1)
	}
	return baseURL
}

// ProcessTestMessage processes message for test mode
// Returns (processedBody, webhookURL, isTestMode)
func (t *TestModeUtils) ProcessTestMessage(body, webhookURL string) (string, string, bool) {
	if !t.isTestMessage(body) {
		return body, webhookURL, false
	}

	// Remove test prefix from message
	processedBody := t.removeTestPrefix(body)

	// Get test webhook URL
	testURL := t.getTestWebhookURL(webhookURL)

	return processedBody, testURL, true
}
