package server

import (
	"context"
	"net/http"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
)

type Standard struct {
	api    service.API
	config *Config
	logger log.Logger
	server *http.Server
}

func NewStandard(cfg *Config, lgr log.Logger, api service.API) (*Standard, error) {
	if lgr == nil {
		return nil, errors.New("logger is missing")
	}
	if api == nil {
		return nil, errors.New("api is missing")
	}
	if cfg == nil {
		return nil, errors.New("config is missing")
	}

	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "config is invalid")
	}

	server := &http.Server{
		Addr: cfg.Address,
	}

	return &Standard{
		logger: lgr,
		api:    api,
		config: cfg,
		server: server,
	}, nil
}

func (s *Standard) Serve() error {
	s.server.Handler = s.api.Handler()

	var err error
	if s.config.TLS {
		s.logger.Infof("Serving HTTPS at %s", s.config.Address)
		err = s.server.ListenAndServeTLS(s.config.TLSCertificateFile, s.config.TLSKeyFile)
	} else {
		s.logger.Infof("Serving HTTP at %s", s.config.Address)
		err = s.server.ListenAndServe()
	}
	return err
}

func (s *Standard) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.Timeout)
	defer cancel()
	s.logger.Info("Shutting down the server")
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Errorf("Error while gracefully shutting down: %v", err)
		return err
	}
	return nil
}
