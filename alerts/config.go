package alerts

// Data models for care team alerts.

import (
	"bytes"
	"context"
	"encoding/json"
	"slices"
	"time"

	"github.com/tidepool-org/platform/data"
	nontypesglucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/data/types/dosingdecision"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
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

// Evaluate alerts in the context of the provided data.
//
// While this method, or the methods it calls, can fail, there's no point in returning an
// error. Instead errors are logged before continuing. This is to ensure that any possible alert
// that should be triggered, will be triggered.
func (c Config) Evaluate(ctx context.Context, gd []*glucose.Glucose, dd []*dosingdecision.DosingDecision) *Notification {
	notification := c.Alerts.Evaluate(ctx, gd, dd)
	if notification != nil {
		notification.FollowedUserID = c.FollowedUserID
		notification.RecipientUserID = c.UserID
	}
	if lgr := log.LoggerFromContext(ctx); lgr != nil {
		lgr.WithField("notification", notification).Info("evaluated alert")
	}

	return notification
}

// LongestDelay of the delays set on enabled alerts.
func (a Alerts) LongestDelay() time.Duration {
	delays := []time.Duration{}
	if a.Low != nil && a.Low.Enabled {
		delays = append(delays, a.Low.Delay.Duration())
	}
	if a.High != nil && a.High.Enabled {
		delays = append(delays, a.High.Delay.Duration())
	}
	if a.NotLooping != nil && a.NotLooping.Enabled {
		delays = append(delays, a.NotLooping.Delay.Duration())
	}
	if a.NoCommunication != nil && a.NoCommunication.Enabled {
		delays = append(delays, a.NoCommunication.Delay.Duration())
	}
	if len(delays) == 0 {
		return 0
	}
	return slices.Max(delays)
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

// Evaluate a user's data to determine if notifications are indicated.
//
// Evaluations are performed according to priority. The process is
// "short-circuited" at the first indicated notification.
func (a Alerts) Evaluate(ctx context.Context,
	gd []*glucose.Glucose, dd []*dosingdecision.DosingDecision) *Notification {

	if a.NoCommunication != nil && a.NoCommunication.Enabled {
		if n := a.NoCommunication.Evaluate(ctx, gd); n != nil {
			return n
		}
	}
	if a.UrgentLow != nil && a.UrgentLow.Enabled {
		if n := a.UrgentLow.Evaluate(ctx, gd); n != nil {
			return n
		}
	}
	if a.Low != nil && a.Low.Enabled {
		if n := a.Low.Evaluate(ctx, gd); n != nil {
			return n
		}
	}
	if a.High != nil && a.High.Enabled {
		if n := a.High.Evaluate(ctx, gd); n != nil {
			return n
		}
	}
	if a.NotLooping != nil && a.NotLooping.Enabled {
		if n := a.NotLooping.Evaluate(ctx, dd); n != nil {
			return n
		}
	}
	return nil
}

// Base describes the minimum specifics of a desired alert.
type Base struct {
	// Enabled controls whether notifications should be sent for this alert.
	Enabled bool `json:"enabled" bson:"enabled"`

	// Activity tracks when events related to the alert occurred.
	Activity `json:"-" bson:"activity,omitempty"`
}

func (b Base) Validate(validator structure.Validator) {
	validator.Bool("enabled", &b.Enabled)
}

func (b Base) Evaluate(ctx context.Context, data []*glucose.Glucose) *Notification {
	if lgr := log.LoggerFromContext(ctx); lgr != nil {
		lgr.Warn("alerts.Base.Evaluate called, this shouldn't happen!")
	}
	return nil
}

type Activity struct {
	// Triggered records the last time this alert was triggered.
	Triggered time.Time `json:"triggered" bson:"triggered"`
	// Sent records the last time this alert was sent.
	Sent time.Time `json:"sent" bson:"sent"`
	// Resolved records the last time this alert was resolved.
	Resolved time.Time `json:"resolved" bson:"resolved"`
}

func (a Activity) IsActive() bool {
	return a.Triggered.After(a.Resolved)
}

func (a Activity) IsSent() bool {
	return a.Sent.After(a.Triggered)
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

// Evaluate urgent low condition.
//
// Assumes data is pre-sorted in descending order by Time.
func (a *UrgentLowAlert) Evaluate(ctx context.Context, data []*glucose.Glucose) (notification *Notification) {
	lgr := log.LoggerFromContext(ctx)
	if len(data) == 0 {
		lgr.Debug("no data to evaluate for urgent low")
		return nil
	}
	datum := data[0]
	okDatum, okThreshold, err := validateGlucoseAlertDatum(datum, a.Threshold)
	if err != nil {
		lgr.WithError(err).Warn("Unable to evaluate urgent low")
		return nil
	}
	defer func() {
		logGlucoseAlertEvaluation(lgr, "urgent low", notification, okDatum, okThreshold)
	}()
	active := okDatum < okThreshold
	if !active {
		if a.IsActive() {
			a.Resolved = time.Now()
		}
		return nil
	}
	if !a.IsActive() {
		a.Triggered = time.Now()
	}
	return &Notification{Message: genGlucoseThresholdMessage("below urgent low")}
}

func validateGlucoseAlertDatum(datum *glucose.Glucose, t Threshold) (float64, float64, error) {
	if datum.Blood.Units == nil || datum.Blood.Value == nil || datum.Blood.Time == nil {
		return 0, 0, errors.Newf("Unable to evaluate datum: Units, Value, or Time is nil")
	}
	threshold := nontypesglucose.NormalizeValueForUnits(&t.Value, datum.Blood.Units)
	if threshold == nil {
		return 0, 0, errors.Newf("Unable to normalize threshold units: normalized to nil")
	}
	return *datum.Blood.Value, *threshold, nil
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

// Evaluate if the device is looping.
func (a NotLoopingAlert) Evaluate(ctx context.Context, decisions []*dosingdecision.DosingDecision) (notifcation *Notification) {
	// TODO will be implemented in the near future.
	return nil
}

// DosingDecisionReasonLoop is specified in a [dosingdecision.DosingDecision] to indicate that
// the decision is part of a loop adjustment (as opposed to bolus or something else).
const DosingDecisionReasonLoop string = "loop"

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

// Evaluate if CGM data is being received by Tidepool.
//
// Assumes data is pre-sorted by Time in descending order.
func (a NoCommunicationAlert) Evaluate(ctx context.Context, data []*glucose.Glucose) *Notification {
	var newest time.Time
	for _, d := range data {
		if d != nil && d.Time != nil && !(*d.Time).IsZero() {
			newest = *d.Time
			break
		}
	}
	if time.Since(newest) > a.Delay.Duration() {
		return &Notification{Message: NoCommunicationMessage}
	}

	return nil
}

const NoCommunicationMessage = "Tidepool is unable to communicate with a user's device"

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

// Evaluate the given data to determine if an alert should be sent.
//
// Assumes data is pre-sorted in descending order by Time.
func (a *LowAlert) Evaluate(ctx context.Context, data []*glucose.Glucose) (notification *Notification) {
	lgr := log.LoggerFromContext(ctx)
	if len(data) == 0 {
		lgr.Debug("no data to evaluate for low")
		return nil
	}
	var eventBegan time.Time
	var okDatum, okThreshold float64
	var err error
	defer func() { logGlucoseAlertEvaluation(lgr, "low", notification, okDatum, okThreshold) }()
	for _, datum := range data {
		okDatum, okThreshold, err = validateGlucoseAlertDatum(datum, a.Threshold)
		if err != nil {
			lgr.WithError(err).Debug("Skipping low alert datum evaluation")
			continue
		}
		active := okDatum < okThreshold
		if !active {
			break
		}
		if (*datum.Time).Before(eventBegan) || eventBegan.IsZero() {
			eventBegan = *datum.Time
		}
	}
	if eventBegan.IsZero() {
		if a.IsActive() {
			a.Resolved = time.Now()
		}
		return nil
	}
	if !a.IsActive() {
		if time.Since(eventBegan) > a.Delay.Duration() {
			a.Triggered = time.Now()
		}
	}
	return &Notification{Message: genGlucoseThresholdMessage("below low")}
}

func genGlucoseThresholdMessage(alertType string) string {
	return "Glucose reading " + alertType + " threshold"
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

// Evaluate the given data to determine if an alert should be sent.
//
// Assumes data is pre-sorted in descending order by Time.
func (a *HighAlert) Evaluate(ctx context.Context, data []*glucose.Glucose) (notification *Notification) {
	lgr := log.LoggerFromContext(ctx)
	if len(data) == 0 {
		lgr.Debug("no data to evaluate for high")
		return nil
	}
	var eventBegan time.Time
	var okDatum, okThreshold float64
	var err error
	defer func() { logGlucoseAlertEvaluation(lgr, "high", notification, okDatum, okThreshold) }()
	for _, datum := range data {
		okDatum, okThreshold, err = validateGlucoseAlertDatum(datum, a.Threshold)
		if err != nil {
			lgr.WithError(err).Debug("Skipping high alert datum evaluation")
			continue
		}
		active := okDatum > okThreshold
		if !active {
			break
		}
		if (*datum.Time).Before(eventBegan) || eventBegan.IsZero() {
			eventBegan = *datum.Time
		}
	}
	if eventBegan.IsZero() {
		if a.IsActive() {
			a.Resolved = time.Now()
		}
		return nil
	}
	if !a.IsActive() {
		if time.Since(eventBegan) > a.Delay.Duration() {
			a.Triggered = time.Now()
		}
	}
	return &Notification{Message: genGlucoseThresholdMessage("above high")}
}

// logGlucoseAlertEvaluation is called during each glucose-based evaluation for record-keeping.
func logGlucoseAlertEvaluation(lgr log.Logger, alertType string, notification *Notification,
	value, threshold float64) {

	fields := log.Fields{
		"isAlerting?": notification != nil,
		"threshold":   threshold,
		"value":       value,
	}
	lgr.WithFields(fields).Info(alertType)
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

// ValueWithUnits binds a value with its units.
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
	v.String("units", &t.Units).OneOf(nontypesglucose.MgdL, nontypesglucose.MmolL)
	// This is a sanity check. Client software will likely further constrain these values. The
	// broadness of these values allows clients to change their own min and max values
	// independently, and it sidesteps rounding and conversion conflicts between the backend and
	// clients.
	var max, min float64
	switch t.Units {
	case nontypesglucose.MgdL, nontypesglucose.Mgdl:
		max = nontypesglucose.MgdLMaximum
		min = nontypesglucose.MgdLMinimum
		v.Float64("value", &t.Value).InRange(min, max)
	case nontypesglucose.MmolL, nontypesglucose.Mmoll:
		max = nontypesglucose.MmolLMaximum
		min = nontypesglucose.MmolLMinimum
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

// Notification gathers information necessary for sending an alert notification.
type Notification struct {
	// Message communicates the alert to the recipient.
	Message         string
	RecipientUserID string
	FollowedUserID  string
}
