package api

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type (
	User struct {
		ID             string   `json:"userid,omitempty"`
		Username       string   `json:"username,omitempty"`
		Emails         []string `json:"emails,omitempty"`
		Roles          []string `json:"roles,omitempty"`
		TermsAccepted  string   `json:"termsAccepted,omitempty"`
		EmailVerified  bool     `json:"emailVerified,omitempty"`
		PasswordExists bool     `json:"passwordExists,omitempty"`
	}
	Users []*User

	UsersQuery struct {
		Role *string `json:"role,omitempty"`
	}

	UserUpdates struct {
		Username      *string   `json:"username,omitempty"`
		Emails        *[]string `json:"emails,omitempty"`
		Password      *string   `json:"password,omitempty"`
		Roles         *[]string `json:"roles,omitempty"`
		TermsAccepted *string   `json:"termsAccepted,omitempty"`
		EmailVerified *bool     `json:"emailVerified,omitempty"`
	}

	UserUpdater interface {
		Update(user *User, userUpdates *UserUpdates) error
	}

	AddRolesUserUpdater struct {
		Roles []string
	}

	RemoveRolesUserUpdater struct {
		Roles []string
	}

	UserDelete struct {
		Password *string `json:"password,omitempty"`
	}
)

func (a *API) GetUserByID(userID string) (*User, error) {
	if userID == "" {
		var err error
		userID, err = a.fetchSessionUserID()
		if err != nil {
			return nil, err
		}
	}

	return a.asUser(a.request("GET", a.joinPaths("auth", "user", userID),
		requestFuncs{a.addSessionToken()},
		responseFuncs{a.expectStatusCode(http.StatusOK)}))
}

func (a *API) GetUserByEmail(email string) (*User, error) {
	if email == "" {
		return nil, errors.New("Email is missing")
	}

	return a.asUser(a.request("GET", a.joinPaths("auth", "user", email),
		requestFuncs{a.addSessionToken()},
		responseFuncs{a.expectStatusCode(http.StatusOK)}))
}

func (a *API) FindUsers(query *UsersQuery) (Users, error) {
	if query == nil {
		return nil, errors.New("Query is missing")
	}

	queryMap := map[string]string{}
	if query.Role != nil {
		queryMap["role"] = *query.Role
	}
	return a.asUsers(a.request("GET", a.addQuery(a.joinPaths("auth", "users"), queryMap),
		requestFuncs{a.addSessionToken()},
		responseFuncs{a.expectStatusCode(http.StatusOK)}))
}

func (a *API) UpdateUserByID(userID string, userUpdates *UserUpdates) (*User, error) {
	if userUpdates == nil {
		return nil, errors.New("Updates is missing")
	}

	if userID == "" {
		var err error
		userID, err = a.fetchSessionUserID()
		if err != nil {
			return nil, err
		}
	}

	userUpdate := map[string]UserUpdates{"update": *userUpdates}
	return a.asUser(a.request("PUT", a.joinPaths("auth", "user", userID),
		requestFuncs{a.addSessionToken(), a.addObjectBody(userUpdate)},
		responseFuncs{a.expectStatusCode(http.StatusOK)}))
}

func (a *API) UpdateUserByObject(user *User, userUpdates *UserUpdates) (*User, error) {
	if user == nil {
		return nil, errors.New("User is missing")
	}
	if user.ID == "" {
		return nil, errors.New("User id is missing")
	}

	return a.UpdateUserByID(user.ID, userUpdates)
}

func (a *API) ApplyUpdatersToUserByID(userID string, updaters []UserUpdater) (*User, error) {
	if userID == "" {
		var err error
		userID, err = a.fetchSessionUserID()
		if err != nil {
			return nil, err
		}
	}

	user, err := a.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	return a.ApplyUpdatersToUserByObject(user, updaters)
}

