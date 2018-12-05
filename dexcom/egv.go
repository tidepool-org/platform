package dexcom

import (
	"strconv"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	EGVUnitMgdL       = "mg/dL"
	EGVUnitMgdLMinute = "mg/dL/min"

	EGVValueMgdLMaximum = 400.0
	EGVValueMgdLMinimum = 40.0

	EGVStatusHigh             = "high"
	EGVStatusLow              = "low"
	EGVStatusOK               = "ok"
	EGVStatusOutOfCalibration = "outOfCalibration"
	EGVStatusSensorNoise      = "sensorNoise"

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
		EGVUnitMgdLMinute,
	}
}

func EGVsResponseUnits() []string {
	return []string{
		EGVUnitMgdL,
	}
}

func EGVStatuses() []string {
	return []string{
		EGVStatusHigh,
		EGVStatusLow,
		EGVStatusOK,
		EGVStatusOutOfCalibration,
		EGVStatusSensorNoise,
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
	SystemTime       *Time    `json:"systemTime,omitempty"`
	DisplayTime      *Time    `json:"displayTime,omitempty"`
	Unit             *string  `json:"unit,omitempty"`
	Value            *float64 `json:"value,omitempty"`
	RealTimeValue    *float64 `json:"realtimeValue,omitempty"`
	SmoothedValue    *float64 `json:"smoothedValue,omitempty"`
	Status           *string  `json:"status,omitempty"`
	Trend            *string  `json:"trend,omitempty"`
	TrendRate        *float64 `json:"trendRate,omitempty"`
	TransmitterID    *string  `json:"transmitterId,omitempty"`
	TransmitterTicks *int     `json:"transmitterTicks,omitempty"`
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
	e.SystemTime = TimeFromRaw(parser.Time("systemTime", TimeFormat))
	e.DisplayTime = TimeFromRaw(parser.Time("displayTime", TimeFormat))
	e.Value = parser.Float64("value")
	e.RealTimeValue = parser.Float64("realtimeValue")
	e.SmoothedValue = parser.Float64("smoothedValue")
	e.Status = parser.String("status")
	e.Trend = parser.String("trend")
	e.TrendRate = parser.Float64("trendRate")
	e.TransmitterID = parser.String("transmitterId")
	e.TransmitterTicks = parser.Int("transmitterTicks")
}

func (e *EGV) Validate(validator structure.Validator) {
	// HACK: Dexcom - pin out of range values
	e.Value = pinEGVValue(e.Value, e.Unit)
	e.RealTimeValue = pinEGVValue(e.RealTimeValue, e.Unit)
	e.SmoothedValue = pinEGVValue(e.SmoothedValue, e.Unit)

	validator = validator.WithMeta(e)
	validator.Time("systemTime", e.SystemTime.Raw()).Exists().NotZero().BeforeNow(SystemTimeNowThreshold)
	validator.Time("displayTime", e.DisplayTime.Raw()).Exists().NotZero()
	validator.String("unit", e.Unit).OneOf(EGVsResponseUnits()...)
	if e.Unit != nil {
		switch *e.Unit {
		case EGVUnitMgdL:
			validator.Float64("value", e.Value).Exists().InRange(EGVValueMgdLMinimum-1, EGVValueMgdLMaximum+1)
			validator.Float64("realtimeValue", e.RealTimeValue).Exists().InRange(EGVValueMgdLMinimum-1, EGVValueMgdLMaximum+1)
			validator.Float64("smoothedValue", e.SmoothedValue).InRange(EGVValueMgdLMinimum-1, EGVValueMgdLMaximum+1)
		}
	}
	validator.String("status", e.Status).OneOf(EGVStatuses()...)
	validator.String("trend", e.Trend).OneOf(EGVTrends()...)
	validator.String("transmitterId", e.TransmitterID).Using(TransmitterIDValidator)
	validator.Int("transmitterTicks", e.TransmitterTicks).GreaterThanOrEqualTo(EGVTransmitterTickMinimum)
}

func pinEGVValue(value *float64, unit *string) *float64 {
	if value != nil && unit != nil {
		switch *unit {
		case EGVUnitMgdL:
			if *value < EGVValueMgdLMinimum {
				return pointer.FromFloat64(EGVValueMgdLMinimum - 1)
			} else if *value > EGVValueMgdLMaximum {
				return pointer.FromFloat64(EGVValueMgdLMaximum + 1)
			}
		}
	}
	return value
}
