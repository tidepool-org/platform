package status

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

type TypeStatusArray []TypeStatusInterface

type TypeStatusInterface interface {
	statusObject()
}

func (t *TypeStatusArray) ParseArray(parser structure.ArrayParser) {
}

func NewStatusArray() *TypeStatusArray {
	return &TypeStatusArray{}
}

func ParseStatusArray(parser structure.ObjectParser) *TypeStatusArray {
	if !parser.Exists() {
		return nil
	}
	datum := NewParseStatusArray()
	parser.Parse(datum)
	return datum
}

func NewParseStatusArray() *TypeStatusArray {
	return &TypeStatusArray{}
}

func (t *TypeStatusArray) Parse(parser structure.ObjectParser) {
	for _, reference := range parser.References() {
		*t = append(*t, ParseStatus(parser.WithReferenceObjectParser(reference)))
	}
}

func (t *TypeStatusArray) Validate(validator structure.Validator) {
}

func (t *TypeStatusArray) Normalize(normalizer data.Normalizer) {}

func ParseStatus(parser structure.ObjectParser) TypeStatusInterface {
	if !parser.Exists() {
		return nil
	}
	datum := ParseBatteryStruct(parser.WithReferenceObjectParser("battery"))
	if datum == nil {
		if datum = ParseAlertsStruct(parser.WithReferenceObjectParser("alerts")); datum == nil {
			if datum = ParseReservoirRemainingStruct(parser.WithReferenceObjectParser("reservoirRemaining")); datum == nil {
				if datum = ParseSignalStrengthStruct(parser.WithReferenceObjectParser("signalStrength")); datum == nil {
					if datum = ParseForecastStruct(parser.WithReferenceObjectParser("forecast")); datum == nil {
						return nil
					}
				}
			}
		}
	}
	return datum
}