func (a *API) ApplyUpdatersToUserByObject(user *User, updaters []UserUpdater) (*User, error) {
	if user == nil {
		return nil, errors.New("User is missing")
	}
	if len(updaters) == 0 {
		return nil, errors.New("Updaters is missing")
	}

	userUpdates := &UserUpdates{}
	for _, updater := range updaters {
		if err := updater.Update(user, userUpdates); err != nil {
			return nil, fmt.Errorf("Failure applying updater to user: %s", err.Error())
		}
	}

	if !userUpdates.HasUpdates() {
		return user, nil
	}

	return a.UpdateUserByID(user.ID, userUpdates)
}

func (a *API) DeleteUserByID(userID string, password string) error {
	if userID == "" {
		var err error
		userID, err = a.fetchSessionUserID()
		if err != nil {
			return err
		}
	}

	userDelete := &UserDelete{}
	if password != "" {
		userDelete.Password = &password
	}

	return a.asEmpty(a.request("DELETE", a.joinPaths("userservices", "v1", "users", userID),
		requestFuncs{a.addSessionToken(), a.addObjectBody(userDelete)},
		responseFuncs{a.expectStatusCode(http.StatusOK)}))
}

func (a *API) asUser(responseBody io.Reader, err error) (*User, error) {
	responseString, err := a.asString(responseBody, err)
	if err != nil {
		return nil, err
	}

	var responseUser *User
	if len(responseString) > 0 {
		responseUser = &User{}
		if err = json.Unmarshal([]byte(responseString), responseUser); err != nil {
			return nil, fmt.Errorf("Error decoding JSON User from response body: %s", err.Error())
		}
	}

	return responseUser, nil
}

func (a *API) asUsers(responseBody io.Reader, err error) (Users, error) {
	responseString, err := a.asString(responseBody, err)
	if err != nil {
		return nil, err
	}

	var responseUsers Users
	if len(responseString) > 0 {
		responseUsers = Users{}
		if err = json.Unmarshal([]byte(responseString), &responseUsers); err != nil {
			return nil, fmt.Errorf("Error decoding JSON Users from response body: %s", err.Error())
		}
	}

	return responseUsers, nil
}

func (u *UserUpdates) HasUpdates() bool {
	return u.Username != nil || u.Emails != nil || u.Password != nil || u.Roles != nil || u.TermsAccepted != nil || u.EmailVerified != nil
}

func NewAddRolesUserUpdater(roles []string) (*AddRolesUserUpdater, error) {
	if len(roles) == 0 {
		return nil, errors.New("Roles is missing")
	}

	return &AddRolesUserUpdater{Roles: roles}, nil
}

func (u *AddRolesUserUpdater) Update(user *User, userUpdates *UserUpdates) error {
	var originalRoles *[]string
	if userUpdates.Roles != nil {
		originalRoles = userUpdates.Roles
	} else if user.Roles != nil {
		originalRoles = &user.Roles
	} else {
		originalRoles = &[]string{}
	}

	addRoles := subtractStringArray(u.Roles, *originalRoles)
	if len(addRoles) == 0 {
		return nil
	}

	updatedRoles := append(append([]string{}, *originalRoles...), addRoles...)
	userUpdates.Roles = &updatedRoles
	return nil
}

func NewRemoveRolesUserUpdater(roles []string) (*RemoveRolesUserUpdater, error) {
	if len(roles) == 0 {
		return nil, errors.New("Roles is missing")
	}

	return &RemoveRolesUserUpdater{Roles: roles}, nil
}

func (u *RemoveRolesUserUpdater) Update(user *User, userUpdates *UserUpdates) error {
	var originalRoles *[]string
	if userUpdates.Roles != nil {
		originalRoles = userUpdates.Roles
	} else if user.Roles != nil {
		originalRoles = &user.Roles
	} else {
		return nil
	}

	updatedRoles := subtractStringArray(*originalRoles, u.Roles)
	if len(updatedRoles) == len(*originalRoles) {
		return nil
	}

	userUpdates.Roles = &updatedRoles
	return nil
}
