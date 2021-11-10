package controller

import (
	"github.com/tidepool-org/platform/data"
	dataTypes "github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	Type = "controllerSettings"
)

type Controller struct {
	dataTypes.Base `bson:",inline"`

	Device        *Device        `json:"device,omitempty" bson:"device,omitempty"`
	Notifications *Notifications `json:"notifications,omitempty" bson:"notifications,omitempty"`
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

	c.Device = ParseDevice(parser.WithReferenceObjectParser("device"))
	c.Notifications = ParseNotifications(parser.WithReferenceObjectParser("notifications"))
}

func (c *Controller) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(c.Meta())
	}

	c.Base.Validate(validator)

	if c.Type != "" {
		validator.String("type", &c.Type).EqualTo(Type)
	}

	if c.Device != nil {
		c.Device.Validate(validator.WithReference("device"))
	}
	if c.Notifications != nil {
		c.Notifications.Validate(validator.WithReference("notifications"))
	}

	if c.Device == nil && c.Notifications == nil {
		validator.ReportError(structureValidator.ErrorValuesNotExistForAny("device", "notifications"))
	}
}

func (c *Controller) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(c.Meta())
	}

	c.Base.Normalize(normalizer)

	if c.Device != nil {
		c.Device.Normalize(normalizer.WithReference("device"))
	}
}
