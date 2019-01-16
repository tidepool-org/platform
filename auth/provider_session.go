package auth

import (
	"context"
	"net/http"
	"regexp"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/oauth"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
)

const (
	ProviderTypeOAuth = "oauth"
)

func ProviderTypes() []string {
	return []string{
		ProviderTypeOAuth,
	}
}

type ProviderSessionAccessor interface {
	ListUserProviderSessions(ctx context.Context, userID string, filter *ProviderSessionFilter, pagination *page.Pagination) (ProviderSessions, error)
	CreateUserProviderSession(ctx context.Context, userID string, create *ProviderSessionCreate) (*ProviderSession, error)
	GetProviderSession(ctx context.Context, id string) (*ProviderSession, error)
	UpdateProviderSession(ctx context.Context, id string, update *ProviderSessionUpdate) (*ProviderSession, error)
	DeleteProviderSession(ctx context.Context, id string) error
}

type ProviderSessionFilter struct {
	Type *string `json:"type,omitempty" bson:"type,omitempty"`
	Name *string `json:"name,omitempty" bson:"name,omitempty"`
}

func NewProviderSessionFilter() *ProviderSessionFilter {
	return &ProviderSessionFilter{}
}

func (p *ProviderSessionFilter) Parse(parser structure.ObjectParser) {
	p.Type = parser.String("type")
	p.Name = parser.String("name")
}

func (p *ProviderSessionFilter) Validate(validator structure.Validator) {
	validator.String("type", p.Type).OneOf(ProviderTypes()...)
	validator.String("name", p.Name).NotEmpty()
}

func (p *ProviderSessionFilter) MutateRequest(req *http.Request) error {
	parameters := map[string]string{}
	if p.Type != nil {
		parameters["type"] = *p.Type
	}
	if p.Name != nil {
		parameters["name"] = *p.Name
	}
	return request.NewParametersMutator(parameters).MutateRequest(req)
}

type ProviderSessionCreate struct {
	Type       string       `json:"type" bson:"type"`
	Name       string       `json:"name" bson:"name"`
	OAuthToken *oauth.Token `json:"oauthToken,omitempty" bson:"oauthToken,omitempty"`
}

func NewProviderSessionCreate() *ProviderSessionCreate {
	return &ProviderSessionCreate{}
}

func (p *ProviderSessionCreate) Parse(parser structure.ObjectParser) {
	if ptr := parser.String("type"); ptr != nil {
		p.Type = *ptr
	}
	if ptr := parser.String("name"); ptr != nil {
		p.Name = *ptr
	}
	if oauthTokenParser := parser.WithReferenceObjectParser("oauthToken"); oauthTokenParser.Exists() {
		p.OAuthToken = oauth.NewToken()
		p.OAuthToken.Parse(oauthTokenParser)
		oauthTokenParser.NotParsed()
	}
}

func (p *ProviderSessionCreate) Validate(validator structure.Validator) {
	validator.String("type", &p.Type).OneOf(ProviderTypes()...)
	validator.String("name", &p.Name).NotEmpty()
	switch p.Type {
	case ProviderTypeOAuth:
		if oauthTokenValidator := validator.WithReference("oauthToken"); p.OAuthToken != nil {
			p.OAuthToken.Validate(oauthTokenValidator)
		} else {
			oauthTokenValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

type ProviderSessionUpdate struct {
	OAuthToken *oauth.Token `json:"oauthToken,omitempty" bson:"oauthToken,omitempty"`
}

func NewProviderSessionUpdate() *ProviderSessionUpdate {
	return &ProviderSessionUpdate{}
}

func (p *ProviderSessionUpdate) Parse(parser structure.ObjectParser) {
	if oauthTokenParser := parser.WithReferenceObjectParser("oauthToken"); oauthTokenParser.Exists() {
		p.OAuthToken = oauth.NewToken()
		p.OAuthToken.Parse(oauthTokenParser)
		oauthTokenParser.NotParsed()
	}
}

func (p *ProviderSessionUpdate) Validate(validator structure.Validator) {
	if p.OAuthToken != nil {
		p.OAuthToken.Validate(validator.WithReference("oauthToken"))
	}
}

func (p *ProviderSessionUpdate) IsEmpty() bool {
	return p.OAuthToken == nil
}

func NewProviderSessionID() string {
	return id.Must(id.New(16))
}

func IsValidProviderSessionID(value string) bool {
	return ValidateProviderSessionID(value) == nil
}

func ProviderSessionIDValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateProviderSessionID(value))
}

func ValidateProviderSessionID(value string) error {
	if value == "" {
		return structureValidator.ErrorValueEmpty()
	} else if !providerSessionIDExpression.MatchString(value) {
		return ErrorValueStringAsProviderSessionIDNotValid(value)
	}
	return nil
}

func ErrorValueStringAsProviderSessionIDNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as provider session id", value)
}

