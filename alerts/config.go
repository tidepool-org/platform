package alerts

// Data models for care team alerts.

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"slices"
	"time"

	"github.com/tidepool-org/platform/data"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/data/types/dosingdecision"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	logjson "github.com/tidepool-org/platform/log/json"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
)

// Config wraps Alerts to include user relationships.
//
// As a wrapper type, Config provides a clear demarcation of what a user controls (Alerts)
// and what is set by the service (the other values in Config).
type Config struct {
	// UserID receives the configured alerts and owns this Config.
	UserID string `json:"userId" bson:"userId"`

	// FollowedUserID is the user whose data generates alerts, and has granted
	// UserID permission to that data.
	FollowedUserID string `json:"followedUserId" bson:"followedUserId"`

	// UploadID identifies the device dataset for which these alerts apply.
	UploadID string `json:"uploadId" bson:"uploadId,omitempty"`

	// Alerts collects the user settings for each type of alert, and tracks their statuses.
	Alerts `bson:"alerts,omitempty"`

	Activity `bson:"activity,omitempty" json:"activity,omitempty"`
}

// Alerts is a wrapper to collect the user-modifiable parts of a Config.
type Alerts struct {
	UrgentLow       *UrgentLowAlert       `json:"urgentLow,omitempty" bson:"urgentLow,omitempty"`
	Low             *LowAlert             `json:"low,omitempty" bson:"low,omitempty"`
	High            *HighAlert            `json:"high,omitempty" bson:"high,omitempty"`
	NotLooping      *NotLoopingAlert      `json:"notLooping,omitempty" bson:"notLooping,omitempty"`
	NoCommunication *NoCommunicationAlert `bson:"noCommunication,omitempty" json:"noCommunication,omitempty"`
}

type Activity struct {
	UrgentLow       AlertActivity `json:"urgentLow,omitempty" bson:"urgentLow,omitempty"`
	Low             AlertActivity `json:"low,omitempty" bson:"low,omitempty"`
	High            AlertActivity `json:"high,omitempty" bson:"high,omitempty"`
	NotLooping      AlertActivity `json:"notLooping,omitempty" bson:"notLooping,omitempty"`
	NoCommunication AlertActivity `json:"noCommunication,omitempty" bson:"noCommunication,omitempty"`
}

func (c Config) Validate(validator structure.Validator) {
	validator.String("userID", &c.UserID).Using(user.IDValidator)
	validator.String("followedUserID", &c.FollowedUserID).Using(user.IDValidator)
	validator.String("uploadID", &c.UploadID).Exists().Using(data.SetIDValidator)
	if c.Alerts.UrgentLow != nil {
		c.Alerts.UrgentLow.Validate(validator)
	}
	if c.Alerts.Low != nil {
		c.Alerts.Low.Validate(validator)
	}
	if c.Alerts.High != nil {
		c.Alerts.High.Validate(validator)
	}
	if c.Alerts.NotLooping != nil {
		c.Alerts.NotLooping.Validate(validator)
	}
	if c.Alerts.NoCommunication != nil {
		c.Alerts.NoCommunication.Validate(validator)
	}
}

