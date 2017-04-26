package parser

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/errors"
)

func ParseDatum(parser data.ObjectParser, factory data.Factory) (*data.Datum, error) {
	if parser == nil {
		return nil, errors.New("parser", "parser is missing")
	}
	if factory == nil {
		return nil, errors.New("parser", "factory is missing")
	}

	if parser.Object() == nil {
		return nil, nil
	}

	inspector, err := NewObjectParserInspector(parser)
	if err != nil {
		return nil, err
	}

	datum, err := factory.Init(inspector)
	if err != nil {
		return nil, err
	}
	if datum == nil {
		return nil, nil
	}

	if err = datum.Parse(parser); err != nil {
		return nil, err
	}

	return &datum, nil
}

func ParseDatumArray(parser data.ArrayParser) (*[]data.Datum, error) {
	if parser == nil {
		return nil, errors.New("parser", "parser is missing")
	}

	array := parser.Array()
	if array == nil {
		return nil, nil
	}

	datumArray := []data.Datum{}

	for index := range *array {
		if datum := parser.ParseDatum(index); datum != nil && *datum != nil {
			datumArray = append(datumArray, *datum)
		}
	}

	return &datumArray, nil
}
