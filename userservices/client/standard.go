package client

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
)

type Standard struct {
	logger             log.Logger
	name               string
	config             *Config
	httpClient         *http.Client
	closingChannel     chan chan bool
	serverTokenTimeout time.Duration
	serverTokenMutex   sync.Mutex
	serverTokenSafe    string
}

const (
	TidepoolServerNameHeaderName          = "X-Tidepool-Server-Name"
	TidepoolServerSecretHeaderName        = "X-Tidepool-Server-Secret"
	TidepoolAuthenticationTokenHeaderName = "X-Tidepool-Session-Token"

	ServerTokenTimeoutOnFailureFirst = time.Second
	ServerTokenTimeoutOnFailureLast  = 60 * time.Second
)

func NewStandard(logger log.Logger, name string, config *Config) (*Standard, error) {
	if logger == nil {
		return nil, app.Error("client", "logger is missing")
	}
	if name == "" {
		return nil, app.Error("client", "name is missing")
	}
	if config == nil {
		return nil, app.Error("client", "config is missing")
	}

	config = config.Clone()
	if err := config.Validate(); err != nil {
		return nil, app.ExtError(err, "client", "config is invalid")
	}

	httpClient := &http.Client{
		Timeout: time.Duration(config.RequestTimeout) * time.Second,
	}
	serverTokenTimeout := time.Duration(config.ServerTokenTimeout) * time.Second

	return &Standard{
		logger:             logger,
		name:               name,
		config:             config,
		httpClient:         httpClient,
		serverTokenTimeout: serverTokenTimeout,
	}, nil
}

func (s *Standard) Start() error {
	if s.closingChannel == nil {
		closingChannel := make(chan chan bool)
		s.closingChannel = closingChannel

		serverTokenTimeout := s.timeoutServerToken(0)

		go func() {
			for {
				timer := time.After(serverTokenTimeout)
				select {
				case closedChannel := <-closingChannel:
					closedChannel <- true
					close(closedChannel)
					return
				case <-timer:
					serverTokenTimeout = s.timeoutServerToken(serverTokenTimeout)
				}
			}
		}()
	}

	return nil
}

func (s *Standard) Close() {
	if s.closingChannel != nil {
		closingChannel := s.closingChannel
		s.closingChannel = nil

		closedChannel := make(chan bool)
		closingChannel <- closedChannel
		close(closingChannel)
		<-closedChannel
	}
}

func (s *Standard) ValidateAuthenticationToken(context service.Context, authenticationToken string) (AuthenticationDetails, error) {
	if context == nil {
		return nil, app.Error("client", "context is missing")
	}
	if authenticationToken == "" {
		return nil, app.Error("client", "authentication token is missing")
	}

	if s.closingChannel == nil {
		return nil, app.Error("client", "client is closed")
	}

	context.Logger().WithField("authentication-token", authenticationToken).Debug("Validating authentication token")

	var authentication struct {
		IsServer bool
		UserID   string
	}

	if err := s.sendRequest(context, "GET", s.buildURL("auth", "token", authenticationToken), &authentication); err != nil {
		return nil, err
	}

	if !authentication.IsServer && authentication.UserID == "" {
		return nil, app.Error("client", "user id is missing")
	}

	return &authenticationDetails{
		token:    authenticationToken,
		isServer: authentication.IsServer,
		userID:   authentication.UserID,
	}, nil
}

func (s *Standard) GetUserPermissions(context service.Context, requestUserID string, targetUserID string) (Permissions, error) {
	if context == nil {
		return nil, app.Error("client", "context is missing")
	}
	if requestUserID == "" {
		return nil, app.Error("client", "request user id is missing")
	}
	if targetUserID == "" {
		return nil, app.Error("client", "target user id is missing")
	}

	if s.closingChannel == nil {
		return nil, app.Error("client", "client is closed")
	}

	context.Logger().WithFields(log.Fields{"request-user-id": requestUserID, "target-user-id": targetUserID}).Debug("Get user permissions")

	permissions := Permissions{}
	if err := s.sendRequest(context, "GET", s.buildURL("access", targetUserID, requestUserID), &permissions); err != nil {
		if unexpectedResponseError, ok := err.(*UnexpectedResponseError); ok {
			if unexpectedResponseError.StatusCode == http.StatusNotFound {
				return nil, NewUnauthorizedError()
			}
		}
		return nil, err
	}

	// Fix missing view and upload permissions for an owner
	if permission, ok := permissions[OwnerPermission]; ok {
		if _, ok = permissions[UploadPermission]; !ok {
			permissions[UploadPermission] = permission
		}
		if _, ok = permissions[ViewPermission]; !ok {
			permissions[ViewPermission] = permission
		}
	}

	return permissions, nil
}

