package oura

const (
	DataTypeDailyActivity          = "daily_activity"
	DataTypeDailyCardiovascularAge = "daily_cardiovascular_age"
	DataTypeDailyReadiness         = "daily_readiness"
	DataTypeDailyResilience        = "daily_resilience"
	DataTypeDailySleep             = "daily_sleep"
	DataTypeDailySpO2              = "daily_spo2"
	DataTypeDailyStress            = "daily_stress"
	DataTypeEnhancedTag            = "enhanced_tag"
	DataTypeHeartRate              = "heartrate" // Explicitly missing underscore word separator
	DataTypeRestModePeriod         = "rest_mode_period"
	DataTypeRingConfiguration      = "ring_configuration"
	DataTypeSession                = "session"
	DataTypeSleep                  = "sleep"
	DataTypeSleepTime              = "sleep_time"
	DataTypeVO2Max                 = "vo2_max"
	DataTypeWorkout                = "workout"

	EventTypeCreate = "create"
	EventTypeUpdate = "update"
	EventTypeDelete = "delete"

	ProviderName = "oura"
)

func DataTypes() []string {
	return []string{
		DataTypeDailyActivity,
		DataTypeDailyCardiovascularAge,
		DataTypeDailyReadiness,
		DataTypeDailyResilience,
		DataTypeDailySleep,
		DataTypeDailySpO2,
		DataTypeDailyStress,
		DataTypeEnhancedTag,
		DataTypeHeartRate,
		DataTypeRestModePeriod,
		DataTypeRingConfiguration,
		DataTypeSession,
		DataTypeSleep,
		DataTypeSleepTime,
		DataTypeVO2Max,
		DataTypeWorkout,
	}
}

func EventTypes() []string {
	return []string{
		EventTypeCreate,
		EventTypeUpdate,
		EventTypeDelete,
	}
}

type Datum struct {
	ID string `json:"id,omitempty"`
}

type Data []*Datum
