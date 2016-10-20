package basal

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import (
	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
)

type Suppressed struct {
	Type         *string  `json:"type,omitempty" bson:"type,omitempty"`
	DeliveryType *string  `json:"deliveryType,omitempty" bson:"deliveryType,omitempty"`
	Rate         *float64 `json:"rate,omitempty" bson:"rate,omitempty"`
	ScheduleName *string  `json:"scheduleName,omitempty" bson:"scheduleName,omitempty"`

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

	s.Suppressed = ParseSuppressed(parser.NewChildObjectParser("suppressed"))
}

func (s *Suppressed) Validate(validator data.Validator, allowedDeliveryTypes []string) {
	validator.ValidateString("type", s.Type).Exists().EqualTo("basal")
	validator.ValidateString("deliveryType", s.DeliveryType).Exists().OneOf(allowedDeliveryTypes)
	validator.ValidateFloat("rate", s.Rate).Exists().InRange(0.0, 100.0)

	if s.DeliveryType != nil && app.StringArrayContains(allowedDeliveryTypes, *s.DeliveryType) {
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
}

func suppressedAsInterface(suppressed *Suppressed) *interface{} {
	if suppressed == nil {
		return nil
	}
	var value interface{} = *suppressed
	return &value
}
