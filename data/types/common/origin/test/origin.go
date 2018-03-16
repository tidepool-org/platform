package test

import (
	"github.com/tidepool-org/platform/data/types/common/origin"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
	testInternet "github.com/tidepool-org/platform/test/internet"
)

func NewOrigin() *origin.Origin {
	datum := origin.NewOrigin()
	datum.ID = pointer.String(test.NewText(1, 100))
	datum.Name = pointer.String(testInternet.NewReverseDomain())
	datum.Time = pointer.Time(test.NewTime())
	datum.Type = pointer.String(test.RandomStringFromArray(origin.Types()))
	datum.Version = pointer.String(test.NewText(1, 100))
	return datum
}

func CloneOrigin(datum *origin.Origin) *origin.Origin {
	if datum == nil {
		return nil
	}
	clone := origin.NewOrigin()
	clone.ID = test.CloneString(datum.ID)
	clone.Name = test.CloneString(datum.Name)
	clone.Time = test.CloneTime(datum.Time)
	clone.Type = test.CloneString(datum.Type)
	clone.Version = test.CloneString(datum.Version)
	return clone
}
