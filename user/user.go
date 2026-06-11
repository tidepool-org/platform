package user

import (
	"context"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/tidepool-org/platform/pointer"

	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/permission"
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

var (
	rolesMap = map[string]any{
		RoleBrokered:         struct{}{},
		RoleCarePartner:      struct{}{},
		RoleClinic:           struct{}{},
		RoleClinician:        struct{}{},
		RoleCustodialAccount: struct{}{},
		RoleDemo:             struct{}{},
		RolePatient:          struct{}{},
	}

	idExpression           = regexp.MustCompile(`^([0-9a-f]{10}|[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12})$`)
	custodialAccountRegexp = regexp.MustCompile(`(?i)unclaimed-custodial-automation\+\d+@tidepool\.org`)
)

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
	UserID               *string   `json:"userid,omitempty" bson:"userid,omitempty"`
	Username             *string   `json:"username,omitempty" bson:"username,omitempty"`
	EmailVerified        *bool     `json:"emailVerified,omitempty" bson:"emailVerified,omitempty"`
	TermsAccepted        *string   `json:"termsAccepted,omitempty" bson:"termsAccepted,omitempty"`
	Roles                *[]string `json:"roles,omitempty" bson:"roles,omitempty"`
	Emails               []string  `json:"emails,omitempty" bson:"emails,omitempty"`
	PwHash               string    `json:"-" bson:"pwhash,omitempty"`
	Hash                 string    `json:"-" bson:"userhash,omitempty"`
	IsMigrated           bool      `json:"-" bson:"-"`
	IsUnclaimedCustodial bool      `json:"-" bson:"-"`
	Enabled              bool      `json:"-" bson:"-"`
	CreatedTime          string    `json:"createdTime,omitempty" bson:"createdTime,omitempty"`
	CreatedUserID        string    `json:"createdUserId,omitempty" bson:"createdUserId,omitempty"`
	ModifiedTime         string    `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
	ModifiedUserID       string    `json:"modifiedUserId,omitempty" bson:"modifiedUserId,omitempty"`
	DeletedTime          string    `json:"deletedTime,omitempty" bson:"deletedTime,omitempty"`
	DeletedUserID        string    `json:"deletedUserId,omitempty" bson:"deletedUserId,omitempty"`
	Profile              *Profile  `json:"profile,omitempty" bson:"-"`
	PasswordExists       *bool     `json:"passwordExists,omitempty" bson:"-"`
}

// TrustUser is the user object returned for the /v1/users/:userId/users route.
type TrustUser struct {
	User
	TrustPermissions
}

type TrustUserArray []*TrustUser

type TrustPermissions struct {
	TrusteePermissions *permission.Permission `json:"trusteePermissions,omitempty"`
	TrustorPermissions *permission.Permission `json:"trustorPermissions,omitempty"`
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
		return slices.Contains(*u.Roles, role)
	}
	return false
}

func (u *User) IsPatient() bool {
	if u.Roles == nil || len(*u.Roles) == 0 || u.HasRole(RolePatient) {
		return true
	}
	return false
}

func IsUnclaimedCustodialEmail(email string) bool {
	return custodialAccountRegexp.MatchString(email)
}

func (u *User) Sanitize(details request.AuthDetails) error {
	if details == nil || (!details.IsService() && details.UserID() != *u.UserID) {
		u.Username = nil
		u.EmailVerified = nil
		u.TermsAccepted = nil
		u.Roles = nil
		u.PasswordExists = nil
	}
	return nil
}

func (u *User) Email() string {
	if u.Username != nil {
		return strings.ToLower(*u.Username)
	}
	return ""
}

func (u *TrustUser) Sanitize(details request.AuthDetails) error {
	if details == nil || (!details.IsService() && details.UserID() != *u.UserID) {
		// Note that a TrustUser includes some fields in the user that [User.Sanitize] wouldn't.
		u.PasswordExists = nil
		if (u.TrustorPermissions == nil || len(*u.TrustorPermissions) == 0) && u.User.Profile != nil {
			// Clear out patient fields
			u.User.Profile.Birthday = ""
			u.User.Profile.DiagnosisDate = ""
			u.User.Profile.DiagnosisType = ""
			u.User.Profile.TargetDevices = nil
			u.User.Profile.TargetTimezone = ""
			u.User.Profile.About = ""
			u.User.Profile.MRN = ""
			u.User.Profile.BiologicalSex = ""
		}
	}
	return nil
}

func (us TrustUserArray) Sanitize(details request.AuthDetails) error {
	for _, u := range us {
		if err := u.Sanitize(details); err != nil {
			return err
		}
	}
	return nil
}

// IsClinic returns true if the user is legacy clinic Account
func (u *User) IsClinic() bool {
	return u.HasRole(RoleClinic)
}

func (u *User) IsCustodialAccount() bool {
	return u.HasRole(RoleCustodialAccount)
}

// IsClinician returns true if the user is a clinician
func (u *User) IsClinician() bool {
	return u.HasRole(RoleClinician)
}

func (u *User) AreTermsAccepted() bool {
	if u.TermsAccepted == nil {
		return false
	}
	_, err := TimestampToUnixString(*u.TermsAccepted)
	return err == nil
}

func (u *User) IsEnabled() bool {
	if u.IsMigrated {
		return u.Enabled
	}
	return u.PwHash != "" && !u.IsDeleted()
}

func (u *User) IsDeleted() bool {
	// mdb only?
	return u.DeletedTime != ""
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

// IsValidUserID return true if the string is in a human readable uuid hex 8-4-4-4-12 format or legacy alphanumeric 10 characters
func IsValidUserID(id string) bool {
	return idExpression.MatchString(id)
}
