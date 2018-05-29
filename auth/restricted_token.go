package auth

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
)

const MaximumExpirationDuration = time.Hour

var pathExpression = regexp.MustCompile("^/.*$")

type RestrictedTokenAccessor interface {
	ListUserRestrictedTokens(ctx context.Context, userID string, filter *RestrictedTokenFilter, pagination *page.Pagination) (RestrictedTokens, error)
	CreateUserRestrictedToken(ctx context.Context, userID string, create *RestrictedTokenCreate) (*RestrictedToken, error)
	GetRestrictedToken(ctx context.Context, id string) (*RestrictedToken, error)
	UpdateRestrictedToken(ctx context.Context, id string, update *RestrictedTokenUpdate) (*RestrictedToken, error)
	DeleteRestrictedToken(ctx context.Context, id string) error
}

type RestrictedTokenFilter struct{}

func NewRestrictedTokenFilter() *RestrictedTokenFilter {
	return &RestrictedTokenFilter{}
}

func (r *RestrictedTokenFilter) Parse(parser structure.ObjectParser) {
}

func (r *RestrictedTokenFilter) Validate(validator structure.Validator) {
}

func (r *RestrictedTokenFilter) Mutate(req *http.Request) error {
	return nil
}

type RestrictedTokenCreate struct {
	Paths          *[]string  `json:"paths,omitempty"`
	ExpirationTime *time.Time `json:"expirationTime,omitempty"`
}

func NewRestrictedTokenCreate() *RestrictedTokenCreate {
	return &RestrictedTokenCreate{
		ExpirationTime: pointer.FromTime(time.Now().Add(MaximumExpirationDuration).Truncate(time.Second)),
	}
}

func (r *RestrictedTokenCreate) Parse(parser structure.ObjectParser) {
	r.Paths = parser.StringArray("paths")
	r.ExpirationTime = parser.Time("expirationTime", time.RFC3339)
}

func (r *RestrictedTokenCreate) Validate(validator structure.Validator) {
	validator.StringArray("paths", r.Paths).LengthInRange(1, 10).EachMatches(pathExpression)
	validator.Time("expirationTime", r.ExpirationTime).Before(time.Now().Add(MaximumExpirationDuration))
}

func (r *RestrictedTokenCreate) Normalize(normalizer structure.Normalizer) {
	if r.ExpirationTime != nil {
		r.ExpirationTime = pointer.FromTime((*r.ExpirationTime).Truncate(time.Second))
	}
}

type RestrictedTokenUpdate struct {
	Paths          *[]string  `json:"paths,omitempty"`
	ExpirationTime *time.Time `json:"expirationTime,omitempty"`
}

func NewRestrictedTokenUpdate() *RestrictedTokenUpdate {
	return &RestrictedTokenUpdate{}
}

func (r *RestrictedTokenUpdate) HasUpdates() bool {
	return r.Paths != nil || r.ExpirationTime != nil
}

func (r *RestrictedTokenUpdate) Parse(parser structure.ObjectParser) {
	r.Paths = parser.StringArray("paths")
	r.ExpirationTime = parser.Time("expirationTime", time.RFC3339)
}

func (r *RestrictedTokenUpdate) Validate(validator structure.Validator) {
	validator.StringArray("paths", r.Paths).LengthInRange(1, 10).EachMatches(pathExpression)
	validator.Time("expirationTime", r.ExpirationTime).Before(time.Now().Add(MaximumExpirationDuration))
}

func (r *RestrictedTokenUpdate) Normalize(normalizer structure.Normalizer) {
	if r.ExpirationTime != nil {
		r.ExpirationTime = pointer.FromTime((*r.ExpirationTime).Truncate(time.Second))
	}
}

func NewRestrictedTokenID() string {
	return id.Must(id.New(16))
}

func IsValidRestrictedTokenID(value string) bool {
	return ValidateRestrictedTokenID(value) == nil
}

func RestrictedTokenIDValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateRestrictedTokenID(value))
}

