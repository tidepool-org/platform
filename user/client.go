package user

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
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/log"
)

type (

	//Client is a generic client interface that we will implement and mock
	Client interface {
		Start() error
		Close()
		CheckToken(token string) (*TokenData, error)
		GetUser(userID string) (*Data, error)
		GetUserPermissons(userID string) (*UsersPermissions, error)
		GetUserGroupID(userID string) (string, error)
	}

	// ServicesClient manages the local data for a client. A client is intended to be shared among multiple
	// goroutines so it's OK to treat it as a singleton (and probably a good idea).
	ServicesClient struct {
		logger log.Logger
		config *Config

		// store a reference to the http client so we can reuse it
		httpClient *http.Client

		mut sync.Mutex

		// stores the most recently received server token
		serverToken string

		// Channel to communicate that the object has been closed
		closed chan chan bool
	}

	Config struct {
		Address       string `json:"address"`
		Secret        string `json:"secret"`
		TokenDuration int    `json:"tokenDuration"`
	}

	//Data is the data structure returned when we get a user
	Data struct {
		UserID         string   `json:"userid,omitempty"`         // the tidepool-assigned user ID
		Username       string   `json:"username,omitempty"`       // the user-assigned name for the login (usually an email address)
		Emails         []string `json:"emails,omitempty"`         // the array of email addresses associated with this account
		PasswordExists bool     `json:"passwordExists,omitempty"` // Does a password exist for the user?
		Roles          []string `json:"roles,omitempty"`          // User roles
		TermsAccepted  string   `json:"termsAccepted,omitempty"`  // When were the terms accepted
		EmailVerified  bool     `json:"emailVerified,omitempty"`  // the user has verified the email used as part of signup
	}

	//TokenData is the data structure returned from a successful CheckToken query.
	TokenData struct {
		UserID   string // the UserID stored in the token
		IsServer bool   // true or false depending on whether the token was a servertoken
	}

	//Permission for a user account
	Permission map[string]interface{}
	//Permissions for a user account
	Permissions map[string]Permission
	//UsersPermissions map of Permissions by userID
	UsersPermissions map[string]Permissions
)

const (
	xTidepoolServerName   = "x-tidepool-server-name"
	xTidepoolServerSecret = "x-tidepool-server-secret"
	xTidepoolSessionToken = "x-tidepool-session-token"

	userPath        = "/auth"
	permissionsPath = "/access"
	metaDataPath    = "/metadata"
)

func (c *Config) Validate() error {
	if c.Address == "" {
		return app.Error("dataservices", "address is not specified")
	}
	if c.Secret == "" {
		return app.Error("dataservices", "secret is not specified")
	}
	if c.TokenDuration == 0 {
		c.TokenDuration = 60
	}
	return nil
}

func NewServicesClient(logger log.Logger) (*ServicesClient, error) {
	clientConfig := &Config{}
	if err := config.Load("userservices_client", clientConfig); err != nil {
		return nil, app.ExtError(err, "user", "unable to load config")
	}
	if err := clientConfig.Validate(); err != nil {
		return nil, app.ExtError(err, "user", "config is not valid")
	}

	return &ServicesClient{
		logger:     logger,
		config:     clientConfig,
		httpClient: http.DefaultClient,
		closed:     make(chan chan bool),
	}, nil
}

// Start starts the client and makes it ready for use.  This must be done before using any of the functionality
// that requires a server token
func (client *ServicesClient) Start() error {
	if err := client.serverLogin(); err != nil {
		// TODO: Is this right?
		client.logger.WithError(err).Error("Problem with initial server token acquisition")
	}

	go func() {
		for {
			timer := time.After(time.Duration(client.config.TokenDuration) * time.Minute)
			select {
			case twoWay := <-client.closed:
				twoWay <- true
				return
			case <-timer:
				if err := client.serverLogin(); err != nil {
					client.logger.WithError(err).Error("Error when refreshing server login")
				}
			}
		}
	}()
	return nil
}

//Close that will close the service connection
func (client *ServicesClient) Close() {
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
func (client *ServicesClient) serverLogin() error {
	host, err := client.getHost(userPath)
	if err != nil {
		return err
	}
	if host == nil {
		return errors.New("No known user-api hosts.")
	}

	host.Path += "/serverlogin"

	req, _ := http.NewRequest("POST", host.String(), nil)
	req.Header.Add(xTidepoolServerName, "dataservices")
	req.Header.Add(xTidepoolServerSecret, client.config.Secret)

	res, err := client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("Failure to obtain a server token %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Unknown response %d from service[%s]", res.StatusCode, req.URL)
	}
	token := res.Header.Get(xTidepoolSessionToken)

	client.mut.Lock()
	defer client.mut.Unlock()
	client.serverToken = token

	return nil
}

