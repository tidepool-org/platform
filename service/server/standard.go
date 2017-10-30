package server

import (
	"net/http"

	graceful "gopkg.in/tylerb/graceful.v1"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
)

type Standard struct {
	logger log.Logger
	api    service.API
	config *Config
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

	return &Standard{
		logger: lgr,
		api:    api,
		config: cfg,
	}, nil
}

func (s *Standard) Serve() error {
	server := &graceful.Server{
		Server: &http.Server{
			Addr:    s.config.Address,
			Handler: s.api.Handler(),
		},
		Timeout: s.config.Timeout,
	}

	var err error
	if s.config.TLS {
		s.logger.Infof("Serving HTTPS at %s", s.config.Address)
		err = server.ListenAndServeTLS(s.config.TLSCertificateFile, s.config.TLSKeyFile)
	} else {
		s.logger.Infof("Serving HTTP at %s", s.config.Address)
		err = server.ListenAndServe()
	}
	return err
}
