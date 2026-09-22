package v1

import (
	abbottServiceApiV1 "github.com/tidepool-org/platform-plugin-abbott/abbott/service/api/v1"
	tandemServiceApiV1 "github.com/tidepool-org/platform-plugin-tandem/tandem/service/api/v1"

	"github.com/tidepool-org/platform/data/service"
	ouraServiceApiV1 "github.com/tidepool-org/platform/oura/service/api/v1"
	"github.com/tidepool-org/platform/service/api"
)

func Routes() []service.Route {
	routes := []service.Route{
		service.Post("/v1/datasets/:dataSetId/data", DataSetsDataCreate, api.RequireAuth),
		service.Delete("/v1/datasets/:dataSetId", DataSetsDelete, api.RequireAuth),
		service.Put("/v1/datasets/:dataSetId", DataSetsUpdate, api.RequireAuth),
		service.Delete("/v1/users/:userId/data", UsersDataDelete, api.RequireAuth),
		service.Post("/v1/users/:userId/datasets", UsersDataSetsCreate, api.RequireAuth),
		service.Get("/v1/users/:userId/datasets", UsersDataSetsGet, api.RequireAuth),

		service.Post("/v1/data_sets/:dataSetId/data", DataSetsDataCreate, api.RequireAuth),
		service.Delete("/v1/data_sets/:dataSetId/data", DataSetsDataDelete, api.RequireAuth),
		service.Delete("/v1/data_sets/:dataSetId", DataSetsDelete, api.RequireAuth),
		service.Put("/v1/data_sets/:dataSetId", DataSetsUpdate, api.RequireAuth),
		service.Get("/v1/time", TimeGet),
		service.Post("/v1/users/:userId/data_sets", UsersDataSetsCreate, api.RequireAuth),

		service.Get("/v1/partners/:partner/sector/:environment", PartnersSector),
		service.Get("/v1/partners/:partner/sector", PartnersSector), // DEPRECATED: Remove once Abbott sandbox client uses environment in path

		service.Post("/v1/partners/twiist/data/:tidepoolLinkId", NewTwiistDataCreateHandler(DataSetsDataCreate), api.RequireAuth),
	}

	routes = append(routes, MetricsRoutes()...)
	routes = append(routes, DataSetsRoutes()...)
	routes = append(routes, SourcesRoutes()...)
	routes = append(routes, SummaryRoutes()...)
	routes = append(routes, AlertsRoutes()...)
	routes = append(routes, NotificationsRoutes()...)
	routes = append(routes, abbottServiceApiV1.Routes()...)
	routes = append(routes, ouraServiceApiV1.Routes()...)
	routes = append(routes, TandemRoutes()...)

	return routes
}

// TandemRoutes adapts the Tandem plugin routes, which are declared without this package so the plugin does
// not depend on the other plugins through it.
func TandemRoutes() []service.Route {
	var routes []service.Route
	for _, route := range tandemServiceApiV1.Routes() {
		handler := func(context service.Context) { route.Handler(context) }
		routes = append(routes, service.MakeRoute(route.Method, route.Path, handler, route.Middleware...))
	}
	return routes
}
