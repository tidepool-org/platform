package dexcom

import (
	"sort"

	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	DataRangeResponseRecordType    = "dataRange"
	DataRangeResponseRecordVersion = "3.0"
)

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
	d.RecordType = parser.String("recordType")
	d.RecordVersion = parser.String("recordVersion")
	d.UserID = parser.String("userId")
	d.Calibrations = ParseDataRange(parser.WithReferenceObjectParser("calibrations"))
	d.EGVs = ParseDataRange(parser.WithReferenceObjectParser("egvs"))
	d.Events = ParseDataRange(parser.WithReferenceObjectParser("events"))
}

func (d *DataRangeResponse) Validate(validator structure.Validator) {
	validator.String("recordType", d.RecordType).Exists().EqualTo(DataRangeResponseRecordType)
	validator.String("recordVersion", d.RecordVersion).Exists().EqualTo(DataRangeResponseRecordVersion)
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

func (d *DataRangeResponse) Normalize(normalizer structure.Normalizer) {
	if d.Calibrations != nil {
		d.Calibrations.Normalize(normalizer.WithReference("calibrations"))
	}
	if d.EGVs != nil {
		d.EGVs.Normalize(normalizer.WithReference("egvs"))
	}
	if d.Events != nil {
		d.Events.Normalize(normalizer.WithReference("events"))
	}
}

func (d *DataRangeResponse) DataRange() *DataRange {
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

	sort.Sort(BySystemTimeRaw(startMoments))
	sort.Sort(sort.Reverse(BySystemTimeRaw(endMoments)))

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
	parser = parser.WithMeta(d)

	d.Start = ParseMoment(parser.WithReferenceObjectParser("start"))
	d.End = ParseMoment(parser.WithReferenceObjectParser("end"))
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

	// End after start
	if d.Start != nil && d.Start.SystemTime != nil && d.End != nil && d.End.SystemTime != nil {
		validator.Time("end", &d.End.SystemTime.Time).After(d.Start.SystemTime.Time)
	}
}

func (d *DataRange) Normalize(normalizer structure.Normalizer) {
	if d.Start != nil {
		d.Start.Normalize(normalizer.WithReference("start"))
	}
	if d.End != nil {
		d.End.Normalize(normalizer.WithReference("end"))
	}
}
