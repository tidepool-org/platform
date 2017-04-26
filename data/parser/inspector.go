package parser

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/service"
)

type ObjectParserInspector struct {
	parser data.ObjectParser
}

func NewObjectParserInspector(parser data.ObjectParser) (*ObjectParserInspector, error) {
	if parser == nil {
		return nil, errors.New("parser", "parser is missing")
	}

	return &ObjectParserInspector{
		parser: parser,
	}, nil
}

func (o *ObjectParserInspector) GetProperty(key string) *string {
	return o.parser.ParseString(key)
}

func (o *ObjectParserInspector) NewMissingPropertyError(key string) error {
	o.parser.AppendError(key, service.ErrorValueNotExists())
	return nil
}

func (o *ObjectParserInspector) NewInvalidPropertyError(key string, value string, allowedValues []string) error {
	o.parser.AppendError(key, service.ErrorValueStringNotOneOf(value, allowedValues))
	return nil
}
