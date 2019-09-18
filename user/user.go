package user

import (
	"context"
	"regexp"
	"time"

	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	RoleClinic = "clinic"
)

func Roles() []string {
	return []string{
		RoleClinic,
	}
}

type Client interface {
	Get(ctx context.Context, id string) (*User, error)
	Delete(ctx context.Context, id string, deleet *Delete, condition *request.Condition) (bool, error)
}

type Delete struct {
	Password *string `json:"password,omitempty"`
}

func NewDelete() *Delete {
	return &Delete{}
}

func (d *Delete) Parse(parser structure.ObjectParser) {
	d.Password = parser.String("password")
}

func (d *Delete) Validate(validator structure.Validator) {
	validator.String("password", d.Password).NotEmpty()
}

type User struct {
	UserID        *string    `json:"userid,omitempty" bson:"userid,omitempty"`     // TODO: Rename ID/id
	Username      *string    `json:"username,omitempty" bson:"username,omitempty"` // TODO: Rename Email/email
	PasswordHash  *string    `json:"-" bson:"pwhash,omitempty"`
	Authenticated *bool      `json:"authenticated,omitempty" bson:"authenticated,omitempty"` // TODO: Rename EmaiLVerified/emailVerified
	TermsAccepted *string    `json:"termsAccepted,omitempty" bson:"termsAccepted,omitempty"`
	Roles         *[]string  `json:"roles,omitempty" bson:"roles,omitempty"`
	CreatedTime   *time.Time `json:"createdTime,omitempty" bson:"createdTime,omitempty"`
	ModifiedTime  *time.Time `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
	DeletedTime   *time.Time `json:"deletedTime,omitempty" bson:"deletedTime,omitempty"`
	Revision      *int       `json:"revision,omitempty" bson:"revision,omitempty"`
}

func (u *User) Parse(parser structure.ObjectParser) {
	u.UserID = parser.String("userid")
	u.Username = parser.String("username")
	u.Authenticated = parser.Bool("authenticated")
	u.TermsAccepted = parser.String("termsAccepted")
	u.Roles = parser.StringArray("roles")
	u.CreatedTime = parser.Time("createdTime", time.RFC3339Nano)
	u.ModifiedTime = parser.Time("modifiedTime", time.RFC3339Nano)
	u.DeletedTime = parser.Time("deletedTime", time.RFC3339Nano)
	u.Revision = parser.Int("revision")
}

func (u *User) Validate(validator structure.Validator) {
	validator.String("userid", u.UserID).Exists().Using(IDValidator)
	validator.String("username", u.Username).Exists().NotEmpty()
	validator.String("termsAccepted", u.TermsAccepted).AsTime(time.RFC3339Nano).NotZero()
	validator.StringArray("roles", u.Roles).EachOneOf(Roles()...).EachUnique()
	validator.Time("createdTime", u.CreatedTime).Exists().NotZero().BeforeNow(time.Second)
	validator.Time("modifiedTime", u.ModifiedTime).NotZero().After(pointer.ToTime(u.CreatedTime)).BeforeNow(time.Second)
	validator.Time("deletedTime", u.DeletedTime).NotZero().After(pointer.ToTime(u.CreatedTime)).BeforeNow(time.Second)
	validator.Int("revision", u.Revision).Exists().GreaterThanOrEqualTo(0)
}

func (u *User) HasRole(role string) bool {
	if u.Roles != nil {
		for _, r := range *u.Roles {
			if r == role {
				return true
			}
		}
	}
	return false
}

func (u *User) Sanitize(details request.Details) error {
	if details == nil || (!details.IsService() && details.UserID() != *u.UserID) {
		u.Username = nil
		u.Authenticated = nil
		u.TermsAccepted = nil
		u.Roles = nil
	}
	return nil
}

type UserArray []*User

func (u UserArray) Sanitize(details request.Details) error {
	for _, datum := range u {
		if err := datum.Sanitize(details); err != nil {
			return err
		}
	}
	return nil
}

func NewID() string {
	return id.Must(id.New(5))
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

var idExpression = regexp.MustCompile(`^([0-9a-f]{10}|[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12})$`)
