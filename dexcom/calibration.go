package dexcom

import (
	"strconv"
	"time"

	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type CalibrationsResponse struct {
	Calibrations []*Calibration `json:"calibrations,omitempty"`
}

func NewCalibrationsResponse() *CalibrationsResponse {
	return &CalibrationsResponse{}
}

func (c *CalibrationsResponse) Parse(parser structure.ObjectParser) {
	if calibrationsParser := parser.WithReferenceArrayParser("calibrations"); calibrationsParser.Exists() {
		for _, reference := range calibrationsParser.References() {
			if calibrationParser := calibrationsParser.WithReferenceObjectParser(reference); calibrationParser.Exists() {
				calibration := NewCalibration()
				calibration.Parse(calibrationParser)
				calibrationParser.NotParsed()
				c.Calibrations = append(c.Calibrations, calibration)
			}
		}
		calibrationsParser.NotParsed()
	}
}

func (c *CalibrationsResponse) Validate(validator structure.Validator) {
	validator = validator.WithMeta(c)
	validator = validator.WithReference("calibrations")
	for index, calibration := range c.Calibrations {
		if calibrationValidator := validator.WithReference(strconv.Itoa(index)); calibration != nil {
			calibration.Validate(calibrationValidator)
		} else {
			calibrationValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

type Calibration struct {
	SystemTime    time.Time `json:"systemTime,omitempty"`
	DisplayTime   time.Time `json:"displayTime,omitempty"`
	Unit          string    `json:"unit,omitempty"`
	Value         float64   `json:"value,omitempty"`
	TransmitterID *string   `json:"transmitterId,omitempty"`
}

func NewCalibration() *Calibration {
	return &Calibration{}
}

func (c *Calibration) Parse(parser structure.ObjectParser) {
	if ptr := parser.Time("systemTime", DateTimeFormat); ptr != nil {
		c.SystemTime = *ptr
	}
	if ptr := parser.Time("displayTime", DateTimeFormat); ptr != nil {
		c.DisplayTime = *ptr
	}
	if ptr := parser.String("unit"); ptr != nil {
		c.Unit = *ptr
	}
	if ptr := parser.Float64("value"); ptr != nil {
		c.Value = *ptr
	}
	c.TransmitterID = parser.String("transmitterId")
}

func (c *Calibration) Validate(validator structure.Validator) {
	validator = validator.WithMeta(c)
	validator.Time("systemTime", &c.SystemTime).BeforeNow(NowThreshold)
	validator.Time("displayTime", &c.DisplayTime).NotZero()
	validator.String("unit", &c.Unit).OneOf(UnitMgdL) // TODO: Add UnitMmolL
	switch c.Unit {
	case UnitMgdL:
		validator.Float64("value", &c.Value).InRange(20, 600)
	case UnitMmolL:
		// TODO: Add value validation
	}
	validator.String("transmitterId", c.TransmitterID).Matches(transmitterIDExpression)
}
