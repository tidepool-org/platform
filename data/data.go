package data

import (
	"regexp"
	"strconv"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/origin"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type SelectorDeduplicator struct {
	Hash *string `json:"hash,omitempty" bson:"hash,omitempty"`
}

func ParseSelectorDeduplicator(parser structure.ObjectParser) *SelectorDeduplicator {
	if !parser.Exists() {
		return nil
	}
	datum := NewSelectorDeduplicator()
	parser.Parse(datum)
	return datum
}

func NewSelectorDeduplicator() *SelectorDeduplicator {
	return &SelectorDeduplicator{}
}

func (s *SelectorDeduplicator) Parse(parser structure.ObjectParser) {
	s.Hash = parser.String("hash")
}

func (s *SelectorDeduplicator) Validate(validator structure.Validator) {
	validator.String("hash", s.Hash).Exists().NotEmpty().LengthLessThanOrEqualTo(DeduplicatorHashLengthMaximum)
}

func (s *SelectorDeduplicator) Matches(other *SelectorDeduplicator) bool {
	return s != nil && other != nil && s.Hash != nil && other.Hash != nil && *s.Hash == *other.Hash
}

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
	validator.String("id", s.ID).Exists().NotEmpty().LengthLessThanOrEqualTo(origin.IDLengthMaximum)
	validator.String("time", s.Time).AsTime(time.RFC3339Nano).NotZero()
}

func (s *SelectorOrigin) Matches(other *SelectorOrigin) bool {
	return s != nil && other != nil && s.ID != nil && other.ID != nil && *s.ID == *other.ID
}

func (s *SelectorOrigin) NewerThan(other *SelectorOrigin) bool {
	if s == nil || s.Time == nil { // Must not be missing
		return false
	} else if sTime, err := time.Parse(time.RFC3339Nano, *s.Time); err != nil { // Must parse
		return false
	} else if other == nil || other.Time == nil { // Must not be missing
		return true
	} else if otherTime, err := time.Parse(time.RFC3339Nano, *other.Time); err != nil { // Must parse
		return true
	} else {
		return sTime.After(otherTime) // Must be newer
	}
}

type Selector struct {
	ID           *string               `json:"id,omitempty" bson:"id,omitempty"`
	Time         *time.Time            `json:"time,omitempty" bson:"time,omitempty"` // Inclusive, currently NOT used in database query
	Deduplicator *SelectorDeduplicator `json:"deduplicator,omitempty" bson:"_deduplicator,omitempty"`
	Origin       *SelectorOrigin       `json:"origin,omitempty" bson:"origin,omitempty"`
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
	s.Deduplicator = ParseSelectorDeduplicator(parser.WithReferenceObjectParser("deduplicator"))
	s.Origin = ParseSelectorOrigin(parser.WithReferenceObjectParser("origin"))
}

func (s *Selector) Validate(validator structure.Validator) {
	if s.ID != nil {
		validator.String("id", s.ID).Using(IDValidator)
		if s.Deduplicator != nil {
			validator.WithReference("deduplicator").ReportError(structureValidator.ErrorValueExists())
		}
		if s.Origin != nil {
			validator.WithReference("origin").ReportError(structureValidator.ErrorValueExists())
		}
	} else if s.Deduplicator != nil {
		s.Deduplicator.Validate(validator.WithReference("deduplicator"))
		if s.Origin != nil {
			validator.WithReference("origin").ReportError(structureValidator.ErrorValueExists())
		}
	} else if s.Origin != nil {
		s.Origin.Validate(validator.WithReference("origin"))
	} else {
		validator.ReportError(structureValidator.ErrorValuesNotExistForOne("id", "deduplicator", "origin"))
	}

	if timeValidator := validator.Time("time", s.Time); s.ID != nil {
		timeValidator.NotZero()
	} else {
		timeValidator.NotExists()
	}
}

func (s *Selector) Matches(other *Selector) bool {
	if s == nil || other == nil { // Must not be missing
		return false
	} else if s.ID != nil { // If id matters, then must match
		return other.ID != nil && *s.ID == *other.ID
	} else if s.Deduplicator != nil { // If deduplicator matters, then must match
		return other.Deduplicator != nil && s.Deduplicator.Matches(other.Deduplicator)
	} else if s.Origin != nil { // If origin matters, then must match
		return other.Origin != nil && s.Origin.Matches(other.Origin)
	} else {
		return false
	}
}

func (s *Selector) NewerThan(other *Selector) bool {
	if s == nil { // Must not be missing
		return false
	} else if other == nil { // Must not be missing
		return true
	} else if s.ID != nil { // If id matters, then must be newer
		return s.Time != nil && (other.Time == nil || s.Time.After(*other.Time))
	} else if s.Deduplicator != nil { // If deduplicator matters, then must be newer
		return true
	} else if s.Origin != nil { // If origin matters, then must be newer
		return other.Origin == nil || s.Origin.NewerThan(other.Origin)
	} else {
		return false
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

func (s *Selectors) Append(other *Selectors) *Selectors {
	selectors := Selectors{}
	selectors = append(selectors, *s...)
	selectors = append(selectors, *other...)
	return &selectors
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

type DataByTime Data

func (d DataByTime) Len() int {
	return len(d)
}

func (d DataByTime) Less(left int, right int) bool {
	if leftTime := d[left].GetTime(); leftTime == nil {
		return true
	} else if rightTime := d[right].GetTime(); rightTime == nil {
		return false
	} else {
		return leftTime.Before(*rightTime)
	}
}

func (d DataByTime) Swap(left int, right int) {
	d[left], d[right] = d[right], d[left]
}

// UserDataStatus is used to track the state of the user's data at the start of a summary calculation
type UserDataStatus struct {
	FirstData time.Time
	LastData  time.Time

	EarliestModified time.Time

	LastUpload time.Time

	LastUpdated     time.Time
	NextLastUpdated time.Time
}
