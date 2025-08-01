package v1

import (
	abbottServiceApiV1 "github.com/tidepool-org/platform-plugin-abbott/abbott/service/api/v1"

	"github.com/tidepool-org/platform/data/service"
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

		service.Post("/v1/partners/twiist/data/:tidepoolLinkId", NewTwiistDataCreateHandler(DataSetsDataCreate), api.RequireUser),
	}

	routes = append(routes, DataSetsRoutes()...)
	routes = append(routes, SourcesRoutes()...)
	routes = append(routes, SummaryRoutes()...)
	routes = append(routes, AlertsRoutes()...)
	routes = append(routes, abbottServiceApiV1.Routes()...)

	return routes
}
