package dexcom

import (
	"strconv"

	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	CalibrationsResponseRecordType    = "calibrations"
	CalibrationsResponseRecordVersion = "3.0"

	CalibrationUnitUnknown = "unknown"
	CalibrationUnitMgdL    = dataBloodGlucose.MgdL
	CalibrationUnitMmolL   = dataBloodGlucose.MmolL
	CalibrationUnitDefault = CalibrationUnitMgdL

	CalibrationValueMgdLMaximum = dataBloodGlucose.MgdLMaximum
	CalibrationValueMgdLMinimum = dataBloodGlucose.MgdLMinimum

	CalibrationValueMmolLMaximum = dataBloodGlucose.MmolLMaximum
	CalibrationValueMmolLMinimum = dataBloodGlucose.MmolLMinimum

	CalibrationTransmitterTickMinimum = 0
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
	Records       *Calibrations `json:"records,omitempty"`
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
	c.RecordType = parser.String("recordType")
	c.RecordVersion = parser.String("recordVersion")
	c.UserID = parser.String("userId")
	c.Records = ParseCalibrations(parser.WithReferenceArrayParser("records"))
}

func (c *CalibrationsResponse) Validate(validator structure.Validator) {
	validator.String("recordType", c.RecordType).Exists().EqualTo(CalibrationsResponseRecordType)
	validator.String("recordVersion", c.RecordVersion).Exists().EqualTo(CalibrationsResponseRecordVersion)
	validator.String("userId", c.UserID).Exists().NotEmpty()

	// Only validate that the records exist, remaining validation will occur later on a per-record basis
	if c.Records == nil {
		validator.WithReference("records").ReportError(structureValidator.ErrorValueNotExists())
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
	RecordID              *string  `json:"recordId,omitempty"`
	SystemTime            *Time    `json:"systemTime,omitempty"`
	DisplayTime           *Time    `json:"displayTime,omitempty"`
	Unit                  *string  `json:"unit,omitempty"`
	Value                 *float64 `json:"value,omitempty"`
	TransmitterGeneration *string  `json:"transmitterGeneration,omitempty"`
	TransmitterID         *string  `json:"transmitterId,omitempty"`
	TransmitterTicks      *int     `json:"transmitterTicks,omitempty"`
	DisplayDevice         *string  `json:"displayDevice,omitempty"`
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
	parser = parser.WithMeta(c)

	c.RecordID = parser.String("recordId")
	c.SystemTime = ParseTime(parser, "systemTime")
	c.DisplayTime = ParseTime(parser, "displayTime")
	c.Unit = ParseStringOrDefault(parser, "unit", CalibrationUnitDefault)
	c.Value = parser.Float64("value")
	c.TransmitterGeneration = parser.String("transmitterGeneration")
	c.TransmitterID = parser.String("transmitterId")
	c.TransmitterTicks = parser.Int("transmitterTicks")
	c.DisplayDevice = parser.String("displayDevice")
}

func (c *Calibration) Validate(validator structure.Validator) {
	validator = validator.WithMeta(c)

	validator.String("recordId", c.RecordID).Exists().NotEmpty()
	validator.Time("systemTime", c.SystemTime.Raw()).Exists().NotZero().BeforeNow(SystemTimeNowThreshold)
	validator.Time("displayTime", c.DisplayTime.Raw()).Exists().NotZero()
	validator.String("unit", c.Unit).Exists().OneOf(CalibrationUnits()...)
	if validator := validator.Float64("value", c.Value); c.Unit != nil {
		switch *c.Unit {
		case CalibrationUnitUnknown:
			validator.Exists()
		case CalibrationUnitMgdL:
			validator.Exists().InRange(CalibrationValueMgdLMinimum, CalibrationValueMgdLMaximum)
		case CalibrationUnitMmolL:
			validator.Exists().InRange(CalibrationValueMmolLMinimum, CalibrationValueMmolLMaximum)
		}
	}
	validator.String("transmitterGeneration", c.TransmitterGeneration).Exists().OneOf(DeviceTransmitterGenerations()...)
	validator.String("transmitterId", c.TransmitterID).Exists().Using(TransmitterIDValidator)
	validator.Int("transmitterTicks", c.TransmitterTicks).Exists().GreaterThanOrEqualTo(CalibrationTransmitterTickMinimum)
	validator.String("displayDevice", c.DisplayDevice).Exists().OneOf(DeviceDisplayDevices()...)

	// Log various warnings
	logger := validator.Logger().WithField("meta", c)
	if c.Unit != nil && *c.Unit == CalibrationUnitUnknown {
		logger.Warnf("Unit is '%s'", *c.Unit)
	}
	if c.TransmitterGeneration != nil && *c.TransmitterGeneration == DeviceTransmitterGenerationUnknown {
		logger.Warnf("TransmitterGeneration is '%s'", *c.TransmitterGeneration)
	}
	if c.TransmitterID != nil && *c.TransmitterID == "" {
		logger.Warnf("TransmitterID is empty", *c.TransmitterID)
	}
	if c.TransmitterTicks != nil && *c.TransmitterTicks == CalibrationTransmitterTickMinimum {
		logger.Warnf("TransmitterTicks is %d", *c.TransmitterTicks, CalibrationTransmitterTickMinimum)
	}
	if c.DisplayDevice != nil && *c.DisplayDevice == DeviceDisplayDeviceUnknown {
		logger.Warnf("DisplayDevice is '%s'", *c.DisplayDevice)
	}
}
