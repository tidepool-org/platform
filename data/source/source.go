package source

import (
	"context"
	"net/http"
	"regexp"
	"slices"
	"time"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
)

const (
	MetadataLengthMaximum = 4 * 1024

	StateConnected    = "connected"
	StateDisconnected = "disconnected"
	StateError        = "error"
)

func States() []string {
	return []string{
		StateConnected,
		StateDisconnected,
		StateError,
	}
}

//go:generate mockgen -destination=./test/mock.go -package test . Client

type Client interface {
	List(ctx context.Context, filter *Filter, pagination *page.Pagination) (SourceArray, error)
	Create(ctx context.Context, userID string, create *Create) (*Source, error)
	DeleteAll(ctx context.Context, userID string) error

	Get(ctx context.Context, id string) (*Source, error)
	Update(ctx context.Context, id string, condition *request.Condition, update *Update) (*Source, error)
	Delete(ctx context.Context, id string, condition *request.Condition) (bool, error)
}

type Filter struct {
	ProviderType       *[]string `json:"providerType,omitempty"`
	ProviderName       *[]string `json:"providerName,omitempty"`
	ProviderSessionID  *[]string `json:"providerSessionId,omitempty"`
	ProviderExternalID *[]string `json:"providerExternalId,omitempty"`
	State              *[]string `json:"state,omitempty"`
	UserID             *string   `json:"userId,omitempty"`
}

func NewFilter() *Filter {
	return &Filter{}
}

func (f *Filter) Parse(parser structure.ObjectParser) {
	f.ProviderType = parser.StringArray("providerType")
	f.ProviderName = parser.StringArray("providerName")
	f.ProviderSessionID = parser.StringArray("providerSessionId")
	f.ProviderExternalID = parser.StringArray("providerExternalId")
	f.State = parser.StringArray("state")
	// The user id ought to be provided in the request path.
	// Ignore user ids provided in the query parameters
	_ = parser.String("userId")
}

func (f *Filter) Validate(validator structure.Validator) {
	validator.StringArray("providerType", f.ProviderType).NotEmpty().EachOneOf(auth.ProviderTypes()...).EachUnique()
	validator.StringArray("providerName", f.ProviderName).NotEmpty().EachUsing(auth.ProviderNameValidator).EachUnique()
	validator.StringArray("providerSessionId", f.ProviderSessionID).NotEmpty().EachUsing(auth.ProviderSessionIDValidator).EachUnique()
	validator.StringArray("providerExternalId", f.ProviderExternalID).NotEmpty().EachUsing(auth.ProviderExternalIDValidator).EachUnique()
	validator.StringArray("state", f.State).NotEmpty().EachOneOf(States()...).EachUnique()
	validator.String("userId", f.UserID).Using(user.IDValidator)
}

func (f *Filter) MutateRequest(req *http.Request) error {
	parameters := map[string][]string{}
	if f.ProviderType != nil {
		parameters["providerType"] = *f.ProviderType
	}
	if f.ProviderName != nil {
		parameters["providerName"] = *f.ProviderName
	}
	if f.ProviderSessionID != nil {
		parameters["providerSessionId"] = *f.ProviderSessionID
	}
	if f.ProviderExternalID != nil {
		parameters["providerExternalId"] = *f.ProviderExternalID
	}
	if f.State != nil {
		parameters["state"] = *f.State
	}
	return request.NewArrayParametersMutator(parameters).MutateRequest(req)
}

type Create struct {
	ProviderType       *string        `json:"providerType,omitempty"`
	ProviderName       *string        `json:"providerName,omitempty"`
	ProviderSessionID  *string        `json:"providerSessionId,omitempty"`
	ProviderExternalID *string        `json:"providerExternalId,omitempty"`
	Metadata           map[string]any `json:"metadata,omitempty"`
}

func NewCreate() *Create {
	return &Create{}
}

func (c *Create) Parse(parser structure.ObjectParser) {
	c.ProviderType = parser.String("providerType")
	c.ProviderName = parser.String("providerName")
	c.ProviderSessionID = parser.String("providerSessionId")
	c.ProviderExternalID = parser.String("providerExternalId")
	if ptr := parser.Object("metadata"); ptr != nil {
		c.Metadata = *ptr
	}
}

func (c *Create) Validate(validator structure.Validator) {
	validator.String("providerType", c.ProviderType).Exists().OneOf(auth.ProviderTypes()...)
	validator.String("providerName", c.ProviderName).Exists().Using(auth.ProviderNameValidator)
	validator.String("providerSessionId", c.ProviderSessionID).Using(auth.ProviderSessionIDValidator)
	validator.String("providerExternalId", c.ProviderExternalID).Using(auth.ProviderExternalIDValidator)
	validator.Object("metadata", &c.Metadata).SizeLessThanOrEqualTo(MetadataLengthMaximum)
}

