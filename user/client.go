// This is a client module to support server-side use of user-api.
//
// NOTE: This client was largly ported from `github.com/tidepool-org/go-common` and will be re-written
package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/tidepool-org/platform/config"
	log "github.com/tidepool-org/platform/logger"
)

type (

	//Generic client interface that we will implement and mock
	Client interface {
		Start() error
		Close()
		CheckToken(token string) *UserTokenData
		GetUser(userID, token string) (*UserData, error)
		GetUserPermissons(userID, token string) (*UsersPermissions, error)
	}

	// UserApiClient manages the local data for a client. A client is intended to be shared among multiple
	// goroutines so it's OK to treat it as a singleton (and probably a good idea).
	UserServicesClient struct {
		// store a reference to the http client so we can reuse it
		httpClient *http.Client

		// Configuration for the client
		config *ClientConfig

		//secret used along with the name to obtain a server token
		secret string

		mut sync.Mutex

		// stores the most recently received server token
		serverToken string

		// Channel to communicate that the object has been closed
		closed chan chan bool
	}

	ClientConfig struct {
		Host                 string `json:"host"`                 // URL of the user client host e.g. "http://localhost:9107"
		Name                 string `json:"name"`                 // The name of this server for use in obtaining a server token
		TokenRefreshInterval string `json:"tokenRefreshInterval"` // The amount of time between refreshes of the server token
		TokenRefreshDuration time.Duration
	}

	// UserData is the data structure returned when we get a user
	UserData struct {
		UserID         string   `json:"userid,omitempty"`         // the tidepool-assigned user ID
		Username       string   `json:"username,omitempty"`       // the user-assigned name for the login (usually an email address)
		Emails         []string `json:"emails,omitempty"`         // the array of email addresses associated with this account
		PasswordExists bool     `json:"passwordExists,omitempty"` // Does a password exist for the user?
		Roles          []string `json:"roles,omitempty"`          // User roles
		TermsAccepted  string   `json:"termsAccepted,omitempty"`  // When were the terms accepted
		EmailVerified  bool     `json:"emailVerified,omitempty"`  // the user has verified the email used as part of signup
	}

	// TokenData is the data structure returned from a successful CheckToken query.
	UserTokenData struct {
		UserID   string // the UserID stored in the token
		IsServer bool   // true or false depending on whether the token was a servertoken
	}

	//permissons types
	Permission       map[string]interface{}
	Permissions      map[string]Permission
	UsersPermissions map[string]Permissions
)

const (
	x_tidepool_server_name   = "x-tidepool-server-name"
	x_tidepool_server_secret = "x-tidepool-server-secret"
	x_tidepool_session_token = "x-tidepool-session-token"
	tidepool_client_secret   = "TIDEPOOL_USER_CLIENT_SECRET"

	user_path        = "/auth"
	permissions_path = "/access"
)

func NewUserServicesClient() *UserServicesClient {

	var clientConfig *ClientConfig

	config.FromJson(&clientConfig, "userclient.json")

	if clientConfig.Name == "" {
		panic("UserServicesClient requires a name to be set")
	}
	if clientConfig.Host == "" {
		panic("UserServicesClient requires a host to be set")
	}

	dur, err := time.ParseDuration(clientConfig.TokenRefreshInterval)
	if err != nil {
		log.Logging.Error("err getting the duration ", err.Error())
	}
	clientConfig.TokenRefreshDuration = dur

	secret, err := config.FromEnv(tidepool_client_secret)
	if err != nil {
		log.Logging.Error("err getting client secret ", err.Error())
	}

	return &UserServicesClient{
		httpClient: http.DefaultClient,
		config:     clientConfig,
		secret:     secret,
		closed:     make(chan chan bool),
	}
}

// Start starts the client and makes it ready for us.  This must be done before using any of the functionality
// that requires a server token
func (client *UserServicesClient) Start() error {
	if err := client.serverLogin(); err != nil {
		log.Logging.Error("Problem with initial server token acquisition:", err.Error())
	}

	go func() {
		for {
			timer := time.After(time.Duration(client.config.TokenRefreshDuration))
			select {
			case twoWay := <-client.closed:
				twoWay <- true
				return
			case <-timer:
				if err := client.serverLogin(); err != nil {
					log.Logging.Error("Error when refreshing server login:", err.Error())
				}
			}
		}
	}()
	return nil
}

func (client *UserServicesClient) Close() {
	twoWay := make(chan bool)
	client.closed <- twoWay
	<-twoWay

	client.mut.Lock()
	defer client.mut.Unlock()
	client.serverToken = ""
}

