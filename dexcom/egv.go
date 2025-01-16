package dexcom

import (
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	dataTypesSettingsCGM "github.com/tidepool-org/platform/data/types/settings/cgm"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	EGVsResponseRecordType    = "egv"
	EGVsResponseRecordVersion = "3.0"

	EGVUnitUnknown = "unknown"
	EGVUnitMgdL    = dataBloodGlucose.MgdL
	EGVUnitMmolL   = dataBloodGlucose.MmolL

	EGVRateUnitUnknown     = "unknown"
	EGVRateUnitMgdLMinute  = "mg/dL/min"
	EGVRateUnitMmolLMinute = "mmol/L/min"

	EGVValueMgdLMaximum  = dataBloodGlucose.MgdLMaximum
	EGVValueMgdLMinimum  = dataBloodGlucose.MgdLMinimum
	EGVValueMmolLMaximum = dataBloodGlucose.MmolLMaximum
	EGVValueMmolLMinimum = dataBloodGlucose.MmolLMinimum

	EGVStatusUnknown = "unknown"
	EGVStatusHigh    = "high"
	EGVStatusLow     = "low"
	EGVStatusOK      = "ok"

	EGVTrendUnknown        = "unknown"
	EGVTrendNone           = "none"
	EGVTrendDoubleUp       = "doubleUp"
	EGVTrendSingleUp       = "singleUp"
	EGVTrendFortyFiveUp    = "fortyFiveUp"
	EGVTrendFlat           = "flat"
	EGVTrendFortyFiveDown  = "fortyFiveDown"
	EGVTrendSingleDown     = "singleDown"
	EGVTrendDoubleDown     = "doubleDown"
	EGVTrendNotComputable  = "notComputable"
	EGVTrendRateOutOfRange = "rateOutOfRange"

	EGVTransmitterTickMinimum = 0

	EGVValuePinnedMgdLMaximum  = dataTypesSettingsCGM.HighAlertLevelMgdLMaximum
	EGVValuePinnedMgdLMinimum  = dataTypesSettingsCGM.UrgentLowAlertLevelMgdLMinimum
	EGVValuePinnedMmolLMaximum = dataTypesSettingsCGM.HighAlertLevelMmolLMaximum
	EGVValuePinnedMmolLMinimum = dataTypesSettingsCGM.UrgentLowAlertLevelMmolLMinimum
)

func EGVUnits() []string {
	return []string{
		EGVUnitUnknown,
		EGVUnitMgdL,
		EGVUnitMmolL,
	}
}

func EGVRateUnits() []string {
	return []string{
		EGVRateUnitUnknown,
		EGVRateUnitMgdLMinute,
		EGVRateUnitMmolLMinute,
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
		EGVTrendUnknown,
		EGVTrendNone,
		EGVTrendDoubleUp,
		EGVTrendSingleUp,
		EGVTrendFortyFiveUp,
		EGVTrendFlat,
		EGVTrendFortyFiveDown,
		EGVTrendSingleDown,
		EGVTrendDoubleDown,
		EGVTrendNotComputable,
		EGVTrendRateOutOfRange,
	}
}

type EGVsResponse struct {
	RecordType    *string `json:"recordType,omitempty"`
	RecordVersion *string `json:"recordVersion,omitempty"`
	UserID        *string `json:"userId,omitempty"`
	Records       *EGVs   `json:"records,omitempty"`
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
	parser = parser.WithMeta(e)

	e.RecordType = parser.String("recordType")
	e.RecordVersion = parser.String("recordVersion")
	e.UserID = parser.String("userId")
	e.Records = ParseEGVs(parser.WithReferenceArrayParser("records"))
}

