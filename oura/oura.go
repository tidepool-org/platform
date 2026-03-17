package oura

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/oauth"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/times"
)

const (
	DataTypeDailyActivity          = "daily_activity"
	DataTypeDailyCardiovascularAge = "daily_cardiovascular_age"
	DataTypeDailyReadiness         = "daily_readiness"
	DataTypeDailyResilience        = "daily_resilience"
	DataTypeDailySleep             = "daily_sleep"
	DataTypeDailySpO2              = "daily_spo2"
	DataTypeDailyStress            = "daily_stress"
	DataTypeEnhancedTag            = "enhanced_tag"
	DataTypeHeartRate              = "heartrate" // Missing underscore per Oura documentation
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
	PartnerName  = ProviderName

	PartnerPathPrefix = "/v1/partners/" + PartnerName

	TimeRangeFormat            = time.RFC3339
	TimeRangeTruncatedDuration = time.Second
	TimeRangeMaximumYears      = 10
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

//go:generate mockgen -source=oura.go -destination=test/oura_mocks.go -package=test Client
type Client interface {
	ListSubscriptions(ctx context.Context) (Subscriptions, error)
	CreateSubscription(ctx context.Context, create *CreateSubscription) (*Subscription, error)
	RenewSubscription(ctx context.Context, id string) (*Subscription, error)
	DeleteSubscription(ctx context.Context, id string) error

	RevokeOAuthToken(ctx context.Context, oauthToken *auth.OAuthToken) error

	GetPersonalInfo(ctx context.Context, tokenSource oauth.TokenSource) (*PersonalInfo, error)

	GetDatum(ctx context.Context, dataType string, dataID string, tokenSource oauth.TokenSource) (*Datum, error)
	GetData(ctx context.Context, dataType string, timeRange times.TimeRange, tokenSource oauth.TokenSource) (Data, error)
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
	validator.String("event_type", c.EventType).Exists().OneOf(EventTypes()...)
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
	validator.String("event_type", s.EventType).Exists().OneOf(EventTypes()...)
	validator.String("data_type", s.DataType).Exists().OneOf(DataTypes()...)
	validator.Time("expiration_time", s.ExpirationTime).Exists().NotZero()
}

type Subscriptions []*Subscription

type PersonalInfo struct {
	ID            *string  `json:"id,omitempty"`
	Age           *int     `json:"age,omitempty"`
	Weight        *float64 `json:"weight,omitempty"`
	Height        *float64 `json:"height,omitempty"`
	BiologicalSex *string  `json:"biological_sex,omitempty"`
	Email         *string  `json:"email,omitempty"`
}

func (p *PersonalInfo) Parse(parser structure.ObjectParser) {
	p.ID = parser.String("id")
	p.Age = parser.Int("age")
	p.Weight = parser.Float64("weight")
	p.Height = parser.Float64("height")
	p.BiologicalSex = parser.String("biological_sex")
	p.Email = parser.String("email")
}

func (p *PersonalInfo) Validate(validator structure.Validator) {
	validator.String("id", p.ID).Exists().NotEmpty()
}

type Datum struct {
	ID string `json:"id,omitempty"`
}

type Data []*Datum
