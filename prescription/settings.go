package prescription

import (
	"github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type InitialSettings struct {
	BloodGlucoseUnits          string                             `json:"bloodGlucoseUnits,omitempty" bson:"bloodGlucoseUnits,omitempty"`
	BasalRateSchedule          *pump.BasalRateStartArray          `json:"basalRateSchedule,omitempty" bson:"basalRateSchedule,omitempty"`
	BloodGlucoseTargetSchedule *pump.BloodGlucoseTargetStartArray `json:"bloodGlucoseTargetSchedule,omitempty" bson:"bloodGlucoseTargetSchedule,omitempty"`
	CarbohydrateRatioSchedule  *pump.CarbohydrateRatioStartArray  `json:"carbohydrateRatioSchedule,omitempty" bson:"carbohydrateRatioSchedule,omitempty"`
	InsulinSensitivitySchedule *pump.InsulinSensitivityStartArray `json:"insulinSensitivitySchedule,omitempty" bson:"insulinSensitivitySchedule,omitempty"`
	BasalRateMaximum           *pump.BasalRateMaximum             `json:"basalRateMaximum,omitempty" bson:"basalRateMaximum,omitempty"`
	BolusAmountMaximum         *pump.BolusAmountMaximum           `json:"bolusAmountMaximum,omitempty" bson:"bolusAmountMaximum,omitempty"`
	PumpID                     *primitive.ObjectID                `json:"pumpId" bson:"pumpId"`
	CgmID                      *primitive.ObjectID                `json:"cgmId" bson:"cgmId"`
	// TODO: Add Suspend threshold - Dependent on latest data model changes
	// TODO: Add Insulin model - Dependent on latest data model changes
}

func (i *InitialSettings) Validate(validator structure.Validator) {
	validator.String("bloodGlucoseUnits", &i.BloodGlucoseUnits).OneOf(glucose.Units()...)
	if i.BasalRateSchedule != nil {
		i.BasalRateSchedule.Validate(validator.WithReference("basalSchedule"))
	}
	if i.BloodGlucoseTargetSchedule != nil {
		i.BloodGlucoseTargetSchedule.Validate(validator.WithReference("bloodGlucoseTargetSchedule"), &i.BloodGlucoseUnits)
	}
	if i.CarbohydrateRatioSchedule != nil {
		i.CarbohydrateRatioSchedule.Validate(validator.WithReference("carbohydrateRatioSchedule"))
	}
	if i.InsulinSensitivitySchedule != nil {
		i.InsulinSensitivitySchedule.Validate(validator.WithReference("insulinSensitivitySchedule"), &i.BloodGlucoseUnits)
	}
	if i.BasalRateMaximum != nil {
		i.BasalRateMaximum.Validate(validator.WithReference("basalRateMaximum"))
	}
	if i.BolusAmountMaximum != nil {
		i.BolusAmountMaximum.Validate(validator.WithReference("bolusAmountMaximum"))
	}
	if i.PumpID != nil {
		id := i.PumpID.Hex()
		validator.String("pumpId", &id).Hexadecimal().LengthEqualTo(24)
	}
	if i.CgmID != nil {
		id := i.CgmID.Hex()
		validator.String("cgmId", &id).Hexadecimal().LengthEqualTo(24)
	}
}

func (i *InitialSettings) ValidateAllRequired(validator structure.Validator) {
	if i.BasalRateSchedule == nil {
		validator.WithReference("basalSchedule").ReportError(structureValidator.ErrorValueEmpty())
	}
	if i.BloodGlucoseTargetSchedule == nil {
		validator.WithReference("bloodGlucoseTargetSchedule").ReportError(structureValidator.ErrorValueEmpty())
	}
	if i.CarbohydrateRatioSchedule == nil {
		validator.WithReference("carbohydrateRatioSchedule").ReportError(structureValidator.ErrorValueEmpty())
	}
	if i.InsulinSensitivitySchedule == nil {
		validator.WithReference("insulinSensitivitySchedule").ReportError(structureValidator.ErrorValueEmpty())
	}
	if i.BasalRateMaximum == nil {
		validator.WithReference("basalRateMaximum").ReportError(structureValidator.ErrorValueEmpty())
	}
	if i.BolusAmountMaximum == nil {
		validator.WithReference("bolusAmountMaximum").ReportError(structureValidator.ErrorValueEmpty())
	}
	if i.PumpID == nil {
		validator.WithReference("pumpId").ReportError(structureValidator.ErrorValueEmpty())
	}
	if i.CgmID == nil {
		validator.WithReference("cgmId").ReportError(structureValidator.ErrorValueEmpty())
	}
	// TODO: Validate Suspend Threshold and Insulin Type
}
