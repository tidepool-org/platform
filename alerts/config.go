package alerts

// Data models for care team alerts.

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/tidepool-org/platform/data"
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

	// FollowedUserID is the user whose data generates alerts, and has granted
	// UserID permission to that data.
	FollowedUserID string `json:"followedUserId" bson:"followedUserId"`

	// UploadID identifies the device dataset for which these alerts apply.
	UploadID string `json:"uploadId" bson:"uploadId,omitempty"`

	Alerts `bson:",inline,omitempty"`
}

// Alerts models a user's desired alerts.
type Alerts struct {
	UrgentLow       *UrgentLowAlert       `json:"urgentLow,omitempty" bson:"urgentLow,omitempty"`
	Low             *LowAlert             `json:"low,omitempty" bson:"low,omitempty"`
	High            *HighAlert            `json:"high,omitempty" bson:"high,omitempty"`
	NotLooping      *NotLoopingAlert      `json:"notLooping,omitempty" bson:"notLooping,omitempty"`
	NoCommunication *NoCommunicationAlert `json:"noCommunication,omitempty" bson:"noCommunication,omitempty"`
}

func (c Config) Validate(validator structure.Validator) {
	validator.String("userID", &c.UserID).Using(user.IDValidator)
	validator.String("followedUserID", &c.FollowedUserID).Using(user.IDValidator)
	validator.String("uploadID", &c.UploadID).Exists().Using(data.SetIDValidator)
	c.Alerts.Validate(validator)
}

func (a Alerts) Validate(validator structure.Validator) {
	if a.UrgentLow != nil {
		a.UrgentLow.Validate(validator)
	}
	if a.Low != nil {
		a.Low.Validate(validator)
	}
	if a.High != nil {
		a.High.Validate(validator)
	}
	if a.NotLooping != nil {
		a.NotLooping.Validate(validator)
	}
	if a.NoCommunication != nil {
		a.NoCommunication.Validate(validator)
	}
}

// Base describes the minimum specifics of a desired alert.
type Base struct {
	// Enabled controls whether notifications should be sent for this alert.
	Enabled bool `json:"enabled" bson:"enabled"`
}

func (b Base) Validate(validator structure.Validator) {
	validator.Bool("enabled", &b.Enabled)
}

const (
	// RepeatMin is the minimum duration for a repeat setting (if not 0).
	RepeatMin = 15 * time.Minute
	// RepeatMax is the maximum duration for a repeat setting.
	RepeatMax = 4 * time.Hour
	// RepeatDisabled specifies that a repeat is not desired.
	RepeatDisabled = 0 * time.Second
)

func validateRepeat(value time.Duration, errorReporter structure.ErrorReporter) {
	if value == RepeatDisabled {
		return
	}
	if value < RepeatMin {
		errorReporter.ReportError(validator.ErrorValueNotGreaterThanOrEqualTo(value, RepeatMin))
	}
	if value > RepeatMax {
		errorReporter.ReportError(validator.ErrorValueNotLessThanOrEqualTo(value, RepeatMax))
	}
}

// UrgentLowAlert extends Base with a threshold.
type UrgentLowAlert struct {
	Base `bson:",inline"`
	// Threshold is compared the current value to determine if an alert should
	// be triggered.
	Threshold `json:"threshold" bson:"threshold"`
}

func (a UrgentLowAlert) Validate(validator structure.Validator) {
	a.Base.Validate(validator)
	a.Threshold.Validate(validator)
}

// NotLoopingAlert extends Base with a delay.
type NotLoopingAlert struct {
	Base  `bson:",inline"`
	Delay DurationMinutes `json:"delay,omitempty"`
}

func (a NotLoopingAlert) Validate(validator structure.Validator) {
	a.Base.Validate(validator)
	dur := a.Delay.Duration()
	validator.Duration("delay", &dur).InRange(0, 2*time.Hour)
}

// NoCommunicationAlert extends Base with a delay.
type NoCommunicationAlert struct {
	Base  `bson:",inline"`
	Delay DurationMinutes `json:"delay,omitempty"`
}

func (a NoCommunicationAlert) Validate(validator structure.Validator) {
	a.Base.Validate(validator)
	dur := a.Delay.Duration()
	validator.Duration("delay", &dur).InRange(0, 6*time.Hour)
}

// LowAlert extends Base with threshold and a delay.
type LowAlert struct {
	Base `bson:",inline"`
	// Threshold is compared the current value to determine if an alert should
	// be triggered.
	Threshold `json:"threshold"`
	Delay     DurationMinutes `json:"delay,omitempty"`
	// Repeat is measured in minutes.
	//
	// A value of 0 (the default) disables repeat notifications.
	Repeat DurationMinutes `json:"repeat,omitempty" bson:"repeat"`
}

func (a LowAlert) Validate(validator structure.Validator) {
	a.Base.Validate(validator)
	delayDur := a.Delay.Duration()
	validator.Duration("delay", &delayDur).InRange(0, 2*time.Hour)
	a.Threshold.Validate(validator)
	repeatDur := a.Repeat.Duration()
	validator.Duration("repeat", &repeatDur).Using(validateRepeat)
}

// HighAlert extends Base with a threshold and a delay.
type HighAlert struct {
	Base `bson:",inline"`
	// Threshold is compared the current value to determine if an alert should
	// be triggered.
	Threshold `json:"threshold"`
	Delay     DurationMinutes `json:"delay,omitempty"`
	// Repeat is measured in minutes.
	//
	// A value of 0 (the default) disables repeat notifications.
	Repeat DurationMinutes `json:"repeat,omitempty" bson:"repeat"`
}

func (a HighAlert) Validate(validator structure.Validator) {
	a.Base.Validate(validator)
	a.Threshold.Validate(validator)
	delayDur := a.Delay.Duration()
	validator.Duration("delay", &delayDur).InRange(0, 6*time.Hour)
	repeatDur := a.Repeat.Duration()
	validator.Duration("repeat", &repeatDur).Using(validateRepeat)
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
func (t Threshold) Validate(v structure.Validator) {
	v.String("units", &t.Units).OneOf(glucose.MgdL, glucose.MmolL)
	// This is a sanity check. Client software will likely further constrain these values. The
	// broadness of these values allows clients to change their own min and max values
	// independently, and it sidesteps rounding and conversion conflicts between the backend and
	// clients.
	var max, min float64
	switch t.Units {
	case glucose.MgdL, glucose.Mgdl:
		max = glucose.MgdLMaximum
		min = glucose.MgdLMinimum
		v.Float64("value", &t.Value).InRange(min, max)
	case glucose.MmolL, glucose.Mmoll:
		max = glucose.MmolLMaximum
		min = glucose.MmolLMinimum
		v.Float64("value", &t.Value).InRange(min, max)
	default:
		v.WithReference("value").ReportError(validator.ErrorValueNotValid())
	}
}

// Repository abstracts persistent storage for Config data.
type Repository interface {
	Get(ctx context.Context, conf *Config) (*Config, error)
	Upsert(ctx context.Context, conf *Config) error
	Delete(ctx context.Context, conf *Config) error
	List(ctx context.Context, userID string) ([]*Config, error)

	EnsureIndexes() error
}