// serverLogin issues a request to the server for a login, using the stored
// secret that was passed in on the creation of the client object. If
// successful, it stores the returned token in ServerToken.
func (client *UserServicesClient) serverLogin() error {
	host := client.getUserHost()
	if host == nil {
		return errors.New("No known user-api hosts.")
	}

	host.Path += "/serverlogin"

	req, _ := http.NewRequest("POST", host.String(), nil)
	req.Header.Add(x_tidepool_server_name, client.config.Name)
	req.Header.Add(x_tidepool_server_secret, client.secret)

	res, err := client.httpClient.Do(req)
	if err != nil {
		return errors.New(fmt.Sprintf("Failure to obtain a server token %v", err))
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("Unknown response %d from service[%s]", res.StatusCode, req.URL))
	}
	token := res.Header.Get(x_tidepool_session_token)

	client.mut.Lock()
	defer client.mut.Unlock()
	client.serverToken = token

	return nil
}

// CheckToken tests a token with the user-api to make sure it's current;
// if so, it returns the data encoded in the token.
func (client *UserServicesClient) CheckToken(token string) *UserTokenData {
	host := client.getUserHost()
	if host == nil {
		log.Logging.Error("No known user-api hosts.")
		return nil
	}

	host.Path += "/token/" + token

	req, _ := http.NewRequest("GET", host.String(), nil)
	req.Header.Add(x_tidepool_session_token, client.serverToken)

	res, err := client.httpClient.Do(req)
	if err != nil {
		log.Logging.Error("Error checking token", err.Error())
		return nil
	}

	switch res.StatusCode {
	case http.StatusOK:
		var td UserTokenData
		if err = json.NewDecoder(res.Body).Decode(&td); err != nil {
			log.Logging.Error("Error parsing JSON results", err.Error())
			return nil
		}
		return &td
	case http.StatusNoContent:
		return nil
	default:
		log.Logging.Error("Unknown response ", res.StatusCode, req.URL)
		return nil
	}
}

// Get user details for the given user
// In this case the userID could be the actual ID or an email address
func (client *UserServicesClient) GetUser(userID, token string) (*UserData, error) {
	host := client.getUserHost()
	if host == nil {
		return nil, errors.New("No known user-api hosts.")
	}

	host.Path += fmt.Sprintf("user/%s", userID)

	req, _ := http.NewRequest("GET", host.String(), nil)
	req.Header.Add(x_tidepool_session_token, token)

	res, err := client.httpClient.Do(req)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failure to get a user \n\n %v", err))
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		var cd UserData
		if err = json.NewDecoder(res.Body).Decode(&cd); err != nil {
			log.Logging.Error("Error parsing JSON results:", err.Error())
			return nil, err
		}
		return &cd, nil
	case http.StatusNoContent:
		return &UserData{}, nil
	default:
		return nil, errors.New(fmt.Sprintf("Unknown response %d from service[%s]", res.StatusCode, req.URL))
	}
}

func (client *UserServicesClient) GetUserPermissons(userID, token string) (*UsersPermissions, error) {
	host := client.getPermissionsHost()
	if host == nil {
		return nil, errors.New("No known user-api hosts.")
	}

	host.Path += fmt.Sprintf("/groups/%s", userID)

	req, _ := http.NewRequest("GET", host.String(), nil)
	req.Header.Add(x_tidepool_session_token, token)

	res, err := client.httpClient.Do(req)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failure to get a groups for user \n\n %v", err))
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		var perms UsersPermissions
		if err = json.NewDecoder(res.Body).Decode(&perms); err != nil {
			log.Logging.Error("Error parsing JSON results:", err.Error())
			return nil, err
		}
		return &perms, nil
	case http.StatusNoContent:
		return &UsersPermissions{}, nil
	default:
		return nil, errors.New(fmt.Sprintf("Unknown response %d from service[%s]", res.StatusCode, req.URL))
	}
}

func (client *UserServicesClient) getPermissionsHost() *url.URL {
	theUrl, err := url.Parse(client.config.Host + permissions_path)
	if err != nil {
		log.Logging.Error("Unable to parse permissions urlString:", client.config.Host+permissions_path)
		return nil
	}
	return theUrl
}

func (client *UserServicesClient) getUserHost() *url.URL {
	theUrl, err := url.Parse(client.config.Host + user_path)
	if err != nil {
		log.Logging.Error("Unable to parse user urlString:", client.config.Host+user_path)
		return nil
	}
	return theUrl
}
