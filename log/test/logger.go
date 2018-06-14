package test

import "github.com/tidepool-org/platform/log"

type Logger struct {
	*Serializer
	log.Logger
}

func NewLogger() *Logger {
	serializer := NewSerializer()
	logger, _ := log.NewLogger(serializer, log.DefaultLevelRanks(), log.DebugLevel)
	return &Logger{
		Serializer: serializer,
		Logger:     logger,
	}
}
