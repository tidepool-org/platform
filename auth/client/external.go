package client

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"go.uber.org/fx"

	"github.com/kelseyhightower/envconfig"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/client"
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

var ExternalClientModule = fx.Provide(func(name ServiceName, loader ExternalConfigLoader, logger log.Logger, lifecycle fx.Lifecycle) (auth.ExternalAccessor, error) {
	cfg := NewExternalConfig()
	cfg.Config.UserAgent = string(name)
	if err := cfg.Load(loader); err != nil {
		return nil, err
	}
	external, err := NewExternal(cfg, platform.AuthorizeAsService, string(name), logger)
	if err != nil {
		return nil, err
	}

	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return external.Start()
		},
		OnStop: func(ctx context.Context) error {
			external.Close()
			return nil
		},
	})

	return external, nil
})

func ProvideServiceName(name string) fx.Option {
	return fx.Supply(ServiceName(name))
}

func ProvideExternalLoader(reporter config.Reporter) ExternalConfigLoader {
	scoped := reporter.WithScopes("auth", "client", "external")
	return NewExternalConfigReporterLoader(scoped)
}

type ServiceName string

type ExternalConfig struct {
	*platform.Config
	ServerSessionTokenSecret  string        `envconfig:"TIDEPOOL_AUTH_CLIENT_EXTERNAL_SERVER_SESSION_TOKEN_SECRET"`
	ServerSessionTokenTimeout time.Duration `envconfig:"TIDEPOOL_AUTH_CLIENT_EXTERNAL_SERVER_SESSION_TOKEN_TIMEOUT" default:"1h"`
}

func NewExternalConfig() *ExternalConfig {
	return &ExternalConfig{
		Config:                    platform.NewConfig(),
		ServerSessionTokenTimeout: ServerSessionTokenTimeout,
	}
}

const ServerSessionTokenTimeout = time.Hour

func (e *ExternalConfig) Load(loader ExternalConfigLoader) error {
	return loader.Load(e)
}

