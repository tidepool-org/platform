package server

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import (
	"net/http"
	"time"

	graceful "gopkg.in/tylerb/graceful.v1"

	"github.com/tidepool-org/platform/app"
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
		return nil, app.Error("server", "logger is missing")
	}
	if api == nil {
		return nil, app.Error("server", "api is missing")
	}
	if config == nil {
		return nil, app.Error("server", "config is missing")
	}

	config = config.Clone()
	if err := config.Validate(); err != nil {
		return nil, app.ExtError(err, "server", "config is invalid")
	}

	return &Standard{
		logger: logger,
		api:    api,
		config: config,
	}, nil
}

func (s *Standard) Serve() error {
	server := &graceful.Server{
		Timeout: time.Duration(s.config.Timeout) * time.Second,
		Server: &http.Server{
			Addr:    s.config.Address,
			Handler: s.api.Handler(),
		},
	}

	var err error
	if s.config.TLS != nil {
		err = server.ListenAndServeTLS(s.config.TLS.CertificateFile, s.config.TLS.KeyFile)
	} else {
		err = server.ListenAndServe()
	}
	return err
}
