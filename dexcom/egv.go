package dexcom

import (
	"strconv"
	"time"

	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type EGVsResponse struct {
	Unit     *string `json:"unit,omitempty"`
	RateUnit *string `json:"rateUnit,omitempty"`
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
	e.Unit = parser.String("unit")
	e.RateUnit = parser.String("rateUnit")
	e.EGVs = ParseEGVs(parser.WithReferenceArrayParser("egvs"), e.Unit)
}

func (e *EGVsResponse) Validate(validator structure.Validator) {
	validator.String("unit", e.Unit).Exists().OneOf(UnitMgdL)            // TODO: Add UnitMmolL
	validator.String("rateUnit", e.RateUnit).Exists().OneOf(UnitMgdLMin) // TODO: Add UnitMmolLMin
	if egvsValidator := validator.WithReference("egvs"); e.EGVs != nil {
		egvsValidator.Validate(e.EGVs)
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
			egvValidator.Validate(egv)
		} else {
			egvValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

type EGV struct {
	SystemTime       *time.Time `json:"systemTime,omitempty"`
	DisplayTime      *time.Time `json:"displayTime,omitempty"`
	Unit             *string    `json:"unit,omitempty"`
	Value            *float64   `json:"value,omitempty"`
	Status           *string    `json:"status,omitempty"`
	Trend            *string    `json:"trend,omitempty"`
	TrendRate        *float64   `json:"trendRate,omitempty"`
	TransmitterID    *string    `json:"transmitterId,omitempty"`
	TransmitterTicks *int       `json:"transmitterTicks,omitempty"`
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
	e.SystemTime = parser.Time("systemTime", DateTimeFormat)
	e.DisplayTime = parser.Time("displayTime", DateTimeFormat)
	e.Value = parser.Float64("value")
	e.Status = parser.String("status")
	e.Trend = parser.String("trend")
	e.TrendRate = parser.Float64("trendRate")
	e.TransmitterID = parser.String("transmitterId")
	e.TransmitterTicks = parser.Int("transmitterTicks")
}

func (e *EGV) Validate(validator structure.Validator) {
	validator = validator.WithMeta(e)
	validator.Time("systemTime", e.SystemTime).Exists().NotZero().BeforeNow(NowThreshold)
	validator.Time("displayTime", e.DisplayTime).Exists().NotZero()
	validator.String("unit", e.Unit).OneOf(UnitMgdL) // TODO: Add UnitMmolL
	if e.Unit != nil {
		switch *e.Unit {
		case UnitMgdL:
			if e.Value != nil {
				if *e.Value < EGVValueMinMgdL {
					*e.Value = EGVValueMinMgdL - 1
				} else if *e.Value > EGVValueMaxMgdL {
					*e.Value = EGVValueMaxMgdL + 1
				}
			}
			validator.Float64("value", e.Value).Exists().InRange(EGVValueMinMgdL-1, EGVValueMaxMgdL+1)
		case UnitMmolL:
			// TODO: Add value validation
		}
	}
	validator.String("status", e.Status).OneOf(StatusHigh, StatusLow, StatusOK, StatusOutOfCalibration, StatusSensorNoise)
	validator.String("trend", e.Trend).OneOf(TrendDoubleUp, TrendSingleUp, TrendFortyFiveUp, TrendFlat, TrendFortyFiveDown, TrendSingleDown, TrendDoubleDown, TrendNone, TrendNotComputable, TrendRateOutOfRange)
	validator.String("transmitterId", e.TransmitterID).Matches(transmitterIDExpression)
	validator.Int("transmitterTicks", e.TransmitterTicks).GreaterThanOrEqualTo(0)
}
