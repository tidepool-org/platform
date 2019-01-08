package factory

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/basal"
	dataTypesBasalAutomated "github.com/tidepool-org/platform/data/types/basal/automated"
	dataTypesBasalScheduled "github.com/tidepool-org/platform/data/types/basal/scheduled"
	dataTypesBasalSuspend "github.com/tidepool-org/platform/data/types/basal/suspend"
	dataTypesBasalTemporary "github.com/tidepool-org/platform/data/types/basal/temporary"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var deliveryTypes = []string{
	dataTypesBasalAutomated.DeliveryType,
	dataTypesBasalScheduled.DeliveryType,
	dataTypesBasalSuspend.DeliveryType,
	dataTypesBasalTemporary.DeliveryType,
}

func NewBasalDatum(parser structure.ObjectParser) data.Datum {
	if !parser.Exists() {
		return nil
	}

	if value := parser.String("type"); value == nil {
		parser.WithReferenceErrorReporter("type").ReportError(structureValidator.ErrorValueNotExists())
		return nil
	} else if *value != basal.Type {
		parser.WithReferenceErrorReporter("type").ReportError(structureValidator.ErrorValueStringNotOneOf(*value, []string{basal.Type}))
		return nil
	}

	value := parser.String("deliveryType")
	if value == nil {
		parser.WithReferenceErrorReporter("deliveryType").ReportError(structureValidator.ErrorValueNotExists())
		return nil
	}

	switch *value {
	case dataTypesBasalAutomated.DeliveryType:
		return dataTypesBasalAutomated.New()
	case dataTypesBasalScheduled.DeliveryType:
		return dataTypesBasalScheduled.New()
	case dataTypesBasalSuspend.DeliveryType:
		return dataTypesBasalSuspend.New()
	case dataTypesBasalTemporary.DeliveryType:
		return dataTypesBasalTemporary.New()
	}

	parser.WithReferenceErrorReporter("deliveryType").ReportError(structureValidator.ErrorValueStringNotOneOf(*value, deliveryTypes))
	return nil
}

func ParseBasalDatum(parser structure.ObjectParser) *data.Datum {
	datum := NewBasalDatum(parser)
	if datum == nil {
		return nil
	}

	datum.Parse(parser)
	return &datum
}
