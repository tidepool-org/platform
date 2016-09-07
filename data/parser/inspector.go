package parser

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import (
	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/service"
)

type ObjectParserInspector struct {
	parser data.ObjectParser
}

func NewObjectParserInspector(parser data.ObjectParser) (*ObjectParserInspector, error) {
	if parser == nil {
		return nil, app.Error("parser", "parser is missing")
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
