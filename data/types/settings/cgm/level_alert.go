package cgm

import (
	"math"

	"github.com/tidepool-org/platform/structure"
)

const (
	LevelAlertUnitsMgdL  = "mg/dL"
	LevelAlertUnitsMmolL = "mmol/L"

	HighAlertLevelMgdLMaximum       = 400.0
	HighAlertLevelMgdLMinimum       = 100.0
	HighAlertLevelMmolLMaximum      = 22.20299
	HighAlertLevelMmolLMinimum      = 5.55075
	LowAlertLevelMgdLMaximum        = 150.0
	LowAlertLevelMgdLMinimum        = 50.0
	LowAlertLevelMmolLMaximum       = 8.32612
	LowAlertLevelMmolLMinimum       = 2.77537
	UrgentLowAlertLevelMgdLMaximum  = 80.0
	UrgentLowAlertLevelMgdLMinimum  = 40.0
	UrgentLowAlertLevelMmolLMaximum = 4.44060
	UrgentLowAlertLevelMmolLMinimum = 2.22030
)

func LevelAlertUnits() []string {
	return []string{
		LevelAlertUnitsMgdL,
		LevelAlertUnitsMmolL,
	}
}

type LevelAlert struct {
	Alert `bson:",inline"`
	Level *float64 `json:"level,omitempty" bson:"level,omitempty"`
	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
}

func (l *LevelAlert) Parse(parser structure.ObjectParser) {
	l.Alert.Parse(parser)
	l.Level = parser.Float64("level")
	l.Units = parser.String("units")
}

func (l *LevelAlert) Validate(validator structure.Validator) {
	l.Alert.Validate(validator)
	if unitsValidator := validator.String("units", l.Units); l.Level != nil {
		unitsValidator.Exists().OneOf(LevelAlertUnits()...)
	} else {
		unitsValidator.NotExists()
	}
}

type HighAlert struct {
	LevelAlert `bson:",inline"`
}

func ParseHighAlert(parser structure.ObjectParser) *HighAlert {
	if !parser.Exists() {
		return nil
	}
	datum := NewHighAlert()
	parser.Parse(datum)
	return datum
}

func NewHighAlert() *HighAlert {
	return &HighAlert{}
}

func (h *HighAlert) Validate(validator structure.Validator) {
	h.LevelAlert.Validate(validator)
	validator.Float64("level", h.Level).InRange(HighAlertLevelRangeForUnits(h.Units))
}

func HighAlertLevelRangeForUnits(units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case LevelAlertUnitsMgdL:
			return HighAlertLevelMgdLMinimum, HighAlertLevelMgdLMaximum
		case LevelAlertUnitsMmolL:
			return HighAlertLevelMmolLMinimum, HighAlertLevelMmolLMaximum
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}

type LowAlert struct {
	LevelAlert `bson:",inline"`
}

func ParseLowAlert(parser structure.ObjectParser) *LowAlert {
	if !parser.Exists() {
		return nil
	}
	datum := NewLowAlert()
	parser.Parse(datum)
	return datum
}

func NewLowAlert() *LowAlert {
	return &LowAlert{}
}

func (l *LowAlert) Validate(validator structure.Validator) {
	l.LevelAlert.Validate(validator)
	validator.Float64("level", l.Level).InRange(LowAlertLevelRangeForUnits(l.Units))
}

func LowAlertLevelRangeForUnits(units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case LevelAlertUnitsMgdL:
			return LowAlertLevelMgdLMinimum, LowAlertLevelMgdLMaximum
		case LevelAlertUnitsMmolL:
			return LowAlertLevelMmolLMinimum, LowAlertLevelMmolLMaximum
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}

type UrgentLowAlert struct {
	LevelAlert `bson:",inline"`
}

func ParseUrgentLowAlert(parser structure.ObjectParser) *UrgentLowAlert {
	if !parser.Exists() {
		return nil
	}
	datum := NewUrgentLowAlert()
	parser.Parse(datum)
	return datum
}

func NewUrgentLowAlert() *UrgentLowAlert {
	return &UrgentLowAlert{}
}

func (u *UrgentLowAlert) Validate(validator structure.Validator) {
	u.LevelAlert.Validate(validator)
	validator.Float64("level", u.Level).InRange(UrgentLowAlertLevelRangeForUnits(u.Units))
}

func UrgentLowAlertLevelRangeForUnits(units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case LevelAlertUnitsMgdL:
			return UrgentLowAlertLevelMgdLMinimum, UrgentLowAlertLevelMgdLMaximum
		case LevelAlertUnitsMmolL:
			return UrgentLowAlertLevelMmolLMinimum, UrgentLowAlertLevelMmolLMaximum
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}
