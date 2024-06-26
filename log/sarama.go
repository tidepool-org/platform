package log

import (
	"fmt"
	"strings"

	"github.com/Shopify/sarama"
)

// NewSarama returns a [Logger] adapted to implement [sarama.StdLogger].
func NewSarama(l Logger) sarama.StdLogger {
	return &SaramaLogger{Logger: l.WithField("SARAMA", "1")}
}

// SaramaLogger wraps a [Logger] to implement [sarama.StdLogger].
//
// Sarama doesn't support the concept of logging levels, so all messages will
// use the info level.
type SaramaLogger struct {
	Logger
}

func (l *SaramaLogger) Print(args ...interface{}) {
	l.Logger.Info(fmt.Sprint(args...))
}

func (l *SaramaLogger) Printf(format string, args ...interface{}) {
	// Sarama log messages sent via this method include a newline, which
	// doesn't fit with Logger's style, so remove it.
	l.Logger.Infof(strings.TrimSuffix(format, "\n"), args...)
}

func (l *SaramaLogger) Println(args ...interface{}) {
	l.Logger.Info(fmt.Sprint(args...))
}
