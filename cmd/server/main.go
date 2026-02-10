package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"jabber-bot/internal/api"
	"jabber-bot/internal/config"
	"jabber-bot/internal/webhook"
	"jabber-bot/internal/xmpp"
	"jabber-bot/pkg/logger"

	"go.uber.org/zap"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration first
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger with config
	zapLogger, err := logger.NewWithConfig(cfg.Logging.Level, cfg.Logging.Output, cfg.Logging.FilePath)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer zapLogger.Sync()

	zapLogger.Info("Starting jabber bot")

	zapLogger.Info("Configuration loaded successfully",
		zap.String("xmpp_jid", cfg.XMPP.JID),
		zap.Int("api_port", cfg.API.Port),
		zap.String("webhook_url", cfg.Webhook.URL),
	)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize XMPP manager
	xmppManager := xmpp.NewManager(cfg, zapLogger)

	// Start XMPP manager
	if err := xmppManager.Start(); err != nil {
		zapLogger.Fatal("Failed to start XMPP manager", zap.Error(err))
	}

	// Initialize webhook manager
	webhookManager := webhook.NewManager(cfg, zapLogger, xmppManager)

	// Start webhook manager
	if err := webhookManager.Start(ctx); err != nil {
		zapLogger.Fatal("Failed to start webhook manager", zap.Error(err))
	}

	// Initialize API server
	apiServer := api.NewServer(cfg, zapLogger, xmppManager)

	// Start API server in goroutine
	go func() {
		if err := apiServer.Start(); err != nil {
			zapLogger.Fatal("Failed to start API server", zap.Error(err))
		}
	}()

	zapLogger.Info("Jabber bot started successfully")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	zapLogger.Info("Shutting down...")

	// Stop API server
	if err := apiServer.Stop(); err != nil {
		zapLogger.Error("Error stopping API server", zap.Error(err))
	}

	// Stop webhook manager
	if err := webhookManager.Stop(); err != nil {
		zapLogger.Error("Error stopping webhook manager", zap.Error(err))
	}

	// Stop XMPP manager
	if err := xmppManager.Stop(); err != nil {
		zapLogger.Error("Error stopping XMPP manager", zap.Error(err))
	}

	zapLogger.Info("Application stopped")
}
