package v1

import "github.com/tidepool-org/platform/data/service"

func Routes() []service.Route {
	routes := []service.Route{
		service.MakeRoute("POST", "/v1/datasets/:dataSetId/data", Authenticate(DatasetsDataCreate)),
		service.MakeRoute("DELETE", "/v1/datasets/:dataSetId", Authenticate(DatasetsDelete)),
		service.MakeRoute("PUT", "/v1/datasets/:dataSetId", Authenticate(DatasetsUpdate)),
		service.MakeRoute("DELETE", "/v1/users/:userId/data", Authenticate(UsersDataDelete)),

		service.MakeRoute("POST", "/v1/data_sets/:dataSetId/data", Authenticate(DatasetsDataCreate)),
		service.MakeRoute("DELETE", "/v1/data_sets/:dataSetId", Authenticate(DatasetsDelete)),
		service.MakeRoute("PUT", "/v1/data_sets/:dataSetId", Authenticate(DatasetsUpdate)),
		service.MakeRoute("POST", "/v1/users/:userId/data_sets", Authenticate(UsersDatasetsCreate)),
		service.MakeRoute("GET", "/v1/users/:userId/data_sets", Authenticate(UsersDatasetsGet)),
	}
	return append(routes, DataSourcesRoutes()...)
}