var providerSessionIDExpression = regexp.MustCompile("^[0-9a-z]{32}$")

const ProviderNameLengthMaximum = 100

func IsValidProviderName(value string) bool {
	return ValidateProviderName(value) == nil
}

func ProviderNameValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateProviderName(value))
}

func ValidateProviderName(value string) error {
	if value == "" {
		return structureValidator.ErrorValueEmpty()
	} else if length := len(value); length > ProviderNameLengthMaximum {
		return structureValidator.ErrorLengthNotLessThanOrEqualTo(length, ProviderNameLengthMaximum)
	}
	return nil
}

type ProviderSession struct {
	ID           string       `json:"id" bson:"id"`
	UserID       string       `json:"userId" bson:"userId"`
	Type         string       `json:"type" bson:"type"`
	Name         string       `json:"name" bson:"name"`
	OAuthToken   *oauth.Token `json:"oauthToken,omitempty" bson:"oauthToken,omitempty"`
	CreatedTime  time.Time    `json:"createdTime" bson:"createdTime"`
	ModifiedTime *time.Time   `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
}

func NewProviderSession(userID string, create *ProviderSessionCreate) (*ProviderSession, error) {
	if userID == "" {
		return nil, errors.New("user id is missing")
	}
	if create == nil {
		return nil, errors.New("create is missing")
	} else if err := structureValidator.New().Validate(create); err != nil {
		return nil, errors.Wrap(err, "create is invalid")
	}

	return &ProviderSession{
		ID:          NewProviderSessionID(),
		UserID:      userID,
		Type:        create.Type,
		Name:        create.Name,
		OAuthToken:  create.OAuthToken,
		CreatedTime: time.Now(),
	}, nil
}

func (p *ProviderSession) Parse(parser structure.ObjectParser) {
	if ptr := parser.String("id"); ptr != nil {
		p.ID = *ptr
	}
	if ptr := parser.String("userId"); ptr != nil {
		p.UserID = *ptr
	}
	if ptr := parser.String("type"); ptr != nil {
		p.Type = *ptr
	}
	if ptr := parser.String("name"); ptr != nil {
		p.Name = *ptr
	}
	if oauthTokenParser := parser.WithReferenceObjectParser("oauthToken"); oauthTokenParser.Exists() {
		p.OAuthToken = oauth.NewToken()
		p.OAuthToken.Parse(oauthTokenParser)
		oauthTokenParser.NotParsed()
	}
	if ptr := parser.Time("createdTime", time.RFC3339Nano); ptr != nil {
		p.CreatedTime = *ptr
	}
	p.ModifiedTime = parser.Time("modifiedTime", time.RFC3339Nano)
}

func (p *ProviderSession) Validate(validator structure.Validator) {
	validator.String("id", &p.ID).Using(ProviderSessionIDValidator)
	validator.String("userId", &p.UserID).Using(user.IDValidator)
	validator.String("type", &p.Type).OneOf(ProviderTypes()...)
	validator.String("name", &p.Name).NotEmpty()
	switch p.Type {
	case ProviderTypeOAuth:
		if oauthTokenValidator := validator.WithReference("oauthToken"); p.OAuthToken != nil {
			p.OAuthToken.Validate(oauthTokenValidator)
		} else {
			oauthTokenValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
	validator.Time("createdTime", &p.CreatedTime).NotZero().BeforeNow(time.Second)
	validator.Time("modifiedTime", p.ModifiedTime).After(p.CreatedTime).BeforeNow(time.Second)
}

func (p *ProviderSession) Sanitize(details request.Details) error {
	if details != nil && details.IsService() {
		return nil
	}
	return errors.New("unable to sanitize")
}

type ProviderSessions []*ProviderSession

func (p ProviderSessions) Sanitize(details request.Details) error {
	for _, providerSession := range p {
		if err := providerSession.Sanitize(details); err != nil {
			return err
		}
	}
	return nil
}
