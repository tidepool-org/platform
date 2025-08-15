package test

import (
	"time"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

const (
	SelectorTypeID           = "id"
	SelectorTypeDeduplicator = "deduplicator"
	SelectorTypeOrigin       = "origin"
)

func SelectorTypes() []string {
	return []string{
		SelectorTypeID,
		SelectorTypeDeduplicator,
		SelectorTypeOrigin,
	}
}

func RandomSelectorType() string {
	return test.RandomStringFromArray(SelectorTypes())
}

func RandomDatumID() string {
	return data.NewID()
}

func NewSessionToken() string {
	return test.RandomStringFromRangeAndCharset(256, 256, test.CharsetAlphaNumeric)
}

func NewDeviceID() string {
	return test.RandomStringFromRangeAndCharset(32, 32, test.CharsetText)
}

func RandomSelectorDeduplicator() *data.SelectorDeduplicator {
	datum := data.NewSelectorDeduplicator()
	datum.Hash = pointer.FromString(test.RandomStringFromRangeAndCharset(32, 32, test.CharsetHexadecimalLowercase))
	return datum
}

func CloneSelectorDeduplicator(datum *data.SelectorDeduplicator) *data.SelectorDeduplicator {
	if datum == nil {
		return nil
	}
	clone := data.NewSelectorDeduplicator()
	clone.Hash = pointer.CloneString(datum.Hash)
	return clone
}

func RandomSelectorOrigin() *data.SelectorOrigin {
	datum := data.NewSelectorOrigin()
	datum.ID = pointer.FromString(test.RandomStringFromRangeAndCharset(1, 100, test.CharsetText))
	datum.Time = pointer.FromString(test.RandomTimeBeforeNow().Format(time.RFC3339))
	return datum
}

func CloneSelectorOrigin(datum *data.SelectorOrigin) *data.SelectorOrigin {
	if datum == nil {
		return nil
	}
	clone := data.NewSelectorOrigin()
	clone.ID = pointer.CloneString(datum.ID)
	clone.Time = pointer.CloneString(datum.Time)
	return clone
}

func RandomSelector() *data.Selector {
	return RandomSelectorWithType(RandomSelectorType())
}

func RandomSelectorWithType(selectorType string) *data.Selector {
	datum := data.NewSelector()
	switch selectorType {
	case SelectorTypeID:
		datum.ID = pointer.FromString(RandomDatumID())
		datum.Time = pointer.FromTime(test.RandomTimeBeforeNow())
	case SelectorTypeDeduplicator:
		datum.Deduplicator = RandomSelectorDeduplicator()
	case SelectorTypeOrigin:
		datum.Origin = RandomSelectorOrigin()
	}
	return datum
}

func CloneSelector(datum *data.Selector) *data.Selector {
	if datum == nil {
		return nil
	}
	clone := data.NewSelector()
	clone.ID = pointer.CloneString(datum.ID)
	clone.Time = pointer.CloneTime(datum.Time)
	clone.Deduplicator = CloneSelectorDeduplicator(datum.Deduplicator)
	clone.Origin = CloneSelectorOrigin(datum.Origin)
	return clone
}

func RandomSelectors() *data.Selectors {
	return RandomSelectorsWithType(RandomSelectorType())
}

func RandomSelectorsWithType(selectorType string) *data.Selectors {
	datum := data.NewSelectors()
	for index := test.RandomIntFromRange(1, 3); index > 0; index-- {
		*datum = append(*datum, RandomSelectorWithType(selectorType))
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
