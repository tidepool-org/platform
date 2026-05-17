//nolint:tagliatelle // struct tags derived from Oura API
package oura

import (
	"context"
	"encoding/json"
	"slices"
	"strconv"
	"time"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/crypto"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/oauth"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/times"
)

//go:generate mockgen -source=oura.go -destination=test/oura_mocks.go -package=test -typed

const (
	DataSetClientName    = "org.tidepool.oura.api"
	DataSetClientVersion = "1.0.0"

	DataTypeDailyActivity          = "daily_activity"           // ScopeDaily
	DataTypeDailyCardiovascularAge = "daily_cardiovascular_age" // ScopeHeartHealth
	DataTypeDailyCyclePhases       = "daily_cycle_phases"       // ScopeReproductiveCycle
	DataTypeDailyReadiness         = "daily_readiness"          // ScopeDaily
	DataTypeDailyResilience        = "daily_resilience"         // ScopeStress
	DataTypeDailySleep             = "daily_sleep"              // ScopeDaily
	DataTypeDailySpO2              = "daily_spo2"               // ScopeSpo2
	DataTypeDailyStress            = "daily_stress"             // ScopeDaily
	DataTypeEnhancedTag            = "enhanced_tag"             // ScopeTag
	DataTypeHeartRate              = "heartrate"                // ScopeHeartRate; missing underscore per Oura documentation
	DataTypePersonalInfo           = "personal_info"            // ScopeEmail, ScopePersonal
	DataTypeRestModePeriod         = "rest_mode_period"         // ScopeDaily
	DataTypeRingBatteryLevel       = "ring_battery_level"       // ScopeRingConfiguration
	DataTypeRingConfiguration      = "ring_configuration"       // ScopeRingConfiguration
	DataTypeSession                = "session"                  // ScopeSession
	DataTypeSleep                  = "sleep"                    // ScopeDaily
	DataTypeSleepTime              = "sleep_time"               // ScopeDaily
	DataTypeVO2Max                 = "vo2_max"                  // ScopeHeartHealth
	DataTypeWorkout                = "workout"                  // ScopeWorkout

	// NOTE: Known, but unavailable data types
	// DataTypeBloodGlucose       = "blood_glucose"       // ScopeMetabolic
	// DataTypeFertileWindow      = "fertile_window"      // ScopeReproductiveCycle
	// DataTypeInterbeatInterval  = "interbeat_interval"  // ScopeResearch
	// DataTypeOvulationConfirmed = "ovulation_confirmed" // ScopeReproductiveCycle
	// DataTypePeriodStart        = "period_start"        // ScopeReproductiveCycle
	// DataTypePregnancy          = "pregnancy"           // ScopePregnancy

	DeviceManufacturer = "Oura"

	EventTypeCreate = "create"
	EventTypeUpdate = "update"
	EventTypeDelete = "delete"

	ProviderName = "oura"
	PartnerName  = ProviderName

	PartnerPathPrefix = "/v1/partners/" + PartnerName

	ScopeDaily             = "extapi:daily"
	ScopeEmail             = "extapi:email"
	ScopeHeartHealth       = "extapi:heart_health"
	ScopeHeartRate         = "extapi:heartrate"
	ScopePersonal          = "extapi:personal"
	ScopeReproductiveCycle = "extapi:reproductive_cycle"
	ScopeRingConfiguration = "extapi:ring_configuration"
	ScopeSession           = "extapi:session"
	ScopeSpo2              = "extapi:spo2"
	ScopeStress            = "extapi:stress"
	ScopeTag               = "extapi:tag"
	ScopeWorkout           = "extapi:workout"

	// NOTE: Known, but unavailable scopes (not included in Scopes() nor used to determine data types)
	// ScopeMetabolic = "extapi:metabolic"
	// ScopePregnancy = "extapi:pregnancy"
	// ScopeResearch  = "extapi:research"

	SubscriptionArrayLengthMaximum   = 100
	SubscriptionExpirationTimeFormat = "2006-01-02T15:04:05.999999999" // Assume location UTC

	TimeRangeFormat       = time.RFC3339
	TimeRangeMaximumYears = 10
)

var (
	DeviceManufacturers = []string{DeviceManufacturer}
	DeviceTags          = []string{data.DeviceTagActivityMonitor}
)