// EvaluateData alerts in the context of the provided data.
//
// While this method, or the methods it calls, can fail, there's no point in returning an
// error. Instead errors are logged before continuing. This is to ensure that any possible
// alert that should be triggered, will be triggered.
func (c *Config) EvaluateData(ctx context.Context, gd []*Glucose,
	dd []*DosingDecision) (*Notification, bool) {

	var n *Notification
	var needsUpsert bool

	ul, low, high, nl := EvalResult{}, EvalResult{}, EvalResult{}, EvalResult{}
	if c.Alerts.UrgentLow != nil && c.Alerts.UrgentLow.Enabled {
		ul = c.Alerts.UrgentLow.Evaluate(ctx, gd)
		needsUpsert = needsUpsert || c.Activity.UrgentLow.Update(ul.OutOfRange)
	}
	if c.Alerts.Low != nil && c.Alerts.Low.Enabled {
		low = c.Alerts.Low.Evaluate(ctx, gd)
		needsUpsert = needsUpsert || c.Activity.Low.Update(low.OutOfRange)
	}
	if c.Alerts.High != nil && c.Alerts.High.Enabled {
		high = c.Alerts.High.Evaluate(ctx, gd)
		needsUpsert = needsUpsert || c.Activity.High.Update(high.OutOfRange)
	}
	if c.Alerts.NotLooping != nil && c.Alerts.NotLooping.Enabled {
		nl = c.Alerts.NotLooping.Evaluate(ctx, dd)
		needsUpsert = needsUpsert || c.Activity.NotLooping.Update(nl.OutOfRange)
	}

	if ul.OutOfRange {
		if isReEval(c.Activity.UrgentLow.Sent, ul.NewestTime) {
			return nil, needsUpsert
		}
		msg := genGlucoseThresholdMessage("below urgent low")
		return c.newNotification(msg, &c.Activity.UrgentLow), needsUpsert
	}
	if low.OutOfRange {
		if isReEval(c.Activity.Low.Sent, low.NewestTime) {
			return nil, needsUpsert
		}
		delay := c.Alerts.Low.Delay.Duration()

		if time.Since(low.Started) > delay {
			repeat := c.Alerts.Low.Repeat
			if !c.Activity.Low.IsSent() || mayRepeat(repeat, c.Activity.Low.Sent) {
				msg := genGlucoseThresholdMessage("below low")
				return c.newNotification(msg, &c.Activity.Low), needsUpsert

			}
		}
		return nil, needsUpsert
	}
	if high.OutOfRange {
		if isReEval(c.Activity.High.Sent, high.NewestTime) {
			return nil, needsUpsert
		}
		delay := c.Alerts.High.Delay.Duration()
		if time.Since(high.Started) > delay {
			repeat := c.Alerts.High.Repeat
			if !c.Activity.High.IsSent() || mayRepeat(repeat, c.Activity.High.Sent) {
				msg := genGlucoseThresholdMessage("above high")
				return c.newNotification(msg, &c.Activity.High), needsUpsert
			}
		}
	}
	if nl.OutOfRange {
		// Because not looping doesn't use a threshold, re-evaluations aren't treated any
		// differently.
		delay := c.Alerts.NotLooping.Delay.Duration()
		if delay == 0 {
			delay = NotLoopingRepeat
		}
		if time.Since(c.Activity.NotLooping.Sent) > delay {
			return c.newNotification(NotLoopingMessage, &c.Activity.NotLooping), needsUpsert
		}
	}

	return n, needsUpsert
}

func mayRepeat(repeat DurationMinutes, lastSent time.Time) bool {
	return repeat.Duration() > 0 && time.Since(lastSent) > repeat.Duration()
}

func (c *Config) newNotification(msg string, act *AlertActivity) *Notification {
	return &Notification{
		FollowedUserID:  c.FollowedUserID,
		RecipientUserID: c.UserID,
		Message:         msg,
		Sent: func(t time.Time) {
			if t.After(act.Sent) {
				act.Sent = t
			}
		},
	}
}

func (c Config) LoggerWithFields(lgr log.Logger) log.Logger {
	return lgr.WithFields(log.Fields{
		"userID":         c.UserID,
		"followedUserID": c.FollowedUserID,
		"dataSetID":      c.UploadID,
	})
}

func isReEval(t1, t2 time.Time) bool {
	return t1.After(t2)
}

