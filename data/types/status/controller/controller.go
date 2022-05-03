package controller

import (
	"github.com/tidepool-org/platform/data"
	dataTypes "github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	Type = "controllerStatus"
)

type Controller struct {
	dataTypes.Base `bson:",inline"`

	Battery *Battery `json:"battery,omitempty" bson:"battery,omitempty"`
}

func New() *Controller {
	return &Controller{
		Base: dataTypes.New(Type),
	}
}

func (c *Controller) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(c.Meta())
	}

	c.Base.Parse(parser)

	c.Battery = ParseBattery(parser.WithReferenceObjectParser("battery"))
}

func (c *Controller) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(c.Meta())
	}

	c.Base.Validate(validator)

	if c.Type != "" {
		validator.String("type", &c.Type).EqualTo(Type)
	}

	if batteryValidator := validator.WithReference("battery"); c.Battery != nil {
		c.Battery.Validate(batteryValidator)
	} else {
		batteryValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
}

func (c *Controller) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(c.Meta())
	}

	c.Base.Normalize(normalizer)
}
