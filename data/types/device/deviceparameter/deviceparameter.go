package deviceparameter

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/structure"
)

const (
	SubType                  = "deviceParameter"
	LastUpdateDateTimeFormat = "2006-01-02T15:04:05.000Z"
	ProcessedYes             = "yes"
	ProcessedNo              = "no"
)

func ProcessedValues() []string {
	return []string{
		ProcessedYes,
		ProcessedNo,
	}
}
func LevelValues() []string {
	return []string{"1", "2", "3"}
}

type DeviceParameter struct {
	device.Device  `bson:",inline"`
	Name           *string   `json:"name" bson:"name"`
	Value          *string   `json:"value" bson:"value"`
	Units          *string   `json:"units,omitempty" bson:"units,omitempty"`
	LastUpdateDate *string   `json:"lastUpdateDate" bson:"lastUpdateDate"`
	PreviousValue  *string   `json:"previousValue,omitempty" bson:"previousValue,omitempty"`
	Level          *string   `json:"level,omitempty" bson:"level,omitempty"`
	MinValue       *string   `json:"minValue,omitempty" bson:"minValue,omitempty"`
	MaxValue       *string   `json:"maxValue,omitempty" bson:"maxValue,omitempty"`
	Processed      *string   `json:"processed,omitempty" bson:"processed,omitempty"`
	LinkedSubType  *[]string `json:"linkedSubType,omitempty" bson:"linkedSubType,omitempty"`
}

func New() *DeviceParameter {
	return &DeviceParameter{
		Device: device.New(SubType),
	}
}

func (p *DeviceParameter) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(p.Meta())
	}

	p.Device.Parse(parser)

	p.Name = parser.String("name")
	p.Value = parser.String("value")
	p.Units = parser.String("units")
	p.LastUpdateDate = parser.String("lastUpdateDate")
	p.PreviousValue = parser.String("previousValue")
	p.Level = parser.String("level")
	p.MinValue = parser.String("minValue")
	p.MaxValue = parser.String("maxValue")
	p.Processed = parser.String("processed")
	p.LinkedSubType = parser.StringArray("linkedSubType")
}

func (p *DeviceParameter) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(p.Meta())
	}

	p.Device.Validate(validator)

	if p.SubType != "" {
		validator.String("subType", &p.SubType).EqualTo(SubType)
	}
	validator.String("name", p.Name).Exists().NotEmpty()
	validator.String("value", p.Value).Exists().NotEmpty()
	validator.String("lastUpdateDate", p.LastUpdateDate).Exists().NotEmpty().AsTime(LastUpdateDateTimeFormat)
	validator.String("level", p.Level).Exists().NotEmpty().OneOf(LevelValues()...)

	if p.Processed != nil && len(*p.Processed) > 0 {
		validator.String("processed", p.Processed).Exists().OneOf(ProcessedValues()...)
		if *p.Processed == ProcessedYes {
			validator.StringArray("linkedSubType", p.LinkedSubType).Exists().NotEmpty()
		}
	}
}

// IsValid returns true if there is no error in the validator
func (p *DeviceParameter) IsValid(validator structure.Validator) bool {
	return !(validator.HasError())
}

func (p *DeviceParameter) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(p.Meta())
	}

	p.Device.Normalize(normalizer)
}
