package client

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/request"

	"github.com/mdblp/shoreline/token"
)

const (
	TidepoolServerNameHeaderName   = "X-Tidepool-Server-Name"
	TidepoolServerSecretHeaderName = "X-Tidepool-Server-Secret"

	ServerSessionTokenTimeoutOnFailureFirst = 1 * time.Second
	ServerSessionTokenTimeoutOnFailureLast  = 60 * time.Second
)

type ExternalConfig struct {
	AuthenticationConfig     *platform.Config
	UserSessionTokenSecret   string
	ServerSessionTokenSecret string
}

func NewExternalConfig() *ExternalConfig {
	return &ExternalConfig{
		AuthenticationConfig: platform.NewConfig(),
	}
}

func (e *ExternalConfig) Load(configReporter config.Reporter) error {
	if err := e.AuthenticationConfig.Load(configReporter); err != nil {
		return err
	}
	e.AuthenticationConfig.Address = configReporter.GetWithDefault("authentication_address", "")
	e.ServerSessionTokenSecret = configReporter.GetWithDefault("server_session_token_secret", "")
	e.UserSessionTokenSecret = configReporter.GetWithDefault("user_session_token_secret", "")

	return nil
}

func (e *ExternalConfig) Validate() error {
	if err := e.AuthenticationConfig.Validate(); err != nil {
		return err
	}
	if e.ServerSessionTokenSecret == "" {
		return errors.New("server session token secret is missing")
	}

	return nil
}

type External struct {
	name                     string
	logger                   log.Logger
	serverSessionTokenSecret string
	userSessionTokenSecret   string
}

func NewExternal(cfg *ExternalConfig, name string, lgr log.Logger) (*External, error) {
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

	return &External{
		logger:                   lgr,
		name:                     name,
		serverSessionTokenSecret: cfg.ServerSessionTokenSecret,
		userSessionTokenSecret:   cfg.UserSessionTokenSecret,
	}, nil
}

func (e *External) ValidateSessionToken(ctx context.Context, sessionToken string) (request.Details, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if sessionToken == "" {
		return nil, errors.New("token is missing")
	}

	tokenData, err := token.UnpackSessionTokenAndVerify(sessionToken, e.userSessionTokenSecret)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	if tokenData.IsServer {
		tokenData.UserId = ""
	} else if tokenData.UserId == "" {
		return nil, errors.New("user id is missing")
	}

	return request.NewDetails(request.MethodSessionToken, tokenData.UserId, sessionToken, tokenData.Role), nil
}
