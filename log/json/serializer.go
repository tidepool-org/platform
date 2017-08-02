package json

import (
	"encoding/json"
	"io"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
)

func NewSerializer(writer io.Writer) (log.Serializer, error) {
	if writer == nil {
		return nil, errors.New("json", "writer is missing")
	}

	return &serializer{
		writer: writer,
	}, nil
}

type serializer struct {
	writer io.Writer
}

func (s *serializer) Serialize(fields log.Fields) error {
	if fields == nil {
		return errors.New("json", "fields are missing")
	}

	bytes, err := json.Marshal(fields)
	if err != nil {
		return errors.Wrapf(err, "json", "unable to serialize fields")
	}

	bytes = append(bytes, []byte("\n")...)

	_, err = s.writer.Write(bytes)
	if err != nil {
		return errors.Wrapf(err, "json", "unable to write serialized field")
	}

	return nil
}
