package null

import "github.com/tidepool-org/platform/log"

// CONCURRENCY: SAFE

func NewLogger() log.Logger {
	logger, _ := log.NewLogger(NewSerializer(), log.DefaultLevelRanks(), log.DefaultLevel()) // Safely ignore error; cannot fail
	return logger
}
