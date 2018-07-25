package factory

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/bolus"
	dataTypesBolusCombination "github.com/tidepool-org/platform/data/types/bolus/combination"
	dataTypesBolusExtended "github.com/tidepool-org/platform/data/types/bolus/extended"
	dataTypesBolusNormal "github.com/tidepool-org/platform/data/types/bolus/normal"
	"github.com/tidepool-org/platform/service"
)

var subTypes = []string{
	dataTypesBolusCombination.SubType,
	dataTypesBolusExtended.SubType,
	dataTypesBolusNormal.SubType,
}

func NewBolusDatum(parser data.ObjectParser) data.Datum {
	if parser.Object() == nil {
		return nil
	}

	if value := parser.ParseString("type"); value == nil {
		parser.AppendError("type", service.ErrorValueNotExists())
		return nil
	} else if *value != bolus.Type {
		parser.AppendError("type", service.ErrorValueStringNotOneOf(*value, []string{bolus.Type}))
		return nil
	}

	value := parser.ParseString("subType")
	if value == nil {
		parser.AppendError("subType", service.ErrorValueNotExists())
		return nil
	}

	switch *value {
	case dataTypesBolusCombination.SubType:
		return dataTypesBolusCombination.New()
	case dataTypesBolusExtended.SubType:
		return dataTypesBolusExtended.New()
	case dataTypesBolusNormal.SubType:
		return dataTypesBolusNormal.New()
	}

	parser.AppendError("subType", service.ErrorValueStringNotOneOf(*value, subTypes))
	return nil
}

func ParseBolusDatum(parser data.ObjectParser) *data.Datum {
	datum := NewBolusDatum(parser)
	if datum == nil {
		return nil
	}

	datum.Parse(parser)
	return &datum
}
