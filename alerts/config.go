package alerts

// Data models for care team alerts.

import (
	"bytes"
	"time"

	"github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/structure"
)

// Config models a user's desired alerts.
type Config struct {
	Key string `json:"key" bson:"_id"`
	// OwnerID links the user that owns and controls these alerts, i.e. the
	// care team member.
	OwnerID string `json:"ownerID"`
	// InvitorID links the user whose data is shared, and will trigger alerts.
	InvitorID string `json:"invitorID"`

	UrgentLow       *WithThreshold `json:"urgentLow,omitempty"`
	Low             *Deluxe        `json:"low,omitempty"`
	High            *Deluxe        `json:"high,omitempty"`
	NotLooping      *WithDelay     `json:"notLooping,omitempty"`
	NoCommunication *WithDelay     `json:"noCommunication,omitempty"`
}

// Base describes the minimum specifics of a desired alert.
type Base struct {
	// Enabled controls whether notifications should be sent for this alert.
	Enabled bool
	// Repeat is measured in minutes.
	Repeat DurationMinutes `json:"repeat"`
}

func (b Base) Validate(validator structure.Validator) {
	validator.Bool("enabled", &b.Enabled)
	dur := b.Repeat.Duration()
	validator.Duration("repeat", &dur)
}

// DelayMixin adds a configurable delay.
type DelayMixin struct {
	// Delay is measured in minutes.
	Delay DurationMinutes `json:"delay,omitempty"`
}

func (d DelayMixin) Validate(validator structure.Validator) {
	dur := d.Delay.Duration()
	validator.Duration("delay", &dur)
}

// ThresholdMixin adds a configurable threshold.
type ThresholdMixin struct {
	// Threshold is compared the current value to determine if an alert should
	// be triggered.
	Threshold `json:"threshold"`
}

func (t ThresholdMixin) Validate(validator structure.Validator) {
	t.Threshold.Validate(validator)
}

// WithThreshold extends Base with ThresholdMixin.
type WithThreshold struct {
	Base
	ThresholdMixin
}

// WithDelay extends Base with DelayMixin.
type WithDelay struct {
	Base
	DelayMixin
}

// Deluxe extends Base with both DelayMixin and ThresholdMixin.
type Deluxe struct {
	Base
	DelayMixin
	ThresholdMixin
}

func (d Deluxe) Validate(validator structure.Validator) {
	d.Base.Validate(validator)
	d.DelayMixin.Validate(validator)
	d.ThresholdMixin.Validate(validator)
}

// DurationMinutes reads a JSON integer and converts it to a time.Duration.
//
// Values are specified in minutes.
type DurationMinutes time.Duration

func (m *DurationMinutes) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) || len(b) == 0 {
		*m = DurationMinutes(0)
		return nil
	}
	d, err := time.ParseDuration(string(b) + "m")
	if err != nil {
		return err
	}
	*m = DurationMinutes(d)
	return nil
}

func (m DurationMinutes) Duration() time.Duration {
	return time.Duration(m)
}

// ValueWithUnits binds a value to its units.
//
// Other types can extend it to parse and validate the Units.
type ValueWithUnits struct {
	Value float64 `json:"value"`
	Units string  `json:"units"`
}

// Threshold is a value measured in either mg/dL or mmol/L.
type Threshold ValueWithUnits

// Validate implements structure.Validatable
func (t Threshold) Validate(validator structure.Validator) {
	validator.String("units", &t.Units).OneOf(glucose.MgdL, glucose.MmolL)
}
