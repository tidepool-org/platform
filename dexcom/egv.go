package dexcom

import (
	"strconv"

	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/types/settings/cgm"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	EGVUnitUnknown     = "unknown"
	EGVUnitMgdL        = "mg/dL"
	EGVUnitMgdLMinute  = "mg/dL/min"
	EGVUnitMmolL       = "mmol/L"
	EGVUnitMmolLMinute = "mmol/L/min"

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
		EGVUnitUnknown,
		EGVUnitMgdLMinute,
		EGVUnitMmolLMinute,
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
	RateUnit *string `json:"rateUnit,omitempty"`
	Unit     *string `json:"unit,omitempty"`
	EGVs     *EGVs   `json:"egvs,omitempty"`
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
	e.RateUnit = parser.String("rateUnit")
	e.Unit = parser.String("unit")
	e.EGVs = ParseEGVs(parser.WithReferenceArrayParser("egvs"), e.Unit)
}

func (e *EGVsResponse) Validate(validator structure.Validator) {
	validator.String("rateUnit", e.RateUnit).Exists().OneOf(EGVsResponseRateUnits()...)
	validator.String("unit", e.Unit).Exists().OneOf(EGVsResponseUnits()...)
	if egvsValidator := validator.WithReference("egvs"); e.EGVs != nil {
		e.EGVs.Validate(egvsValidator)
	} else {
		egvsValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
}

type EGVs []*EGV

func ParseEGVs(parser structure.ArrayParser, unit *string) *EGVs {
	if !parser.Exists() {
		return nil
	}
	datum := NewEGVs()
	datum.Parse(parser, unit)
	parser.NotParsed()
	return datum
}

func NewEGVs() *EGVs {
	return &EGVs{}
}

func (e *EGVs) Parse(parser structure.ArrayParser, unit *string) {
	for _, reference := range parser.References() {
		*e = append(*e, ParseEGV(parser.WithReferenceObjectParser(reference), unit))
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
	Value                 *float64 `json:"value,omitempty"`
	RealTimeValue         *float64 `json:"realtimeValue,omitempty"`
	SmoothedValue         *float64 `json:"smoothedValue,omitempty"`
	Status                *string  `json:"status,omitempty"`
	Trend                 *string  `json:"trend,omitempty"`
	TrendRate             *float64 `json:"trendRate,omitempty"`
	TransmitterID         *string  `json:"transmitterId,omitempty"`
	TransmitterTicks      *int     `json:"transmitterTicks,omitempty"`
	TransmitterGeneration *string  `json:"transmitterGeneration,omitempty"`
	DisplayDevice         *string  `json:"displayDevice,omitempty"`
}

func ParseEGV(parser structure.ObjectParser, unit *string) *EGV {
	if !parser.Exists() {
		return nil
	}
	datum := NewEGV(unit)
	parser.Parse(datum)
	return datum
}

func NewEGV(unit *string) *EGV {
	return &EGV{
		Unit: unit,
	}
}

func (e *EGV) Parse(parser structure.ObjectParser) {
	e.ID = parser.String("recordId")
	e.SystemTime = TimeFromRaw(parser.ForgivingTime("systemTime", TimeFormat))
	e.DisplayTime = TimeFromRaw(parser.ForgivingTime("displayTime", TimeFormat))
	e.Value = parser.Float64("value")
	e.RealTimeValue = parser.Float64("realtimeValue")
	e.SmoothedValue = parser.Float64("smoothedValue")
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
	validator.Time("systemTime", e.SystemTime.Raw()).Exists().NotZero().BeforeNow(SystemTimeNowThreshold)
	validator.Time("displayTime", e.DisplayTime.Raw()).Exists().NotZero()
	validator.String("unit", e.Unit).OneOf(EGVsResponseUnits()...)
	if e.Unit != nil {
		switch *e.Unit {
		case EGVUnitMgdL:
			validator.Float64("value", e.Value).Exists().InRange(EGVValueMgdLMinimum, EGVValueMgdLMaximum)
			validator.Float64("realtimeValue", e.RealTimeValue).Exists().InRange(EGVValueMgdLMinimum, EGVValueMgdLMaximum)
			validator.Float64("smoothedValue", e.SmoothedValue).InRange(EGVValueMgdLMinimum, EGVValueMgdLMaximum)
		case EGVUnitMmolL:
			validator.Float64("value", e.Value).Exists().InRange(EGVValueMmolLMinimum, EGVValueMmolLMaximum)
			validator.Float64("realtimeValue", e.RealTimeValue).Exists().InRange(EGVValueMmolLMinimum, EGVValueMmolLMaximum)
			validator.Float64("smoothedValue", e.SmoothedValue).InRange(EGVValueMmolLMinimum, EGVValueMmolLMaximum)
		}
	}
	validator.String("status", e.Status).OneOf(EGVStatuses()...)
	validator.String("trend", e.Trend).OneOf(EGVTrends()...)
	validator.String("transmitterId", e.TransmitterID).Using(TransmitterIDValidator)
	validator.Int("transmitterTicks", e.TransmitterTicks).GreaterThanOrEqualTo(EGVTransmitterTickMinimum)
	validator.String("transmitterGeneration", e.TransmitterGeneration).OneOf(DeviceTransmitterGenerations()...)
	validator.String("displayDevice", e.DisplayDevice).OneOf(DeviceDisplayDevices()...)
}
