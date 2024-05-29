package cgm

import (
	"regexp"
	"sort"

	"github.com/tidepool-org/platform/data"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	dataTypes "github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	Type = "cgmSettings"

	FirmwareVersionLengthMaximum  = 100
	HardwareVersionLengthMaximum  = 100
	ManufacturerLengthMaximum     = 100
	ManufacturersLengthMaximum    = 10
	ModelLengthMaximum            = 100
	NameLengthMaximum             = 100
	SerialNumberLengthMaximum     = 100
	SoftwareVersionLengthMaximum  = 100
	TransmitterIDExpressionString = "^[0-9a-zA-Z]{5,64}$"
)

type CGM struct {
	dataTypes.Base `bson:",inline"`

	FirmwareVersion *string   `json:"firmwareVersion,omitempty" bson:"firmwareVersion,omitempty"`
	HardwareVersion *string   `json:"hardwareVersion,omitempty" bson:"hardwareVersion,omitempty"`
	Manufacturers   *[]string `json:"manufacturers,omitempty" bson:"manufacturers,omitempty"`
	Model           *string   `json:"model,omitempty" bson:"model,omitempty"`
	Name            *string   `json:"name,omitempty" bson:"name,omitempty"`
	SerialNumber    *string   `json:"serialNumber,omitempty" bson:"serialNumber,omitempty"`
	SoftwareVersion *string   `json:"softwareVersion,omitempty" bson:"softwareVersion,omitempty"`
	TransmitterID   *string   `json:"transmitterId,omitempty" bson:"transmitterId,omitempty"`
	Units           *string   `json:"units,omitempty" bson:"units,omitempty"`

	DefaultAlerts   *Alerts          `json:"defaultAlerts,omitempty" bson:"defaultAlerts,omitempty"`
	ScheduledAlerts *ScheduledAlerts `json:"scheduledAlerts,omitempty" bson:"scheduledAlerts,omitempty"`

	// FUTURE: DEPRECATED
	HighLevelAlert  *HighLevelAlertDEPRECATED  `json:"highAlerts,omitempty" bson:"highAlerts,omitempty"`               // FUTURE: Migrate to DefaultAlerts
	LowLevelAlert   *LowLevelAlertDEPRECATED   `json:"lowAlerts,omitempty" bson:"lowAlerts,omitempty"`                 // FUTURE: Migrate to DefaultAlerts
	OutOfRangeAlert *OutOfRangeAlertDEPRECATED `json:"outOfRangeAlerts,omitempty" bson:"outOfRangeAlerts,omitempty"`   // FUTURE: Migrate to DefaultAlerts
	RateAlerts      *RateAlertsDEPRECATED      `json:"rateOfChangeAlert,omitempty" bson:"rateOfChangeAlert,omitempty"` // FUTURE: Migrate to DefaultAlerts
}

func New() *CGM {
	return &CGM{
		Base: dataTypes.New(Type),
	}
}

func (c *CGM) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(c.Meta())
	}

	c.Base.Parse(parser)

	c.FirmwareVersion = parser.String("firmwareVersion")
	c.HardwareVersion = parser.String("hardwareVersion")
	c.Manufacturers = parser.StringArray("manufacturers")
	c.Model = parser.String("model")
	c.Name = parser.String("name")
	c.SerialNumber = parser.String("serialNumber")
	c.SoftwareVersion = parser.String("softwareVersion")
	c.TransmitterID = parser.String("transmitterId")
	c.Units = parser.String("units")

	c.DefaultAlerts = ParseAlerts(parser.WithReferenceObjectParser("defaultAlerts"))
	c.ScheduledAlerts = ParseScheduledAlerts(parser.WithReferenceArrayParser("scheduledAlerts"))

	c.HighLevelAlert = ParseHighLevelAlertDEPRECATED(parser.WithReferenceObjectParser("highAlerts"))
	c.LowLevelAlert = ParseLowLevelAlertDEPRECATED(parser.WithReferenceObjectParser("lowAlerts"))
	c.OutOfRangeAlert = ParseOutOfRangeAlertDEPRECATED(parser.WithReferenceObjectParser("outOfRangeAlerts"))
	c.RateAlerts = ParseRateAlertsDEPRECATED(parser.WithReferenceObjectParser("rateOfChangeAlerts"))
}

