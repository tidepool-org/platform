package test

import (
	"time"

	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/common/origin"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewOrigin() *origin.Origin {
	datum := origin.NewOrigin()
	datum.ID = pointer.FromString(test.NewText(1, 100))
	datum.Name = pointer.FromString(test.NewText(1, 100))
	datum.Payload = dataTest.NewBlob()
	datum.Time = pointer.FromString(test.RandomTime().Format(time.RFC3339Nano))
	datum.Type = pointer.FromString(test.RandomStringFromArray(origin.Types()))
	datum.Version = pointer.FromString(test.NewText(1, 100))
	return datum
}

func CloneOrigin(datum *origin.Origin) *origin.Origin {
	if datum == nil {
		return nil
	}
	clone := origin.NewOrigin()
	clone.ID = test.CloneString(datum.ID)
	clone.Name = test.CloneString(datum.Name)
	clone.Payload = dataTest.CloneBlob(datum.Payload)
	clone.Time = test.CloneString(datum.Time)
	clone.Type = test.CloneString(datum.Type)
	clone.Version = test.CloneString(datum.Version)
	return clone
}
