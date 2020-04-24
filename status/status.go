package status

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
	"go.uber.org/fx"

	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/version"
)

var RouterModule = fx.Provide(fx.Annotated{
	Group:  "routers",
	Target: NewRouter,
})

type Status struct {
	Version string
	Store   interface{}
}

type StoreStatusReporter interface {
	Status() interface{}
}

type Router struct {
	versionReporter     version.Reporter
	storeStatusReporter StoreStatusReporter
}

type Params struct {
	fx.In

	VersionReporter     version.Reporter
	StoreStatusReporter StoreStatusReporter
}

func NewRouter(p Params) service.Router {
	return &Router{
		versionReporter:     p.VersionReporter,
		storeStatusReporter: p.StoreStatusReporter,
	}
}

func (r *Router) Routes() []*rest.Route {
	return []*rest.Route{
		rest.Get("/status", r.StatusGet),
	}
}

func (r *Router) StatusGet(res rest.ResponseWriter, req *rest.Request) {
	s := &Status{
		Version: r.versionReporter.Long(),
		Store:   r.storeStatusReporter.Status(),
	}

	request.MustNewResponder(res, req).Data(http.StatusOK, s)
}
