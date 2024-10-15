package json

import (
	"io"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/logger"
)

// CONCURRENCY: SAFE IFF writer is safe

func NewLogger(writer io.Writer, levelRanks log.LevelRanks, level log.Level) (log.Logger, error) {
	serializer, err := NewSerializer(writer)
	if err != nil {
		return nil, err
	}

	return logger.New(serializer, levelRanks, level)
}
