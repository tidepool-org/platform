package client

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/request"
)

const (
	TidepoolServerNameHeaderName   = "X-Tidepool-Server-Name"
	TidepoolServerSecretHeaderName = "X-Tidepool-Server-Secret"

	ServerSessionTokenTimeoutOnFailureFirst = 1 * time.Second
	ServerSessionTokenTimeoutOnFailureLast  = 60 * time.Second
)

type ExternalConfig struct {
	AuthenticationConfig      *platform.Config
	AuthorizationConfig       *platform.Config
	ServerSessionTokenSecret  string
	ServerSessionTokenTimeout time.Duration
}

func NewExternalConfig() *ExternalConfig {
	return &ExternalConfig{
		AuthenticationConfig:      platform.NewConfig(),
		AuthorizationConfig:       platform.NewConfig(),
		ServerSessionTokenTimeout: 3600 * time.Second,
	}
}

func (e *ExternalConfig) Load(configReporter config.Reporter) error {
	if err := e.AuthenticationConfig.Load(configReporter); err != nil {
		return err
	}
	if err := e.AuthorizationConfig.Load(configReporter); err != nil {
		return err
	}
	e.AuthenticationConfig.Address = configReporter.GetWithDefault("authentication_address", "")
	e.AuthorizationConfig.Address = configReporter.GetWithDefault("authorization_address", "")
	e.ServerSessionTokenSecret = configReporter.GetWithDefault("server_session_token_secret", "")
	if serverSessionTokenTimeoutString, err := configReporter.Get("server_session_token_timeout"); err == nil {
		var serverSessionTokenTimeoutInteger int64
		serverSessionTokenTimeoutInteger, err = strconv.ParseInt(serverSessionTokenTimeoutString, 10, 0)
		if err != nil {
			return errors.New("server session token timeout is invalid")
		}
		e.ServerSessionTokenTimeout = time.Duration(serverSessionTokenTimeoutInteger) * time.Second
	}

	return nil
}

func (e *ExternalConfig) Validate() error {
	if err := e.AuthenticationConfig.Validate(); err != nil {
		return err
	}
	if err := e.AuthorizationConfig.Validate(); err != nil {
		return err
	}

	if e.ServerSessionTokenSecret == "" {
		return errors.New("server session token secret is missing")
	}
	if e.ServerSessionTokenTimeout <= 0 {
		return errors.New("server session token timeout is invalid")
	}

	return nil
}

type External struct {
	authenticationClient      *platform.Client
	authorizationClient       *platform.Client
	name                      string
	logger                    log.Logger
	serverSessionTokenSecret  string
	serverSessionTokenTimeout time.Duration
	serverSessionTokenMutex   sync.Mutex
	serverSessionTokenSafe    string
	closingChannel            chan chan bool
}

func NewExternal(cfg *ExternalConfig, authorizeAs platform.AuthorizeAs, name string, lgr log.Logger) (*External, error) {
	if cfg == nil {
		return nil, errors.New("config is missing")
	}
	if name == "" {
		return nil, errors.New("name is missing")
	}
	if lgr == nil {
		return nil, errors.New("logger is missing")
	}

	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "config is invalid")
	}

	authenticationClnt, err := platform.NewClient(cfg.AuthenticationConfig, authorizeAs)
	if err != nil {
		return nil, err
	}
	authorizationClnt, err := platform.NewClient(cfg.AuthorizationConfig, authorizeAs)
	if err != nil {
		return nil, err
	}

	return &External{
		authenticationClient:      authenticationClnt,
		authorizationClient:       authorizationClnt,
		logger:                    lgr,
		name:                      name,
		serverSessionTokenSecret:  cfg.ServerSessionTokenSecret,
		serverSessionTokenTimeout: cfg.ServerSessionTokenTimeout,
	}, nil
}

func (e *External) Start() error {
	if e.closingChannel == nil {
		closingChannel := make(chan chan bool)
		e.closingChannel = closingChannel

		serverSessionTokenTimeout := e.timeoutServerSessionToken(0)

		go func() {
			for {
				timer := time.After(serverSessionTokenTimeout)
				select {
				case closedChannel := <-closingChannel:
					closedChannel <- true
					close(closedChannel)
					return
				case <-timer:
					serverSessionTokenTimeout = e.timeoutServerSessionToken(serverSessionTokenTimeout)
				}
			}
		}()
	}

	return nil
}

func (e *External) Close() {
	if e.closingChannel != nil {
		closingChannel := e.closingChannel
		e.closingChannel = nil

		closedChannel := make(chan bool)
		closingChannel <- closedChannel
		close(closingChannel)
		<-closedChannel
	}
}

func (e *External) ServerSessionToken() (string, error) {
	if e.closingChannel == nil {
		return "", errors.New("client is closed")
	}

	serverSessionToken := e.serverSessionToken()
	if serverSessionToken == "" {
		return "", errors.New("unable to obtain server session token")
	}

	return serverSessionToken, nil
}

