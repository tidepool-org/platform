package alert

import (
	"time"

	"github.com/tidepool-org/platform/data"
	dataTypes "github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/structure"
)

const (
	Type = "alert"

	NameLengthMaximum      = 100
	PriorityCritical       = "critical"
	PriorityNormal         = "normal"
	PriorityTimeSensitive  = "timeSensitive"
	SoundName              = "name"
	SoundNameLengthMaximum = 100
	SoundSilence           = "silence"
	SoundVibrate           = "vibrate"
	TriggerDelayed         = "delayed"
	TriggerDelayMaximum    = 60 * 60 * 24
	TriggerDelayMinimum    = 0
	TriggerImmediate       = "immediate"
	TriggerRepeating       = "repeating"
)

func Priorities() []string {
	return []string{
		PriorityCritical,
		PriorityNormal,
		PriorityTimeSensitive,
	}
}

func Sounds() []string {
	return []string{
		SoundName,
		SoundSilence,
		SoundVibrate,
	}
}

func Triggers() []string {
	return []string{
		TriggerDelayed,
		TriggerImmediate,
		TriggerRepeating,
	}
}

type Alert struct {
	dataTypes.Base `bson:",inline"`

	Name             *string    `json:"name,omitempty" bson:"name,omitempty"`
	Priority         *string    `json:"priority,omitempty" bson:"priority,omitempty"`
	Trigger          *string    `json:"trigger,omitempty" bson:"trigger,omitempty"`
	TriggerDelay     *int       `json:"triggerDelay,omitempty" bson:"triggerDelay,omitempty"`
	Sound            *string    `json:"sound,omitempty" bson:"sound,omitempty"`
	SoundName        *string    `json:"soundName,omitempty" bson:"soundName,omitempty"`
	IssuedTime       *time.Time `json:"issuedTime,omitempty" bson:"issuedTime,omitempty"`
	AcknowledgedTime *time.Time `json:"acknowledgedTime,omitempty" bson:"acknowledgedTime,omitempty"`
	RetractedTime    *time.Time `json:"retractedTime,omitempty" bson:"retractedTime,omitempty"`
}

func New() *Alert {
	return &Alert{
		Base: dataTypes.New(Type),
	}
}

func (a *Alert) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(a.Meta())
	}

	a.Base.Parse(parser)

	a.Name = parser.String("name")
	a.Priority = parser.String("priority")
	a.Trigger = parser.String("trigger")
	a.TriggerDelay = parser.Int("triggerDelay")
	a.Sound = parser.String("sound")
	a.SoundName = parser.String("soundName")
	a.IssuedTime = parser.Time("issuedTime", time.RFC3339Nano)
	a.AcknowledgedTime = parser.Time("acknowledgedTime", time.RFC3339Nano)
	a.RetractedTime = parser.Time("retractedTime", time.RFC3339Nano)
}

func (a *Alert) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(a.Meta())
	}

	a.Base.Validate(validator)

	if a.Type != "" {
		validator.String("type", &a.Type).EqualTo(Type)
	}

	validator.String("name", a.Name).Exists().NotEmpty().LengthLessThanOrEqualTo(NameLengthMaximum)
	validator.String("priority", a.Priority).OneOf(Priorities()...)
	validator.String("trigger", a.Trigger).OneOf(Triggers()...)
	if triggerDelayValidator := validator.Int("triggerDelay", a.TriggerDelay); a.Trigger != nil && (*a.Trigger == TriggerDelayed || *a.Trigger == TriggerRepeating) {
		triggerDelayValidator.Exists().InRange(TriggerDelayMinimum, TriggerDelayMaximum)
	} else {
		triggerDelayValidator.NotExists()
	}
	validator.String("sound", a.Sound).OneOf(Sounds()...)
	if soundNameValidator := validator.String("soundName", a.SoundName); a.Sound != nil && *a.Sound == SoundName {
		soundNameValidator.Exists().NotEmpty().LengthLessThanOrEqualTo(SoundNameLengthMaximum)
	} else {
		soundNameValidator.NotExists()
	}
	validator.Time("issuedTime", a.IssuedTime).Exists().NotZero()
	if acknowledgedTimeValidator := validator.Time("acknowledgedTime", a.AcknowledgedTime); a.IssuedTime != nil {
		acknowledgedTimeValidator.After(*a.IssuedTime)
	}
	if retractedTimeValidator := validator.Time("retractedTime", a.RetractedTime); a.IssuedTime != nil {
		retractedTimeValidator.After(*a.IssuedTime)
	}
}

func (a *Alert) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(a.Meta())
	}

	a.Base.Normalize(normalizer)
}
