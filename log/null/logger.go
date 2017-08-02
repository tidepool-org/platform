package null

import "github.com/tidepool-org/platform/log"

func NewLogger() log.Logger {
	logger, _ := log.NewLogger(NewSerializer(), log.DefaultLevels(), log.DefaultLevel()) // Safely ignore error; cannot fail
	return logger
}