func (e *External) ValidateSessionToken(ctx context.Context, token string) (request.Details, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if token == "" {
		return nil, errors.New("token is missing")
	}

	var result struct {
		IsServer bool
		UserID   string
	}
	if err := e.authenticationClient.RequestData(ctx, "GET", e.authenticationClient.ConstructURL("token", token), nil, nil, &result); err != nil {
		return nil, err
	}

	if result.IsServer {
		result.UserID = ""
	} else if result.UserID == "" {
		return nil, errors.New("user id is missing")
	}

	return request.NewDetails(request.MethodSessionToken, result.UserID, token), nil
}

func (e *External) EnsureAuthorized(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is missing")
	}

	if details := request.DetailsFromContext(ctx); details != nil {
		return nil
	}

	return request.ErrorUnauthorized()
}

func (e *External) EnsureAuthorizedService(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is missing")
	}

	if details := request.DetailsFromContext(ctx); details != nil {
		if details.IsService() {
			return nil
		}
	}

	return request.ErrorUnauthorized()
}

func (e *External) EnsureAuthorizedUser(ctx context.Context, targetUserID string, authorizedPermission string) (string, error) {
	if ctx == nil {
		return "", errors.New("context is missing")
	}
	if targetUserID == "" {
		return "", errors.New("target user id is missing")
	}
	if authorizedPermission == "" {
		return "", errors.New("authorized permission is missing")
	}

	if details := request.DetailsFromContext(ctx); details != nil {
		if details.IsService() {
			return "", nil
		}

		authenticatedUserID := details.UserID()
		if authenticatedUserID == targetUserID {
			if authorizedPermission != permission.Custodian {
				return authenticatedUserID, nil
			}
		} else {
			url := e.authorizationClient.ConstructURL("access", targetUserID, authenticatedUserID)
			permissions := permission.Permissions{}
			if err := e.authorizationClient.RequestData(ctx, "GET", url, nil, nil, &permissions); err != nil {
				if !request.IsErrorResourceNotFound(err) {
					return "", errors.Wrap(err, "unable to get user permissions")
				}
			} else {
				permissions = permission.FixOwnerPermissions(permissions)
				if _, ok := permissions[authorizedPermission]; ok {
					return authenticatedUserID, nil
				}
			}
		}
	}

	return "", request.ErrorUnauthorized()
}

func (e *External) timeoutServerSessionToken(serverSessionTokenTimeout time.Duration) time.Duration {
	if err := e.refreshServerSessionToken(); err != nil {
		if serverSessionTokenTimeout == 0 || serverSessionTokenTimeout == e.serverSessionTokenTimeout {
			serverSessionTokenTimeout = ServerSessionTokenTimeoutOnFailureFirst
		} else {
			serverSessionTokenTimeout *= 2
			if serverSessionTokenTimeout > ServerSessionTokenTimeoutOnFailureLast {
				serverSessionTokenTimeout = ServerSessionTokenTimeoutOnFailureLast
			}
		}
		e.logger.WithError(err).WithField("retry", serverSessionTokenTimeout.String()).Warn("Unable to refresh server session token; retrying")
	} else {
		serverSessionTokenTimeout = e.serverSessionTokenTimeout
	}

	return serverSessionTokenTimeout
}

func (e *External) refreshServerSessionToken() error {
	e.logger.Debug("Refreshing server session token")

	requestMethod := "POST"
	requestURL := e.authenticationClient.ConstructURL("serverlogin")
	request, err := http.NewRequest(requestMethod, requestURL, nil)
	if err != nil {
		return errors.Wrapf(err, "unable to create new request for %s %s", requestMethod, requestURL)
	}

	request.Header.Add(TidepoolServerNameHeaderName, e.name)
	request.Header.Add(TidepoolServerSecretHeaderName, e.serverSessionTokenSecret)

	response, err := e.authenticationClient.HTTPClient().Do(request)
	if err != nil {
		return errors.Wrap(err, "unable to refresh server session token")
	}
	if response.Body != nil {
		defer response.Body.Close()
	}

	if response.StatusCode != http.StatusOK {
		return errors.Newf("unexpected response status code %d while refreshing server session token", response.StatusCode)
	}

	serverSessionTokenHeader := response.Header.Get(auth.TidepoolSessionTokenHeaderKey)
	if serverSessionTokenHeader == "" {
		return errors.New("server session token is missing")
	}

	e.setServerSessionToken(serverSessionTokenHeader)

	return nil
}

func (e *External) setServerSessionToken(serverSessionToken string) {
	e.serverSessionTokenMutex.Lock()
	defer e.serverSessionTokenMutex.Unlock()

	e.serverSessionTokenSafe = serverSessionToken
}

func (e *External) serverSessionToken() string {
	e.serverSessionTokenMutex.Lock()
	defer e.serverSessionTokenMutex.Unlock()

	return e.serverSessionTokenSafe
}
