package pump

import (
	"sort"

	"github.com/tidepool-org/platform/data"
	dataTypesCommon "github.com/tidepool-org/platform/data/types/common"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	SleepSchedulesMidnightOffsetMaximum = 86400
	SleepSchedulesMidnightOffsetMinimum = 0
)

type SleepScheduleMap map[string]*SleepSchedule

func ParseSleepScheduleMap(parser structure.ObjectParser) *SleepScheduleMap {
	if !parser.Exists() {
		return nil
	}
	datum := NewSleepScheduleMap()
	parser.Parse(datum)
	return datum
}

func NewSleepScheduleMap() *SleepScheduleMap {
	return &SleepScheduleMap{}
}

func (s *SleepScheduleMap) Parse(parser structure.ObjectParser) {
	for _, reference := range parser.References() {
		s.Set(reference, ParseSleepSchedule(parser.WithReferenceObjectParser(reference)))
	}
}

func (s *SleepScheduleMap) Validate(validator structure.Validator) {
	for _, name := range s.sortedNames() {
		datumValidator := validator.WithReference(name)
		if datum := s.Get(name); datum != nil {
			datum.Validate(datumValidator)
		} else {
			datumValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

func (s *SleepScheduleMap) Normalize(normalizer data.Normalizer) {
	for _, name := range s.sortedNames() {
		datumNormalizer := normalizer.WithReference(name)
		if datum := s.Get(name); datum != nil {
			datum.Normalize(datumNormalizer)
		}
	}
}

func (s *SleepScheduleMap) Get(name string) *SleepSchedule {
	if datum, exists := (*s)[name]; exists {
		return datum
	}
	return nil
}

func (s *SleepScheduleMap) Set(name string, datum *SleepSchedule) {
	(*s)[name] = datum
}

func (s *SleepScheduleMap) sortedNames() []string {
	names := []string{}
	for name := range *s {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

type SleepSchedule struct {
	Enabled *bool     `json:"enabled,omitempty" bson:"enabled,omitempty"`
	Days    *[]string `json:"days,omitempty" bson:"days,omitempty"`
	Start   *int      `json:"start,omitempty" bson:"start,omitempty"`
	End     *int      `json:"end,omitempty" bson:"end,omitempty"`
}

func ParseSleepSchedule(parser structure.ObjectParser) *SleepSchedule {
	if !parser.Exists() {
		return nil
	}
	datum := NewSleepSchedule()
	parser.Parse(datum)
	return datum
}

func NewSleepSchedule() *SleepSchedule {
	return &SleepSchedule{}
}

func (s *SleepSchedule) Parse(parser structure.ObjectParser) {
	s.Enabled = parser.Bool("enabled")
	s.Days = parser.StringArray("days")
	s.Start = parser.Int("start")
	s.End = parser.Int("end")
}

func (s *SleepSchedule) Validate(validator structure.Validator) {
	validator.Bool("enabled", s.Enabled).Exists()
	if s.Enabled != nil {
		if *s.Enabled {
			validator.StringArray("days", s.Days).Exists().EachOneOf(dataTypesCommon.DaysOfWeek()...).EachUnique()
			validator.Int("start", s.Start).Exists().InRange(SleepSchedulesMidnightOffsetMinimum, SleepSchedulesMidnightOffsetMaximum)
			validator.Int("end", s.End).Exists().InRange(SleepSchedulesMidnightOffsetMinimum, SleepSchedulesMidnightOffsetMaximum)
		}
	}
}

func (s *SleepSchedule) Normalize(normalizer data.Normalizer) {
	if s.Days != nil {
		sort.Sort(dataTypesCommon.DaysOfWeekByDayIndex(*s.Days))
	}
}
