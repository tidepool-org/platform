package dexcom

import (
	"strconv"

	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	CalibrationUnitMgdL = "mg/dL"

	CalibrationValueMgdLMaximum = 600.0
	CalibrationValueMgdLMinimum = 20.0
)

func CalibrationUnits() []string {
	return []string{
		CalibrationUnitMgdL,
	}
}

type CalibrationsResponse struct {
	Calibrations *Calibrations `json:"calibrations,omitempty"`
}

func ParseCalibrationsResponse(parser structure.ObjectParser) *CalibrationsResponse {
	if !parser.Exists() {
		return nil
	}
	datum := NewCalibrationsResponse()
	parser.Parse(datum)
	return datum
}

func NewCalibrationsResponse() *CalibrationsResponse {
	return &CalibrationsResponse{}
}

func (c *CalibrationsResponse) Parse(parser structure.ObjectParser) {
	c.Calibrations = ParseCalibrations(parser.WithReferenceArrayParser("calibrations"))
}

func (c *CalibrationsResponse) Validate(validator structure.Validator) {
	if calibrationsValidator := validator.WithReference("calibrations"); c.Calibrations != nil {
		c.Calibrations.Validate(calibrationsValidator)
	} else {
		calibrationsValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
}

type Calibrations []*Calibration

func ParseCalibrations(parser structure.ArrayParser) *Calibrations {
	if !parser.Exists() {
		return nil
	}
	datum := NewCalibrations()
	parser.Parse(datum)
	return datum
}

func NewCalibrations() *Calibrations {
	return &Calibrations{}
}

func (c *Calibrations) Parse(parser structure.ArrayParser) {
	for _, reference := range parser.References() {
		*c = append(*c, ParseCalibration(parser.WithReferenceObjectParser(reference)))
	}
}

func (c *Calibrations) Validate(validator structure.Validator) {
	for index, calibration := range *c {
		if calibrationValidator := validator.WithReference(strconv.Itoa(index)); calibration != nil {
			calibration.Validate(calibrationValidator)
		} else {
			calibrationValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

type Calibration struct {
	SystemTime    *Time    `json:"systemTime,omitempty"`
	DisplayTime   *Time    `json:"displayTime,omitempty"`
	Unit          *string  `json:"unit,omitempty"`
	Value         *float64 `json:"value,omitempty"`
	TransmitterID *string  `json:"transmitterId,omitempty"`
}

func ParseCalibration(parser structure.ObjectParser) *Calibration {
	if !parser.Exists() {
		return nil
	}
	datum := NewCalibration()
	parser.Parse(datum)
	return datum
}

func NewCalibration() *Calibration {
	return &Calibration{}
}

func (c *Calibration) Parse(parser structure.ObjectParser) {
	c.SystemTime = TimeFromRaw(parser.Time("systemTime", TimeFormat))
	c.DisplayTime = TimeFromRaw(parser.Time("displayTime", TimeFormat))
	c.Unit = parser.String("unit")
	c.Value = parser.Float64("value")
	c.TransmitterID = parser.String("transmitterId")
}

func (c *Calibration) Validate(validator structure.Validator) {
	validator = validator.WithMeta(c)
	validator.Time("systemTime", c.SystemTime.Raw()).Exists().NotZero().BeforeNow(SystemTimeNowThreshold)
	validator.Time("displayTime", c.DisplayTime.Raw()).Exists().NotZero()
	validator.String("unit", c.Unit).Exists().OneOf(CalibrationUnits()...)
	if c.Unit != nil {
		switch *c.Unit {
		case CalibrationUnitMgdL:
			validator.Float64("value", c.Value).Exists().InRange(CalibrationValueMgdLMinimum, CalibrationValueMgdLMaximum)
		}
	}
	validator.String("transmitterId", c.TransmitterID).Using(TransmitterIDValidator)
}