func ValidateRestrictedTokenID(value string) error {
	if value == "" {
		return structureValidator.ErrorValueEmpty()
	} else if !restrictedTokenIDExpression.MatchString(value) {
		return ErrorValueStringAsRestrictedTokenIDNotValid(value)
	}
	return nil
}

func ErrorValueStringAsRestrictedTokenIDNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as restricted token id", value)
}

var restrictedTokenIDExpression = regexp.MustCompile("^[0-9a-z]{32}$")

type RestrictedToken struct {
	ID             string     `json:"id" bson:"id"`
	UserID         string     `json:"userId" bson:"userId"`
	Paths          *[]string  `json:"paths,omitempty" bson:"paths,omitempty"`
	ExpirationTime time.Time  `json:"expirationTime" bson:"expirationTime"`
	CreatedTime    time.Time  `json:"createdTime" bson:"createdTime"`
	ModifiedTime   *time.Time `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
}

func NewRestrictedToken(userID string, create *RestrictedTokenCreate) (*RestrictedToken, error) {
	if userID == "" {
		return nil, errors.New("user id is missing")
	}
	if create == nil {
		return nil, errors.New("create is missing")
	} else if err := structureValidator.New().Validate(create); err != nil {
		return nil, errors.Wrap(err, "create is invalid")
	}

	restrictedToken := &RestrictedToken{
		ID:          NewRestrictedTokenID(),
		UserID:      userID,
		Paths:       create.Paths,
		CreatedTime: time.Now().Truncate(time.Second),
	}
	if create.ExpirationTime != nil {
		restrictedToken.ExpirationTime = (*create.ExpirationTime).Truncate(time.Second)
	} else {
		restrictedToken.ExpirationTime = time.Now().Add(MaximumExpirationDuration).Truncate(time.Second)
	}

	return restrictedToken, nil
}

func (r *RestrictedToken) Parse(parser structure.ObjectParser) {
	if ptr := parser.String("id"); ptr != nil {
		r.ID = *ptr
	}
	if ptr := parser.String("userId"); ptr != nil {
		r.UserID = *ptr
	}
	r.Paths = parser.StringArray("paths")
	if ptr := parser.Time("expirationTime", time.RFC3339); ptr != nil {
		r.ExpirationTime = *ptr
	}
	if ptr := parser.Time("createdTime", time.RFC3339); ptr != nil {
		r.CreatedTime = *ptr
	}
	r.ModifiedTime = parser.Time("modifiedTime", time.RFC3339)
}

func (r *RestrictedToken) Validate(validator structure.Validator) {
	validator.String("id", &r.ID).Using(RestrictedTokenIDValidator)
	validator.String("userId", &r.UserID).Using(user.IDValidator)
	validator.StringArray("paths", r.Paths).LengthInRange(1, 10).EachMatches(pathExpression)
	validator.Time("expirationTime", &r.ExpirationTime).Before(time.Now().Add(MaximumExpirationDuration))
	validator.Time("createdTime", &r.CreatedTime).NotZero().BeforeNow(time.Second)
	validator.Time("modifiedTime", r.ModifiedTime).After(r.CreatedTime).BeforeNow(time.Second)
}

func (r *RestrictedToken) Authenticates(req *http.Request) bool {
	if req == nil || req.URL == nil {
		return false
	}
	if time.Now().After(r.ExpirationTime) {
		return false
	}
	if r.Paths != nil {
		escapedPath := req.URL.EscapedPath()
		for _, path := range *r.Paths {
			if path == escapedPath || strings.HasPrefix(escapedPath, strings.TrimSuffix(path, "/")+"/") {
				return true
			}
		}
		return false
	}
	return true
}

func (r *RestrictedToken) Sanitize(details request.Details) error {
	if details != nil && (details.IsService() || details.UserID() == r.UserID) {
		return nil
	}
	return errors.New("unable to sanitize")
}

type RestrictedTokens []*RestrictedToken

func (r RestrictedTokens) Sanitize(details request.Details) error {
	for _, restrictedToken := range r {
		if err := restrictedToken.Sanitize(details); err != nil {
			return err
		}
	}
	return nil
}