func DataTypes() []string {
	return []string{
		DataTypeDailyActivity,
		DataTypeDailyCardiovascularAge,
		DataTypeDailyCyclePhases,
		DataTypeDailyReadiness,
		DataTypeDailyResilience,
		DataTypeDailySleep,
		DataTypeDailySpO2,
		DataTypeDailyStress,
		DataTypeEnhancedTag,
		DataTypeHeartRate,
		DataTypePersonalInfo,
		DataTypeRestModePeriod,
		DataTypeRingBatteryLevel,
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

func EventDataTypes() []string {
	return []string{
		DataTypeDailyActivity,
		DataTypeDailyCardiovascularAge,
		DataTypeDailyCyclePhases,
		DataTypeDailyReadiness,
		DataTypeDailyResilience,
		DataTypeDailySleep,
		DataTypeDailySpO2,
		DataTypeDailyStress,
		DataTypeEnhancedTag,
		DataTypeRestModePeriod,
		DataTypeRingConfiguration,
		DataTypeSession,
		DataTypeSleep,
		DataTypeSleepTime,
		DataTypeVO2Max,
		DataTypeWorkout,
	}
}

func Scopes() []string {
	return []string{
		ScopeDaily,
		ScopeEmail,
		ScopeHeartHealth,
		ScopeHeartRate,
		ScopePersonal,
		ScopeReproductiveCycle,
		ScopeRingConfiguration,
		ScopeSession,
		ScopeSpo2,
		ScopeStress,
		ScopeTag,
		ScopeWorkout,
	}
}

func ScopesForDataTypes(dataTypes []string) []string {
	var scopes []string
	for _, dataType := range dataTypes {
		scopes = append(scopes, ScopesForDataType(dataType)...)
	}
	slices.Sort(scopes)
	return slices.Compact(scopes)
}

func ScopesForDataType(dataType string) []string {
	var scopesForDataType = map[string][]string{
		DataTypeDailyActivity:          {ScopeDaily},
		DataTypeDailyCardiovascularAge: {ScopeHeartHealth},
		DataTypeDailyCyclePhases:       {ScopeReproductiveCycle},
		DataTypeDailyReadiness:         {ScopeDaily},
		DataTypeDailyResilience:        {ScopeStress},
		DataTypeDailySleep:             {ScopeDaily},
		DataTypeDailySpO2:              {ScopeSpo2},
		DataTypeDailyStress:            {ScopeDaily},
		DataTypeEnhancedTag:            {ScopeTag},
		DataTypeHeartRate:              {ScopeHeartRate},
		DataTypePersonalInfo:           {ScopeEmail, ScopePersonal},
		DataTypeRestModePeriod:         {ScopeDaily},
		DataTypeRingBatteryLevel:       {ScopeRingConfiguration},
		DataTypeRingConfiguration:      {ScopeRingConfiguration},
		DataTypeSession:                {ScopeSession},
		DataTypeSleep:                  {ScopeDaily},
		DataTypeSleepTime:              {ScopeDaily},
		DataTypeVO2Max:                 {ScopeHeartHealth},
		DataTypeWorkout:                {ScopeWorkout},
	}
	if scopes, ok := scopesForDataType[dataType]; ok {
		return scopes
	}
	return nil
}

func DataTypeInScopes(dataType string, scopes *[]string) bool {
	if scopes == nil || len(*scopes) == 0 {
		return true
	}
	for _, scope := range *scopes {
		if DataTypeInScope(dataType, scope) {
			return true
		}
	}
	return false
}

func DataTypeInScope(dataType string, scope string) bool {
	return slices.Contains(ScopesForDataType(dataType), scope)
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

	GetData(ctx context.Context, dataType string, timeRange *times.TimeRange, pagination *Pagination, tokenSource oauth.TokenSource) (*DataResponse, error)
	GetDatum(ctx context.Context, dataType string, dataID string, tokenSource oauth.TokenSource) (Datum, error)

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
	validator.String("data_type", c.DataType).Exists().Using(EventDataTypeValidator)
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
	validator.String("data_type", u.DataType).Exists().Using(EventDataTypeValidator)
	validator.String("event_type", u.EventType).Exists().Using(EventTypeValidator)
}

type Subscription struct {
	ID             *string `json:"id,omitempty" bson:"id,omitempty"`
	CallbackURL    *string `json:"callback_url,omitempty" bson:"callback_url,omitempty"`
	DataType       *string `json:"data_type,omitempty" bson:"data_type,omitempty"`
	EventType      *string `json:"event_type,omitempty" bson:"event_type,omitempty"`
	ExpirationTime *string `json:"expiration_time,omitempty" bson:"expiration_time,omitempty"`
}

func ParseSubscription(parser structure.ObjectParser) *Subscription {
	if !parser.Exists() {
		return nil
	}
	datum := &Subscription{}
	datum.Parse(parser)
	return datum
}

func (s *Subscription) Parse(parser structure.ObjectParser) {
	s.ID = parser.String("id")
	s.CallbackURL = parser.String("callback_url")
	s.DataType = parser.String("data_type")
	s.EventType = parser.String("event_type")
	s.ExpirationTime = parser.String("expiration_time")
}

func (s *Subscription) Validate(validator structure.Validator) {
	validator.String("id", s.ID).Exists().Using(DataIDValidator)
	validator.String("callback_url", s.CallbackURL).Exists().NotEmpty()
	validator.String("data_type", s.DataType).Exists().Using(EventDataTypeValidator)
	validator.String("event_type", s.EventType).Exists().Using(EventTypeValidator)
	validator.String("expiration_time", s.ExpirationTime).Exists().AsTime(SubscriptionExpirationTimeFormat).NotZero()
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

type Event struct {
	EventTime *time.Time `json:"event_time,omitempty" bson:"event_time,omitempty"`
	EventType *string    `json:"event_type,omitempty" bson:"event_type,omitempty"`
	UserID    *string    `json:"user_id,omitempty" bson:"user_id,omitempty"`
	ObjectID  *string    `json:"object_id,omitempty" bson:"object_id,omitempty"`
	DataType  *string    `json:"data_type,omitempty" bson:"data_type,omitempty"`
}

func ParseEvent(parser structure.ObjectParser) *Event {
	if !parser.Exists() {
		return nil
	}
	datum := &Event{}
	datum.Parse(parser)
	return datum
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
	validator.String("event_type", e.EventType).Exists().Using(EventTypeValidator)
	validator.String("user_id", e.UserID).Exists().NotEmpty()
	validator.String("object_id", e.ObjectID).Exists().NotEmpty()
	validator.String("data_type", e.DataType).Exists().Using(EventDataTypeValidator)
}

func (e *Event) Hash() (string, error) {
	if bites, err := json.Marshal(e); err != nil {
		return "", errors.Wrap(err, "unable to generate hash")
	} else {
		return crypto.Base64EncodedSHA256Hash(bites), nil
	}
}

const MetadataKeyEvent = "event"

type EventMetadata struct {
	Event *Event `json:"event,omitempty" bson:"event,omitempty"`
}

func ParseEventMetadata(parser structure.ObjectParser) *EventMetadata {
	if !parser.Exists() {
		return nil
	}
	datum := &EventMetadata{}
	datum.Parse(parser)
	return datum
}

func (e *EventMetadata) Parse(parser structure.ObjectParser) {
	e.Event = ParseEvent(parser.WithReferenceObjectParser(MetadataKeyEvent))
}

func (e *EventMetadata) Validate(validator structure.Validator) {
	if e.Event != nil {
		e.Event.Validate(validator.WithReference(MetadataKeyEvent))
	}
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

func (p *PersonalInfo) Hash() (string, error) {
	if bites, err := json.Marshal(p); err != nil {
		return "", errors.Wrap(err, "unable to generate hash")
	} else {
		return crypto.Base64EncodedSHA256Hash(bites), nil
	}
}

type Pagination struct {
	NextToken *string `json:"next_token,omitempty" bson:"next_token,omitempty"`
}

func (p *Pagination) Parse(parser structure.ObjectParser) {
	p.NextToken = parser.String("next_token")
}

func (p *Pagination) Validate(validator structure.Validator) {
	validator.String("next_token", p.NextToken).NotEmpty()
}

func (p *Pagination) HasNext() bool {
	return p.NextToken != nil
}

type DataResponse struct {
	Data       Data `json:"data,omitempty" bson:"data,omitempty"`
	Pagination `bson:",inline"`
}

func (d *DataResponse) Parse(parser structure.ObjectParser) {
	d.Data = ParseData(parser.WithReferenceArrayParser("data"))
	d.Pagination.Parse(parser)
}

func (d *DataResponse) Validate(validator structure.Validator) {
	d.Data.Validate(validator.WithReference("data"))
	d.Pagination.Validate(validator)
}

type DataMap map[string]Data

type Data []Datum

func ParseData(parser structure.ArrayParser) Data {
	if !parser.Exists() {
		return nil
	}
	datum := Data{}
	datum.Parse(parser)
	return datum
}

func (d *Data) Parse(parser structure.ArrayParser) {
	for _, reference := range parser.References() {
		if ptr := parser.Object(reference); ptr != nil {
			*d = append(*d, *ptr)
		}
	}
}

func (d *Data) Validate(validator structure.Validator) {
	for index, datum := range *d {
		if datumValidator := validator.WithReference(strconv.Itoa(index)); datum == nil {
			datumValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

func (d *Data) TimeMaximum() *time.Time {
	var datumTimeMaximum *time.Time
	for _, datum := range *d {
		if datumTime := datum.Time(); datumTime != nil {
			if datumTimeMaximum == nil || datumTime.After(*datumTimeMaximum) {
				datumTimeMaximum = datumTime
			}
		}
	}
	return datumTimeMaximum
}

type Datum map[string]any

func (d Datum) Time() *time.Time {
	if timestampRaw, ok := d["timestamp"]; ok {
		if timestampString, ok := timestampRaw.(string); ok {
			if timestamp, err := time.ParseInLocation(TimeRangeFormat, timestampString, time.UTC); err == nil {
				return pointer.From(timestamp)
			}
		}
	}
	return nil
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

func IsValidEventDataType(value string) bool {
	return ValidateEventDataType(value) == nil
}

func EventDataTypeValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateEventDataType(value))
}

func ValidateEventDataType(value string) error {
	if value == "" {
		return structureValidator.ErrorValueEmpty()
	} else if !slices.Contains(EventDataTypes(), value) {
		return structureValidator.ErrorValueStringNotOneOf(value, EventDataTypes())
	}
	return nil
}
