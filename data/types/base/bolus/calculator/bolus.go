package calculator

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/base/bolus/combination"
	"github.com/tidepool-org/platform/data/types/base/bolus/extended"
	"github.com/tidepool-org/platform/data/types/base/bolus/normal"
)

func ParseBolus(parser data.ObjectParser) data.Datum {
	subType := parser.ParseString("subType")
	if subType == nil {
		return nil
	}

	var datum data.Datum
	var err error

	switch *subType {
	case normal.SubType():
		datum, err = normal.New()
	case extended.SubType():
		datum, err = extended.New()
	case combination.SubType():
		datum, err = combination.New()
	default:
		return nil
	}

	if err != nil {
		return nil // TODO_DATA: Do something with error here
	}
	if datum != nil {
		datum.Parse(parser)
	}
	return datum
}
