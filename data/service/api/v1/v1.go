package v1

import "github.com/tidepool-org/platform/data/service"

func Routes() []service.Route {
	routes := []service.Route{
		service.MakeRoute("POST", "/v1/datasets/:dataSetId/data", EnforceAuthentication(DataSetsDataCreate)),
		service.MakeRoute("DELETE", "/v1/datasets/:dataSetId", EnforceAuthentication(DataSetsDelete)),
		service.MakeRoute("PUT", "/v1/datasets/:dataSetId", EnforceAuthentication(DataSetsUpdate)),
		service.MakeRoute("DELETE", "/v1/users/:userId/data", EnforceAuthentication(UsersDataDelete)),
		service.MakeRoute("POST", "/v1/users/:userId/datasets", EnforceAuthentication(UsersDataSetsCreate)),
		service.MakeRoute("GET", "/v1/users/:userId/datasets", EnforceAuthentication(UsersDataSetsGet)),

		service.MakeRoute("POST", "/v1/data_sets/:dataSetId/data", EnforceAuthentication(DataSetsDataCreate)),
		service.MakeRoute("DELETE", "/v1/data_sets/:dataSetId/data", EnforceAuthentication(DataSetsDataDelete)),
		service.MakeRoute("DELETE", "/v1/data_sets/:dataSetId", EnforceAuthentication(DataSetsDelete)),
		service.MakeRoute("PUT", "/v1/data_sets/:dataSetId", EnforceAuthentication(DataSetsUpdate)),
		service.MakeRoute("GET", "/v1/time", TimeGet),
		service.MakeRoute("POST", "/v1/users/:userId/data_sets", EnforceAuthentication(UsersDataSetsCreate)),
	}

	routes = append(routes, DataSetsRoutes()...)
	routes = append(routes, SourcesRoutes()...)
	routes = append(routes, SummaryRoutes()...)
	routes = append(routes, AlertsRoutes()...)

	return routes
}
