package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/tidepool-org/go-common/clients/disc"
	"github.com/tidepool-org/go-common/clients/status"
	"github.com/tidepool-org/go-common/errors"
)

type (
	//Interface so that we can mock authClient for tests
	Auth interface {
		//userID  -- the Tidepool-assigned userID
		//
		// returns the Auth Sources for the user
		ListUserRestrictedTokens(userID string, token string) (RestrictedTokens, error)
		CreateRestrictedToken(userID string, expirationTime time.Time, paths []string, token string) (*RestrictedToken, error)
		UpdateRestrictedToken(tokenId string, expirationTime time.Time, paths []string, token string) (*RestrictedToken, error)
		DeleteRestrictedToken(tokenId string, token string) error
	}

	AuthClient struct {
		httpClient    *http.Client    // store a reference to the http client so we can reuse it
		hostGetter    disc.HostGetter // The getter that provides the host to talk to for the client
		tokenProvider TokenProvider   // An object that provides tokens for communicating with data
	}

	authClientBuilder struct {
		httpClient    *http.Client    // store a reference to the http client so we can reuse it
		hostGetter    disc.HostGetter // The getter that provides the host to talk to for the client
		tokenProvider TokenProvider   // An object that provides tokens for communicating with data
	}
)

// RestrictedToken is the data structure returned from a successful create restricted token query.
type RestrictedToken struct {
	ID             string     `json:"id"`
	UserID         string     `json:"userId"`
	Paths          *[]string  `json:"paths,omitempty"`
	ExpirationTime time.Time  `json:"expirationTime"`
	CreatedTime    time.Time  `json:"createdTime"`
	ModifiedTime   *time.Time `json:"modifiedTime,omitempty"`
}

type RestrictedTokens []*RestrictedToken

type RestrictedTokenCreate struct {
	Paths          *[]string  `json:"paths,omitempty"`
	ExpirationTime *time.Time `json:"expirationTime"`
}

type RestrictedTokenUpdate struct {
	Paths          *[]string  `json:"paths,omitempty"`
	ExpirationTime *time.Time `json:"expirationTime"`
}

type RestrictedTokenFilter struct{}

func NewAuthClientBuilder() *authClientBuilder {
	return &authClientBuilder{}
}

func (b *authClientBuilder) WithHttpClient(httpClient *http.Client) *authClientBuilder {
	b.httpClient = httpClient
	return b
}

func (b *authClientBuilder) WithHostGetter(hostGetter disc.HostGetter) *authClientBuilder {
	b.hostGetter = hostGetter
	return b
}

func (b *authClientBuilder) WithTokenProvider(tokenProvider TokenProvider) *authClientBuilder {
	b.tokenProvider = tokenProvider
	return b
}

func (b *authClientBuilder) Build() *AuthClient {
	if b.hostGetter == nil {
		panic("authClient requires a hostGetter to be set")
	}
	if b.tokenProvider == nil {
		panic("authClient requires a tokenProvider to be set")
	}

	if b.httpClient == nil {
		b.httpClient = http.DefaultClient
	}

	return &AuthClient{
		httpClient:    b.httpClient,
		hostGetter:    b.hostGetter,
		tokenProvider: b.tokenProvider,
	}
}

// ListUserRestrictedTokens listsrestricted tokens for a given user
func (client *AuthClient) ListUserRestrictedTokens(userID string, token string) (RestrictedTokens, error) {
	host := client.getHost()
	if host == nil {
		return nil, errors.New("No known auth hosts")
	}
	host.Path = path.Join(host.Path, "v1", "users", userID, "restricted_tokens")

	req, _ := http.NewRequest("GET", host.String(), nil)
	req.Header.Add("x-tidepool-session-token", token)

	res, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		retVal := RestrictedTokens{}
		if err := json.NewDecoder(res.Body).Decode(&retVal); err != nil {
			log.Println(err)
			return nil, &status.StatusError{
				Status: status.NewStatusf(res.StatusCode, "ListUserRestrictedTokens Unable to parse response[%s]", req.URL),
			}
		}
		return retVal, nil
	} else if res.StatusCode == 404 {
		return nil, nil
	} else {
		return nil, &status.StatusError{
			Status: status.NewStatusf(res.StatusCode, "Unexpected response code from service[%s]", req.URL),
		}
	}
}

