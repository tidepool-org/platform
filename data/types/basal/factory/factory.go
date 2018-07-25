package factory

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/basal"
	dataTypesBasalAutomated "github.com/tidepool-org/platform/data/types/basal/automated"
	dataTypesBasalScheduled "github.com/tidepool-org/platform/data/types/basal/scheduled"
	dataTypesBasalSuspend "github.com/tidepool-org/platform/data/types/basal/suspend"
	dataTypesBasalTemporary "github.com/tidepool-org/platform/data/types/basal/temporary"
	"github.com/tidepool-org/platform/service"
)

var deliveryTypes = []string{
	dataTypesBasalAutomated.DeliveryType,
	dataTypesBasalScheduled.DeliveryType,
	dataTypesBasalSuspend.DeliveryType,
	dataTypesBasalTemporary.DeliveryType,
}

func NewBasalDatum(parser data.ObjectParser) data.Datum {
	if parser.Object() == nil {
		return nil
	}

	if value := parser.ParseString("type"); value == nil {
		parser.AppendError("type", service.ErrorValueNotExists())
		return nil
	} else if *value != basal.Type {
		parser.AppendError("type", service.ErrorValueStringNotOneOf(*value, []string{basal.Type}))
		return nil
	}

	value := parser.ParseString("deliveryType")
	if value == nil {
		parser.AppendError("deliveryType", service.ErrorValueNotExists())
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

	parser.AppendError("deliveryType", service.ErrorValueStringNotOneOf(*value, deliveryTypes))
	return nil
}

func ParseBasalDatum(parser data.ObjectParser) *data.Datum {
	datum := NewBasalDatum(parser)
	if datum == nil {
		return nil
	}

	datum.Parse(parser)
	return &datum
}