func (s *Standard) GetUserGroupID(context service.Context, userID string) (string, error) {
	if context == nil {
		return "", app.Error("client", "context is missing")
	}
	if userID == "" {
		return "", app.Error("client", "user id is missing")
	}

	if s.closingChannel == nil {
		return "", app.Error("client", "client is closed")
	}

	context.Logger().WithField("user-id", userID).Debug("Getting user group id")

	var uploadsPair struct {
		ID    string
		Value string
	}

	if err := s.sendRequest(context, "GET", s.buildURL("metadata", userID, "private", "uploads"), &uploadsPair); err != nil {
		return "", err
	}

	groupID := uploadsPair.ID
	if groupID == "" {
		return "", app.Error("client", "group id is missing")
	}

	return groupID, nil
}

// TODO: Current user related APIs return http.StatusUnauthorized for BOTH bad server token
// AND bad session token. Since a bad server token is unlikely (though possible) we MUST assume
// that it means bad session token.

func (s *Standard) sendRequest(context service.Context, requestMethod string, requestURL string, responseObject interface{}) error {

	serverToken := s.serverToken()
	if serverToken == "" {
		return app.Errorf("client", "unable to obtain server token for %s %s", requestMethod, requestURL)
	}

	request, err := http.NewRequest(requestMethod, requestURL, nil)
	if err != nil {
		return app.ExtErrorf(err, "client", "unable to create new request for %s %s", requestMethod, requestURL)
	}

	if err = service.CopyRequestTrace(context.Request(), request); err != nil {
		return app.ExtErrorf(err, "client", "unable to copy request trace")
	}

	request.Header.Add(TidepoolAuthenticationTokenHeaderName, serverToken)

	response, err := s.httpClient.Do(request)
	if err != nil {
		return app.ExtErrorf(err, "client", "unable to perform request %s %s", requestMethod, requestURL)
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		if responseObject != nil {
			if err = json.NewDecoder(response.Body).Decode(responseObject); err != nil {
				return app.ExtErrorf(err, "client", "error decoding JSON response from %s %s", request.Method, request.URL.String())
			}
		}
		return nil
	case http.StatusUnauthorized:
		return NewUnauthorizedError()
	default:
		return NewUnexpectedResponseError(response, request)
	}
}

func (s *Standard) timeoutServerToken(serverTokenTimeout time.Duration) time.Duration {
	if err := s.refreshServerToken(); err != nil {
		if serverTokenTimeout == 0 || serverTokenTimeout == s.serverTokenTimeout {
			serverTokenTimeout = ServerTokenTimeoutOnFailureFirst
		} else {
			serverTokenTimeout *= 2
			if serverTokenTimeout > ServerTokenTimeoutOnFailureLast {
				serverTokenTimeout = ServerTokenTimeoutOnFailureLast
			}
		}
		s.logger.WithError(err).WithField("retry", serverTokenTimeout.String()).Error("Unable to refresh server token; retrying")
	} else {
		serverTokenTimeout = s.serverTokenTimeout
	}

	return serverTokenTimeout
}

func (s *Standard) refreshServerToken() error {

	s.logger.Debug("Refreshing server token")

	requestMethod := "POST"
	requestURL := s.buildURL("auth", "serverlogin")
	request, err := http.NewRequest(requestMethod, requestURL, nil)
	if err != nil {
		return app.ExtErrorf(err, "client", "unable to create new request for %s %s", requestMethod, requestURL)
	}

	request.Header.Add(TidepoolServerNameHeaderName, s.name)
	request.Header.Add(TidepoolServerSecretHeaderName, s.config.ServerTokenSecret)

	response, err := s.httpClient.Do(request)
	if err != nil {
		return app.ExtError(err, "client", "failure requesting new server token")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return app.Errorf("client", "unexpected response status code %d while requesting new server token", response.StatusCode)
	}

	serverTokenHeader := response.Header.Get(TidepoolAuthenticationTokenHeaderName)
	if serverTokenHeader == "" {
		return app.Error("client", "server token is missing")
	}

	s.setServerToken(serverTokenHeader)

	return nil
}

func (s *Standard) setServerToken(serverToken string) {
	s.serverTokenMutex.Lock()
	defer s.serverTokenMutex.Unlock()

	s.serverTokenSafe = serverToken
}

func (s *Standard) serverToken() string {
	s.serverTokenMutex.Lock()
	defer s.serverTokenMutex.Unlock()

	return s.serverTokenSafe
}

func (s *Standard) buildURL(paths ...string) string {
	return strings.Join(append([]string{s.config.Address}, paths...), "/")
}

type authenticationDetails struct {
	token    string
	isServer bool
	userID   string
}

func (a *authenticationDetails) Token() string {
	return a.token
}

func (a *authenticationDetails) IsServer() bool {
	return a.isServer
}

func (a *authenticationDetails) UserID() string {
	return a.userID
}
