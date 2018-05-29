package user

import (
	"regexp"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type User struct {
	ID                string             `json:"userid,omitempty" bson:"userid,omitempty"`
	Email             string             `json:"username,omitempty" bson:"username,omitempty"`
	Emails            []string           `json:"emails,omitempty" bson:"emails,omitempty"`
	Roles             []string           `json:"roles,omitempty" bson:"roles,omitempty"`
	TermsAcceptedTime string             `json:"termsAccepted,omitempty" bson:"termsAccepted,omitempty"`
	EmailVerified     bool               `json:"emailVerified" bson:"authenticated"`
	PasswordHash      string             `json:"-" bson:"pwhash,omitempty"`
	Hash              string             `json:"-" bson:"userhash,omitempty"`
	Private           map[string]*IDHash `json:"-" bson:"private,omitempty"`
	CreatedTime       string             `json:"createdTime,omitempty" bson:"createdTime,omitempty"`
	CreatedUserID     string             `json:"createdUserId,omitempty" bson:"createdUserId,omitempty"`
	ModifiedTime      string             `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
	ModifiedUserID    string             `json:"modifiedUserId,omitempty" bson:"modifiedUserId,omitempty"`
	DeletedTime       string             `json:"deletedTime,omitempty" bson:"deletedTime,omitempty"`
	DeletedUserID     string             `json:"deletedUserId,omitempty" bson:"deletedUserId,omitempty"`

	ProfileID *string `json:"-" bson:"-"`
}

type IDHash struct {
	ID   string `json:"id"`
	Hash string `json:"hash"`
}

func (u *User) HasRole(role string) bool {
	for _, userRole := range u.Roles {
		if userRole == role {
			return true
		}
	}
	return false
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

func ErrorValueStringAsIDNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as user id", value)
}

var idExpression = regexp.MustCompile("^[0-9a-f]{10}$")
