package basal

import (
	"github.com/tidepool-org/platform/data"
)

type Suppressed struct {
	Type         *string        `json:"type,omitempty" bson:"type,omitempty"`
	DeliveryType *string        `json:"deliveryType,omitempty" bson:"deliveryType,omitempty"`
	Rate         *float64       `json:"rate,omitempty" bson:"rate,omitempty"`
	ScheduleName *string        `json:"scheduleName,omitempty" bson:"scheduleName,omitempty"`
	Annotations  *[]interface{} `json:"annotations,omitempty" bson:"annotations,omitempty"`

	Suppressed *Suppressed `json:"suppressed,omitempty" bson:"suppressed,omitempty"`
}

func ParseSuppressed(parser data.ObjectParser) *Suppressed {
	var suppressed *Suppressed
	if parser.Object() != nil {
		suppressed = NewSuppressed()
		suppressed.Parse(parser)
		parser.ProcessNotParsed()
	}
	return suppressed
}

func NewSuppressed() *Suppressed {
	return &Suppressed{}
}

func (s *Suppressed) Parse(parser data.ObjectParser) {
	s.Type = parser.ParseString("type")
	s.DeliveryType = parser.ParseString("deliveryType")
	s.Rate = parser.ParseFloat("rate")
	s.ScheduleName = parser.ParseString("scheduleName")
	s.Annotations = parser.ParseInterfaceArray("annotations")

	s.Suppressed = ParseSuppressed(parser.NewChildObjectParser("suppressed"))
}

func (s *Suppressed) Validate(validator data.Validator, allowedDeliveryTypes []string) {
	validator.ValidateString("type", s.Type).Exists().EqualTo("basal")
	validator.ValidateString("deliveryType", s.DeliveryType).Exists().OneOf(allowedDeliveryTypes)
	validator.ValidateFloat("rate", s.Rate).Exists().InRange(0.0, 100.0)

	if s.HasDeliveryTypeOneOf(allowedDeliveryTypes) {
		scheduleNameValidator := validator.ValidateString("scheduleName", s.ScheduleName)
		suppressedValidator := validator.ValidateInterface("suppressed", suppressedAsInterface(s.Suppressed))
		if *s.DeliveryType == "scheduled" {
			scheduleNameValidator.NotEmpty()
			suppressedValidator.NotExists()
		} else {
			scheduleNameValidator.NotExists()
			suppressedValidator.Exists()
			if s.Suppressed != nil {
				s.Suppressed.Validate(validator.NewChildValidator("suppressed"), []string{"scheduled"})
			}
		}
	}

	// validator.ValidateInterfaceArray("annotations", s.Annotations)    // TODO: Any validations? Optional? Size?
}

func (s *Suppressed) HasDeliveryTypeOneOf(deliveryTypes []string) bool {
	if s.DeliveryType == nil {
		return false
	}

	for _, deliveryType := range deliveryTypes {
		if deliveryType == *s.DeliveryType {
			return true
		}
	}

	return false
}

func suppressedAsInterface(suppressed *Suppressed) *interface{} {
	if suppressed == nil {
		return nil
	}
	var value interface{} = *suppressed
	return &value
}
