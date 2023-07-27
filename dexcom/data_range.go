package dexcom

import (
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type DataRange struct {
	Start *DateRange `json:"start,omitempty"`
	End   *DateRange `json:"end,omitempty"`
}

type DateRange struct {
	SystemTime  *Time `json:"systemTime,omitempty"`
	DisplayTime *Time `json:"displayTime,omitempty"`
}

type DataRangeResponse struct {
	RecordType    *string    `json:"recordType,omitempty"`
	RecordVersion *string    `json:"recordVersion,omitempty"`
	UserID        *string    `json:"userId,omitempty"`
	Calibrations  *DataRange `json:"calibrations,omitempty"`
	Egvs          *DataRange `json:"egvs,omitempty"`
	Events        *DataRange `json:"events,omitempty"`
}

func ParseDataRangeResponse(parser structure.ObjectParser) *DataRangeResponse {
	if !parser.Exists() {
		return nil
	}
	datum := NewDataRangeResponse()
	parser.Parse(datum)
	return datum
}

func NewDataRangeResponse() *DataRangeResponse {
	return &DataRangeResponse{}
}

func (d *DataRangeResponse) Parse(parser structure.ObjectParser) {
	d.UserID = parser.String("userId")
	d.RecordType = parser.String("recordType")
	d.RecordVersion = parser.String("recordVersion")
	d.Calibrations = ParseDataRange(parser.WithReferenceObjectParser("calibrations"))
	d.Egvs = ParseDataRange(parser.WithReferenceObjectParser("egvs"))
	d.Events = ParseDataRange(parser.WithReferenceObjectParser("events"))
}

func (d *DataRangeResponse) Validate(validator structure.Validator) {
	if calibrationsValidator := validator.WithReference("calibrations"); d.Calibrations != nil {
		d.Calibrations.Validate(calibrationsValidator)
	} else {
		calibrationsValidator.ReportError(structureValidator.ErrorValueNotExists())
	}

	if egvsValidator := validator.WithReference("egvs"); d.Egvs != nil {
		d.Egvs.Validate(egvsValidator)
	} else {
		egvsValidator.ReportError(structureValidator.ErrorValueNotExists())
	}

	if eventsValidator := validator.WithReference("events"); d.Events != nil {
		d.Events.Validate(eventsValidator)
	} else {
		eventsValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
}

func ParseDataRange(parser structure.ObjectParser) *DataRange {
	if !parser.Exists() {
		return nil
	}
	datum := NewDataRange()
	parser.Parse(datum)
	return datum
}

func NewDataRange() *DataRange {
	return &DataRange{}
}

func (d *DataRange) Parse(parser structure.ObjectParser) {
	d.Start = ParseNewDateRange(parser.WithReferenceObjectParser("start"))
	d.End = ParseNewDateRange(parser.WithReferenceObjectParser("end"))
}

func (d *DataRange) Validate(validator structure.Validator) {
	validator = validator.WithMeta(d)
	if startValidator := validator.WithReference("start"); d.Start != nil {
		d.Start.Validate(startValidator)
	} else {
		startValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
	if endValidator := validator.WithReference("end"); d.End != nil {
		d.End.Validate(endValidator)
	} else {
		endValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
}

func ParseNewDateRange(parser structure.ObjectParser) *DateRange {
	if !parser.Exists() {
		return nil
	}
	datum := NewDateRange()
	parser.Parse(datum)
	return datum
}

func NewDateRange() *DateRange {
	return &DateRange{}
}

func (c *DateRange) Parse(parser structure.ObjectParser) {
	c.SystemTime = TimeFromString(parser.String("systemTime"))
	c.DisplayTime = TimeFromString(parser.String("displayTime"))
}

func (d *DateRange) Validate(validator structure.Validator) {
	validator = validator.WithMeta(d)
	validator.Time("systemTime", d.SystemTime.Raw()).Exists().NotZero().BeforeNow(SystemTimeNowThreshold)
	validator.Time("displayTime", d.DisplayTime.Raw()).Exists().NotZero()
}
