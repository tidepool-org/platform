package dexcom

import (
	"strconv"

	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/types/settings/cgm"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	EGVUnitUnknown = "unknown"
	EGVUnitMgdL    = dataBloodGlucose.MgdL
	EGVUnitMmolL   = dataBloodGlucose.MmolL

	EGVRateUnitUnknown     = "unknown"
	EGVRateUnitMgdLMinute  = "mg/dL/min"
	EGVRateUnitMmolLMinute = "mmol/L/min"

	EGVValueMgdLMaximum        = dataBloodGlucose.MgdLMaximum
	EGVValueMgdLMinimum        = dataBloodGlucose.MgdLMinimum
	EGVValuePinnedMgdLMaximum  = 400.0
	EGVValuePinnedMgdLMinimum  = 40.0
	EGVValueMmolLMaximum       = dataBloodGlucose.MmolLMaximum
	EGVValueMmolLMinimum       = dataBloodGlucose.MmolLMinimum
	EGVValuePinnedMmolLMaximum = cgm.HighAlertLevelMmolLMaximum
	EGVValuePinnedMmolLMinimum = cgm.UrgentLowAlertLevelMmolLMinimum

	EGVStatusUnknown = "unknown"
	EGVStatusHigh    = "high"
	EGVStatusLow     = "low"
	EGVStatusOK      = "ok"

	EGVTrendDoubleUp       = "doubleUp"
	EGVTrendSingleUp       = "singleUp"
	EGVTrendFortyFiveUp    = "fortyFiveUp"
	EGVTrendFlat           = "flat"
	EGVTrendFortyFiveDown  = "fortyFiveDown"
	EGVTrendSingleDown     = "singleDown"
	EGVTrendDoubleDown     = "doubleDown"
	EGVTrendNone           = "none"
	EGVTrendNotComputable  = "notComputable"
	EGVTrendRateOutOfRange = "rateOutOfRange"

	EGVTransmitterTickMinimum = 0
)

func EGVsResponseRateUnits() []string {
	return []string{
		EGVRateUnitUnknown,
		EGVRateUnitMgdLMinute,
		EGVRateUnitMmolLMinute,
	}
}

func EGVsResponseUnits() []string {
	return []string{
		EGVUnitUnknown,
		EGVUnitMgdL,
		EGVUnitMmolL,
	}
}

func EGVStatuses() []string {
	return []string{
		EGVStatusUnknown,
		EGVStatusHigh,
		EGVStatusLow,
		EGVStatusOK,
	}
}

func EGVTrends() []string {
	return []string{
		EGVTrendDoubleUp,
		EGVTrendSingleUp,
		EGVTrendFortyFiveUp,
		EGVTrendFlat,
		EGVTrendFortyFiveDown,
		EGVTrendSingleDown,
		EGVTrendDoubleDown,
		EGVTrendNone,
		EGVTrendNotComputable,
		EGVTrendRateOutOfRange,
	}
}

type EGVsResponse struct {
	RecordType    *string `json:"recordType,omitempty"`
	RecordVersion *string `json:"recordVersion,omitempty"`
	UserID        *string `json:"userId,omitempty"`
	EGVs          *EGVs   `json:"records,omitempty"`
}

func ParseEGVsResponse(parser structure.ObjectParser) *EGVsResponse {
	if !parser.Exists() {
		return nil
	}
	datum := NewEGVsResponse()
	parser.Parse(datum)
	return datum
}

func NewEGVsResponse() *EGVsResponse {
	return &EGVsResponse{}
}

func (e *EGVsResponse) Parse(parser structure.ObjectParser) {
	e.UserID = parser.String("userId")
	e.RecordType = parser.String("recordType")
	e.RecordVersion = parser.String("recordVersion")
	e.EGVs = ParseEGVs(parser.WithReferenceArrayParser("records"))
}

