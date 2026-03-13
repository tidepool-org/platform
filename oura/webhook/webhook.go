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
