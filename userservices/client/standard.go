package client

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
	config             *Config
	httpClient         *http.Client
	closingChannel     chan chan bool
	serverTokenTimeout time.Duration
	serverTokenMutex   sync.Mutex
	serverTokenSafe    string
}

const (
	TidepoolServerNameHeaderName       = "x-tidepool-server-name"
	TidepoolServerSecretHeaderName     = "x-tidepool-server-secret"
	TidepoolUserSessionTokenHeaderName = "x-tidepool-session-token"

	ServerTokenTimeoutOnFailureFirst = time.Second
	ServerTokenTimeoutOnFailureLast  = time.Minute
)

func NewStandard(config *Config, logger log.Logger) (*Standard, error) {
	if config == nil {
		return nil, app.Error("standard", "config is missing")
	}
	if logger == nil {
		return nil, app.Error("standard", "logger is missing")
	}

	if err := config.Validate(); err != nil {
		return nil, app.ExtError(err, "standard", "config is not valid")
	}

	httpClient := &http.Client{
		Timeout: time.Duration(config.RequestTimeout) * time.Second,
	}
	closingChannel := make(chan chan bool)
	serverTokenTimeout := time.Duration(config.ServerTokenTimeout) * time.Minute

	return &Standard{
		logger:             logger,
		config:             config,
		httpClient:         httpClient,
		closingChannel:     closingChannel,
		serverTokenTimeout: serverTokenTimeout,
	}, nil
}

func (s *Standard) Start() error {
	serverTokenTimeout := s.timeoutServerToken(0)

	go func() {
		for {
			timer := time.After(serverTokenTimeout)
			select {
			case closedChannel := <-s.closingChannel:
				closedChannel <- true
				return
			case <-timer:
				serverTokenTimeout = s.timeoutServerToken(serverTokenTimeout)
			}
		}
	}()

	return nil
}

func (s *Standard) Close() {
	closedChannel := make(chan bool)
	s.closingChannel <- closedChannel
	<-closedChannel

	s.setServerToken("")
}

func (s *Standard) ValidateUserSession(context *service.Context, sessionToken string) (string, error) {
	if context == nil {
		return "", app.Error("standard", "context is missing")
	}
	if sessionToken == "" {
		return "", app.Error("standard", "session token is missing")
	}

	var sessionTokenData struct {
		UserID   string
		IsServer bool
	}

	context.Logger().WithField("session-token", sessionToken).Debug("Validating user session")

	if err := s.sendRequest(context, "GET", s.buildURL("auth", "token", sessionToken), &sessionTokenData); err != nil {
		return "", err
	}

	userID := sessionTokenData.UserID
	if userID == "" {
		return "", app.Error("standard", "user id is missing")
	}

	return userID, nil
}

func (s *Standard) ValidateTargetUserPermissions(context *service.Context, requestUserID string, targetUserID string, targetPermissions Permissions) error {
	if context == nil {
		return app.Error("standard", "context is missing")
	}
	if requestUserID == "" {
		return app.Error("standard", "request user id is missing")
	}
	if targetUserID == "" {
		return app.Error("standard", "target user id is missing")
	}
	if len(targetPermissions) == 0 {
		return app.Error("standard", "target permissions is empty")
	}

	context.Logger().WithFields(log.Fields{"request-user-id": requestUserID, "target-user-id": targetUserID, "target-permissions": targetPermissions}).Debug("Validating target user permissions")

	actualPermissions := Permissions{}
	if err := s.sendRequest(context, "GET", s.buildURL("access", targetUserID, requestUserID), &actualPermissions); err != nil {
		if unexpectedResponseError, ok := err.(*UnexpectedResponseError); ok {
			if unexpectedResponseError.StatusCode == http.StatusNotFound {
				return NewUnauthorizedError()
			}
		}
		return err
	}

	if _, ok := actualPermissions["root"]; ok {
		return nil
	}

	for key := range targetPermissions {
		if _, ok := actualPermissions[key]; !ok {
			return NewUnauthorizedError()
		}
	}

	return nil
}

func (s *Standard) GetUserGroupID(context *service.Context, userID string) (string, error) {
	if context == nil {
		return "", app.Error("standard", "context is missing")
	}
	if userID == "" {
		return "", app.Error("standard", "user id is missing")
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
		return "", app.Error("standard", "group id is missing")
	}

	return groupID, nil
}

// TODO: Current user related APIs return http.StatusUnauthorized for BOTH bad server token
// AND bad session token. Since a bad server token is unlikely (though possible) we MUST assume
// that it means bad session token.

func (s *Standard) sendRequest(context *service.Context, method string, url string, responseObject interface{}) error {

	serverToken := s.serverToken()
	if serverToken == "" {
		return app.Errorf("standard", "unable to obtain server token for %s %s", method, url)
	}

	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		return app.ExtErrorf(err, "standard", "unable to create new request for %s %s", method, url)
	}

	if err := service.CopyRequestTrace(context.Request(), request); err != nil {
		return app.ExtErrorf(err, "standard", "unable to copy request trace")
	}

	request.Header.Add(TidepoolUserSessionTokenHeaderName, serverToken)

	response, err := s.httpClient.Do(request)
	if err != nil {
		return app.ExtErrorf(err, "standard", "unable to perform request %s %s", method, url)
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		if responseObject != nil {
			if err = json.NewDecoder(response.Body).Decode(responseObject); err != nil {
				return app.ExtErrorf(err, "standard", "error decoding JSON response from %s %s", request.Method, request.URL.String())
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

	method := "POST"
	url := s.buildURL("auth", "serverlogin")
	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		return app.ExtErrorf(err, "standard", "unable to create new request for %s %s", method, url)
	}

	request.Header.Add(TidepoolServerNameHeaderName, "dataservices")
	request.Header.Add(TidepoolServerSecretHeaderName, s.config.ServerTokenSecret)

	response, err := s.httpClient.Do(request)
	if err != nil {
		return app.ExtError(err, "standard", "failure requesting new server token")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return app.Errorf("standard", "unexpected response status code %d while requesting new server token", response.StatusCode)
	}

	serverTokenHeader := response.Header.Get(TidepoolUserSessionTokenHeaderName)
	if serverTokenHeader == "" {
		return app.Error("standard", "server token is missing")
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
