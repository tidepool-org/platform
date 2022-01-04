package factory

import (
	"github.com/tidepool-org/platform/data"
	dataTypesActivityPhysical "github.com/tidepool-org/platform/data/types/activity/physical"
	dataTypesBasal "github.com/tidepool-org/platform/data/types/basal"
	dataTypesBasalFactory "github.com/tidepool-org/platform/data/types/basal/factory"
	dataTypesBloodGlucoseContinuous "github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	dataTypesBloodGlucoseSelfmonitored "github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
	dataTypesBloodKetone "github.com/tidepool-org/platform/data/types/blood/ketone"
	dataTypesBolus "github.com/tidepool-org/platform/data/types/bolus"
	dataTypesBolusFactory "github.com/tidepool-org/platform/data/types/bolus/factory"
	dataTypesCalculator "github.com/tidepool-org/platform/data/types/calculator"
	dataTypesDevice "github.com/tidepool-org/platform/data/types/device"
	dataTypesDeviceFactory "github.com/tidepool-org/platform/data/types/device/factory"
	dataTypesFood "github.com/tidepool-org/platform/data/types/food"
	dataTypesInsulin "github.com/tidepool-org/platform/data/types/insulin"
	dataTypesSettingsBasalSecurity "github.com/tidepool-org/platform/data/types/settings/basalsecurity"
	dataTypesSettingsCGM "github.com/tidepool-org/platform/data/types/settings/cgm"
	dataTypesSettingsPump "github.com/tidepool-org/platform/data/types/settings/pump"
	dataTypesStateReported "github.com/tidepool-org/platform/data/types/state/reported"
	dataTypesUpload "github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var types = []string{
	dataTypesActivityPhysical.Type,
	dataTypesBasal.Type,
	dataTypesSettingsBasalSecurity.Type,
	dataTypesBloodGlucoseContinuous.Type,
	dataTypesBloodGlucoseSelfmonitored.Type,
	dataTypesBloodKetone.Type,
	dataTypesBolus.Type,
	dataTypesCalculator.Type,
	dataTypesDevice.Type,
	dataTypesFood.Type,
	dataTypesInsulin.Type,
	dataTypesSettingsCGM.Type,
	dataTypesSettingsPump.Type,
	dataTypesStateReported.Type,
	dataTypesUpload.Type,
}

func NewDatum(parser structure.ObjectParser) data.Datum {
	if !parser.Exists() {
		return nil
	}

	value := parser.String("type")
	if value == nil {
		parser.WithReferenceErrorReporter("type").ReportError(structureValidator.ErrorValueNotExists())
		return nil
	}

	switch *value {
	case dataTypesActivityPhysical.Type:
		return dataTypesActivityPhysical.New()
	case dataTypesBasal.Type:
		return dataTypesBasalFactory.NewBasalDatum(parser)
	case dataTypesBloodGlucoseContinuous.Type:
		return dataTypesBloodGlucoseContinuous.New()
	case dataTypesBloodGlucoseSelfmonitored.Type:
		return dataTypesBloodGlucoseSelfmonitored.New()
	case dataTypesBloodKetone.Type:
		return dataTypesBloodKetone.New()
	case dataTypesBolus.Type:
		return dataTypesBolusFactory.NewBolusDatum(parser)
	case dataTypesCalculator.Type:
		return dataTypesCalculator.New()
	case dataTypesDevice.Type:
		return dataTypesDeviceFactory.NewDeviceDatum(parser)
	case dataTypesFood.Type:
		return dataTypesFood.New()
	case dataTypesInsulin.Type:
		return dataTypesInsulin.New()
	case dataTypesSettingsBasalSecurity.Type:
		return dataTypesSettingsBasalSecurity.New()
	case dataTypesSettingsCGM.Type:
		return dataTypesSettingsCGM.New()
	case dataTypesSettingsPump.Type:
		return dataTypesSettingsPump.New()
	case dataTypesStateReported.Type:
		return dataTypesStateReported.New()
	case dataTypesUpload.Type:
		return dataTypesUpload.New()
	}

	parser.WithReferenceErrorReporter("type").ReportError(structureValidator.ErrorValueStringNotOneOf(*value, types))
	return nil
}

func ParseDatum(parser structure.ObjectParser) *data.Datum {
	datum := NewDatum(parser)
	if datum == nil {
		return nil
	}

	datum.Parse(parser)
	return &datum
}
