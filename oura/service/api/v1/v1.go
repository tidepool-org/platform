package v1

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth"
	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/oura"
	ouraWebhook "github.com/tidepool-org/platform/oura/webhook"
	"github.com/tidepool-org/platform/work"
)

type Dependencies struct {
	AuthClient auth.Client
	OuraClient oura.Client
	WorkClient work.Client
}

func (d Dependencies) Validate() error {
	if d.AuthClient == nil {
		return errors.New("auth client is missing")
	}
	if d.OuraClient == nil {
		return errors.New("oura client is missing")
	}
	if d.WorkClient == nil {
		return errors.New("work client is missing")
	}
	return nil
}

type Router struct {
	Dependencies
}

func NewRouter(dependencies Dependencies) (*Router, error) {
	if err := dependencies.Validate(); err != nil {
		return nil, errors.Wrap(err, "dependencies is invalid")
	}

	return &Router{
		Dependencies: dependencies,
	}, nil
}

func (r *Router) Routes() []*rest.Route {
	webhookPath := oura.PartnerPathPrefix + ouraWebhook.EventPath
	return []*rest.Route{
		rest.Get(webhookPath, r.Subscription),
		rest.Post(webhookPath, r.Event),
	}
}

// NOTE: Below is temporary glue code to adapt deprecated data service routes to new router style.
// FUTURE: Remove once all routes have been converted to new router style.

func Routes() []dataService.Route {
	webhookPath := oura.PartnerPathPrefix + ouraWebhook.EventPath
	return []dataService.Route{
		dataService.Get(webhookPath, Subscription),
		dataService.Post(webhookPath, Event),
	}
}

func Subscription(context dataService.Context) {
	NewRouterFromContext(context).Subscription(context.Response(), context.Request())
}

func Event(context dataService.Context) {
	NewRouterFromContext(context).Event(context.Response(), context.Request())
}

func NewRouterFromContext(context dataService.Context) *Router {
	router, _ := NewRouter(Dependencies{
		AuthClient: context.AuthClient(),
		OuraClient: context.OuraClient(),
		WorkClient: context.WorkClient(),
	})
	return router
}
