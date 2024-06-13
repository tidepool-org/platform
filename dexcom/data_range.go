package dexcom

import (
	"errors"
	"time"

	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type DataRange struct {
	Start *Times `json:"start,omitempty"`
	End   *Times `json:"end,omitempty"`
}

type Times struct {
	SystemTime  *Time `json:"systemTime,omitempty"`
	DisplayTime *Time `json:"displayTime,omitempty"`
}

type DataRangeResponse struct {
	RecordType    *string    `json:"recordType,omitempty"`
	RecordVersion *string    `json:"recordVersion,omitempty"`
	UserID        *string    `json:"userId,omitempty"`
	Calibrations  *DataRange `json:"calibrations,omitempty"`
	EGVs          *DataRange `json:"egvs,omitempty"`
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
	d.EGVs = ParseDataRange(parser.WithReferenceObjectParser("egvs"))
	d.Events = ParseDataRange(parser.WithReferenceObjectParser("events"))
}

func (d *DataRangeResponse) GetOldestStartDate() (time.Time, error) {
	now := time.Now().UTC()
	var err error

	oldest, err := d.Calibrations.getOlderTime(now)
	if err != nil {
		return time.Time{}, err
	}
	oldest, err = d.Events.getOlderTime(oldest)
	if err != nil {
		return time.Time{}, err
	}

	oldest, err = d.EGVs.getOlderTime(oldest)
	if err != nil {
		return time.Time{}, err
	}

	if oldest.Compare(now) == 0 {
		return time.Time{}, errors.New("the oldest start date should before now")
	}

	return oldest, nil
}

func (d *DataRangeResponse) Validate(validator structure.Validator) {
	if calibrationsValidator := validator.WithReference("calibrations"); d.Calibrations != nil {
		d.Calibrations.Validate(calibrationsValidator)
	} else {
		calibrationsValidator.ReportError(structureValidator.ErrorValueNotExists())
	}

	if egvsValidator := validator.WithReference("egvs"); d.EGVs != nil {
		d.EGVs.Validate(egvsValidator)
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
	return &DataRange{
		Start: NewTimes(),
		End:   NewTimes(),
	}
}

func (d *DataRange) getOlderTime(t time.Time) (time.Time, error) {
	if d.Start != nil {
		if d.Start.DisplayTime != nil {
			if !d.Start.DisplayTime.IsZero() {
				if d.Start.DisplayTime.Before(t) {
					return d.Start.DisplayTime.Time, nil
				}
				return t, nil
			}
		}
	}
	return time.Time{}, errors.New("invalid start display time")
}

func (d *DataRange) Parse(parser structure.ObjectParser) {
	d.Start = ParseNewTimes(parser.WithReferenceObjectParser("start"))
	d.End = ParseNewTimes(parser.WithReferenceObjectParser("end"))
}

func (d *DataRange) Validate(validator structure.Validator) error {
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
	return validator.Error()
}

func ParseNewTimes(parser structure.ObjectParser) *Times {
	if !parser.Exists() {
		return nil
	}
	datum := NewTimes()
	parser.Parse(datum)
	return datum
}

func NewTimes() *Times {
	return &Times{
		SystemTime:  NewTime(),
		DisplayTime: NewTime(),
	}
}

func (c *Times) Parse(parser structure.ObjectParser) {
	c.SystemTime = TimeFromString(parser.String("systemTime"))
	c.DisplayTime = TimeFromString(parser.String("displayTime"))
}

func (d *Times) Validate(validator structure.Validator) {
	validator = validator.WithMeta(d)
	validator.Time("systemTime", d.SystemTime.Raw()).Exists().NotZero().BeforeNow(SystemTimeNowThreshold)
	validator.Time("displayTime", d.DisplayTime.Raw()).Exists().NotZero()
}
