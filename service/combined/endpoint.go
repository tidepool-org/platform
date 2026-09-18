package combined

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/server"
)

// endpoint participates in a component's existing serving lifecycle without
// opening another listener. Serve signals that its routes and middleware are ready.
type endpoint struct {
	router    service.Router
	ready     chan struct{}
	stopped   chan struct{}
	readyOnce sync.Once
	stopOnce  sync.Once
}

func newEndpoint() *endpoint {
	return &endpoint{ready: make(chan struct{}), stopped: make(chan struct{})}
}

func (e *endpoint) factory(_ *server.Config, _ log.Logger, api service.API) (server.Server, error) {
	router, ok := api.(service.Router)
	if !ok {
		return nil, fmt.Errorf("API %T does not expose its routes", api)
	}
	e.router = router
	return e, nil
}

func (e *endpoint) Serve() error {
	e.readyOnce.Do(func() { close(e.ready) })
	<-e.stopped
	return http.ErrServerClosed
}

func (e *endpoint) Shutdown() error {
	e.stopOnce.Do(func() { close(e.stopped) })
	return nil
}
