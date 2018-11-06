package data

import (
	"regexp"
	"strconv"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	SelectorOriginIDLengthMaximum = 100
)

type SelectorOrigin struct {
	ID *string `json:"id,omitempty"`
}

func ParseSelectorOrigin(parser structure.ObjectParser) *SelectorOrigin {
	if !parser.Exists() {
		return nil
	}
	datum := NewSelectorOrigin()
	parser.Parse(datum)
	return datum
}

func NewSelectorOrigin() *SelectorOrigin {
	return &SelectorOrigin{}
}

func (s *SelectorOrigin) Parse(parser structure.ObjectParser) {
	s.ID = parser.String("id")
}

func (s *SelectorOrigin) Validate(validator structure.Validator) {
	validator.String("id", s.ID).Exists().NotEmpty().LengthLessThanOrEqualTo(SelectorOriginIDLengthMaximum)
}

type Selector struct {
	ID     *string         `json:"id,omitempty"`
	Origin *SelectorOrigin `json:"origin,omitempty"`
}

func ParseSelector(parser structure.ObjectParser) *Selector {
	if !parser.Exists() {
		return nil
	}
	datum := NewSelector()
	parser.Parse(datum)
	return datum
}

func NewSelector() *Selector {
	return &Selector{}
}

func (s *Selector) Parse(parser structure.ObjectParser) {
	s.ID = parser.String("id")
	s.Origin = ParseSelectorOrigin(parser.WithReferenceObjectParser("origin"))
}

func (s *Selector) Validate(validator structure.Validator) {
	if (s.ID != nil) == (s.Origin != nil) {
		validator.ReportError(structureValidator.ErrorValuesNotExistForOne("id", "origin"))
	} else if s.ID != nil {
		validator.String("id", s.ID).Using(IDValidator)
	} else {
		s.Origin.Validate(validator.WithReference("origin"))
	}
}

type Selectors []*Selector

func ParseSelectors(parser structure.ArrayParser) *Selectors {
	if !parser.Exists() {
		return nil
	}
	datum := NewSelectors()
	parser.Parse(datum)
	return datum
}

func NewSelectors() *Selectors {
	return &Selectors{}
}

func (s *Selectors) Parse(parser structure.ArrayParser) {
	for _, reference := range parser.References() {
		*s = append(*s, ParseSelector(parser.WithReferenceObjectParser(reference)))
	}
}

func (s *Selectors) Validate(validator structure.Validator) {
	if len(*s) == 0 {
		validator.ReportError(structureValidator.ErrorValueEmpty())
	}
	for index, selector := range *s {
		if selectorValidator := validator.WithReference(strconv.Itoa(index)); selector != nil {
			selector.Validate(selectorValidator)
		} else {
			selectorValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

func NewID() string {
	return id.Must(id.New(16))
}

func IsValidID(value string) bool {
	return ValidateID(value) == nil
}

func IDValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateID(value))
}

func ValidateID(value string) error {
	if value == "" {
		return structureValidator.ErrorValueEmpty()
	} else if !idExpression.MatchString(value) {
		return ErrorValueStringAsIDNotValid(value)
	}
	return nil
}

func ErrorValueStringAsIDNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as data id", value)
}

var idExpression = regexp.MustCompile("^[0-9a-z]{32}$") // TODO: Want just "[0-9a-f]{32}" (Jellyfish uses [0-9a-z])
