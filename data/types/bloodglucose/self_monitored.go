package bloodglucose

import (
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

const SelfMonitoredName = "smbg"

type SelfMonitored struct {
	Value      *float64 `json:"value" bson:"value" valid:"-"`
	Units      *string  `json:"units" bson:"units" valid:"-"`
	types.Base `bson:",inline"`
}

func BuildSelfMonitored(datum types.Datum, errs validate.ErrorProcessing) *SelfMonitored {

	selfMonitored := &SelfMonitored{
		Value: datum.ToFloat64(types.BloodGlucoseValueField.Name, errs),
		Units: datum.ToString(types.MmolOrMgUnitsField.Name, errs),
		Base:  types.BuildBase(datum, errs),
	}

	bgValidator := types.NewBloodGlucoseValidation(selfMonitored.Value, selfMonitored.Units)
	selfMonitored.Value, selfMonitored.Units = bgValidator.ValidateAndConvertBloodGlucoseValue(errs)

	types.GetPlatformValidator().Struct(selfMonitored, errs)

	return selfMonitored
}
