package bloodglucose

import (
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

const ContinuousName = "cbg"

type Continuous struct {
	Value      *float64 `json:"value" bson:"value" valid:"bloodglucosevalue"`
	Units      *string  `json:"units" bson:"units" valid:"mmolmgunits"`
	types.Base `bson:",inline"`
}

func BuildContinuous(datum types.Datum, errs validate.ErrorProcessing) *Continuous {

	continuous := &Continuous{
		Value: datum.ToFloat64(types.BloodGlucoseValueField.Name, errs),
		Units: datum.ToString(types.MmolOrMgUnitsField.Name, errs),
		Base:  types.BuildBase(datum, errs),
	}

	continuous.Units = NormalizeUnitName(continuous.Units)
	continuous.Value = ConvertMgToMmol(continuous.Value, continuous.Units)

	types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(continuous, errs)
	return continuous
}
