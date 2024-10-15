package null

import (
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/logger"
)

// CONCURRENCY: SAFE

func NewLogger() log.Logger {
	lgr, _ := logger.New(NewSerializer(), log.DefaultLevelRanks(), log.DefaultLevel()) // Safely ignore error; cannot fail
	return lgr
}
