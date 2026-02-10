package webhook

import (
	"context"
	"fmt"

	"jabber-bot/internal/config"
	"jabber-bot/internal/models"

	"go.uber.org/zap"
)

// XMPPManagerInterface defines the interface for XMPP manager operations
type XMPPManagerInterface interface {
	GetWebhookChannel() <-chan models.Message
}

// Manager manages webhook service integration with XMPP manager
type Manager struct {
	config         *config.Config
	logger         *zap.Logger
	webhookService *Service
	xmppManager    XMPPManagerInterface
}

// NewManager creates new webhook manager
func NewManager(cfg *config.Config, logger *zap.Logger, xmppManager XMPPManagerInterface) *Manager {
	return &Manager{
		config:         cfg,
		logger:         logger,
		webhookService: NewService(cfg, logger),
		xmppManager:    xmppManager,
	}
}

// Start starts webhook manager
func (m *Manager) Start(ctx context.Context) error {
	m.logger.Info("Starting webhook manager")

	// Start webhook service
	if err := m.webhookService.Start(); err != nil {
		return fmt.Errorf("failed to start webhook service: %w", err)
	}

	// Start message processor
	go m.processXMPPMessages(ctx)

	m.logger.Info("Webhook manager started successfully")
	return nil
}

// Stop stops webhook manager
func (m *Manager) Stop() error {
	m.logger.Info("Stopping webhook manager")

	// Stop webhook service
	if err := m.webhookService.Stop(); err != nil {
		m.logger.Error("Error stopping webhook service", zap.Error(err))
		return err
	}

	m.logger.Info("Webhook manager stopped")
	return nil
}

// GetService returns webhook service
func (m *Manager) GetService() *Service {
	return m.webhookService
}

// processXMPPMessages processes messages from XMPP manager
func (m *Manager) processXMPPMessages(ctx context.Context) {
	m.logger.Info("Starting XMPP message processor for webhooks")
	defer m.logger.Info("XMPP message processor stopped")

	messageChan := m.xmppManager.GetWebhookChannel()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-messageChan:
			if !ok {
				return
			}
			m.handleIncomingMessage(msg)
		}
	}
}

// handleIncomingMessage handles incoming XMPP message
func (m *Manager) handleIncomingMessage(msg models.Message) {
	m.logger.Info("Processing incoming message for webhook",
		zap.String("from", msg.From),
		zap.String("to", msg.To),
		zap.String("type", msg.Type),
	)

	// Send to webhook service
	err := m.webhookService.SendMessage(msg)
	if err != nil {
		m.logger.Error("Failed to send message to webhook service",
			zap.Error(err),
			zap.String("from", msg.From),
			zap.String("to", msg.To),
		)
	}
}

// GetStatus returns webhook manager status
func (m *Manager) GetStatus() map[string]interface{} {
	stats := m.webhookService.GetStats()

	return map[string]interface{}{
		"running":      m.webhookService.isRunning(),
		"healthy":      m.webhookService.IsHealthy(),
		"queue_length": m.webhookService.GetQueueLength(),
		"webhook_url":  m.config.Webhook.URL,
		"total_sent":   stats.TotalSent,
		"total_failed": stats.TotalFailed,
		"last_sent":    stats.LastSent,
		"last_failure": stats.LastFailure,
		"last_error":   stats.LastError,
	}
}
