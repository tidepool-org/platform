package temporary

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
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/base/basal"
)

type Temporary struct {
	basal.Basal `bson:",inline"`

	Duration         *int     `json:"duration,omitempty" bson:"duration,omitempty"`
	ExpectedDuration *int     `json:"expectedDuration,omitempty" bson:"expectedDuration,omitempty"`
	Rate             *float64 `json:"rate,omitempty" bson:"rate,omitempty"`
	Percent          *float64 `json:"percent,omitempty" bson:"percent,omitempty"`

	Suppressed *basal.Suppressed `json:"suppressed,omitempty" bson:"suppressed,omitempty"`
}

func DeliveryType() string {
	return "temp"
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
	t.Basal.DeliveryType = DeliveryType()

	t.Duration = nil
	t.ExpectedDuration = nil
	t.Rate = nil
	t.Percent = nil

	t.Suppressed = nil
}

func (t *Temporary) Parse(parser data.ObjectParser) error {
	if err := t.Basal.Parse(parser); err != nil {
		return err
	}

	t.Duration = parser.ParseInteger("duration")
	t.ExpectedDuration = parser.ParseInteger("expectedDuration")
	t.Rate = parser.ParseFloat("rate")
	t.Percent = parser.ParseFloat("percent")

	t.Suppressed = basal.ParseSuppressed(parser.NewChildObjectParser("suppressed"))

	return nil
}

func (t *Temporary) Validate(validator data.Validator) error {
	if err := t.Basal.Validate(validator); err != nil {
		return err
	}

	validator.ValidateInteger("duration", t.Duration).Exists().InRange(0, 86400000)

	expectedDurationValidator := validator.ValidateInteger("expectedDuration", t.ExpectedDuration)
	if t.Duration != nil {
		expectedDurationValidator.InRange(*t.Duration, 86400000)
	} else {
		expectedDurationValidator.InRange(0, 86400000)
	}

	validator.ValidateFloat("rate", t.Rate).Exists().InRange(0.0, 20.0)
	validator.ValidateFloat("percent", t.Percent).InRange(0.0, 10.0)

	if t.Suppressed != nil {
		t.Suppressed.Validate(validator.NewChildValidator("suppressed"), []string{"scheduled"})
	}

	return nil
}
