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
)

type (

	//Client is a generic client interface that we will implement and mock
	Client interface {
		Start() error
		Close()
		CheckToken(token string) *TokenData
		GetUser(userID, token string) (*Data, error)
		GetUserPermissons(userID, token string) (*UsersPermissions, error)
	}

	// ServicesClient manages the local data for a client. A client is intended to be shared among multiple
	// goroutines so it's OK to treat it as a singleton (and probably a good idea).
	ServicesClient struct {
		// store a reference to the http client so we can reuse it
		httpClient *http.Client

		// ClientConfig for the client
		config *ClientConfig

		//secret used along with the name to obtain a server token
		secret string

		mut sync.Mutex

		// stores the most recently received server token
		serverToken string

		// Channel to communicate that the object has been closed
		closed chan chan bool
	}

	//ClientConfig for initialising ServicesClient
	ClientConfig struct {
		Host                 string `json:"host"`                 // URL of the user client host e.g. "http://localhost:9107"
		Name                 string `json:"name"`                 // The name of this server for use in obtaining a server token
		TokenRefreshInterval string `json:"tokenRefreshInterval"` // The amount of time between refreshes of the server token
		TokenRefreshDuration time.Duration
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
	tidepoolClientSecret  = "TIDEPOOL_USER_CLIENT_SECRET"

	userPath        = "/auth"
	permissionsPath = "/access"
)

//NewServicesClient returns and initailised ServicesClient instance
func NewServicesClient() *ServicesClient {

	var clientConfig *ClientConfig

	config.FromJSON(&clientConfig, "userclient.json")

	if clientConfig.Name == "" {
		panic("ServicesClient requires a name to be set")
	}
	if clientConfig.Host == "" {
		panic("ServicesClient requires a host to be set")
	}

	dur, err := time.ParseDuration(clientConfig.TokenRefreshInterval)
	if err != nil {
		log.Error("err getting the duration ", err.Error())
	}
	clientConfig.TokenRefreshDuration = dur

	secret, err := config.FromEnv(tidepoolClientSecret)
	if err != nil {
		log.Error("err getting client secret ", err.Error())
	}

	return &ServicesClient{
		httpClient: http.DefaultClient,
		config:     clientConfig,
		secret:     secret,
		closed:     make(chan chan bool),
	}
}

// Start starts the client and makes it ready for use.  This must be done before using any of the functionality
// that requires a server token
func (client *ServicesClient) Start() error {
	if err := client.serverLogin(); err != nil {
		log.Error("Problem with initial server token acquisition:", err.Error())
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
					log.Error("Error when refreshing server login:", err.Error())
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
	host := client.getUserHost()
	if host == nil {
		return errors.New("No known user-api hosts.")
	}

	host.Path += "/serverlogin"

	req, _ := http.NewRequest("POST", host.String(), nil)
	req.Header.Add(xTidepoolServerName, client.config.Name)
	req.Header.Add(xTidepoolServerSecret, client.secret)

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

//CheckToken tests a token with the user-api to make sure it's current;
//if so, it returns the data encoded in the token.
func (client *ServicesClient) CheckToken(token string) *TokenData {
	host := client.getUserHost()
	if host == nil {
		log.Error("No known user-api hosts.")
		return nil
	}

	host.Path += "/token/" + token

	req, _ := http.NewRequest("GET", host.String(), nil)
	req.Header.Add(xTidepoolSessionToken, client.serverToken)

	res, err := client.httpClient.Do(req)
	if err != nil {
		log.Error("Error checking token", err.Error())
		return nil
	}

	switch res.StatusCode {
	case http.StatusOK:
		var td TokenData
		if err = json.NewDecoder(res.Body).Decode(&td); err != nil {
			log.Error("Error parsing JSON results", err.Error())
			return nil
		}
		return &td
	case http.StatusNoContent:
		return nil
	default:
		log.Error("Unknown response ", res.StatusCode, req.URL)
		return nil
	}
}

//GetUser details for the given user
//In this case the userID could be the actual ID or an email address
func (client *ServicesClient) GetUser(userID, token string) (*Data, error) {
	host := client.getUserHost()
	if host == nil {
		return nil, errors.New("No known user-api hosts.")
	}

	host.Path += fmt.Sprintf("user/%s", userID)

	req, _ := http.NewRequest("GET", host.String(), nil)
	req.Header.Add(xTidepoolSessionToken, token)

	res, err := client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failure to get a user \n\n %v", err)
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		var cd Data
		if err = json.NewDecoder(res.Body).Decode(&cd); err != nil {
			log.Error("Error parsing JSON results:", err.Error())
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
func (client *ServicesClient) GetUserPermissons(userID, token string) (*UsersPermissions, error) {
	host := client.getPermissionsHost()
	if host == nil {
		return nil, errors.New("No known user-api hosts.")
	}

	host.Path += fmt.Sprintf("/groups/%s", userID)

	req, _ := http.NewRequest("GET", host.String(), nil)
	req.Header.Add(xTidepoolSessionToken, token)

	res, err := client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failure to get a groups for user \n\n %v", err)
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		var perms UsersPermissions
		if err = json.NewDecoder(res.Body).Decode(&perms); err != nil {
			log.Error("Error parsing JSON results:", err.Error())
			return nil, err
		}
		return &perms, nil
	case http.StatusNoContent:
		return &UsersPermissions{}, nil
	default:
		return nil, fmt.Errorf("Unknown response %d from service[%s]", res.StatusCode, req.URL)
	}
}

func (client *ServicesClient) getPermissionsHost() *url.URL {
	theURL, err := url.Parse(client.config.Host + permissionsPath)
	if err != nil {
		log.Error("Unable to parse permissions urlString:", client.config.Host+permissionsPath)
		return nil
	}
	return theURL
}

func (client *ServicesClient) getUserHost() *url.URL {
	theURL, err := url.Parse(client.config.Host + userPath)
	if err != nil {
		log.Error("Unable to parse user urlString:", client.config.Host+userPath)
		return nil
	}
	return theURL
}
