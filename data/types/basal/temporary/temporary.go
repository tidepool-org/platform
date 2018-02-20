package temporary

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/basal"
	"github.com/tidepool-org/platform/data/types/basal/scheduled"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
)

const (
	DurationMaximum = 604800000
	DurationMinimum = 0
	PercentMaximum  = 10.0
	PercentMinimum  = 0.0
	RateMaximum     = 100.0
	RateMinimum     = 0.0
)

func SuppressedDeliveryTypes() []string {
	return []string{
		scheduled.DeliveryType(),
	}
}

type Temporary struct {
	basal.Basal `bson:",inline"`

	Duration         *int              `json:"duration,omitempty" bson:"duration,omitempty"`
	DurationExpected *int              `json:"expectedDuration,omitempty" bson:"expectedDuration,omitempty"`
	Percent          *float64          `json:"percent,omitempty" bson:"percent,omitempty"`
	Rate             *float64          `json:"rate,omitempty" bson:"rate,omitempty"`
	Suppressed       *basal.Suppressed `json:"suppressed,omitempty" bson:"suppressed,omitempty"`
}

func DeliveryType() string {
	return "temp" // TODO: Rename Type to "basal/temporary"; remove DeliveryType
}

func NewDatum() data.Datum {
	return New()
}

func New() *Temporary {
	return &Temporary{}
}

func Init() *Temporary {
	temporary := New()
	temporary.Init()
	return temporary
}

func (t *Temporary) Init() {
	t.Basal.Init()
	t.DeliveryType = DeliveryType()

	t.Duration = nil
	t.DurationExpected = nil
	t.Percent = nil
	t.Rate = nil
	t.Suppressed = nil
}

func (t *Temporary) Parse(parser data.ObjectParser) error {
	if err := t.Basal.Parse(parser); err != nil {
		return err
	}

	t.Duration = parser.ParseInteger("duration")
	t.DurationExpected = parser.ParseInteger("expectedDuration")
	t.Percent = parser.ParseFloat("percent")
	t.Rate = parser.ParseFloat("rate")
	t.Suppressed = basal.ParseSuppressed(parser.NewChildObjectParser("suppressed"))

	return nil
}

func (t *Temporary) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(t.Meta())
	}

	t.Basal.Validate(validator)

	if t.DeliveryType != "" {
		validator.String("deliveryType", &t.DeliveryType).EqualTo(DeliveryType())
	}

	validator.Int("duration", t.Duration).Exists().InRange(DurationMinimum, DurationMaximum)
	expectedDurationValidator := validator.Int("expectedDuration", t.DurationExpected)
	if t.Duration != nil && *t.Duration >= DurationMinimum && *t.Duration <= DurationMaximum {
		expectedDurationValidator.InRange(*t.Duration, DurationMaximum)
	} else {
		expectedDurationValidator.InRange(DurationMinimum, DurationMaximum)
	}
	validator.Float64("percent", t.Percent).InRange(PercentMinimum, PercentMaximum)
	validator.Float64("rate", t.Rate).Exists().InRange(RateMinimum, RateMaximum)
	if t.Suppressed != nil {
		t.Suppressed.Validate(validator.WithReference("suppressed"), pointer.StringArray(SuppressedDeliveryTypes()))
	}
}

func (t *Temporary) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(t.Meta())
	}

	t.Basal.Normalize(normalizer)

	if t.Suppressed != nil {
		t.Suppressed.Normalize(normalizer.WithReference("suppressed"))
	}
}
