package data

import (
	"regexp"
	"strconv"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	SelectorOriginIDLengthMaximum = 100
	DataSourceIDHeaderKey         = "x-tidepool-data-source-id"
)

type SelectorOrigin struct {
	ID   *string `json:"id,omitempty" bson:"id,omitempty"`
	Time *string `json:"time,omitempty" bson:"time,omitempty"` // Inclusive, currently NOT used in database query
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
	s.Time = parser.String("time")
}

func (s *SelectorOrigin) Validate(validator structure.Validator) {
	validator.String("id", s.ID).Exists().NotEmpty().LengthLessThanOrEqualTo(SelectorOriginIDLengthMaximum)
	validator.String("time", s.Time).AsTime(time.RFC3339Nano).NotZero()
}

func (s *SelectorOrigin) Includes(other *SelectorOrigin) bool {
	if s == nil || other == nil { // Must not be missing
		return false
	} else if s.ID != nil && (other.ID == nil || *s.ID != *other.ID) { // If id matters, then must include
		return false
	} else if s.Time == nil { // If time does not matter, success
		return true
	} else if other.Time == nil { // Must exist
		return false
	} else if sTime, err := time.Parse(time.RFC3339Nano, *s.Time); err != nil || sTime.IsZero() { // Must parse
		return false
	} else if otherTime, err := time.Parse(time.RFC3339Nano, *other.Time); err != nil || otherTime.IsZero() { // Must parse
		return false
	} else if otherTime.Before(sTime) { // Must include
		return false
	} else {
		return true
	}
}

type Selector struct {
	ID     *string         `json:"id,omitempty" bson:"id,omitempty"`
	Time   *time.Time      `json:"time,omitempty" bson:"time,omitempty"` // Inclusive, currently NOT used in database query
	Origin *SelectorOrigin `json:"origin,omitempty" bson:"origin,omitempty"`
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
	s.Time = parser.Time("time", TimeFormat)
	s.Origin = ParseSelectorOrigin(parser.WithReferenceObjectParser("origin"))
}

func (s *Selector) Validate(validator structure.Validator) {
	if (s.ID != nil) == (s.Origin != nil) {
		validator.ReportError(structureValidator.ErrorValuesNotExistForOne("id", "origin"))
	} else if s.ID != nil {
		validator.String("id", s.ID).Using(IDValidator)
		validator.Time("time", s.Time).NotZero()
	} else {
		validator.Time("time", s.Time).NotExists()
		s.Origin.Validate(validator.WithReference("origin"))
	}
}

func (s *Selector) Includes(other *Selector) bool {
	if s == nil || other == nil { // Must not be missing
		return false
	} else if s.ID != nil && (other.ID == nil || *s.ID != *other.ID) { // If id matters, then must include
		return false
	} else if s.Time != nil && (other.Time == nil || other.Time.Before(*s.Time)) { // If time matters, then must include
		return false
	} else if s.Origin != nil && (other.Origin == nil || !s.Origin.Includes(other.Origin)) { // If origin matters, then must include
		return false
	} else {
		return true
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

func (s *Selectors) Filter(predicate func(*Selector) bool) *Selectors {
	filtered := Selectors{}
	for _, selector := range *s {
		if predicate(selector) {
			filtered = append(filtered, selector)
		}
	}
	return &filtered
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

// UserDataStatus is used to track the state of the user's data at the start of a summary calculation
type UserDataStatus struct {
	FirstData time.Time
	LastData  time.Time

	EarliestModified time.Time

	LastUpload time.Time

	LastUpdated     time.Time
	NextLastUpdated time.Time
}
