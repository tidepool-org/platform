package factory

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/bolus"
	dataTypesBolusCombination "github.com/tidepool-org/platform/data/types/bolus/combination"
	dataTypesBolusExtended "github.com/tidepool-org/platform/data/types/bolus/extended"
	dataTypesBolusNormal "github.com/tidepool-org/platform/data/types/bolus/normal"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var subTypes = []string{
	dataTypesBolusCombination.SubType,
	dataTypesBolusExtended.SubType,
	dataTypesBolusNormal.SubType,
}

func NewBolusDatum(parser structure.ObjectParser) data.Datum {
	if !parser.Exists() {
		return nil
	}

	if value := parser.String("type"); value == nil {
		parser.WithReferenceErrorReporter("type").ReportError(structureValidator.ErrorValueNotExists())
		return nil
	} else if *value != bolus.Type {
		parser.WithReferenceErrorReporter("type").ReportError(structureValidator.ErrorValueStringNotOneOf(*value, []string{bolus.Type}))
		return nil
	}

	value := parser.String("subType")
	if value == nil {
		parser.WithReferenceErrorReporter("subType").ReportError(structureValidator.ErrorValueNotExists())
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

	parser.WithReferenceErrorReporter("subType").ReportError(structureValidator.ErrorValueStringNotOneOf(*value, subTypes))
	return nil
}

func ParseBolusDatum(parser structure.ObjectParser) *data.Datum {
	datum := NewBolusDatum(parser)
	if datum == nil {
		return nil
	}

	datum.Parse(parser)
	return &datum
}
