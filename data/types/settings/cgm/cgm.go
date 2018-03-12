package cgm

import (
	"regexp"

	"github.com/tidepool-org/platform/data"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	Type = "cgmSettings"

	TransmitterIDExpressionString = "^[0-9A-Z]{5,6}$"
)

type CGM struct {
	types.Base `bson:",inline"`

	HighLevelAlert  *HighLevelAlert  `json:"highAlerts,omitempty" bson:"highAlerts,omitempty"`               // TODO: Rename highLevelAlert
	LowLevelAlert   *LowLevelAlert   `json:"lowAlerts,omitempty" bson:"lowAlerts,omitempty"`                 // TODO: Rename lowLevelAlert
	OutOfRangeAlert *OutOfRangeAlert `json:"outOfRangeAlerts,omitempty" bson:"outOfRangeAlerts,omitempty"`   // TODO: Rename outOfRangeAlert
	RateAlerts      *RateAlerts      `json:"rateOfChangeAlert,omitempty" bson:"rateOfChangeAlert,omitempty"` // TODO: Split into separate fallRateAlert, riseRateAlert
	TransmitterID   *string          `json:"transmitterId,omitempty" bson:"transmitterId,omitempty"`
	Units           *string          `json:"units,omitempty" bson:"units,omitempty"`
}

func New() *CGM {
	return &CGM{
		Base: types.New(Type),
	}
}

func (c *CGM) Parse(parser data.ObjectParser) error {
	parser.SetMeta(c.Meta())

	if err := c.Base.Parse(parser); err != nil {
		return err
	}

	c.HighLevelAlert = ParseHighLevelAlert(parser.NewChildObjectParser("highAlerts"))
	c.LowLevelAlert = ParseLowLevelAlert(parser.NewChildObjectParser("lowAlerts"))
	c.OutOfRangeAlert = ParseOutOfRangeAlert(parser.NewChildObjectParser("outOfRangeAlerts"))
	c.RateAlerts = ParseRateAlerts(parser.NewChildObjectParser("rateOfChangeAlerts"))
	c.TransmitterID = parser.ParseString("transmitterId")
	c.Units = parser.ParseString("units")

	return nil
}

func (c *CGM) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(c.Meta())
	}

	c.Base.Validate(validator)

	if c.Type != "" {
		validator.String("type", &c.Type).EqualTo(Type)
	}

	if c.HighLevelAlert != nil {
		c.HighLevelAlert.Validate(validator.WithReference("highAlerts"), c.Units)
	} else {
		validator.WithReference("highAlerts").ReportError(structureValidator.ErrorValueNotExists())
	}
	if c.LowLevelAlert != nil {
		c.LowLevelAlert.Validate(validator.WithReference("lowAlerts"), c.Units)
	} else {
		validator.WithReference("lowAlerts").ReportError(structureValidator.ErrorValueNotExists())
	}
	if c.OutOfRangeAlert != nil {
		c.OutOfRangeAlert.Validate(validator.WithReference("outOfRangeAlerts"))
	} else {
		validator.WithReference("outOfRangeAlerts").ReportError(structureValidator.ErrorValueNotExists())
	}
	if c.RateAlerts != nil {
		c.RateAlerts.Validate(validator.WithReference("rateOfChangeAlerts"), c.Units)
	} else {
		validator.WithReference("rateOfChangeAlerts").ReportError(structureValidator.ErrorValueNotExists())
	}
	validator.String("transmitterId", c.TransmitterID).Exists().Using(ValidateTransmitterID)
	validator.String("units", c.Units).Exists().OneOf(dataBloodGlucose.Units()...)
}

func (c *CGM) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(c.Meta())
	}

	c.Base.Normalize(normalizer)

	if c.HighLevelAlert != nil {
		c.HighLevelAlert.Normalize(normalizer.WithReference("highAlerts"), c.Units)
	}
	if c.LowLevelAlert != nil {
		c.LowLevelAlert.Normalize(normalizer.WithReference("lowAlerts"), c.Units)
	}
	if c.OutOfRangeAlert != nil {
		c.OutOfRangeAlert.Normalize(normalizer.WithReference("outOfRangeAlerts"))
	}
	if c.RateAlerts != nil {
		c.RateAlerts.Normalize(normalizer.WithReference("rateOfChangeAlerts"), c.Units)
	}
	if normalizer.Origin() == structure.OriginExternal {
		c.Units = dataBloodGlucose.NormalizeUnits(c.Units)
	}
}

var transmitterIDExpression = regexp.MustCompile(TransmitterIDExpressionString)

func ValidateTransmitterID(value string, errorReporter structure.ErrorReporter) {
	if value == "" {
		errorReporter.ReportError(structureValidator.ErrorValueEmpty())
	} else if !transmitterIDExpression.MatchString(value) {
		errorReporter.ReportError(ErrorValueStringAsTransmitterIDNotValid(value))
	}
}

func ErrorValueStringAsTransmitterIDNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as transmitter id", value)
}