func (e *ExternalConfig) Validate() error {
	if err := e.Config.Validate(); err != nil {
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
	client                    *platform.Client
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

	clnt, err := platform.NewClient(cfg.Config, authorizeAs)
	if err != nil {
		return nil, err
	}

	return &External{
		client:                    clnt,
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

func (e *External) ValidateSessionToken(ctx context.Context, token string) (request.AuthDetails, error) {
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
	if err := e.client.RequestData(ctx, "GET", e.client.ConstructURL("auth", "token", token), nil, nil, &result); err != nil {
		return nil, err
	}

	if result.IsServer {
		result.UserID = ""
	} else if result.UserID == "" {
		return nil, errors.New("user id is missing")
	}

	return request.NewAuthDetails(request.MethodSessionToken, result.UserID, token), nil
}

func (e *External) EnsureAuthorized(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is missing")
	}

	if details := request.GetAuthDetails(ctx); details != nil {
		return nil
	}

	return request.ErrorUnauthorized()
}

func (e *External) EnsureAuthorizedService(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is missing")
	}

	if details := request.GetAuthDetails(ctx); details != nil {
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

	if details := request.GetAuthDetails(ctx); details != nil {
		if details.IsService() {
			return "", nil
		}

		authenticatedUserID := details.UserID()
		if authenticatedUserID == targetUserID {
			if authorizedPermission != permission.Custodian {
				return authenticatedUserID, nil
			}
		} else {
			url := e.client.ConstructURL("access", targetUserID, authenticatedUserID)
			permissions := permission.Permissions{}
			if err := e.client.RequestData(ctx, "GET", url, nil, nil, &permissions); err != nil {
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

// GetUserPermissions implements permission.Client
// TODO: consolidate this, this is the same as the permissions client but is
// known to be available to service/service.Authenticated. This is here for
// convenience to not require PermissionsClient within existing Authenticated
// services
func (e *External) GetUserPermissions(ctx context.Context, requestUserID string, targetUserID string) (permission.Permissions, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if requestUserID == "" {
		return nil, errors.New("request user id is missing")
	}
	if targetUserID == "" {
		return nil, errors.New("target user id is missing")
	}

	url := e.client.ConstructURL("access", targetUserID, requestUserID)
	result := permission.Permissions{}
	if err := e.client.RequestData(ctx, "GET", url, nil, nil, &result); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, request.ErrorUnauthorized()
		}
		return nil, err
	}

	return permission.FixOwnerPermissions(result), nil
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
	requestURL := e.client.ConstructURL("auth", "serverlogin")
	request, err := http.NewRequest(requestMethod, requestURL, nil)
	if err != nil {
		return errors.Wrapf(err, "unable to create new request for %s %s", requestMethod, requestURL)
	}

	request.Header.Add(TidepoolServerNameHeaderName, e.name)
	request.Header.Add(TidepoolServerSecretHeaderName, e.serverSessionTokenSecret)

	response, err := e.client.HTTPClient().Do(request)
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

// ExternalConfigLoader abstracts the method by which config values are loaded.
type ExternalConfigLoader interface {
	// Load sets config values for the properties of ExternalConfig.
	Load(*ExternalConfig) error
}

// externalConfigReporterLoader adapts a config.Reporter to implement ConfigLoader.
type externalConfigReporterLoader struct {
	Reporter config.Reporter
	platform.ConfigLoader
}

func NewExternalConfigReporterLoader(reporter config.Reporter) *externalConfigReporterLoader {
	return &externalConfigReporterLoader{
		Reporter:     reporter,
		ConfigLoader: platform.NewConfigReporterLoader(reporter),
	}
}

// Load implements ConfigLoader.
func (l *externalConfigReporterLoader) Load(cfg *ExternalConfig) error {
	if err := l.ConfigLoader.Load(cfg.Config); err != nil {
		return err
	}
	cfg.ServerSessionTokenSecret = l.Reporter.GetWithDefault("server_session_token_secret", "")
	if serverSessionTokenTimeoutString, err := l.Reporter.Get("server_session_token_timeout"); err == nil {
		var serverSessionTokenTimeoutInteger int64
		serverSessionTokenTimeoutInteger, err = strconv.ParseInt(serverSessionTokenTimeoutString, 10, 0)
		if err != nil {
			return errors.New("server session token timeout is invalid")
		}
		cfg.ServerSessionTokenTimeout = time.Duration(serverSessionTokenTimeoutInteger) * time.Second
	}

	return nil
}

// externalEnvconfigLoader adapts envconfig to implement ConfigLoader.
type externalEnvconfigLoader struct {
	platform.ConfigLoader
}

// NewExternalEnvconfigLoader loads values via envconfig.
//
// If loader is nil, it defaults to envconfig for platform values.
func NewExternalEnvconfigLoader(loader platform.ConfigLoader) *externalEnvconfigLoader {
	if loader == nil {
		loader = platform.NewEnvconfigLoader(nil)
	}
	return &externalEnvconfigLoader{
		ConfigLoader: loader,
	}
}

// Load implements ConfigLoader.
func (l *externalEnvconfigLoader) Load(cfg *ExternalConfig) error {
	eeCfg := &struct {
		Address string `envconfig:"TIDEPOOL_AUTH_CLIENT_EXTERNAL_ADDRESS" required:"true"`
		*ExternalConfig
	}{ExternalConfig: cfg}
	if err := envconfig.Process(client.EnvconfigEmptyPrefix, eeCfg); err != nil {
		return err
	}
	// Override the client.Config.Address. It's not possible to change the
	// envconfig tag on the config.Client at runtime. In addition, we don't
	// want to use the envconfig.Prefix so that the code is more easily
	// searched. The results is that we have to override this value.
	cfg.Config.Config.Address = eeCfg.Address
	if err := l.ConfigLoader.Load(cfg.Config); err != nil {
		return err
	}
	return nil
}