// CreateRestrictedToken creates a restricted token for a given user
func (client *AuthClient) CreateRestrictedToken(userID string, expirationTime time.Time, paths []string, token string) (*RestrictedToken, error) {
	host := client.getHost()
	if host == nil {
		return nil, errors.New("No known auth hosts")
	}
	host.Path = path.Join(host.Path, "v1", "users", userID, "restricted_tokens")

	payload := RestrictedTokenCreate{
		Paths:          &paths,
		ExpirationTime: &expirationTime,
	}

	if jsonToken, err := json.Marshal(payload); err != nil {
		return nil, fmt.Errorf("unable to marshal payload: %w", err)
	} else {
		req, _ := http.NewRequest("POST", host.String(), bytes.NewBuffer(jsonToken))
		req.Header.Add("x-tidepool-session-token", token)

		res, err := client.httpClient.Do(req)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't create restricted token")
		}
		defer res.Body.Close()

		switch res.StatusCode {
		case http.StatusCreated:
			var td RestrictedToken
			if err = json.NewDecoder(res.Body).Decode(&td); err != nil {
				return nil, errors.Wrap(err, "Error parsing JSON results")
			}
			return &td, nil
		default:
			return nil, &status.StatusError{
				Status: status.NewStatusf(res.StatusCode, "Unexpected response code from service[%s]", req.URL),
			}
		}
	}
}

// UpdateRestrictedToken updates a restricted token
func (client *AuthClient) UpdateRestrictedToken(tokenID string, expirationTime time.Time, paths []string, token string) (*RestrictedToken, error) {
	host := client.getHost()
	if host == nil {
		return nil, errors.New("No known auth hosts")
	}
	host.Path = path.Join(host.Path, "v1", "restricted_tokens", tokenID)

	payload := RestrictedTokenUpdate{
		Paths:          &paths,
		ExpirationTime: &expirationTime,
	}

	if jsonToken, err := json.Marshal(payload); err != nil {
		return nil, fmt.Errorf("unable to marshal payload: %w", err)
	} else {
		req, _ := http.NewRequest("PUT", host.String(), bytes.NewBuffer(jsonToken))
		req.Header.Add("x-tidepool-session-token", token)

		res, err := client.httpClient.Do(req)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't update restricted token")
		}
		defer res.Body.Close()

		switch res.StatusCode {
		case http.StatusOK:
			var td RestrictedToken
			if err = json.NewDecoder(res.Body).Decode(&td); err != nil {
				return nil, errors.Wrap(err, "Error parsing JSON results")
			}
			return &td, nil
		default:
			return nil, &status.StatusError{
				Status: status.NewStatusf(res.StatusCode, "Unexpected response code from service[%s]", req.URL),
			}
		}
	}
}

// DeleteRestrictedToken deletes a restricted token
func (client *AuthClient) DeleteRestrictedToken(tokenID string, token string) error {
	host := client.getHost()
	if host == nil {
		return errors.New("No known auth hosts")
	}
	host.Path = path.Join(host.Path, "v1", "restricted_tokens", tokenID)

	req, _ := http.NewRequest("DELETE", host.String(), nil)
	req.Header.Add("x-tidepool-session-token", token)

	res, err := client.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "couldn't delete restricted token")
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusNoContent:
		return nil
	default:
		return &status.StatusError{
			Status: status.NewStatusf(res.StatusCode, "Unknown response code from service[%s]", req.URL),
		}
	}
}

func (client *AuthClient) getHost() *url.URL {
	if hostArr := client.hostGetter.HostGet(); len(hostArr) > 0 {
		cpy := new(url.URL)
		*cpy = hostArr[0]
		return cpy
	} else {
		return nil
	}
}
