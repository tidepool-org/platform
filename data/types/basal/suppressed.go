package basal

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	RateMaximum = 100.0
	RateMinimum = 0.0
)

// TODO: Separate into distinct DeliveryType Suppressed types (eg. TemporarySuppressed, ScheduledSuppressed)

type Suppressed struct {
	Type         *string `json:"type,omitempty" bson:"type,omitempty"`
	DeliveryType *string `json:"deliveryType,omitempty" bson:"deliveryType,omitempty"`

	Annotations  *data.BlobArray `json:"annotations,omitempty" bson:"annotations,omitempty"`
	Rate         *float64        `json:"rate,omitempty" bson:"rate,omitempty"`
	ScheduleName *string         `json:"scheduleName,omitempty" bson:"scheduleName,omitempty"`
	Suppressed   *Suppressed     `json:"suppressed,omitempty" bson:"suppressed,omitempty"`
}

func ParseSuppressed(parser data.ObjectParser) *Suppressed {
	if parser.Object() == nil {
		return nil
	}
	suppressed := NewSuppressed()
	suppressed.Parse(parser)
	parser.ProcessNotParsed()
	return suppressed
}

func NewSuppressed() *Suppressed {
	return &Suppressed{}
}

func (s *Suppressed) Parse(parser data.ObjectParser) {
	s.Type = parser.ParseString("type")
	s.DeliveryType = parser.ParseString("deliveryType")

	s.Annotations = data.ParseBlobArray(parser.NewChildArrayParser("annotations"))
	s.Rate = parser.ParseFloat("rate")
	s.ScheduleName = parser.ParseString("scheduleName")
	s.Suppressed = ParseSuppressed(parser.NewChildObjectParser("suppressed"))
}

func (s *Suppressed) Validate(validator structure.Validator, allowedDeliveryTypes *[]string) {
	validator.String("type", s.Type).Exists().EqualTo("basal")
	validator.String("deliveryType", s.DeliveryType).Exists().OneOf(*allowedDeliveryTypes...)

	if s.Annotations != nil {
		s.Annotations.Validate(validator.WithReference("annotations"))
	}
	validator.Float64("rate", s.Rate).Exists().InRange(RateMinimum, RateMaximum)
	if s.DeliveryType != nil {
		if suppressedDeliveryTypes, allowed := FindAndRemoveDeliveryType(*allowedDeliveryTypes, *s.DeliveryType); allowed {
			scheduleNameValidator := validator.String("scheduleName", s.ScheduleName)
			suppressedValidator := validator.WithReference("suppressed")
			if *s.DeliveryType == "scheduled" {
				scheduleNameValidator.NotEmpty()
				if s.Suppressed != nil {
					suppressedValidator.ReportError(structureValidator.ErrorValueExists())
				}
			} else {
				scheduleNameValidator.NotExists()
				if s.Suppressed != nil {
					s.Suppressed.Validate(suppressedValidator, &suppressedDeliveryTypes)
				} else {
					suppressedValidator.ReportError(structureValidator.ErrorValueNotExists())
				}
			}
		}
	}
}

func (s *Suppressed) Normalize(normalizer data.Normalizer) {
	if s.Suppressed != nil {
		s.Suppressed.Normalize(normalizer.WithReference("suppressed"))
	}
}

func FindAndRemoveDeliveryType(deliveryTypes []string, deliveryType string) ([]string, bool) {
	if len(deliveryTypes) == 0 {
		return deliveryTypes, false
	}
	result := []string{}
	found := false
	for _, dt := range deliveryTypes {
		if dt != deliveryType {
			result = append(result, dt)
		} else {
			found = true
		}
	}
	return result, found
}
