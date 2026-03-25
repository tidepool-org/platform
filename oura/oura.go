package oura

import (
	"context"
	"slices"
	"strconv"
	"time"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/oauth"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/times"
)

//go:generate mockgen -source=oura.go -destination=test/oura_mocks.go -package=test -typed

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

	SubscriptionArrayLengthMaximum   = 100
	SubscriptionExpirationTimeFormat = "2006-01-02T15:04:05.999999999" // Assume location UTC

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

type BaseClient interface {
	ClientID() string
	ClientSecret() string

	PartnerURL() string
	PartnerSecret() string
}

type Client interface {
	BaseClient

	ListSubscriptions(ctx context.Context) (Subscriptions, error)
	CreateSubscription(ctx context.Context, create *CreateSubscription) (*Subscription, error)
	UpdateSubscription(ctx context.Context, id string, update *UpdateSubscription) (*Subscription, error)
	RenewSubscription(ctx context.Context, id string) (*Subscription, error)
	DeleteSubscription(ctx context.Context, id string) error

	GetPersonalInfo(ctx context.Context, tokenSource oauth.TokenSource) (*PersonalInfo, error)

	GetData(ctx context.Context, dataType string, timeRange times.TimeRange, tokenSource oauth.TokenSource) (*Data, error)
	GetDatum(ctx context.Context, dataType string, dataID string, tokenSource oauth.TokenSource) (*Datum, error)

	RevokeOAuthToken(ctx context.Context, oauthToken *auth.OAuthToken) error
}

type CreateSubscription struct {
	CallbackURL       *string `json:"callback_url,omitempty" bson:"callback_url,omitempty"`
	VerificationToken *string `json:"verification_token,omitempty" bson:"verification_token,omitempty"`
	DataType          *string `json:"data_type,omitempty" bson:"data_type,omitempty"`
	EventType         *string `json:"event_type,omitempty" bson:"event_type,omitempty"`
}

func (c *CreateSubscription) Parse(parser structure.ObjectParser) {
	c.CallbackURL = parser.String("callback_url")
	c.VerificationToken = parser.String("verification_token")
	c.DataType = parser.String("data_type")
	c.EventType = parser.String("event_type")
}

func (c *CreateSubscription) Validate(validator structure.Validator) {
	validator.String("callback_url", c.CallbackURL).Exists().NotEmpty()
	validator.String("verification_token", c.VerificationToken).Exists().NotEmpty()
	validator.String("data_type", c.DataType).Exists().Using(DataTypeValidator)
	validator.String("event_type", c.EventType).Exists().Using(EventTypeValidator)
}

type UpdateSubscription struct {
	CallbackURL       *string `json:"callback_url,omitempty" bson:"callback_url,omitempty"`
	VerificationToken *string `json:"verification_token,omitempty" bson:"verification_token,omitempty"`
	DataType          *string `json:"data_type,omitempty" bson:"data_type,omitempty"`
	EventType         *string `json:"event_type,omitempty" bson:"event_type,omitempty"`
}

func (u *UpdateSubscription) Parse(parser structure.ObjectParser) {
	u.CallbackURL = parser.String("callback_url")
	u.VerificationToken = parser.String("verification_token")
	u.DataType = parser.String("data_type")
	u.EventType = parser.String("event_type")
}

func (u *UpdateSubscription) Validate(validator structure.Validator) {
	validator.String("callback_url", u.CallbackURL).Exists().NotEmpty()
	validator.String("verification_token", u.VerificationToken).Exists().NotEmpty()
	validator.String("data_type", u.DataType).Exists().Using(DataTypeValidator)
	validator.String("event_type", u.EventType).Exists().Using(EventTypeValidator)
}

func ParseSubscription(parser structure.ObjectParser) *Subscription {
	if !parser.Exists() {
		return nil
	}
	datum := &Subscription{}
	datum.Parse(parser)
	return datum
}

