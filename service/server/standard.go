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

func NewStandard(logger log.Logger, api service.API, config *Config) (*Standard, error) {
	if logger == nil {
		return nil, errors.New("server", "logger is missing")
	}
	if api == nil {
		return nil, errors.New("server", "api is missing")
	}
	if config == nil {
		return nil, errors.New("server", "config is missing")
	}

	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "server", "config is invalid")
	}

	return &Standard{
		logger: logger,
		api:    api,
		config: config,
	}, nil
}

func (s *Standard) Serve() error {
	server := &graceful.Server{
		Timeout: s.config.Timeout,
		Server: &http.Server{
			Addr:    s.config.Address,
			Handler: s.api.Handler(),
		},
	}

	var err error
	if s.config.TLS {
		err = server.ListenAndServeTLS(s.config.TLSCertificateFile, s.config.TLSKeyFile)
	} else {
		err = server.ListenAndServe()
	}
	return err
}
