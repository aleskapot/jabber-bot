package filemanager

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	"jabber-bot/internal/config"

	"go.uber.org/zap"
)

type Manager struct {
	config     *config.Config
	logger     *zap.Logger
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup
}

func NewManager(cfg *config.Config, logger *zap.Logger) *Manager {
	return &Manager{
		config: cfg,
		logger: logger,
	}
}

func (m *Manager) Start(ctx context.Context) error {
	if !m.config.FileTransfer.Enabled || m.config.FileTransfer.RetainHours <= 0 {
		m.logger.Debug("File cleanup is disabled")
		return nil
	}

	m.logger.Info("Starting file cleanup manager",
		zap.String("storage_path", m.config.FileTransfer.StoragePath),
		zap.Int("retain_hours", m.config.FileTransfer.RetainHours),
	)

	processorCtx, cancel := context.WithCancel(ctx)
	m.cancelFunc = cancel

	m.wg.Add(1)
	go m.runCleanup(processorCtx)

	return nil
}

func (m *Manager) Stop() error {
	m.logger.Info("Stopping file cleanup manager")

	if m.cancelFunc != nil {
		m.cancelFunc()
	}

	m.wg.Wait()
	m.logger.Info("File cleanup manager stopped")

	return nil
}

func (m *Manager) runCleanup(ctx context.Context) {
	defer m.wg.Done()

	storagePath := m.config.FileTransfer.StoragePath
	retainDuration := time.Duration(m.config.FileTransfer.RetainHours) * time.Hour

	m.cleanupFiles(storagePath, retainDuration)

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("File cleanup routine stopped")
			return
		case <-ticker.C:
			m.cleanupFiles(storagePath, retainDuration)
		}
	}
}

func (m *Manager) cleanupFiles(storagePath string, retainDuration time.Duration) {
	entries, err := os.ReadDir(storagePath)
	if err != nil {
		if !os.IsNotExist(err) {
			m.logger.Error("Failed to read storage directory",
				zap.String("path", storagePath),
				zap.Error(err),
			)
		}
		return
	}

	cutoffTime := time.Now().Add(-retainDuration)
	var deletedCount, errorCount int

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			m.logger.Error("Failed to get file info",
				zap.String("file", entry.Name()),
				zap.Error(err),
			)
			errorCount++
			continue
		}

		if info.ModTime().Before(cutoffTime) {
			filePath := filepath.Join(storagePath, entry.Name())
			if err := os.Remove(filePath); err != nil {
				m.logger.Error("Failed to delete old file",
					zap.String("file", filePath),
					zap.Error(err),
				)
				errorCount++
			} else {
				m.logger.Info("Deleted old file",
					zap.String("file", filePath),
					zap.Time("modified", info.ModTime()),
				)
				deletedCount++
			}
		}
	}

	if deletedCount > 0 || errorCount > 0 {
		m.logger.Info("File cleanup completed",
			zap.Int("deleted", deletedCount),
			zap.Int("errors", errorCount),
		)
	}
}