// TODO pass in a logger
func (c *Config) EvaluateNoCommunication(ctx context.Context, last time.Time) (
	*Notification, bool) {

	if c.Alerts.NoCommunication == nil || !c.Alerts.NoCommunication.Enabled {
		return nil, false
	}

	lgr := c.LoggerWithFields(log.LoggerFromContext(ctx))
	ctx = log.NewContextWithLogger(ctx, lgr)
	nc := c.Alerts.NoCommunication.Evaluate(ctx, last)
	needsUpsert := c.Activity.NoCommunication.Update(nc.OutOfRange)
	// TODO check re-eval? I don't think so
	delay := c.Alerts.NoCommunication.Delay.Duration()
	if delay == 0 {
		delay = DefaultNoCommunicationDelay
	}
	if time.Since(nc.Started) > delay && time.Since(c.Activity.NoCommunication.Sent) > delay {
		n := c.newNotification(NoCommunicationMessage, &c.Activity.NoCommunication)
		return n, needsUpsert
	}
	return nil, needsUpsert
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
	if len(delays) == 0 {
		return 0
	}
	return slices.Max(delays)
}

// Base describes the minimum specifics of a desired alert.
type Base struct {
	// Enabled controls whether notifications should be sent for this alert.
	Enabled bool `json:"enabled" bson:"enabled"`
}

func (b Base) Validate(validator structure.Validator) {
	validator.Bool("enabled", &b.Enabled)
}

func (b Base) Evaluate(ctx context.Context, data []*Glucose) *Notification {
	if lgr := log.LoggerFromContext(ctx); lgr != nil {
		lgr.Warn("alerts.Base.Evaluate called, this shouldn't happen!")
	}
	return nil
}

func (b Base) lgr(ctx context.Context) log.Logger {
	var lgr log.Logger = log.LoggerFromContext(ctx)
	if lgr == nil {
		// NewLogger can only fail if os.Stderr is nil.
		lgr, _ = logjson.NewLogger(os.Stderr, log.DefaultLevelRanks(), log.DefaultLevel())
	}
	return lgr
}

type AlertActivity struct {
	// Triggered records the last time this alert was triggered.
	Triggered time.Time `json:"triggered" bson:"triggered"`
	// Sent records the last time this alert was sent.
	Sent time.Time `json:"sent" bson:"sent"`
	// Resolved records the last time this alert was resolved.
	Resolved time.Time `json:"resolved" bson:"resolved"`
}

func (a AlertActivity) IsActive() bool {
	return a.Triggered.After(a.Resolved)
}

func (a AlertActivity) IsSent() bool {
	return a.Sent.After(a.Triggered)
}

