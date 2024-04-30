package user

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	RoleClinic           = "clinic"
	RoleClinician        = "clinician"
	RoleCustodialAccount = "custodial_account"
	RoleMigratedClinic   = "migrated_clinic"
	RolePatient          = "patient"
	RoleBrokered         = "brokered"
)

func Roles() []string {
	return []string{
		RoleClinic,
	}
}

type Client interface {
	Get(ctx context.Context, id string) (*User, error)
}

type User struct {
	UserID               *string             `json:"userid,omitempty" bson:"userid,omitempty"`
	Username             *string             `json:"username,omitempty" bson:"username,omitempty"`
	EmailVerified        *bool               `json:"emailVerified,omitempty" bson:"emailVerified,omitempty"`
	TermsAccepted        *string             `json:"termsAccepted,omitempty" bson:"termsAccepted,omitempty"`
	Roles                *[]string           `json:"roles,omitempty" bson:"roles,omitempty"`
	Emails               []string            `json:"emails,omitempty" bson:"emails,omitempty"`
	PwHash               string              `json:"-" bson:"pwhash,omitempty"`
	Hash                 string              `json:"-" bson:"userhash,omitempty"`
	IsMigrated           bool                `json:"-" bson:"-"`
	IsUnclaimedCustodial bool                `json:"-" bson:"-"`
	Enabled              bool                `json:"-" bson:"-"`
	CreatedTime          string              `json:"createdTime,omitempty" bson:"createdTime,omitempty"`
	CreatedUserID        string              `json:"createdUserId,omitempty" bson:"createdUserId,omitempty"`
	ModifiedTime         string              `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
	ModifiedUserID       string              `json:"modifiedUserId,omitempty" bson:"modifiedUserId,omitempty"`
	DeletedTime          string              `json:"deletedTime,omitempty" bson:"deletedTime,omitempty"`
	DeletedUserID        string              `json:"deletedUserId,omitempty" bson:"deletedUserId,omitempty"`
	Attributes           map[string][]string `json:"-"`
	Profile              *UserProfile        `json:"-"`
	FirstName            string              `json:"firstName,omitempty"`
	LastName             string              `json:"lastName,omitempty"`
}

func (u *User) Parse(parser structure.ObjectParser) {
	u.UserID = parser.String("userid")
	u.Username = parser.String("username")
	u.EmailVerified = parser.Bool("emailVerified")
	u.TermsAccepted = parser.String("termsAccepted")
	u.Roles = parser.StringArray("roles")
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
	if u.Roles == nil || len(*u.Roles) == 0 {
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

func (u *User) Email() string {
	if u.Username != nil {
		return strings.ToLower(*u.Username)
	}
	return ""
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

var idExpression = regexp.MustCompile(`^([0-9a-f]{10}|[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12})$`)

func IsValidUserID(id string) bool {
	ok, _ := regexp.MatchString(`^([a-fA-F0-9]{10})$|^([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12})$`, id)
	return ok
}
