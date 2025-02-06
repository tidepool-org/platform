package dexcom

import (
	"sort"

	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	DataRangesResponseRecordType    = "dataRange"
	DataRangesResponseRecordVersion = "3.0"
)

type DataRangesResponse struct {
	RecordType    *string    `json:"recordType,omitempty"`
	RecordVersion *string    `json:"recordVersion,omitempty"`
	UserID        *string    `json:"userId,omitempty"`
	Calibrations  *DataRange `json:"calibrations,omitempty"`
	EGVs          *DataRange `json:"egvs,omitempty"`
	Events        *DataRange `json:"events,omitempty"`
}

func ParseDataRangesResponse(parser structure.ObjectParser) *DataRangesResponse {
	if !parser.Exists() {
		return nil
	}
	datum := NewDataRangesResponse()
	parser.Parse(datum)
	return datum
}

func NewDataRangesResponse() *DataRangesResponse {
	return &DataRangesResponse{}
}

func (d *DataRangesResponse) Parse(parser structure.ObjectParser) {
	parser = parser.WithMeta(d)

	d.RecordType = parser.String("recordType")
	d.RecordVersion = parser.String("recordVersion")
	d.UserID = parser.String("userId")
	d.Calibrations = ParseDataRange(parser.WithReferenceObjectParser("calibrations"))
	d.EGVs = ParseDataRange(parser.WithReferenceObjectParser("egvs"))
	d.Events = ParseDataRange(parser.WithReferenceObjectParser("events"))
}

func (d *DataRangesResponse) Validate(validator structure.Validator) {
	validator = validator.WithMeta(d)

	validator.String("recordType", d.RecordType).Exists().EqualTo(DataRangesResponseRecordType)
	validator.String("recordVersion", d.RecordVersion).Exists().EqualTo(DataRangesResponseRecordVersion)
	validator.String("userId", d.UserID).Exists().NotEmpty()
	if calibrationsValidator := validator.WithReference("calibrations"); d.Calibrations != nil {
		d.Calibrations.Validate(calibrationsValidator)
	}
	if egvsValidator := validator.WithReference("egvs"); d.EGVs != nil {
		d.EGVs.Validate(egvsValidator)
	}
	if eventsValidator := validator.WithReference("events"); d.Events != nil {
		d.Events.Validate(eventsValidator)
	}
}

func (d *DataRangesResponse) DataRange() *DataRange {
	var startMoments Moments
	var endMoments Moments

	if d.Calibrations != nil {
		startMoments = append(startMoments, d.Calibrations.Start)
		endMoments = append(endMoments, d.Calibrations.End)
	}
	if d.EGVs != nil {
		startMoments = append(startMoments, d.EGVs.Start)
		endMoments = append(endMoments, d.EGVs.End)
	}
	if d.Events != nil {
		startMoments = append(startMoments, d.Events.Start)
		endMoments = append(endMoments, d.Events.End)
	}

	startMoments = startMoments.Compact()
	endMoments = endMoments.Compact()
	if len(startMoments) == 0 || len(endMoments) == 0 {
		return nil
	}

	sort.Stable(MomentsBySystemTimeRaw(startMoments))
	sort.Stable(sort.Reverse(MomentsBySystemTimeRaw(endMoments)))

	return &DataRange{
		Start: startMoments[0],
		End:   endMoments[0],
	}
}

type DataRange struct {
	Start *Moment `json:"start,omitempty"`
	End   *Moment `json:"end,omitempty"`
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
	d.Start = ParseMoment(parser.WithReferenceObjectParser("start"))
	d.End = ParseMoment(parser.WithReferenceObjectParser("end"))
}

func (d *DataRange) Validate(validator structure.Validator) {
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

	// End after start
	if d.Start != nil && d.Start.SystemTime != nil && d.End != nil && d.End.SystemTime != nil {
		validator.Time("end", &d.End.SystemTime.Time).After(d.Start.SystemTime.Time)
	}
}
