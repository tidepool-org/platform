package user

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
