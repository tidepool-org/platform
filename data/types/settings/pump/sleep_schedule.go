package pump

import (
	"strconv"

	"github.com/tidepool-org/platform/data/types/common"
	"github.com/tidepool-org/platform/structure"

	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	SleepSchedulesMidnightOffsetMaximum = 86400
	SleepSchedulesMidnightOffsetMinimum = 0

	SleepSchedulesLengthMaximum = 10
	SleepSchedulesLengthMinimum = 0
)

type SleepSchedules []*SleepSchedule

func ParseSleepSchedules(parser structure.ArrayParser) *SleepSchedules {
	if !parser.Exists() {
		return nil
	}
	datum := NewSleepSchedules()
	parser.Parse(datum)
	return datum
}

func NewSleepSchedules() *SleepSchedules {
	return &SleepSchedules{}
}

func (s *SleepSchedules) Parse(parser structure.ArrayParser) {
	for _, reference := range parser.References() {
		*s = append(*s, ParseSleepSchedule(parser.WithReferenceObjectParser(reference)))
	}
}

func (s *SleepSchedules) Validate(validator structure.Validator) {
	length := len(*s)

	if length > SleepSchedulesLengthMaximum {
		validator.ReportError(structureValidator.ErrorLengthNotLessThanOrEqualTo(length, SleepSchedulesLengthMaximum))
	}

	for index, datum := range *s {
		if datumValidator := validator.WithReference(strconv.Itoa(index)); datum != nil {
			datum.Validate(datumValidator)
		} else {
			datumValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
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
			validator.StringArray("days", s.Days).Exists().EachOneOf(common.Days()...).EachUnique()
			validator.Int("start", s.Start).Exists().InRange(SleepSchedulesMidnightOffsetMinimum, SleepSchedulesMidnightOffsetMaximum)
			validator.Int("end", s.End).Exists().InRange(SleepSchedulesMidnightOffsetMinimum, SleepSchedulesMidnightOffsetMaximum)
		}
	}
}
