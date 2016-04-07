package bloodglucose

import (
	"github.com/tidepool-org/platform/data/types"

	"github.com/tidepool-org/platform/validate"
)

const SelfMonitoredName = "smbg"

type SelfMonitored struct {
	Value      *float64 `json:"value" bson:"value" valid:"bloodglucosevalue"`
	Units      *string  `json:"units" bson:"units" valid:"bloodglucoseunits"`
	types.Base `bson:",inline"`
}

func BuildSelfMonitored(datum types.Datum, errs validate.ErrorProcessing) *SelfMonitored {

	selfMonitored := &SelfMonitored{
		Value: datum.ToFloat64(valueField.Name, errs),
		Units: datum.ToString(unitsField.Name, errs),
		Base:  types.BuildBase(datum, errs),
	}

	selfMonitored.Units = normalizeUnitName(selfMonitored.Units)
	selfMonitored.Value = convertMgToMmol(selfMonitored.Value, selfMonitored.Units)

	types.GetPlatformValidator().SetErrorReasons(failureReasons).Struct(selfMonitored, errs)
	return selfMonitored
}
