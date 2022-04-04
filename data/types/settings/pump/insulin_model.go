package pump

import (
	"github.com/tidepool-org/platform/structure"
)

const (
	InsulinModelActionDelayMaximum          = 86400
	InsulinModelActionDelayMinimum          = 0
	InsulinModelActionDurationMaximum       = 86400
	InsulinModelActionDurationMinimum       = 0
	InsulinModelActionPeakOffsetMaximum     = InsulinModelActionDurationMaximum
	InsulinModelActionPeakOffsetMinimum     = 0
	InsulinModelModelTypeFiasp              = "fiasp"
	InsulinModelModelTypeOther              = "other"
	InsulinModelModelTypeOtherLengthMaximum = 100
	InsulinModelModelTypeRapidAdult         = "rapidAdult"
	InsulinModelModelTypeRapidChild         = "rapidChild"
	InsulinModelModelTypeWalsh              = "walsh"
)

func InsulinModelModelTypes() []string {
	return []string{
		InsulinModelModelTypeFiasp,
		InsulinModelModelTypeOther,
		InsulinModelModelTypeRapidAdult,
		InsulinModelModelTypeRapidChild,
		InsulinModelModelTypeWalsh,
	}
}

type InsulinModel struct {
	ModelType        *string `json:"modelType,omitempty" bson:"modelType,omitempty"`
	ModelTypeOther   *string `json:"modelTypeOther,omitempty" bson:"modelTypeOther,omitempty"`
	ActionDelay      *int    `json:"actionDelay,omitempty" bson:"actionDelay,omitempty"`
	ActionDuration   *int    `json:"actionDuration,omitempty" bson:"actionDuration,omitempty"`
	ActionPeakOffset *int    `json:"actionPeakOffset,omitempty" bson:"actionPeakOffset,omitempty"`
}

func ParseInsulinModel(parser structure.ObjectParser) *InsulinModel {
	if !parser.Exists() {
		return nil
	}
	datum := NewInsulinModel()
	parser.Parse(datum)
	return datum
}

func NewInsulinModel() *InsulinModel {
	return &InsulinModel{}
}

func (i *InsulinModel) Parse(parser structure.ObjectParser) {
	i.ModelType = parser.String("modelType")
	i.ModelTypeOther = parser.String("modelTypeOther")
	i.ActionDelay = parser.Int("actionDelay")
	i.ActionDuration = parser.Int("actionDuration")
	i.ActionPeakOffset = parser.Int("actionPeakOffset")
}

func (i *InsulinModel) Validate(validator structure.Validator) {
	validator.String("modelType", i.ModelType).OneOf(InsulinModelModelTypes()...)
	if i.ModelType != nil && *i.ModelType == InsulinModelModelTypeOther {
		validator.String("modelTypeOther", i.ModelTypeOther).Exists().NotEmpty().LengthLessThanOrEqualTo(InsulinModelModelTypeOtherLengthMaximum)
	} else {
		validator.String("modelTypeOther", i.ModelTypeOther).NotExists()
	}
	validator.Int("actionDelay", i.ActionDelay).InRange(InsulinModelActionDelayMinimum, InsulinModelActionDelayMaximum)
	validator.Int("actionDuration", i.ActionDuration).InRange(InsulinModelActionDurationMinimum, InsulinModelActionDurationMaximum)
	actionPeakOffsetValidator := validator.Int("actionPeakOffset", i.ActionPeakOffset)
	if i.ActionDuration != nil && *i.ActionDuration >= InsulinModelActionDurationMinimum && *i.ActionDuration <= InsulinModelActionDurationMaximum {
		actionPeakOffsetValidator.InRange(InsulinModelActionPeakOffsetMinimum, *i.ActionDuration)
	} else {
		actionPeakOffsetValidator.InRange(InsulinModelActionPeakOffsetMinimum, InsulinModelActionPeakOffsetMaximum)
	}
}
