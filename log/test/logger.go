package test

import (
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/logger"
)

type Logger struct {
	*Serializer
	log.Logger
}

func NewLogger() *Logger {
	serializer := NewSerializer()
	lgr, _ := logger.New(serializer, log.DefaultLevelRanks(), log.DebugLevel)
	return &Logger{
		Serializer: serializer,
		Logger:     lgr,
	}
}
