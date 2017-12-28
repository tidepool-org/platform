package service

import (
	authClient "github.com/tidepool-org/platform/auth/client"
	"github.com/tidepool-org/platform/errors"
)

type Authenticated struct {
	*Service
	authClient *authClient.Client
}

func NewAuthenticated(prefix string) (*Authenticated, error) {
	svc, err := New(prefix)
	if err != nil {
		return nil, err
	}

	return &Authenticated{
		Service: svc,
	}, nil
}

func (a *Authenticated) Initialize() error {
	if err := a.Service.Initialize(); err != nil {
		return err
	}

	return a.initializeAuthClient()
}

func (a *Authenticated) Terminate() {
	a.terminateAuthClient()

	a.Service.Terminate()
}

func (a *Authenticated) initializeAuthClient() error {
	a.Logger().Debug("Loading auth client config")

	cfg := authClient.NewConfig()
	cfg.UserAgent = a.UserAgent()
	cfg.ExternalConfig.UserAgent = a.UserAgent()
	if err := cfg.Load(a.ConfigReporter().WithScopes("auth", "client")); err != nil {
		return errors.Wrap(err, "unable to load auth client config")
	}

	a.Logger().Debug("Creating auth client")

	clnt, err := authClient.NewClient(cfg, a.Name(), a.Logger())
	if err != nil {
		return errors.Wrap(err, "unable to create auth client")
	}
	a.authClient = clnt

	a.Logger().Debug("Starting auth client")

	if err = a.authClient.Start(); err != nil {
		return errors.Wrap(err, "unable to start auth client")
	}

	a.SetAuthClient(a.authClient)

	return nil
}

func (a *Authenticated) terminateAuthClient() {
	if a.authClient != nil {
		a.Logger().Debug("Closing auth client")
		a.authClient.Close()

		a.Logger().Debug("Destroying auth client")
		a.authClient = nil

		a.SetAuthClient(nil)
	}
}
