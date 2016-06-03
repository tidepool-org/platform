package calculator

import (
	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/base"
	"github.com/tidepool-org/platform/data/types/base/bolus"
	"github.com/tidepool-org/platform/data/types/base/bolus/combination"
	"github.com/tidepool-org/platform/data/types/base/bolus/extended"
	"github.com/tidepool-org/platform/data/types/base/bolus/normal"
)

func ParseBolus(parser data.ObjectParser) (data.Datum, error) {
	if parser.Object() == nil {
		return nil, nil
	}

	datumType := parser.ParseString("type")
	if datumType == nil {
		parser.AppendError("type", base.ErrorValueMissing())
		return nil, nil
	}

	if *datumType != bolus.Type() {
		parser.AppendError("type", base.ErrorTypeInvalid(*datumType))
		return nil, nil
	}

	subType := parser.ParseString("subType")
	if subType == nil {
		parser.AppendError("subType", base.ErrorValueMissing())
		return nil, nil
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
		parser.AppendError("subType", base.ErrorSubTypeInvalid(*subType))
		return nil, nil
	}

	if err != nil {
		return nil, err
	}
	if datum == nil {
		return nil, app.Error("calculator", "datum is missing")
	}

	if err := datum.Parse(parser); err != nil {
		return nil, err
	}

	parser.ProcessNotParsed()

	return datum, nil
}
