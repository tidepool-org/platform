package webhook

import (
	"time"

	"github.com/tidepool-org/platform/oura"
	"github.com/tidepool-org/platform/structure"
)

// All, but heartrate
func DataTypes() []string {
	return []string{
		oura.DataTypeDailyActivity,
		oura.DataTypeDailyCardiovascularAge,
		oura.DataTypeDailyReadiness,
		oura.DataTypeDailyResilience,
		oura.DataTypeDailySleep,
		oura.DataTypeDailySpO2,
		oura.DataTypeDailyStress,
		oura.DataTypeEnhancedTag,
		oura.DataTypeRestModePeriod,
		oura.DataTypeRingConfiguration,
		oura.DataTypeSession,
		oura.DataTypeSleep,
		oura.DataTypeSleepTime,
		oura.DataTypeVO2Max,
		oura.DataTypeWorkout,
	}
}

type Event struct {
	EventTime *time.Time `json:"event_time,omitempty"`
	EventType *string    `json:"event_type,omitempty"`
	UserID    *string    `json:"user_id,omitempty"`
	ObjectID  *string    `json:"object_id,omitempty"`
	DataType  *string    `json:"data_type,omitempty"`
}

func ParseEvent(parser structure.ObjectParser) *Event {
	if !parser.Exists() {
		return nil
	}
	datum := NewEvent()
	parser.Parse(datum)
	return datum

}

func NewEvent() *Event {
	return &Event{}
}

func (e *Event) Parse(parser structure.ObjectParser) {
	e.EventTime = parser.Time("event_time", time.RFC3339Nano)
	e.EventType = parser.String("event_type")
	e.UserID = parser.String("user_id")
	e.ObjectID = parser.String("object_id")
	e.DataType = parser.String("data_type")
}

func (e *Event) Validate(validator structure.Validator) {
	validator.Time("event_time", e.EventTime).Exists().NotZero()
	validator.String("event_type", e.EventType).Exists().OneOf(oura.EventTypes()...)
	validator.String("user_id", e.UserID).Exists().NotEmpty()
	validator.String("object_id", e.ObjectID).Exists().NotEmpty()
	validator.String("data_type", e.DataType).Exists().OneOf(DataTypes()...)
}

type CreateSubscription struct {
	CallbackURL       *string `json:"callback_url,omitempty"`
	VerificationToken *string `json:"verification_token,omitempty"`
	EventType         *string `json:"event_type,omitempty"`
	DataType          *string `json:"data_type,omitempty"`
}

func (c *CreateSubscription) Parse(parser structure.ObjectParser) {
	c.CallbackURL = parser.String("callback_url")
	c.VerificationToken = parser.String("verification_token")
	c.EventType = parser.String("event_type")
	c.DataType = parser.String("data_type")
}

func (c *CreateSubscription) Validate(validator structure.Validator) {
	validator.String("callback_url", c.CallbackURL).Exists().NotEmpty()
	validator.String("verification_token", c.VerificationToken).Exists().NotEmpty()
	validator.String("event_type", c.EventType).Exists().OneOf(oura.EventTypes()...)
	validator.String("data_type", c.DataType).Exists().OneOf(DataTypes()...)
}

type Subscription struct {
	ID             *string    `json:"id,omitempty"`
	CallbackURL    *string    `json:"callback_url,omitempty"`
	EventType      *string    `json:"event_type,omitempty"`
	DataType       *string    `json:"data_type,omitempty"`
	ExpirationTime *time.Time `json:"expiration_time,omitempty"`
}

func (s *Subscription) Parse(parser structure.ObjectParser) {
	s.ID = parser.String("id")
	s.CallbackURL = parser.String("callback_url")
	s.EventType = parser.String("event_type")
	s.DataType = parser.String("data_type")
	s.ExpirationTime = parser.Time("expiration_time", time.RFC3339Nano)
}

func (s *Subscription) Validate(validator structure.Validator) {
	validator.String("id", s.ID).Exists().NotEmpty()
	validator.String("callback_url", s.CallbackURL).Exists().NotEmpty()
	validator.String("event_type", s.EventType).Exists().OneOf(oura.EventTypes()...)
	validator.String("data_type", s.DataType).Exists().OneOf(DataTypes()...)
	validator.Time("expiration_time", s.ExpirationTime).Exists().NotZero()
}