type Update struct {
	ProviderSessionID  *string              `json:"providerSessionId,omitempty"`
	ProviderExternalID *string              `json:"providerExternalId,omitempty"`
	State              *string              `json:"state,omitempty"`
	Metadata           map[string]any       `json:"metadata,omitempty"`
	Error              *errors.Serializable `json:"error,omitempty"`
	DataSetIDs         *[]string            `json:"dataSetIds,omitempty"`
	EarliestDataTime   *time.Time           `json:"earliestDataTime,omitempty"`
	LatestDataTime     *time.Time           `json:"latestDataTime,omitempty"`
	LastImportTime     *time.Time           `json:"lastImportTime,omitempty"`
}

func NewUpdate() *Update {
	return &Update{}
}

func (u *Update) Parse(parser structure.ObjectParser) {
	u.ProviderSessionID = parser.String("providerSessionId")
	u.ProviderExternalID = parser.String("providerExternalId")
	u.State = parser.String("state")
	if ptr := parser.Object("metadata"); ptr != nil {
		u.Metadata = *ptr
	}
	if parser.ReferenceExists("error") {
		serializable := &errors.Serializable{}
		serializable.Parse("error", parser)
		if serializable.Error != nil {
			u.Error = serializable
		}
	}
	u.DataSetIDs = parser.StringArray("dataSetIds")
	u.EarliestDataTime = parser.Time("earliestDataTime", time.RFC3339Nano)
	u.LatestDataTime = parser.Time("latestDataTime", time.RFC3339Nano)
	u.LastImportTime = parser.Time("lastImportTime", time.RFC3339Nano)
}

func (u *Update) Validate(validator structure.Validator) {
	if providerSessionIDValidator := validator.String("providerSessionId", u.ProviderSessionID); u.State == nil || *u.State != StateConnected {
		providerSessionIDValidator.NotExists()
	} else {
		providerSessionIDValidator.Exists().Using(auth.ProviderSessionIDValidator)
	}
	validator.String("providerExternalId", u.ProviderExternalID).Using(auth.ProviderExternalIDValidator)
	validator.String("state", u.State).OneOf(States()...)
	validator.Object("metadata", &u.Metadata).SizeLessThanOrEqualTo(MetadataLengthMaximum)
	if u.Error != nil {
		u.Error.Validate(validator.WithReference("error"))
	}
	validator.StringArray("dataSetIds", u.DataSetIDs).NotEmpty().EachUsing(data.SetIDValidator).EachUnique()
	validator.Time("earliestDataTime", u.EarliestDataTime).NotZero().BeforeNow(time.Second)
	validator.Time("latestDataTime", u.LatestDataTime).NotZero().After(pointer.ToTime(u.EarliestDataTime)).BeforeNow(time.Second)
	validator.Time("lastImportTime", u.LastImportTime).NotZero().BeforeNow(time.Second)
}

func (u *Update) Normalize(normalizer structure.Normalizer) {
	if u.Error != nil {
		u.Error.Normalize(normalizer.WithReference("error"))
	}
}

func (u *Update) IsEmpty() bool {
	return u.ProviderSessionID == nil && u.ProviderExternalID == nil && u.State == nil && u.Metadata == nil && u.Error == nil && u.DataSetIDs == nil && u.EarliestDataTime == nil && u.LatestDataTime == nil && u.LastImportTime == nil
}

