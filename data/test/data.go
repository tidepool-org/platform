package test

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomID() string {
	return data.NewID()
}

func NewSessionToken() string {
	return test.NewString(256, test.CharsetAlphaNumeric)
}

func NewDeviceID() string {
	return test.NewString(32, test.CharsetText)
}

func RandomDeleteOrigin() *data.DeleteOrigin {
	datum := data.NewDeleteOrigin()
	datum.ID = pointer.FromString(test.RandomStringFromRangeAndCharset(1, 100, test.CharsetText))
	return datum
}

func CloneDeleteOrigin(datum *data.DeleteOrigin) *data.DeleteOrigin {
	if datum == nil {
		return nil
	}
	clone := data.NewDeleteOrigin()
	clone.ID = pointer.CloneString(datum.ID)
	return clone
}

func RandomDelete() *data.Delete {
	datum := data.NewDelete()
	datum.ID = pointer.FromString(RandomID())
	datum.Origin = RandomDeleteOrigin()
	return datum
}

func CloneDelete(datum *data.Delete) *data.Delete {
	if datum == nil {
		return nil
	}
	clone := data.NewDelete()
	clone.ID = pointer.CloneString(datum.ID)
	clone.Origin = CloneDeleteOrigin(datum.Origin)
	return clone
}

func RandomDeletes() *data.Deletes {
	datum := data.NewDeletes()
	for index := test.RandomIntFromRange(1, 3); index > 0; index-- {
		*datum = append(*datum, RandomDelete())
	}
	return datum
}

func CloneDeletes(datum *data.Deletes) *data.Deletes {
	if datum == nil {
		return nil
	}
	clone := data.NewDeletes()
	for _, d := range *datum {
		*clone = append(*clone, CloneDelete(d))
	}
	return clone
}
