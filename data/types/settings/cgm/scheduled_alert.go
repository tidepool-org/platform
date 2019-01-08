package cgm

import (
	"strconv"

	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	ScheduledAlertsLengthMaximum = 10

	ScheduledAlertNameLengthMaximum = 100

	ScheduledAlertDaysSunday    = "sunday"
	ScheduledAlertDaysMonday    = "monday"
	ScheduledAlertDaysTuesday   = "tuesday"
	ScheduledAlertDaysWednesday = "wednesday"
	ScheduledAlertDaysThursday  = "thursday"
	ScheduledAlertDaysFriday    = "friday"
	ScheduledAlertDaysSaturday  = "saturday"

	ScheduledAlertStartMaximum = 86400000
	ScheduledAlertStartMinimum = 0

	ScheduledAlertEndMaximum = 86400000
	ScheduledAlertEndMinimum = 0
)

func ScheduledAlertDays() []string {
	return []string{
		ScheduledAlertDaysSunday,
		ScheduledAlertDaysMonday,
		ScheduledAlertDaysTuesday,
		ScheduledAlertDaysWednesday,
		ScheduledAlertDaysThursday,
		ScheduledAlertDaysFriday,
		ScheduledAlertDaysSaturday,
	}
}

type ScheduledAlerts []*ScheduledAlert

func ParseScheduledAlerts(parser structure.ArrayParser) *ScheduledAlerts {
	if !parser.Exists() {
		return nil
	}
	datum := NewScheduledAlerts()
	parser.Parse(datum)
	return datum
}

func NewScheduledAlerts() *ScheduledAlerts {
	return &ScheduledAlerts{}
}

func (s *ScheduledAlerts) Parse(parser structure.ArrayParser) {
	for _, reference := range parser.References() {
		*s = append(*s, ParseScheduledAlert(parser.WithReferenceObjectParser(reference)))
	}
}

func (s *ScheduledAlerts) Validate(validator structure.Validator) {
	if length := len(*s); length == 0 {
		validator.ReportError(structureValidator.ErrorValueEmpty())
	} else if length > ScheduledAlertsLengthMaximum {
		validator.ReportError(structureValidator.ErrorLengthNotLessThanOrEqualTo(length, ScheduledAlertsLengthMaximum))
	}

	for index, datum := range *s {
		if datumValidator := validator.WithReference(strconv.Itoa(index)); datum != nil {
			datum.Validate(datumValidator)
		} else {
			datumValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

type ScheduledAlert struct {
	Name   *string   `json:"name,omitempty" bson:"name,omitempty"`
	Days   *[]string `json:"days,omitempty" bson:"days,omitempty"`
	Start  *int      `json:"start,omitempty" bson:"start,omitempty"`
	End    *int      `json:"end,omitempty" bson:"end,omitempty"`
	Alerts *Alerts   `json:"alerts,omitempty" bson:"alerts,omitempty"`
}

func ParseScheduledAlert(parser structure.ObjectParser) *ScheduledAlert {
	if !parser.Exists() {
		return nil
	}
	datum := NewScheduledAlert()
	parser.Parse(datum)
	return datum
}

func NewScheduledAlert() *ScheduledAlert {
	return &ScheduledAlert{}
}

func (s *ScheduledAlert) Parse(parser structure.ObjectParser) {
	s.Name = parser.String("name")
	s.Days = parser.StringArray("days")
	s.Start = parser.Int("start")
	s.End = parser.Int("end")
	s.Alerts = ParseAlerts(parser.WithReferenceObjectParser("alerts"))
}

func (s *ScheduledAlert) Validate(validator structure.Validator) {
	validator.String("name", s.Name).NotEmpty().LengthLessThanOrEqualTo(ScheduledAlertNameLengthMaximum)
	validator.StringArray("days", s.Days).Exists().EachOneOf(ScheduledAlertDays()...).EachUnique()
	validator.Int("start", s.Start).Exists().InRange(ScheduledAlertStartMinimum, ScheduledAlertStartMaximum)
	validator.Int("end", s.End).Exists().InRange(ScheduledAlertEndMinimum, ScheduledAlertEndMaximum)
	if alertsValidator := validator.WithReference("alerts"); s.Alerts != nil {
		s.Alerts.Validate(alertsValidator)
	} else {
		alertsValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
}
