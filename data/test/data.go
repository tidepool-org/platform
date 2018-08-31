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

func RandomSelectorOrigin() *data.SelectorOrigin {
	datum := data.NewSelectorOrigin()
	datum.ID = pointer.FromString(test.RandomStringFromRangeAndCharset(1, 100, test.CharsetText))
	return datum
}

func CloneSelectorOrigin(datum *data.SelectorOrigin) *data.SelectorOrigin {
	if datum == nil {
		return nil
	}
	clone := data.NewSelectorOrigin()
	clone.ID = pointer.CloneString(datum.ID)
	return clone
}

func RandomSelector() *data.Selector {
	datum := data.NewSelector()
	datum.ID = pointer.FromString(RandomID())
	datum.Origin = RandomSelectorOrigin()
	return datum
}

func CloneSelector(datum *data.Selector) *data.Selector {
	if datum == nil {
		return nil
	}
	clone := data.NewSelector()
	clone.ID = pointer.CloneString(datum.ID)
	clone.Origin = CloneSelectorOrigin(datum.Origin)
	return clone
}

func RandomSelectors() *data.Selectors {
	datum := data.NewSelectors()
	for index := test.RandomIntFromRange(1, 3); index > 0; index-- {
		*datum = append(*datum, RandomSelector())
	}
	return datum
}

func CloneSelectors(datum *data.Selectors) *data.Selectors {
	if datum == nil {
		return nil
	}
	clone := data.NewSelectors()
	for _, d := range *datum {
		*clone = append(*clone, CloneSelector(d))
	}
	return clone
}
