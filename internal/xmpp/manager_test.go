package xmpp

import (
	"fmt"
	"testing"
	"time"

	"jabber-bot/internal/config"
	"jabber-bot/internal/models"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestManager_MergeChannels(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}
	manager := NewManager(cfg, logger)

	// Create test channels
	ch1 := make(chan models.Message, 10)
	ch2 := make(chan models.Message, 10)

	// Send test messages
	ch1 <- models.Message{From: "sender1", Body: "message1"}
	ch2 <- models.Message{From: "sender2", Body: "message2"}

	// Merge channels
	merged := manager.mergeChannels(ch1, ch2)
	defer close(ch1)
	defer close(ch2)

	// Read merged messages (order may vary)
	var messages []models.Message
	for i := 0; i < 2; i++ {
		select {
		case msg := <-merged:
			messages = append(messages, msg)
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for merged message")
		}
	}

	// Verify we got both messages
	assert.Len(t, messages, 2)

	messageFrom1 := messages[0].From == "sender1" || messages[1].From == "sender1"
	messageFrom2 := messages[0].From == "sender2" || messages[1].From == "sender2"
	assert.True(t, messageFrom1, "Should have message from sender1")
	assert.True(t, messageFrom2, "Should have message from sender2")
}

func TestManager_WebhookChannel_ThreadSafety(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}
	manager := NewManager(cfg, logger)

	// Test concurrent access to webhook channel
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			msg := models.Message{
				From: fmt.Sprintf("sender%d", id),
				Body: fmt.Sprintf("message%d", id),
			}

			select {
			case manager.webhookChan <- msg:
				// Message sent
			case <-time.After(100 * time.Millisecond):
				t.Errorf("Timeout sending message %d", id)
			}

			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all messages were received
	assert.Len(t, manager.webhookChan, 10)

	// Clear channel
	for len(manager.webhookChan) > 0 {
		<-manager.webhookChan
	}
}
