package test

import (
	"github.com/tidepool-org/platform/data/types/bolus/biphasic"
	dataTypesBolusNormalTest "github.com/tidepool-org/platform/data/types/bolus/normal/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewBiphasic() *biphasic.Biphasic {
	datum := biphasic.New()
	datum.Normal = *dataTypesBolusNormalTest.NewNormal()
	datum.Type = "bolus"
	datum.SubType = "biphasic"
	datum.Part = pointer.FromString(test.RandomStringFromArray(biphasic.Parts()))
	datum.EventID = pointer.FromString("123456789")
	datum.LinkedBolus = NewLinkedBolus()
	return datum
}

func CloneBiphasic(datum *biphasic.Biphasic) *biphasic.Biphasic {
	if datum == nil {
		return nil
	}
	clone := biphasic.New()
	clone.Normal = *dataTypesBolusNormalTest.CloneNormal(&datum.Normal)
	clone.SubType = datum.SubType
	clone.Part = pointer.CloneString(datum.Part)
	clone.EventID = pointer.CloneString(datum.EventID)
	clone.LinkedBolus = CloneLinkedBolus(datum.LinkedBolus)
	return clone
}
