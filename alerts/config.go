package alerts

// Data models for care team alerts.

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/user"
)

// Config models a user's desired alerts.
type Config struct {
	// UserID receives the alerts, and owns this Config.
	UserID string `json:"userId" bson:"userId"`
	// FollowedID is the user whose data generates alerts, and has granted
	// UserID permission to that data.
	FollowedID      string                 `json:"followedId" bson:"followedId"`
	UrgentLow       *WithThreshold         `json:"urgentLow,omitempty" bson:"urgentLow,omitempty"`
	Low             *WithDelayAndThreshold `json:"low,omitempty" bson:"low,omitempty"`
	High            *WithDelayAndThreshold `json:"high,omitempty" bson:"high,omitempty"`
	NotLooping      *WithDelay             `json:"notLooping,omitempty" bson:"notLooping,omitempty"`
	NoCommunication *WithDelay             `json:"noCommunication,omitempty" bson:"noCommunication,omitempty"`
}

func (c Config) Validate(validator structure.Validator) {
	validator.String("UserID", &c.UserID).Using(user.IDValidator)
	validator.String("FollowedID", &c.FollowedID).Using(user.IDValidator)
	if c.Low != nil {
		c.Low.Validate(validator)
	}
	if c.UrgentLow != nil {
		c.UrgentLow.Validate(validator)
	}
	if c.High != nil {
		c.High.Validate(validator)
	}
	if c.NotLooping != nil {
		c.NotLooping.Validate(validator)
	}
	if c.NoCommunication != nil {
		c.NoCommunication.Validate(validator)
	}
}

// Base describes the minimum specifics of a desired alert.
type Base struct {
	// Enabled controls whether notifications should be sent for this alert.
	Enabled bool `json:"enabled"`
	// Repeat is measured in minutes.
	Repeat DurationMinutes `json:"repeat,omitempty"`
}

func (b Base) Validate(validator structure.Validator) {
	validator.Bool("enabled", &b.Enabled)
	dur := b.Repeat.Duration()
	validator.Duration("repeat", &dur).GreaterThan(0 * time.Minute)
}

// DelayMixin adds a configurable delay.
type DelayMixin struct {
	// Delay is measured in minutes.
	Delay DurationMinutes `json:"delay,omitempty"`
}

func (d DelayMixin) Validate(validator structure.Validator) {
	dur := d.Delay.Duration()
	validator.Duration("delay", &dur).GreaterThan(0 * time.Minute)
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
	Base           `bson:",inline"`
	ThresholdMixin `bson:",inline"`
}

func (d WithThreshold) Validate(validator structure.Validator) {
	d.Base.Validate(validator)
	d.ThresholdMixin.Validate(validator)
}

// WithDelay extends Base with DelayMixin.
type WithDelay struct {
	Base       `bson:",inline"`
	DelayMixin `bson:",inline"`
}

func (d WithDelay) Validate(validator structure.Validator) {
	d.Base.Validate(validator)
	d.DelayMixin.Validate(validator)
}

// WithDelayAndThreshold extends Base with both DelayMixin and ThresholdMixin.
type WithDelayAndThreshold struct {
	Base           `bson:",inline"`
	DelayMixin     `bson:",inline"`
	ThresholdMixin `bson:",inline"`
}

func (d WithDelayAndThreshold) Validate(validator structure.Validator) {
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

func (m *DurationMinutes) MarshalJSON() ([]byte, error) {
	minutes := time.Duration(*m) / time.Minute
	return json.Marshal(minutes)
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

// Repository abstracts persistent storage for Config data.
type Repository interface {
	Get(ctx context.Context, conf *Config) (*Config, error)
	Upsert(ctx context.Context, conf *Config) error
	Delete(ctx context.Context, conf *Config) error

	EnsureIndexes() error
}