func (a *AlertActivity) Update(outOfRange bool) bool {
	changed := false
	if outOfRange && !a.IsActive() {
		a.Triggered = time.Now()
		changed = true
	} else if !outOfRange && a.IsActive() {
		a.Resolved = time.Now()
		changed = true
	}
	return changed
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

type EvalResult struct {
	Name        string
	Started     time.Time
	Threshold   float64
	NewestTime  time.Time
	NewestValue float64
	Evaluator   func(dv, tv float64) bool `json:"-"`
	OutOfRange  bool
}

func (r EvalResult) String() string {
	b, err := json.Marshal(r)
	if err != nil {
		return "<error marshaling EvalResult>"
	}
	return string(b)
}

func (r *EvalResult) Process(ctx context.Context, t Threshold, data []*Glucose) {
	for _, datum := range data {
		dv, tv, err := normalizeUnits(datum, t)
		if err != nil {
			r.lgr(ctx).WithError(err).Info("Unable to normalize datum")
			continue
		}

		if datum.Time == nil {
			r.lgr(ctx).Warn("Unable to process: Time == nil; that shouldn't be possible")
			continue
		}

		outOfRange := r.Evaluator(dv, tv)

		if r.NewestValue == 0 {
			r.NewestValue = dv
			r.NewestTime = *datum.Time
			r.OutOfRange = outOfRange
			r.Threshold = tv
			r.logGlucoseEval(ctx)
		}

		if !outOfRange {
			break
		}

		if datum.Time != nil && (r.Started.IsZero() || datum.Time.Before(r.Started)) {
			r.Started = *datum.Time
		}
	}
}

// Evaluate urgent low condition.
//
// Assumes data is pre-sorted in descending order by Time.
func (a *UrgentLowAlert) Evaluate(ctx context.Context, data []*Glucose) EvalResult {
	er := EvalResult{
		Name:      "urgent low",
		Evaluator: func(dv, tv float64) bool { return dv < tv },
	}
	er.Process(ctx, a.Threshold, data)
	return er
}

func (r EvalResult) logGlucoseEval(ctx context.Context) {
	fields := log.Fields{
		"isAlerting?": r.Evaluator(r.NewestValue, r.Threshold),
		"threshold":   r.Threshold,
		"value":       r.NewestValue,
	}
	r.lgr(ctx).WithFields(fields).Info(r.Name)
}

func (r EvalResult) lgr(ctx context.Context) log.Logger {
	var lgr log.Logger = log.LoggerFromContext(ctx)
	if lgr == nil {
		// NewLogger can only fail if os.Stderr is nil.
		lgr, _ = logjson.NewLogger(os.Stderr, log.DefaultLevelRanks(), log.DefaultLevel())
	}
	return lgr
}

func normalizeUnits(datum *Glucose, t Threshold) (float64, float64, error) {
	if datum == nil || datum.Blood.Units == nil || datum.Blood.Value == nil {
		return 0, 0, errors.Newf("Unable to evaluate datum: Units or Value is nil")
	}

	// Both units are the same, no need to convert either.
	if t.Units == *datum.Blood.Units {
		return *datum.Blood.Value, t.Value, nil
	}

	// The units don't match. There exists a known good function that converts to MmolL, so
	// we'll convert whichever value isn't in MmolL to MmolL.

	if dataBloodGlucose.IsMmolL(t.Units) {
		n := dataBloodGlucose.NormalizeValueForUnits(datum.Blood.Value, datum.Blood.Units)
		return *n, t.Value, nil
	} else if dataBloodGlucose.IsMmolL(*datum.Blood.Units) {
		n := dataBloodGlucose.NormalizeValueForUnits(&t.Value, &t.Units)
		return *datum.Blood.Value, *n, nil
	}

	// This shouldn't happen. It indicates a new, third glucose unit is in use.
	return 0, 0, errors.New("Unable to handle unit conversion, neither is MmolL")
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
func (a *NotLoopingAlert) Evaluate(ctx context.Context, decisions []*DosingDecision) EvalResult {
	er := EvalResult{}
	for _, decision := range decisions {
		if decision.Reason == nil || *decision.Reason != DosingDecisionReasonLoop {
			continue
		}
		if decision.Time == nil {
			a.lgr(ctx).Warn("Unable to process: Time == nil; that shouldn't be possible")
			continue
		}
		if !decision.Time.IsZero() {
			er.NewestTime = *decision.Time
			break
		}
	}
	delay := a.Delay.Duration()
	if delay == 0 {
		delay = DefaultNotLoopingDelay
	}
	er.OutOfRange = time.Since(er.NewestTime) > delay
	logNotLoopingEvaluation(a.lgr(ctx), er.OutOfRange, time.Since(er.NewestTime), delay)

	return er
}

// DefaultNotLoopingDelay is used when the delay has a Zero value (its default).
const DefaultNotLoopingDelay = 30 * time.Minute

func logNotLoopingEvaluation(lgr log.Logger, isAlerting bool, since, threshold time.Duration) {
	fields := log.Fields{
		"isAlerting?": isAlerting,
		"value":       since,
		"threshold":   threshold,
	}
	lgr.WithFields(fields).Info("not looping")
}

const NotLoopingMessage = "Loop is not able to loop"

// DosingDecisionReasonLoop is specified in a [DosingDecision] to indicate
// that the decision is part of a loop adjustment (as opposed to bolus or something else).
const DosingDecisionReasonLoop string = "loop"

// NotLoopingRepeat is the interval between sending notifications when not looping.
const NotLoopingRepeat = 5 * time.Minute

// NoCommunicationAlert is configured to send notifications when no data is received.
//
// It differs fundamentally from DataAlerts in that it is polled instead of being triggered
// when data is received.
type NoCommunicationAlert struct {
	Base `bson:",inline"`
	// Delay represents the time after which a No Communication alert should be sent.
	//
	// A value of 0 is the default, and is treated as five minutes.
	Delay DurationMinutes `json:"delay,omitempty"`
}

func (a NoCommunicationAlert) Validate(validator structure.Validator) {
	a.Base.Validate(validator)
	dur := a.Delay.Duration()
	validator.Duration("delay", &dur).InRange(0, 6*time.Hour)
}

// Evaluate if the time since data was last received warrants a notification.
func (a *NoCommunicationAlert) Evaluate(ctx context.Context, lastReceived time.Time) EvalResult {
	er := EvalResult{}

	if lastReceived.IsZero() {
		a.lgr(ctx).Info("Unable to evaluate no communication: time is Zero")
		return er
	}

	delay := a.Delay.Duration()
	if delay == 0 {
		delay = DefaultNoCommunicationDelay
	}
	er.OutOfRange = time.Since(lastReceived) > delay
	er.Started = lastReceived
	er.NewestTime = lastReceived
	a.lgr(ctx).WithField("isAlerting?", er.OutOfRange).Info("no communication")

	return er
}

const DefaultNoCommunicationDelay = 5 * time.Minute

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
func (a *LowAlert) Evaluate(ctx context.Context, data []*Glucose) EvalResult {
	er := EvalResult{
		Name:      "low",
		Evaluator: func(dv, tv float64) bool { return dv < tv },
	}
	er.Process(ctx, a.Threshold, data)
	return er
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
func (a *HighAlert) Evaluate(ctx context.Context, data []*Glucose) EvalResult {
	er := &EvalResult{
		Name:      "high",
		Evaluator: func(dv, tv float64) bool { return dv > tv },
	}
	er.Process(ctx, a.Threshold, data)
	return *er
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
	v.String("units", &t.Units).OneOf(dataBloodGlucose.MgdL, dataBloodGlucose.MmolL)
	// This is a sanity check. Client software will likely further constrain these
	// values. The broadness of these values allows clients to change their own min and max
	// values independently, and it sidesteps rounding and conversion conflicts between the
	// backend and clients.
	var max, min float64
	switch t.Units {
	case dataBloodGlucose.MgdL, dataBloodGlucose.Mgdl:
		max = dataBloodGlucose.MgdLMaximum
		min = dataBloodGlucose.MgdLMinimum
		v.Float64("value", &t.Value).InRange(min, max)
	case dataBloodGlucose.MmolL, dataBloodGlucose.Mmoll:
		max = dataBloodGlucose.MmolLMaximum
		min = dataBloodGlucose.MmolLMinimum
		v.Float64("value", &t.Value).InRange(min, max)
	default:
		v.WithReference("value").ReportError(validator.ErrorValueNotValid())
	}
}

// Repository abstracts persistent storage in the alerts collection for Config data.
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
	Sent            func(time.Time)
}

// LastCommunicationsRepository encapsulates queries of the [LastCommunication] records
// collection for use with alerts.
type LastCommunicationsRepository interface {
	// RecordReceivedDeviceData upserts the time of last communication from a user.
	RecordReceivedDeviceData(context.Context, LastCommunication) error
	// OverdueCommunications lists records for those users that haven't communicated for a
	// time.
	OverdueCommunications(context.Context) ([]LastCommunication, error)

	EnsureIndexes() error
}

// DosingDecision is an alias of convenience.
type DosingDecision = dosingdecision.DosingDecision

// Glucose is an alias of convenience.
type Glucose = glucose.Glucose