type Source struct {
	ID                 *string              `json:"id,omitempty" bson:"id,omitempty"`
	UserID             *string              `json:"userId,omitempty" bson:"userId,omitempty"`
	ProviderType       *string              `json:"providerType,omitempty" bson:"providerType,omitempty"`
	ProviderName       *string              `json:"providerName,omitempty" bson:"providerName,omitempty"`
	ProviderSessionID  *string              `json:"providerSessionId,omitempty" bson:"providerSessionId,omitempty"`
	ProviderExternalID *string              `json:"providerExternalId,omitempty" bson:"providerExternalId,omitempty"`
	State              *string              `json:"state,omitempty" bson:"state,omitempty"`
	Metadata           map[string]any       `json:"metadata,omitempty" bson:"metadata,omitempty"`
	Error              *errors.Serializable `json:"error,omitempty" bson:"error,omitempty"`
	DataSetIDs         *[]string            `json:"dataSetIds,omitempty" bson:"dataSetIds,omitempty"`
	EarliestDataTime   *time.Time           `json:"earliestDataTime,omitempty" bson:"earliestDataTime,omitempty"`
	LatestDataTime     *time.Time           `json:"latestDataTime,omitempty" bson:"latestDataTime,omitempty"`
	LastImportTime     *time.Time           `json:"lastImportTime,omitempty" bson:"lastImportTime,omitempty"`
	CreatedTime        *time.Time           `json:"createdTime,omitempty" bson:"createdTime,omitempty"`
	ModifiedTime       *time.Time           `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
	Revision           *int                 `json:"revision,omitempty" bson:"revision,omitempty"`
}

func (s *Source) Parse(parser structure.ObjectParser) {
	s.ID = parser.String("id")
	s.UserID = parser.String("userId")
	s.ProviderType = parser.String("providerType")
	s.ProviderName = parser.String("providerName")
	s.ProviderSessionID = parser.String("providerSessionId")
	s.ProviderExternalID = parser.String("providerExternalId")
	s.State = parser.String("state")
	if ptr := parser.Object("metadata"); ptr != nil {
		s.Metadata = *ptr
	}
	if parser.ReferenceExists("error") {
		serializable := &errors.Serializable{}
		serializable.Parse("error", parser)
		if serializable.Error != nil {
			s.Error = serializable
		}
	}
	s.DataSetIDs = parser.StringArray("dataSetIds")
	s.EarliestDataTime = parser.Time("earliestDataTime", time.RFC3339Nano)
	s.LatestDataTime = parser.Time("latestDataTime", time.RFC3339Nano)
	s.LastImportTime = parser.Time("lastImportTime", time.RFC3339Nano)
	s.CreatedTime = parser.Time("createdTime", time.RFC3339Nano)
	s.ModifiedTime = parser.Time("modifiedTime", time.RFC3339Nano)
	s.Revision = parser.Int("revision")
}

func (s *Source) Validate(validator structure.Validator) {
	validator.String("id", s.ID).Exists().Using(IDValidator)
	validator.String("userId", s.UserID).Exists().Using(user.IDValidator)
	validator.String("providerType", s.ProviderType).Exists().OneOf(auth.ProviderTypes()...)
	validator.String("providerName", s.ProviderName).Exists().Using(auth.ProviderNameValidator)
	if providerSessionIDValidator := validator.String("providerSessionId", s.ProviderSessionID); s.State == nil {
		providerSessionIDValidator.Using(auth.ProviderSessionIDValidator)
	} else if *s.State != StateDisconnected {
		providerSessionIDValidator.Exists().Using(auth.ProviderSessionIDValidator)
	} else {
		providerSessionIDValidator.NotExists()
	}
	validator.String("providerExternalId", s.ProviderExternalID).Using(auth.ProviderExternalIDValidator)
	validator.String("state", s.State).Exists().OneOf(States()...)
	validator.Object("metadata", &s.Metadata).SizeLessThanOrEqualTo(MetadataLengthMaximum)
	if s.Error != nil {
		s.Error.Validate(validator.WithReference("error"))
	}
	validator.StringArray("dataSetIds", s.DataSetIDs).NotEmpty().EachUsing(data.SetIDValidator).EachUnique()
	validator.Time("earliestDataTime", s.EarliestDataTime).NotZero().BeforeNow(time.Second)
	validator.Time("latestDataTime", s.LatestDataTime).NotZero().After(pointer.ToTime(s.EarliestDataTime)).BeforeNow(time.Second)
	validator.Time("lastImportTime", s.LastImportTime).NotZero().BeforeNow(time.Second)
	validator.Time("createdTime", s.CreatedTime).Exists().NotZero().BeforeNow(time.Second)
	validator.Time("modifiedTime", s.ModifiedTime).NotZero().After(pointer.ToTime(s.CreatedTime)).BeforeNow(time.Second)
	validator.Int("revision", s.Revision).Exists().GreaterThanOrEqualTo(0)
}

func (s *Source) Normalize(normalizer structure.Normalizer) {
	if s.Error != nil {
		s.Error.Normalize(normalizer.WithReference("error"))
	}
}

func (s *Source) Sanitize(details request.AuthDetails) error {
	if details == nil {
		return errors.New("unable to sanitize")
	}

	if details.IsUser() {
		s.ProviderSessionID = nil
		if s.Error != nil && s.Error.Error != nil {
			// TODO: Is there a way to make this a more general use case?
			// TODO: Check all production data source errors for examples.
			if cause := errors.Cause(s.Error.Error); request.IsErrorUnauthenticated(cause) {
				s.Error.Error = cause
			}
			s.Error.Error = errors.Sanitize(s.Error.Error)
		}
	}

	return nil
}

func (s *Source) EnsureMetadata() {
	if s.Metadata == nil {
		s.Metadata = map[string]any{}
	}
}

func (s *Source) HasError() bool {
	return s.Error != nil && s.Error.Error != nil
}

func (s *Source) GetError() error {
	if s.Error != nil {
		return s.Error.Error
	}
	return nil
}

type SourceArray []*Source

func (s SourceArray) Sanitize(details request.AuthDetails) error {
	for _, datum := range s {
		if err := datum.Sanitize(details); err != nil {
			return err
		}
	}
	return nil
}

func (s *Source) AddDataSetID(dataSetID string) bool {
	if s.DataSetIDs == nil {
		s.DataSetIDs = &[]string{}
	}
	if slices.Contains(*s.DataSetIDs, dataSetID) {
		return false
	}
	*s.DataSetIDs = append(*s.DataSetIDs, dataSetID)
	return true
}

func NewID() string {
	return id.Must(id.New(16))
}

func IsValidID(value string) bool {
	return ValidateID(value) == nil
}

func IDValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateID(value))
}

func ValidateID(value string) error {
	if value == "" {
		return structureValidator.ErrorValueEmpty()
	} else if !idExpression.MatchString(value) {
		return ErrorValueStringAsIDNotValid(value)
	}
	return nil
}

var idExpression = regexp.MustCompile("^[0-9a-z]{32}$")

func ErrorValueStringAsIDNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as data source id", value)
}