type Subscription struct {
	ID             *string    `json:"id,omitempty" bson:"id,omitempty"`
	CallbackURL    *string    `json:"callback_url,omitempty" bson:"callback_url,omitempty"`
	DataType       *string    `json:"data_type,omitempty" bson:"data_type,omitempty"`
	EventType      *string    `json:"event_type,omitempty" bson:"event_type,omitempty"`
	ExpirationTime *time.Time `json:"expiration_time,omitempty" bson:"expiration_time,omitempty"`
}

func (s *Subscription) Parse(parser structure.ObjectParser) {
	s.ID = parser.String("id")
	s.CallbackURL = parser.String("callback_url")
	s.DataType = parser.String("data_type")
	s.EventType = parser.String("event_type")
	s.ExpirationTime = parser.Time("expiration_time", SubscriptionExpirationTimeFormat)
}

func (s *Subscription) Validate(validator structure.Validator) {
	validator.String("id", s.ID).Exists().Using(DataIDValidator)
	validator.String("callback_url", s.CallbackURL).Exists().NotEmpty()
	validator.String("data_type", s.DataType).Exists().Using(DataTypeValidator)
	validator.String("event_type", s.EventType).Exists().Using(EventTypeValidator)
	validator.Time("expiration_time", s.ExpirationTime).Exists().NotZero()
}

type Subscriptions []*Subscription

func (s *Subscriptions) Parse(parser structure.ArrayParser) {
	for _, reference := range parser.References() {
		*s = append(*s, ParseSubscription(parser.WithReferenceObjectParser(reference)))
	}
}

func (s *Subscriptions) Validate(validator structure.Validator) {
	if length := len(*s); length > SubscriptionArrayLengthMaximum {
		validator.ReportError(structureValidator.ErrorLengthNotLessThanOrEqualTo(length, SubscriptionArrayLengthMaximum))
	}
	for index, datum := range *s {
		if datumValidator := validator.WithReference(strconv.Itoa(index)); datum != nil {
			datum.Validate(datumValidator)
		} else {
			datumValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

func (s *Subscriptions) Get(dataType string, eventType string) *Subscription {
	for _, subscription := range *s {
		if subscription != nil &&
			subscription.DataType != nil && *subscription.DataType == dataType &&
			subscription.EventType != nil && *subscription.EventType == eventType {
			return subscription
		}
	}
	return nil
}

type PersonalInfo struct {
	ID            *string  `json:"id,omitempty" bson:"id,omitempty"`
	Age           *int     `json:"age,omitempty" bson:"age,omitempty"`
	Weight        *float64 `json:"weight,omitempty" bson:"weight,omitempty"`
	Height        *float64 `json:"height,omitempty" bson:"height,omitempty"`
	BiologicalSex *string  `json:"biological_sex,omitempty" bson:"biological_sex,omitempty"`
	Email         *string  `json:"email,omitempty" bson:"email,omitempty"`
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

type Data []*Datum

type Datum struct {
	ID string `json:"id,omitempty"`
}

func IsValidDataID(value string) bool {
	return ValidateDataID(value) == nil
}

func DataIDValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateDataID(value))
}

func ValidateDataID(value string) error {
	if value == "" {
		return structureValidator.ErrorValueEmpty()
	}
	return nil
}

func IsValidDataType(value string) bool {
	return ValidateDataType(value) == nil
}

func DataTypeValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateDataType(value))
}

func ValidateDataType(value string) error {
	if value == "" {
		return structureValidator.ErrorValueEmpty()
	} else if !slices.Contains(DataTypes(), value) {
		return structureValidator.ErrorValueStringNotOneOf(value, DataTypes())
	}
	return nil
}

func IsValidEventType(value string) bool {
	return ValidateEventType(value) == nil
}

func EventTypeValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateEventType(value))
}

func ValidateEventType(value string) error {
	if value == "" {
		return structureValidator.ErrorValueEmpty()
	} else if !slices.Contains(EventTypes(), value) {
		return structureValidator.ErrorValueStringNotOneOf(value, EventTypes())
	}
	return nil
}
