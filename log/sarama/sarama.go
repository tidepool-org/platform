package sarama

import (
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/IBM/sarama"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
)

// CONCURRENCY: SAFE IFF logger is safe

// Sarama reads its loggers without synchronization, so they are only assigned here and Setup swaps the target
func init() {
	sarama.Logger = &logger{level: log.InfoLevel}
	sarama.DebugLogger = &logger{level: log.DebugLevel}
}

var target atomic.Pointer[log.Logger]

// Setup sends the process-wide Sarama loggers, which otherwise discard everything, to the logger
func Setup(lgr log.Logger) error {
	if lgr == nil {
		return errors.New("logger is missing")
	}

	target.Store(&lgr)
	return nil
}

type logger struct {
	level log.Level
}

func (l *logger) Print(v ...interface{}) {
	l.log(fmt.Sprint(v...))
}

func (l *logger) Printf(format string, v ...interface{}) {
	l.log(fmt.Sprintf(format, v...))
}

func (l *logger) Println(v ...interface{}) {
	l.log(fmt.Sprintln(v...))
}

// Sarama terminates most messages with a newline
func (l *logger) log(message string) {
	if lgr := target.Load(); lgr != nil {
		(*lgr).Log(l.level, strings.TrimSpace(message))
	}
}
