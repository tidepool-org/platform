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
	dataTypesDosingDecision "github.com/tidepool-org/platform/data/types/dosingdecision"
	dataTypesFood "github.com/tidepool-org/platform/data/types/food"
	dataTypesInsulin "github.com/tidepool-org/platform/data/types/insulin"
	dataTypesPumpStatus "github.com/tidepool-org/platform/data/types/pumpstatus"
	dataTypesSettingsApplication "github.com/tidepool-org/platform/data/types/settings/application"
	dataTypesSettingsCGM "github.com/tidepool-org/platform/data/types/settings/cgm"
	dataTypesSettingsPump "github.com/tidepool-org/platform/data/types/settings/pump"
	dataTypesStateReported "github.com/tidepool-org/platform/data/types/state/reported"
	dataTypesUpload "github.com/tidepool-org/platform/data/types/upload"
	dataTypesWater "github.com/tidepool-org/platform/data/types/water"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var types = []string{
	dataTypesActivityPhysical.Type,
	dataTypesBasal.Type,
	dataTypesBloodGlucoseContinuous.Type,
	dataTypesBloodGlucoseSelfmonitored.Type,
	dataTypesBloodKetone.Type,
	dataTypesBolus.Type,
	dataTypesCalculator.Type,
	dataTypesDevice.Type,
	dataTypesDosingDecision.Type,
	dataTypesFood.Type,
	dataTypesInsulin.Type,
	dataTypesPumpStatus.Type,
	dataTypesSettingsApplication.Type,
	dataTypesSettingsCGM.Type,
	dataTypesSettingsPump.Type,
	dataTypesStateReported.Type,
	dataTypesUpload.Type,
	dataTypesWater.Type,
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
	case dataTypesDosingDecision.Type:
		return dataTypesDosingDecision.New()
	case dataTypesFood.Type:
		return dataTypesFood.New()
	case dataTypesInsulin.Type:
		return dataTypesInsulin.New()
	case dataTypesPumpStatus.Type:
		return dataTypesPumpStatus.New()
	case dataTypesSettingsApplication.Type:
		return dataTypesSettingsApplication.New()
	case dataTypesSettingsCGM.Type:
		return dataTypesSettingsCGM.New()
	case dataTypesSettingsPump.Type:
		return dataTypesSettingsPump.New()
	case dataTypesStateReported.Type:
		return dataTypesStateReported.New()
	case dataTypesUpload.Type:
		return dataTypesUpload.New()
	case dataTypesWater.Type:
		return dataTypesWater.New()
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
