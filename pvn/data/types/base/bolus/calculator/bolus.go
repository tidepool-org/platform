package calculator

import (
	"github.com/tidepool-org/platform/pvn/data"
	"github.com/tidepool-org/platform/pvn/data/types/base/bolus/combination"
	"github.com/tidepool-org/platform/pvn/data/types/base/bolus/extended"
	"github.com/tidepool-org/platform/pvn/data/types/base/bolus/normal"
)

func ParseBolus(parser data.ObjectParser) data.Datum {

	subType := parser.ParseString("subType")
	if subType == nil {
		return nil
	}

	var datum data.Datum

	switch *subType {
	case normal.SubType():
		datum = normal.New()
		datum.Parse(parser)
		return datum
	case extended.SubType():
		datum = extended.New()
		datum.Parse(parser)
		return datum
	case combination.SubType():
		datum = combination.New()
		datum.Parse(parser)
		return datum
	default:
		return nil
	}
}
