package dexcom

import (
	"strconv"

	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	CalibrationUnitUnknown = "unknown"
	CalibrationUnitMgdL    = dataBloodGlucose.MgdL
	CalibrationUnitMmolL   = dataBloodGlucose.MmolL

	CalibrationValueMgdLMaximum = dataBloodGlucose.MgdLMaximum
	CalibrationValueMgdLMinimum = dataBloodGlucose.MgdLMinimum

	CalibrationValueMmolLMaximum = dataBloodGlucose.MmolLMaximum
	CalibrationValueMmolLMinimum = dataBloodGlucose.MmolLMinimum
)

func CalibrationUnits() []string {
	return []string{
		CalibrationUnitUnknown,
		CalibrationUnitMgdL,
		CalibrationUnitMmolL,
	}
}

type CalibrationsResponse struct {
	RecordType    *string       `json:"recordType,omitempty"`
	RecordVersion *string       `json:"recordVersion,omitempty"`
	UserID        *string       `json:"userId,omitempty"`
	Calibrations  *Calibrations `json:"records,omitempty"`
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
	c.UserID = parser.String("userId")
	c.RecordType = parser.String("recordType")
	c.RecordVersion = parser.String("recordVersion")
	c.Calibrations = ParseCalibrations(parser.WithReferenceArrayParser("records"))
}

func (c *CalibrationsResponse) Validate(validator structure.Validator) {
	if calibrationsValidator := validator.WithReference("records"); c.Calibrations != nil {
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
	ID                    *string  `json:"recordId,omitempty"`
	SystemTime            *Time    `json:"systemTime,omitempty"`
	DisplayTime           *Time    `json:"displayTime,omitempty"`
	Unit                  *string  `json:"unit,omitempty"`
	Value                 *float64 `json:"value,omitempty"`
	TransmitterID         *string  `json:"transmitterId,omitempty"`
	TransmitterGeneration *string  `json:"transmitterGeneration,omitempty"`
	DisplayDevice         *string  `json:"displayDevice,omitempty"`
	TransmitterTicks      *int     `json:"transmitterTicks,omitempty"`
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
	c.ID = parser.String("recordId")
	c.SystemTime = ParseTime(parser, "systemTime")
	c.DisplayTime = ParseTime(parser, "displayTime")
	c.Unit = StringOrDefault(parser, "unit", CalibrationUnitMgdL)
	c.Value = parser.Float64("value")
	c.TransmitterID = parser.String("transmitterId")
	c.TransmitterGeneration = parser.String("transmitterGeneration")
	c.DisplayDevice = parser.String("displayDevice")
	c.TransmitterTicks = parser.Int("transmitterTicks")
}

func (c *Calibration) Validate(validator structure.Validator) {
	validator = validator.WithMeta(c)
	validator.String("recordId", c.ID).Exists().NotEmpty()
	validator.Time("systemTime", c.SystemTime.Raw()).NotZero().BeforeNow(SystemTimeNowThreshold)
	validator.Time("displayTime", c.DisplayTime.Raw()).NotZero()
	validator.String("transmitterId", c.TransmitterID).Exists().Using(TransmitterIDValidator)
	validator.Int("transmitterTicks", c.TransmitterTicks).Exists()
	validator.String("displayDevice", c.DisplayDevice).Exists().NotEmpty()
	validator.String("transmitterGeneration", c.TransmitterGeneration).Exists().OneOf(DeviceTransmitterGenerations()...)
	validator.String("unit", c.Unit).Exists().OneOf(CalibrationUnits()...)
	if c.Unit != nil {
		switch *c.Unit {
		case CalibrationUnitMgdL:
			validator.Float64("value", c.Value).Exists().InRange(CalibrationValueMgdLMinimum, CalibrationValueMgdLMaximum)
		case CalibrationUnitMmolL:
			validator.Float64("value", c.Value).Exists().InRange(CalibrationValueMmolLMinimum, CalibrationValueMmolLMaximum)
		case CalibrationUnitUnknown:
			validator.Float64("value", c.Value).Exists()
		}
	}
}
