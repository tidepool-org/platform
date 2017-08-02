package null

import (
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
)

func NewSerializer() log.Serializer {
	return &serializer{}
}

type serializer struct{}

func (s *serializer) Serialize(fields log.Fields) error {
	if fields == nil {
		return errors.New("null", "fields are missing")
	}

	return nil
}
