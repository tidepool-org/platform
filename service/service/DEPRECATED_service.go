package service

import (
	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/auth"
	authClient "github.com/tidepool-org/platform/auth/client"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/platform"
)

type DEPRECATEDService struct {
	*application.Application
	secret     string
	authClient *authClient.Client
}

func NewDEPRECATEDService() *DEPRECATEDService {
	return &DEPRECATEDService{
		Application: application.New(),
	}
}

func (d *DEPRECATEDService) Initialize(provider application.Provider) error {
	if err := d.Application.Initialize(provider); err != nil {
		return err
	}

	if err := d.initializeSecret(); err != nil {
		return err
	}
	return d.initializeAuthClient()
}

func (d *DEPRECATEDService) Terminate() {
	if d.authClient != nil {
		d.authClient.Close()
		d.authClient = nil
	}
	d.secret = ""

	d.Application.Terminate()
}

func (d *DEPRECATEDService) Secret() string {
	return d.secret
}

func (d *DEPRECATEDService) AuthClient() auth.Client {
	return d.authClient
}

func (d *DEPRECATEDService) initializeSecret() error {
	d.Logger().Debug("Initializing secret")

	secret := d.ConfigReporter().GetWithDefault("secret", "")
	if secret == "" {
		return errors.New("secret is missing")
	}
	d.secret = secret

	return nil
}

func (d *DEPRECATEDService) initializeAuthClient() error {
	d.Logger().Debug("Loading auth client config")

	cfg := authClient.NewConfig()
	cfg.UserAgent = d.UserAgent()
	cfg.ExternalConfig.UserAgent = d.UserAgent()
	if err := cfg.Load(d.ConfigReporter().WithScopes("auth", "client")); err != nil {
		return errors.Wrap(err, "unable to load auth client config")
	}

	d.Logger().Debug("Creating auth client")

	clnt, err := authClient.NewClient(cfg, platform.AuthorizeAsService, d.Name(), d.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create auth client")
	}
	d.authClient = clnt

	d.Logger().Debug("Starting auth client")

	if err = d.authClient.Start(); err != nil {
		return errors.Wrap(err, "unable to start auth client")
	}

	return nil
}