func (e *EGVsResponse) Validate(validator structure.Validator) {
	validator = validator.WithMeta(e)

	validator.String("recordType", e.RecordType).Exists().EqualTo(EGVsResponseRecordType)
	validator.String("recordVersion", e.RecordVersion).Exists().EqualTo(EGVsResponseRecordVersion)
	validator.String("userId", e.UserID).Exists().NotEmpty()

	// Only validate that the records exist, remaining validation will occur later on a per-record basis
	if e.Records == nil {
		validator.WithReference("records").ReportError(structureValidator.ErrorValueNotExists())
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

type EGV struct {
	RecordID              *string  `json:"recordId,omitempty"`
	SystemTime            *Time    `json:"systemTime,omitempty"`
	DisplayTime           *Time    `json:"displayTime,omitempty"`
	Unit                  *string  `json:"unit,omitempty"`
	Value                 *float64 `json:"value,omitempty"`
	RateUnit              *string  `json:"rateUnit,omitempty"`
	TrendRate             *float64 `json:"trendRate,omitempty"`
	Status                *string  `json:"status,omitempty"`
	Trend                 *string  `json:"trend,omitempty"`
	TransmitterGeneration *string  `json:"transmitterGeneration,omitempty"`
	TransmitterID         *string  `json:"transmitterId,omitempty"`
	TransmitterTicks      *int     `json:"transmitterTicks,omitempty"`
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
	parser = parser.WithMeta(e)

	e.RecordID = parser.String("recordId")
	e.SystemTime = ParseTime(parser, "systemTime")
	e.DisplayTime = ParseTime(parser, "displayTime")
	e.Unit = parser.String("unit")
	e.Value = parser.Float64("value")
	e.RateUnit = parser.String("rateUnit")
	e.TrendRate = parser.Float64("trendRate")
	e.Status = parser.String("status")
	e.Trend = parser.String("trend")
	e.TransmitterGeneration = parser.String("transmitterGeneration")
	e.TransmitterID = parser.String("transmitterId")
	e.TransmitterTicks = parser.Int("transmitterTicks")
	e.DisplayDevice = parser.String("displayDevice")
}

func (e *EGV) Validate(validator structure.Validator) {
	validator = validator.WithMeta(e)

	validator.String("recordId", e.RecordID).Exists().NotEmpty()
	validator.Time("systemTime", e.SystemTime.Raw()).Exists().NotZero()
	validator.Time("displayTime", e.DisplayTime.Raw()).Exists().NotZero()
	validator.String("unit", e.Unit).Exists().OneOf(EGVUnits()...)
	valueValidator := validator.Float64("value", e.Value)
	valueValidator.Exists()
	if e.Unit != nil {
		switch *e.Unit {
		case EGVUnitMgdL:
			valueValidator.InRange(EGVValueMgdLMinimum, EGVValueMgdLMaximum)
		case EGVUnitMmolL:
			valueValidator.InRange(EGVValueMmolLMinimum, EGVValueMmolLMaximum)
		}
	}
	validator.String("rateUnit", e.RateUnit).Exists().OneOf(EGVRateUnits()...)
	validator.Float64("trendRate", e.TrendRate)                  // Dexcom - May not exist
	validator.String("status", e.Status).OneOf(EGVStatuses()...) // Dexcom - May not exist
	validator.String("trend", e.Trend).Exists().OneOf(EGVTrends()...)
	validator.String("transmitterGeneration", e.TransmitterGeneration).Exists().OneOf(DeviceTransmitterGenerations()...)
	validator.String("transmitterId", e.TransmitterID).Exists().Using(TransmitterIDValidator)
	validator.Int("transmitterTicks", e.TransmitterTicks).Exists().GreaterThanOrEqualTo(EGVTransmitterTickMinimum)
	validator.String("displayDevice", e.DisplayDevice).Exists().OneOf(DeviceDisplayDevices()...)

	// Log various warnings
	logger := validator.Logger().WithField("meta", e)
	if e.Unit != nil && *e.Unit == EGVUnitUnknown {
		logger.Warnf("Unit is '%s'", *e.Unit)
	}
	if e.RateUnit != nil && *e.RateUnit == EGVRateUnitUnknown {
		logger.Warnf("RateUnit is '%s'", *e.RateUnit)
	}
	if e.Status != nil && *e.Status == EGVStatusUnknown {
		logger.Warnf("Status is '%s'", *e.Status)
	}
	if e.Trend != nil && *e.Trend == EGVTrendUnknown {
		logger.Warnf("Trend is '%s'", *e.Trend)
	}
	if e.TransmitterID != nil && *e.TransmitterID == "" {
		logger.Warnf("TransmitterID is empty", *e.TransmitterID)
	}
	if e.TransmitterTicks != nil && *e.TransmitterTicks == EGVTransmitterTickMinimum {
		logger.Warnf("TransmitterTicks is %d", *e.TransmitterTicks)
	}
	if e.DisplayDevice != nil && *e.DisplayDevice == DeviceDisplayDeviceUnknown {
		logger.Warnf("DisplayDevice is '%s'", *e.DisplayDevice)
	}
}

func (e *EGV) Normalize(normalizer structure.Normalizer) {}
