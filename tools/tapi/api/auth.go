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

	"github.com/tidepool-org/platform/user"
)

type (
	Token struct {
		IsServer bool   `json:"isserver"`
		UserID   string `json:"userid"`
	}
)

func (a *API) ServerLogin() error {
	return a.asEmpty(a.request("POST", a.joinPaths("auth", "serverlogin"),
		requestFuncs{a.addServerSecret()},
		responseFuncs{a.expectStatusCode(http.StatusOK), a.storeTidepoolSession(true)}))
}

func (a *API) Login(email string, password string) (*user.User, error) {
	if email == "" {
		return nil, errors.New("Email is missing")
	}
	if password == "" {
		return nil, errors.New("Password is missing")
	}

	return a.asUser(a.request("POST", a.joinPaths("auth", "login"),
		requestFuncs{a.addBasicAuthorization(email, password)},
		responseFuncs{a.expectStatusCode(http.StatusOK), a.storeTidepoolSession(false)}))
}

func (a *API) Logout() error {
	return a.asEmpty(a.request("POST", a.joinPaths("auth", "logout"),
		requestFuncs{a.addSessionToken()},
		responseFuncs{a.expectStatusCode(http.StatusOK), a.destroyTidepoolSession()}))
}

func (a *API) RefreshToken() (*Token, error) {
	return a.asToken(a.request("GET", a.joinPaths("auth", "login"),
		requestFuncs{a.addSessionToken()},
		responseFuncs{a.expectStatusCode(http.StatusOK), a.updateTidepoolSession()}))
}

func (a *API) CheckToken(token string) (*Token, error) {
	if token == "" {
		return nil, errors.New("Token is missing")
	}

	return a.asToken(a.request("GET", a.joinPaths("auth", "token", token),
		requestFuncs{a.addSessionToken()},
		responseFuncs{a.expectStatusCode(http.StatusOK)}))
}

func (a *API) asToken(responseBody io.Reader, err error) (*Token, error) {
	responseString, err := a.asString(responseBody, err)
	if err != nil {
		return nil, err
	}

	var token *Token
	if len(responseString) > 0 {
		token = &Token{}
		if err = json.Unmarshal([]byte(responseString), &token); err != nil {
			return nil, fmt.Errorf("Error decoding JSON Token from response body: %s", err.Error())
		}
	}

	return token, nil
}
