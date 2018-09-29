package cgm

import (
	"strconv"

	"github.com/tidepool-org/platform/data"
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

func ParseScheduledAlerts(parser data.ArrayParser) *ScheduledAlerts {
	if parser.Array() == nil {
		return nil
	}
	datum := NewScheduledAlerts()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewScheduledAlerts() *ScheduledAlerts {
	return &ScheduledAlerts{}
}

func (s *ScheduledAlerts) Parse(parser data.ArrayParser) {
	for index := range *parser.Array() {
		*s = append(*s, ParseScheduledAlert(parser.NewChildObjectParser(index)))
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

func ParseScheduledAlert(parser data.ObjectParser) *ScheduledAlert {
	if parser.Object() == nil {
		return nil
	}
	datum := NewScheduledAlert()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewScheduledAlert() *ScheduledAlert {
	return &ScheduledAlert{}
}

func (s *ScheduledAlert) Parse(parser data.ObjectParser) {
	s.Name = parser.ParseString("name")
	s.Days = parser.ParseStringArray("days")
	s.Start = parser.ParseInteger("start")
	s.End = parser.ParseInteger("end")
	s.Alerts = ParseAlerts(parser.NewChildObjectParser("alerts"))
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
