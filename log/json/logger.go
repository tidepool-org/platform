package json

import (
	"io"
	"os"

	"github.com/tidepool-org/platform/log"
)

// CONCURRENCY: SAFE IFF writer is safe

func NewLogger(writer io.Writer, levelRanks log.LevelRanks, level log.Level) (log.Logger, error) {
	serializer, err := NewSerializer(writer)
	if err != nil {
		return nil, err
	}

	return log.NewLogger(serializer, levelRanks, level)
}

// NewWithDefaults calls NewLogger with default values.
func NewWithDefaults() (log.Logger, error) {
	return NewLogger(os.Stderr, log.DefaultLevelRanks(), log.DefaultLevel())
}
