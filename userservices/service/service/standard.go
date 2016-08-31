package service

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
	"github.com/tidepool-org/platform/app"
	metricservicesClient "github.com/tidepool-org/platform/metricservices/client"
	"github.com/tidepool-org/platform/service/server"
	"github.com/tidepool-org/platform/service/service"
	userservicesClient "github.com/tidepool-org/platform/userservices/client"
	"github.com/tidepool-org/platform/userservices/service/api"
	"github.com/tidepool-org/platform/userservices/service/api/v1"
)

type Standard struct {
	*service.Standard
	metricServicesClient *metricservicesClient.Standard
	userServicesClient   *userservicesClient.Standard
	userServicesAPI      *api.Standard
	userServicesServer   *server.Standard
}

func NewStandard() (*Standard, error) {
	standard, err := service.NewStandard("userservices", "TIDEPOOL")
	if err != nil {
		return nil, err
	}

	return &Standard{
		Standard: standard,
	}, nil
}

func (s *Standard) Initialize() error {
	if err := s.Standard.Initialize(); err != nil {
		return err
	}

	if err := s.initializeMetricServicesClient(); err != nil {
		return err
	}
	if err := s.initializeUserServicesClient(); err != nil {
		return err
	}
	if err := s.initializeUserServicesAPI(); err != nil {
		return err
	}
	if err := s.initializeUserServicesServer(); err != nil {
		return err
	}

	return nil
}

func (s *Standard) Terminate() {
	s.userServicesServer = nil
	s.userServicesAPI = nil
	if s.userServicesClient != nil {
		s.userServicesClient.Close()
		s.userServicesClient = nil
	}
	s.metricServicesClient = nil

	s.Standard.Terminate()
}

func (s *Standard) Run() error {
	if s.userServicesServer == nil {
		return app.Error("service", "service not initialized")
	}

	return s.userServicesServer.Serve()
}

func (s *Standard) initializeMetricServicesClient() error {
	s.Logger().Debug("Loading metric services client config")

	metricServicesClientConfig := &metricservicesClient.Config{}
	if err := s.ConfigLoader().Load("metricservices_client", metricServicesClientConfig); err != nil {
		return app.ExtError(err, "service", "unable to load metric services client config")
	}

	s.Logger().Debug("Creating metric services client")

	metricServicesClient, err := metricservicesClient.NewStandard(s.Logger(), s.Name(), metricServicesClientConfig)
	if err != nil {
		return app.ExtError(err, "service", "unable to create metric services client")
	}
	s.metricServicesClient = metricServicesClient

	return nil
}

func (s *Standard) initializeUserServicesClient() error {
	s.Logger().Debug("Loading user services client config")

	userServicesClientConfig := &userservicesClient.Config{}
	if err := s.ConfigLoader().Load("userservices_client", userServicesClientConfig); err != nil {
		return app.ExtError(err, "service", "unable to load user services client config")
	}

	s.Logger().Debug("Creating user services client")

	userServicesClient, err := userservicesClient.NewStandard(s.Logger(), s.Name(), userServicesClientConfig)
	if err != nil {
		return app.ExtError(err, "service", "unable to create user services client")
	}
	s.userServicesClient = userServicesClient

	s.Logger().Debug("Starting user services client")
	if err = s.userServicesClient.Start(); err != nil {
		return app.ExtError(err, "service", "unable to start user services client")
	}

	return nil
}

func (s *Standard) initializeUserServicesAPI() error {
	s.Logger().Debug("Creating user services api")

	userServicesAPI, err := api.NewStandard(s.VersionReporter(), s.EnvironmentReporter(), s.Logger(), s.metricServicesClient, s.userServicesClient)
	if err != nil {
		return app.ExtError(err, "service", "unable to create user services api")
	}
	s.userServicesAPI = userServicesAPI

	s.Logger().Debug("Initializing user services api middleware")

	if err = s.userServicesAPI.InitializeMiddleware(); err != nil {
		return app.ExtError(err, "service", "unable to initialize user services api middleware")
	}

	s.Logger().Debug("Initializing user services api router")

	if err = s.userServicesAPI.InitializeRouter(v1.Routes()); err != nil {
		return app.ExtError(err, "service", "unable to initialize user services api router")
	}

	return nil
}

func (s *Standard) initializeUserServicesServer() error {
	s.Logger().Debug("Loading user services server config")

	userServicesServerConfig := &server.Config{}
	if err := s.ConfigLoader().Load("userservices_server", userServicesServerConfig); err != nil {
		return app.ExtError(err, "service", "unable to load user services server config")
	}

	s.Logger().Debug("Creating user services server")

	userServicesServer, err := server.NewStandard(s.Logger(), s.userServicesAPI, userServicesServerConfig)
	if err != nil {
		return app.ExtError(err, "service", "unable to create user services server")
	}
	s.userServicesServer = userServicesServer

	return nil
}
