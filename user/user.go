package user

import (
	"context"
	"regexp"
	"slices"
	"time"

	"github.com/tidepool-org/platform/pointer"

	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	RoleBrokered         = "brokered"
	RoleCarePartner      = "care_partner"
	RoleClinic           = "clinic"
	RoleClinician        = "clinician"
	RoleCustodialAccount = "custodial_account"
	RoleDemo             = "demo"
	RolePatient          = "patient"
)

var rolesMap = map[string]any{
	RoleBrokered:         struct{}{},
	RoleCarePartner:      struct{}{},
	RoleClinic:           struct{}{},
	RoleClinician:        struct{}{},
	RoleCustodialAccount: struct{}{},
	RoleDemo:             struct{}{},
	RolePatient:          struct{}{},
}

func Roles() []string {
	return []string{
		RoleBrokered,
		RoleCarePartner,
		RoleClinic,
		RoleClinician,
		RoleCustodialAccount,
		RoleDemo,
		RolePatient,
	}
}

//go:generate mockgen -source=user.go -destination=test/user_mocks.go -package=test Client
type Client interface {
	Get(ctx context.Context, id string) (*User, error)
}

type User struct {
	UserID        *string   `json:"userid,omitempty" bson:"userid,omitempty"`
	Username      *string   `json:"username,omitempty" bson:"username,omitempty"`
	EmailVerified *bool     `json:"emailVerified,omitempty" bson:"emailVerified,omitempty"`
	TermsAccepted *string   `json:"termsAccepted,omitempty" bson:"termsAccepted,omitempty"`
	Roles         *[]string `json:"roles,omitempty" bson:"roles,omitempty"`
}

func (u *User) Parse(parser structure.ObjectParser) {
	u.UserID = parser.String("userid")
	u.Username = parser.String("username")
	u.EmailVerified = parser.Bool("emailVerified")
	u.TermsAccepted = parser.String("termsAccepted")
	u.Roles = parser.StringArray("roles")
	if u.Roles != nil {
		u.Roles = pointer.FromAny(slices.DeleteFunc(*u.Roles, func(role string) bool {
			_, validRole := rolesMap[role]
			return !validRole
		}))
	}
	parser.Bool("passwordExists")
	parser.StringArray("emails")
}

func (u *User) Validate(validator structure.Validator) {
	validator.String("userid", u.UserID).Exists().Using(IDValidator)
	validator.String("username", u.Username).Exists().NotEmpty()
	validator.String("termsAccepted", u.TermsAccepted).AsTime(time.RFC3339Nano).NotZero()
	validator.StringArray("roles", u.Roles).EachOneOf(Roles()...).EachUnique()
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

func (u *User) IsPatient() bool {
	if u.Roles == nil || len(*u.Roles) == 0 || u.HasRole(RolePatient) {
		return true
	}
	return false
}

func (u *User) Sanitize(details request.AuthDetails) error {
	if details == nil || (!details.IsService() && details.UserID() != *u.UserID) {
		u.Username = nil
		u.EmailVerified = nil
		u.TermsAccepted = nil
		u.Roles = nil
	}
	return nil
}

type UserArray []*User

func (u UserArray) Sanitize(details request.AuthDetails) error {
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
