package work

import (
	"maps"
	"slices"
	"time"

	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	StringTimeMapReferenceLengthMaximum = 1000
	StringTimeMapLengthMaximum          = 1000
)

type StringTimeMap map[string]*time.Time

func ParseStringTimeMap(parser structure.ObjectParser) *StringTimeMap {
	if !parser.Exists() {
		return nil
	}
	datum := &StringTimeMap{}
	datum.Parse(parser)
	return datum
}

func (s *StringTimeMap) Parse(parser structure.ObjectParser) {
	for _, reference := range parser.References() {
		if *s == nil {
			*s = map[string]*time.Time{}
		}
		(*s)[reference] = parser.Time(reference, time.RFC3339Nano)
	}
}

func (s *StringTimeMap) Validate(validator structure.Validator) {
	if length := len(*s); length > StringTimeMapLengthMaximum {
		validator.ReportError(structureValidator.ErrorLengthNotLessThanOrEqualTo(length, StringTimeMapLengthMaximum))
	}
	for _, reference := range s.SortedKeys() {
		validator.WithReference(reference).String(structure.ReferenceSelf, &reference).LengthLessThanOrEqualTo(StringTimeMapReferenceLengthMaximum)
	}
}

func (s *StringTimeMap) SortedKeys() []string {
	return slices.Sorted(maps.Keys(*s))
}
