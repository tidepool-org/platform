package json

import (
	"io"

	"github.com/tidepool-org/platform/log"
)

func NewLogger(writer io.Writer, levels log.Levels, level log.Level) (log.Logger, error) {
	serializer, err := NewSerializer(writer)
	if err != nil {
		return nil, err
	}

	return log.NewLogger(serializer, levels, level)
}