func (c *CGM) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(c.Meta())
	}

	c.Base.Validate(validator)

	if c.Type != "" {
		validator.String("type", &c.Type).EqualTo(Type)
	}

	validator.String("firmwareVersion", c.FirmwareVersion).NotEmpty().LengthLessThanOrEqualTo(FirmwareVersionLengthMaximum)
	validator.String("hardwareVersion", c.HardwareVersion).NotEmpty().LengthLessThanOrEqualTo(HardwareVersionLengthMaximum)
	validator.StringArray("manufacturers", c.Manufacturers).NotEmpty().LengthLessThanOrEqualTo(ManufacturersLengthMaximum).Each(func(stringValidator structure.String) {
		stringValidator.Exists().NotEmpty().LengthLessThanOrEqualTo(ManufacturerLengthMaximum)
	}).EachUnique()
	validator.String("model", c.Model).NotEmpty().LengthLessThanOrEqualTo(ModelLengthMaximum)
	validator.String("name", c.Name).NotEmpty().LengthLessThanOrEqualTo(NameLengthMaximum)
	validator.String("serialNumber", c.SerialNumber).NotEmpty().LengthLessThanOrEqualTo(SerialNumberLengthMaximum)
	validator.String("softwareVersion", c.SoftwareVersion).NotEmpty().LengthLessThanOrEqualTo(SoftwareVersionLengthMaximum)
	validator.String("transmitterId", c.TransmitterID).Using(TransmitterIDValidator)
	validator.String("units", c.Units).OneOf(dataBloodGlucose.Units()...) // FUTURE: Use locally defined Units

	if c.DefaultAlerts != nil {
		c.DefaultAlerts.Validate(validator.WithReference("defaultAlerts"))
	}
	if c.ScheduledAlerts != nil {
		c.ScheduledAlerts.Validate(validator.WithReference("scheduledAlerts"))
	}

	if c.HighLevelAlert != nil {
		c.HighLevelAlert.Validate(validator.WithReference("highAlerts"), c.Units)
	}
	if c.LowLevelAlert != nil {
		c.LowLevelAlert.Validate(validator.WithReference("lowAlerts"), c.Units)
	}
	if c.OutOfRangeAlert != nil {
		c.OutOfRangeAlert.Validate(validator.WithReference("outOfRangeAlerts"))
	}
	if c.RateAlerts != nil {
		c.RateAlerts.Validate(validator.WithReference("rateOfChangeAlerts"), c.Units)
	}
}

func (c *CGM) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(c.Meta())
	}

	c.Base.Normalize(normalizer)

	units := c.Units

	if normalizer.Origin() == structure.OriginExternal {
		if c.Manufacturers != nil {
			sort.Strings(*c.Manufacturers)
		}

		c.Units = dataBloodGlucose.NormalizeUnits(c.Units) // FUTURE: Do not normalize units after deprecated fields deleted
	}

	if c.HighLevelAlert != nil {
		c.HighLevelAlert.Normalize(normalizer.WithReference("highAlerts"), units)
	}
	if c.LowLevelAlert != nil {
		c.LowLevelAlert.Normalize(normalizer.WithReference("lowAlerts"), units)
	}
	if c.OutOfRangeAlert != nil {
		c.OutOfRangeAlert.Normalize(normalizer.WithReference("outOfRangeAlerts"))
	}
	if c.RateAlerts != nil {
		c.RateAlerts.Normalize(normalizer.WithReference("rateOfChangeAlerts"), units)
	}
}

func (c *CGM) LegacyIdentityFields() ([]string, error) {
	return dataTypes.NewLegacyIdentityBuilder(&c.Base, dataTypes.TypeTimeDeviceIDFormat).Build()
}

func IsValidTransmitterID(value string) bool {
	return ValidateTransmitterID(value) == nil
}

func TransmitterIDValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateTransmitterID(value))
}

func ValidateTransmitterID(value string) error {
	if value == "" {
		// transmitterId is no longer guaranteed from dexcom
		return nil
	} else if !transmitterIDExpression.MatchString(value) {
		return ErrorValueStringAsTransmitterIDNotValid(value)
	}
	return nil
}

func ErrorValueStringAsTransmitterIDNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as transmitter id", value)
}

var transmitterIDExpression = regexp.MustCompile(TransmitterIDExpressionString)
