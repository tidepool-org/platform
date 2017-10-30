package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/tidepool-org/platform/user"
)

type (
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
		Update(updateUser *user.User, userUpdates *UserUpdates) error
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

func (a *API) GetUserByID(userID string) (*user.User, error) {
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

func (a *API) GetUserByEmail(email string) (*user.User, error) {
	if email == "" {
		return nil, errors.New("Email is missing")
	}

	return a.asUser(a.request("GET", a.joinPaths("auth", "user", email),
		requestFuncs{a.addSessionToken()},
		responseFuncs{a.expectStatusCode(http.StatusOK)}))
}

func (a *API) FindUsers(query *UsersQuery) ([]*user.User, error) {
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

func (a *API) UpdateUserByID(userID string, userUpdates *UserUpdates) (*user.User, error) {
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

	userUpdate := map[string]UserUpdates{"updates": *userUpdates}
	return a.asUser(a.request("PUT", a.joinPaths("auth", "user", userID),
		requestFuncs{a.addSessionToken(), a.addObjectBody(userUpdate)},
		responseFuncs{a.expectStatusCode(http.StatusOK)}))
}

func (a *API) UpdateUserByObject(updateUser *user.User, userUpdates *UserUpdates) (*user.User, error) {
	if updateUser == nil {
		return nil, errors.New("User is missing")
	}
	if updateUser.ID == "" {
		return nil, errors.New("User id is missing")
	}

	return a.UpdateUserByID(updateUser.ID, userUpdates)
}

func (a *API) ApplyUpdatersToUserByID(userID string, updaters []UserUpdater) (*user.User, error) {
	if userID == "" {
		var err error
		userID, err = a.fetchSessionUserID()
		if err != nil {
			return nil, err
		}
	}

	updateUser, err := a.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	return a.ApplyUpdatersToUserByObject(updateUser, updaters)
}

func (a *API) ApplyUpdatersToUserByObject(updateUser *user.User, updaters []UserUpdater) (*user.User, error) {
	if updateUser == nil {
		return nil, errors.New("User is missing")
	}
	if len(updaters) == 0 {
		return nil, errors.New("Updaters is missing")
	}

	userUpdates := &UserUpdates{}
	for _, updater := range updaters {
		if err := updater.Update(updateUser, userUpdates); err != nil {
			return nil, fmt.Errorf("Failure applying updater to user: %s", err.Error())
		}
	}

	if !userUpdates.HasUpdates() {
		return updateUser, nil
	}

	return a.UpdateUserByID(updateUser.ID, userUpdates)
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

	return a.asEmpty(a.request("DELETE", a.joinPaths("v1", "users", userID),
		requestFuncs{a.addSessionToken(), a.addObjectBody(userDelete)},
		responseFuncs{a.expectStatusCode(http.StatusOK)}))
}

func (a *API) asUser(responseBody io.Reader, err error) (*user.User, error) {
	responseString, err := a.asString(responseBody, err)
	if err != nil {
		return nil, err
	}

	var responseUser *user.User
	if len(responseString) > 0 {
		responseUser = &user.User{}
		if err = json.Unmarshal([]byte(responseString), responseUser); err != nil {
			return nil, fmt.Errorf("Error decoding JSON User from response body: %s", err.Error())
		}
	}

	return responseUser, nil
}

func (a *API) asUsers(responseBody io.Reader, err error) ([]*user.User, error) {
	responseString, err := a.asString(responseBody, err)
	if err != nil {
		return nil, err
	}

	var responseUsers []*user.User
	if len(responseString) > 0 {
		responseUsers = []*user.User{}
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

func (a *AddRolesUserUpdater) Update(updateUser *user.User, userUpdates *UserUpdates) error {
	var originalRoles *[]string
	if userUpdates.Roles != nil {
		originalRoles = userUpdates.Roles
	} else if updateUser.Roles != nil {
		originalRoles = &updateUser.Roles
	} else {
		originalRoles = &[]string{}
	}

	addRoles := subtractStringArray(a.Roles, *originalRoles)
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

func (r *RemoveRolesUserUpdater) Update(updateUser *user.User, userUpdates *UserUpdates) error {
	var originalRoles *[]string
	if userUpdates.Roles != nil {
		originalRoles = userUpdates.Roles
	} else if updateUser.Roles != nil {
		originalRoles = &updateUser.Roles
	} else {
		return nil
	}

	updatedRoles := subtractStringArray(*originalRoles, r.Roles)
	if len(updatedRoles) == len(*originalRoles) {
		return nil
	}

	userUpdates.Roles = &updatedRoles
	return nil
}
