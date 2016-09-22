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
	"fmt"
	"io"
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
)

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