func (e *EGVsResponse) Validate(validator structure.Validator) {
	if egvsValidator := validator.WithReference("records"); e.EGVs != nil {
		e.EGVs.Validate(egvsValidator)
	} else {
		egvsValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
}

type EGVs []*EGV

func ParseEGVs(parser structure.ArrayParser) *EGVs {
	if !parser.Exists() {
		return nil
	}
	datum := NewEGVs()
	parser.Parse(datum)
	return datum
}

func NewEGVs() *EGVs {
	return &EGVs{}
}

func (e *EGVs) Parse(parser structure.ArrayParser) {
	for _, reference := range parser.References() {
		*e = append(*e, ParseEGV(parser.WithReferenceObjectParser(reference)))
	}
}

func (e *EGVs) Validate(validator structure.Validator) {
	for index, egv := range *e {
		if egvValidator := validator.WithReference(strconv.Itoa(index)); egv != nil {
			egv.Validate(egvValidator)
		} else {
			egvValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

type EGV struct {
	ID                    *string  `json:"recordId,omitempty"`
	SystemTime            *Time    `json:"systemTime,omitempty"`
	DisplayTime           *Time    `json:"displayTime,omitempty"`
	Unit                  *string  `json:"unit,omitempty"`
	RateUnit              *string  `json:"rateUnit,omitempty"`
	Value                 *float64 `json:"value,omitempty"`
	Status                *string  `json:"status,omitempty"`
	Trend                 *string  `json:"trend,omitempty"`
	TrendRate             *float64 `json:"trendRate,omitempty"`
	TransmitterID         *string  `json:"transmitterId,omitempty"`
	TransmitterTicks      *int     `json:"transmitterTicks,omitempty"`
	TransmitterGeneration *string  `json:"transmitterGeneration,omitempty"`
	DisplayDevice         *string  `json:"displayDevice,omitempty"`
}

func ParseEGV(parser structure.ObjectParser) *EGV {
	if !parser.Exists() {
		return nil
	}
	datum := NewEGV()
	parser.Parse(datum)
	return datum
}

func NewEGV() *EGV {
	return &EGV{}
}

func (e *EGV) Parse(parser structure.ObjectParser) {
	e.ID = parser.String("recordId")
	e.SystemTime = ParseTime(parser, "systemTime")
	e.DisplayTime = ParseTime(parser, "displayTime")
	e.Unit = StringOrDefault(parser, "unit", EGVUnitMgdL)
	e.RateUnit = StringOrDefault(parser, "rateUnit", EGVRateUnitMgdLMinute)
	e.Value = parser.Float64("value")
	e.Status = parser.String("status")
	e.Trend = parser.String("trend")
	e.TrendRate = parser.Float64("trendRate")
	e.TransmitterID = parser.String("transmitterId")
	e.TransmitterTicks = parser.Int("transmitterTicks")
	e.TransmitterGeneration = parser.String("transmitterGeneration")
	e.DisplayDevice = parser.String("displayDevice")
}

func (e *EGV) Validate(validator structure.Validator) {
	validator = validator.WithMeta(e)
	validator.String("recordId", e.ID).Exists().NotEmpty()
	validator.Time("systemTime", e.SystemTime.Raw()).NotZero().BeforeNow(SystemTimeNowThreshold)
	validator.Time("displayTime", e.DisplayTime.Raw()).NotZero()
	validator.String("unit", e.Unit).Exists().OneOf(EGVsResponseUnits()...)
	validator.String("rateUnit", e.RateUnit).Exists().OneOf(EGVsResponseRateUnits()...)
	if e.Unit != nil {
		switch *e.Unit {
		case EGVUnitMgdL:
			validator.Float64("value", e.Value).Exists().InRange(EGVValueMgdLMinimum, EGVValueMgdLMaximum)
		case EGVUnitMmolL:
			validator.Float64("value", e.Value).Exists().InRange(EGVValueMmolLMinimum, EGVValueMmolLMaximum)
		}
	}
	validator.Int("transmitterTicks", e.TransmitterTicks).Exists().GreaterThanOrEqualTo(EGVTransmitterTickMinimum)
	validator.String("transmitterGeneration", e.TransmitterGeneration).Exists().OneOf(DeviceTransmitterGenerations()...)
	validator.String("displayDevice", e.DisplayDevice).Exists().OneOf(DeviceDisplayDevices()...)
	validator.String("transmitterId", e.TransmitterID).Using(TransmitterIDValidator)
	validator.String("status", e.Status).OneOf(EGVStatuses()...)
	validator.String("trend", e.Trend).OneOf(EGVTrends()...)
}
