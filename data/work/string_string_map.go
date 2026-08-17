package work

import (
	"maps"
	"slices"

	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	StringStringMapReferenceLengthMaximum = 1000
	StringStringMapValueLengthMaximum     = 1000
	StringStringMapLengthMaximum          = 1000
)

type StringStringMap map[string]*string

func ParseStringStringMap(parser structure.ObjectParser) *StringStringMap {
	if !parser.Exists() {
		return nil
	}
	datum := &StringStringMap{}
	datum.Parse(parser)
	return datum
}

func (s *StringStringMap) Parse(parser structure.ObjectParser) {
	for _, reference := range parser.References() {
		if *s == nil {
			*s = map[string]*string{}
		}
		(*s)[reference] = parser.String(reference)
	}
}

func (s *StringStringMap) Validate(validator structure.Validator) {
	if length := len(*s); length > StringStringMapLengthMaximum {
		validator.ReportError(structureValidator.ErrorLengthNotLessThanOrEqualTo(length, StringStringMapLengthMaximum))
	}
	for _, reference := range s.SortedKeys() {
		validator.WithReference(reference).String(structure.ReferenceSelf, &reference).LengthLessThanOrEqualTo(StringStringMapReferenceLengthMaximum)
		validator.String(reference, (*s)[reference]).LengthLessThanOrEqualTo(StringStringMapValueLengthMaximum)
	}
}

func (s *StringStringMap) SortedKeys() []string {
	return slices.Sorted(maps.Keys(*s))
}
