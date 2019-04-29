package v1

import "github.com/tidepool-org/platform/data/service"

func Routes() []service.Route {
	routes := []service.Route{
		service.MakeRoute("POST", "/v1/datasets/:dataSetId/data", Authenticate(DataSetsDataCreate)),
		service.MakeRoute("DELETE", "/v1/datasets/:dataSetId", Authenticate(DataSetsDelete)),
		service.MakeRoute("PUT", "/v1/datasets/:dataSetId", Authenticate(DataSetsUpdate)),
		service.MakeRoute("DELETE", "/v1/users/:userId/data", Authenticate(UsersDataDelete)),
		service.MakeRoute("POST", "/v1/users/:userId/datasets", Authenticate(UsersDataSetsCreate)),
		service.MakeRoute("GET", "/v1/users/:userId/datasets", Authenticate(UsersDataSetsGet)),

		service.MakeRoute("POST", "/v1/data_sets/:dataSetId/data", Authenticate(DataSetsDataCreate)),
		service.MakeRoute("DELETE", "/v1/data_sets/:dataSetId/data", Authenticate(DataSetsDataDelete)),
		service.MakeRoute("DELETE", "/v1/data_sets/:dataSetId", Authenticate(DataSetsDelete)),
		service.MakeRoute("PUT", "/v1/data_sets/:dataSetId", Authenticate(DataSetsUpdate)),
		service.MakeRoute("GET", "/v1/time", TimeGet),
		service.MakeRoute("POST", "/v1/users/:userId/data_sets", Authenticate(UsersDataSetsCreate)),
	}
	return append(append(routes, DataSetsRoutes()...), SourcesRoutes()...)
}
