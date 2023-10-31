package service

import (
	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/auth"
	authClient "github.com/tidepool-org/platform/auth/client"
	"github.com/tidepool-org/platform/errors"
)

type Service struct {
	*application.Application
	secret     string
	authClient *authClient.Client
}

func NewService() *Service {
	return &Service{
		Application: application.New(),
	}
}

func (d *Service) Initialize(provider application.Provider) error {
	if err := d.Application.Initialize(provider); err != nil {
		return err
	}

	if err := d.initializeSecret(); err != nil {
		return err
	}
	return d.initializeAuthClient()
}

func (d *Service) Terminate() {
	if d.authClient != nil {
		d.authClient = nil
	}
	d.secret = ""

	d.Application.Terminate()
}

func (d *Service) Secret() string {
	return d.secret
}

func (d *Service) AuthClient() auth.Client {
	return d.authClient
}

func (d *Service) initializeSecret() error {
	d.Logger().Debug("Initializing secret")

	secret := d.ConfigReporter().GetWithDefault("secret", "")
	if secret == "" {
		return errors.New("secret is missing")
	}
	d.secret = secret

	return nil
}

func (d *Service) initializeAuthClient() error {
	d.Logger().Debug("Loading auth client config")

	userAgent := d.UserAgent()
	cfg := authClient.NewConfig()
	cfg.ExternalConfig.AuthenticationConfig.UserAgent = userAgent
	if err := cfg.Load(d.ConfigReporter().WithScopes("auth", "client")); err != nil {
		return errors.Wrap(err, "unable to load auth client config")
	}

	d.Logger().Debug("Creating auth client")

	clnt, err := authClient.NewClient(cfg, d.Name(), d.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create auth client")
	}
	d.authClient = clnt

	d.Logger().Debug("Starting auth client")

	return nil
}
