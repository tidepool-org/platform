package dexcom

import (
	"strconv"
	"time"

	"github.com/tidepool-org/platform/structure"
)

type EGVsResponse struct {
	Unit     string `json:"unit,omitempty"`
	RateUnit string `json:"rateUnit,omitempty"`
	EGVs     []*EGV `json:"egvs,omitempty"`
}

func NewEGVsResponse() *EGVsResponse {
	return &EGVsResponse{}
}

func (e *EGVsResponse) Parse(parser structure.ObjectParser) {
	if ptr := parser.String("unit"); ptr != nil {
		e.Unit = *ptr
	}
	if ptr := parser.String("rateUnit"); ptr != nil {
		e.RateUnit = *ptr
	}
	if egvsParser := parser.WithReferenceArrayParser("egvs"); egvsParser.Exists() {
		for _, reference := range egvsParser.References() {
			if egvParser := egvsParser.WithReferenceObjectParser(reference); egvParser.Exists() {
				egv := NewEGV(e.Unit)
				egv.Parse(egvParser)
				egvParser.NotParsed()
				e.EGVs = append(e.EGVs, egv)
			}
		}
		egvsParser.NotParsed()
	}
}

func (e *EGVsResponse) Validate(validator structure.Validator) {
	validator.String("unit", &e.Unit).OneOf(UnitMgdL)            // TODO: Add UnitMmolL
	validator.String("rateUnit", &e.RateUnit).OneOf(UnitMgdLMin) // TODO: Add UnitMmolLMin

	validator = validator.WithReference("egvs")
	for index, egv := range e.EGVs {
		validator.Validating(strconv.Itoa(index), egv).Exists().Validate()
	}
}

type EGV struct {
	SystemTime       time.Time `json:"systemTime,omitempty"`
	DisplayTime      time.Time `json:"displayTime,omitempty"`
	Unit             string    `json:"unit,omitempty"`
	Value            float64   `json:"value,omitempty"`
	Status           *string   `json:"status,omitempty"`
	Trend            *string   `json:"trend,omitempty"`
	TrendRate        *float64  `json:"trendRate,omitempty"`
	TransmitterID    *string   `json:"transmitterId,omitempty"`
	TransmitterTicks *int      `json:"transmitterTicks,omitempty"`
}

func NewEGV(unit string) *EGV {
	return &EGV{
		Unit: unit,
	}
}

func (e *EGV) Parse(parser structure.ObjectParser) {
	if ptr := parser.Time("systemTime", DateTimeFormat); ptr != nil {
		e.SystemTime = *ptr
	}
	if ptr := parser.Time("displayTime", DateTimeFormat); ptr != nil {
		e.DisplayTime = *ptr
	}
	if ptr := parser.Float64("value"); ptr != nil {
		e.Value = *ptr
	}
	e.Status = parser.String("status")
	e.Trend = parser.String("trend")
	e.TrendRate = parser.Float64("trendRate")
	e.TransmitterID = parser.String("transmitterId")
	e.TransmitterTicks = parser.Int("transmitterTicks")
}

func (e *EGV) Validate(validator structure.Validator) {
	validator.Time("systemTime", &e.SystemTime).BeforeNow(NowThreshold)
	validator.Time("displayTime", &e.DisplayTime).NotZero()
	validator.String("unit", &e.Unit).OneOf(UnitMgdL) // TODO: Add UnitMmolL
	switch e.Unit {
	case UnitMgdL:
		if e.Value < EGVValueMinMgdL {
			e.Value = EGVValueMinMgdL - 1
		} else if e.Value > EGVValueMaxMgdL {
			e.Value = EGVValueMaxMgdL + 1
		}
		validator.Float64("value", &e.Value).InRange(EGVValueMinMgdL-1, EGVValueMaxMgdL+1)
	case UnitMmolL:
		// TODO: Add value validation
	}
	validator.String("status", e.Status).OneOf(StatusHigh, StatusLow, StatusOK, StatusOutOfCalibration, StatusSensorNoise)
	validator.String("trend", e.Trend).OneOf(TrendDoubleUp, TrendSingleUp, TrendFortyFiveUp, TrendFlat, TrendFortyFiveDown, TrendSingleDown, TrendDoubleDown, TrendNone, TrendNotComputable, TrendRateOutOfRange)
	validator.String("transmitterId", e.TransmitterID).Matches(TransmitterIDExpression)
	validator.Int("transmitterTicks", e.TransmitterTicks).GreaterThanOrEqualTo(0)
}