//tokenProvide will return a servertoken that is internally provided for service calls
func (client *ServicesClient) tokenProvide() string {
	client.mut.Lock()
	defer client.mut.Unlock()
	return client.serverToken
}

//CheckToken tests a token with the user-api to make sure it's current;
//if so, it returns the data encoded in the token.
func (client *ServicesClient) CheckToken(token string) (*TokenData, error) {
	host, err := client.getHost(userPath)
	if err != nil {
		return nil, err
	}
	if host == nil {
		client.logger.Error(fmt.Sprintf("No known host for %s", userPath))
		return nil, nil
	}

	host.Path += "/token/" + token

	req, _ := http.NewRequest("GET", host.String(), nil)
	req.Header.Add(xTidepoolSessionToken, client.serverToken)

	res, err := client.httpClient.Do(req)
	if err != nil {
		client.logger.WithError(err).Error("Error checking token")
		return nil, err
	}

	switch res.StatusCode {
	case http.StatusOK:
		var td TokenData
		if err = json.NewDecoder(res.Body).Decode(&td); err != nil {
			client.logger.WithError(err).Error("Error parsing JSON results")
			return nil, err
		}
		return &td, nil
	case http.StatusNoContent:
		return nil, nil
	default:
		client.logger.Error(fmt.Sprintf("Unknown response %d %s", res.StatusCode, req.URL))
		return nil, nil
	}
}

//GetUser details for the given user
//In this case the userID could be the actual ID or an email address
func (client *ServicesClient) GetUser(userID string) (*Data, error) {
	host, err := client.getHost(userPath)
	if err != nil {
		return nil, err
	}
	if host == nil {
		return nil, fmt.Errorf("No known %s host", userPath)
	}

	host.Path += fmt.Sprintf("user/%s", userID)

	req, _ := http.NewRequest("GET", host.String(), nil)
	req.Header.Add(xTidepoolSessionToken, client.tokenProvide())

	res, err := client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failure to get a user \n\n %v", err)
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		var cd Data
		if err = json.NewDecoder(res.Body).Decode(&cd); err != nil {
			client.logger.WithError(err).Error("Error parsing JSON results")
			return nil, err
		}
		return &cd, nil
	case http.StatusNoContent:
		return &Data{}, nil
	default:
		return nil, fmt.Errorf("Unknown response %d from service[%s]", res.StatusCode, req.URL)
	}
}

//GetUserPermissons for the given userID
func (client *ServicesClient) GetUserPermissons(userID string) (*UsersPermissions, error) {
	host, err := client.getHost(permissionsPath)
	if err != nil {
		return nil, err
	}
	if host == nil {
		return nil, fmt.Errorf("No known %s host", permissionsPath)
	}

	host.Path += fmt.Sprintf("/groups/%s", userID)

	req, _ := http.NewRequest("GET", host.String(), nil)
	req.Header.Add(xTidepoolSessionToken, client.tokenProvide())

	res, err := client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failure to get a groups for user \n\n %v", err)
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		var perms UsersPermissions
		if err = json.NewDecoder(res.Body).Decode(&perms); err != nil {
			client.logger.WithError(err).Error("Error parsing JSON results")
			return nil, err
		}
		return &perms, nil
	case http.StatusNoContent:
		return &UsersPermissions{}, nil
	default:
		return nil, fmt.Errorf("Unknown response %d from service[%s]", res.StatusCode, req.URL)
	}
}

//GetUserGroupID for the given userID
func (client *ServicesClient) GetUserGroupID(userID string) (string, error) {
	host, err := client.getHost(metaDataPath)
	if err != nil {
		return "", err
	}
	if host == nil {
		return "", fmt.Errorf("No known %s host", metaDataPath)
	}

	host.Path += fmt.Sprintf("%s/private/uploads", userID)

	req, _ := http.NewRequest("GET", host.String(), nil)
	req.Header.Add(xTidepoolSessionToken, client.tokenProvide())

	res, err := client.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Failure to get groupID for user \n\n %v", err)
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		var pair struct {
			ID    string
			Value string
		}
		if err = json.NewDecoder(res.Body).Decode(&pair); err != nil {
			client.logger.WithError(err).Error("Error parsing JSON results")
			return "", err
		}
		return pair.ID, nil
	default:
		return "", fmt.Errorf("Unknown response %d from service[%s]", res.StatusCode, req.URL)
	}
}

func (client *ServicesClient) getHost(pathName string) (*url.URL, error) {
	urlString := client.config.Address + pathName
	theURL, err := url.Parse(urlString)
	if err != nil {
		return nil, app.ExtError(err, "user", "unable to determine host url")
	}
	return theURL, nil
}
