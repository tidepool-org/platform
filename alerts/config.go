package alerts

// Data models for care team alerts.

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
)

// Config wraps Alerts to include user relationships.
//
// As a wrapper type, Config provides a clear demarcation of what a user
// controls (Alerts) and what is set by the service (the other values in
// Config).
type Config struct {
	// UserID receives the configured alerts and owns this Config.
	UserID string `json:"userId" bson:"userId"`

	// FollowedID is the user whose data generates alerts, and has granted
	// UserID permission to that data.
	FollowedID string `json:"followedId" bson:"followedId"`

	Alerts `bson:",inline,omitempty"`
}

// Alerts models a user's desired alerts.
type Alerts struct {
	UrgentLow       *WithThreshold         `json:"urgentLow,omitempty" bson:"urgentLow,omitempty"`
	Low             *WithDelayAndThreshold `json:"low,omitempty" bson:"low,omitempty"`
	High            *WithDelayAndThreshold `json:"high,omitempty" bson:"high,omitempty"`
	NotLooping      *WithDelay             `json:"notLooping,omitempty" bson:"notLooping,omitempty"`
	NoCommunication *WithDelay             `json:"noCommunication,omitempty" bson:"noCommunication,omitempty"`
}

func (c Config) Validate(validator structure.Validator) {
	validator.String("UserID", &c.UserID).Using(user.IDValidator)
	validator.String("FollowedID", &c.FollowedID).Using(user.IDValidator)
	c.Alerts.Validate(validator)
}

func (i Alerts) Validate(validator structure.Validator) {
	if i.Low != nil {
		i.Low.Validate(validator)
	}
	if i.UrgentLow != nil {
		i.UrgentLow.Validate(validator)
	}
	if i.High != nil {
		i.High.Validate(validator)
	}
	if i.NotLooping != nil {
		i.NotLooping.Validate(validator)
	}
	if i.NoCommunication != nil {
		i.NoCommunication.Validate(validator)
	}
}

// Base describes the minimum specifics of a desired alert.
type Base struct {
	// Enabled controls whether notifications should be sent for this alert.
	Enabled bool `json:"enabled"`
	// Repeat is measured in minutes.
	//
	// A value of 0 (the default) disables repeat notifications.
	Repeat DurationMinutes `json:"repeat,omitempty"`
}

func (b Base) Validate(validator structure.Validator) {
	validator.Bool("enabled", &b.Enabled)
	dur := b.Repeat.Duration()
	validator.Duration("repeat", &dur).Using(validateRepeat)
}

const (
	// RepeatMin is the minimum duration for a repeat setting (if not 0).
	RepeatMin = 15 * time.Minute
	// RepeatMax is the maximum duration for a repeat setting.
	RepeatMax = 4 * time.Hour
)

func validateRepeat(value time.Duration, errorReporter structure.ErrorReporter) {
	if value == 0 {
		return
	}
	if value < RepeatMin {
		errorReporter.ReportError(validator.ErrorValueNotGreaterThanOrEqualTo(value, RepeatMin))
	}
	if value > RepeatMax {
		errorReporter.ReportError(validator.ErrorValueNotLessThanOrEqualTo(value, RepeatMax))
	}
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
