package v1

import (
	abbottServiceApiV1 "github.com/tidepool-org/platform-plugin-abbott/abbott/service/api/v1"

	dataService "github.com/tidepool-org/platform/data/service"
	serviceApi "github.com/tidepool-org/platform/service/api"
	workService "github.com/tidepool-org/platform/work/service"
)

func Routes() []dataService.Route {
	routes := []dataService.Route{
		dataService.Post("/v1/datasets/:dataSetId/data", DataSetsDataCreate, serviceApi.RequireAuth),
		dataService.Delete("/v1/datasets/:dataSetId", DataSetsDelete, serviceApi.RequireAuth),
		dataService.Put("/v1/datasets/:dataSetId", DataSetsUpdate, serviceApi.RequireAuth),
		dataService.Delete("/v1/users/:userId/data", UsersDataDelete, serviceApi.RequireAuth),
		dataService.Post("/v1/users/:userId/datasets", UsersDataSetsCreate, serviceApi.RequireAuth),
		dataService.Get("/v1/users/:userId/datasets", UsersDataSetsGet, serviceApi.RequireAuth),

		dataService.Post("/v1/data_sets/:dataSetId/data", DataSetsDataCreate, serviceApi.RequireAuth),
		dataService.Delete("/v1/data_sets/:dataSetId/data", DataSetsDataDelete, serviceApi.RequireAuth),
		dataService.Delete("/v1/data_sets/:dataSetId", DataSetsDelete, serviceApi.RequireAuth),
		dataService.Put("/v1/data_sets/:dataSetId", DataSetsUpdate, serviceApi.RequireAuth),
		dataService.Get("/v1/time", TimeGet),
		dataService.Post("/v1/users/:userId/data_sets", UsersDataSetsCreate, serviceApi.RequireAuth),

		dataService.Get("/v1/partners/:partner/sector", PartnersSector),
	}

	routes = append(routes, DataSetsRoutes()...)
	routes = append(routes, SourcesRoutes()...)
	routes = append(routes, SummaryRoutes()...)
	routes = append(routes, AlertsRoutes()...)
	routes = append(routes, abbottServiceApiV1.Routes()...)

	// TODO: optional inclusion of work load testing Routes
	routes = append(routes, workService.LoadTestRoutes()...)

	return routes
}
