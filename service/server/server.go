package server

import (
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
)

// Server owns a service's serving lifecycle. A combined application can supply
// a Factory which mounts the API in a shared listener instead.
type Server interface {
	Serve() error
	Shutdown() error
}

type Factory func(*Config, log.Logger, service.API) (Server, error)

func New(cfg *Config, logger log.Logger, api service.API) (Server, error) {
	return NewStandard(cfg, logger, api)
}
