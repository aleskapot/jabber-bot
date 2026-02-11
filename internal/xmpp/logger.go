package xmpp

import (
	"fmt"

	"go.uber.org/zap"
)

// xmppLoggerAdapter adapts zap logger to XMPP logger interface and io.Writer
type xmppLoggerAdapter struct {
	logger *zap.Logger
}

func (l *xmppLoggerAdapter) Printf(format string, v ...interface{}) {
	l.logger.Debug(fmt.Sprintf(format, v...))
}

func (l *xmppLoggerAdapter) Println(v ...interface{}) {
	l.logger.Debug(fmt.Sprint(v...))
}

func (l *xmppLoggerAdapter) Write(p []byte) (n int, err error) {
	l.logger.Debug("XMPP stream", zap.String("data", string(p)))
	return len(p), nil
}
