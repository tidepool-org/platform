package v1

import "github.com/tidepool-org/platform/dataservices/service"

func Routes() []service.Route {
	return []service.Route{
		service.MakeRoute("POST", "/v1/datasets/:datasetid/data", Authenticate(DatasetsDataCreate)),
		service.MakeRoute("DELETE", "/v1/datasets/:datasetid", Authenticate(DatasetsDelete)),
		service.MakeRoute("PUT", "/v1/datasets/:datasetid", Authenticate(DatasetsUpdate)),
		service.MakeRoute("DELETE", "/v1/users/:userid/data", Authenticate(UsersDataDelete)),
		service.MakeRoute("POST", "/v1/users/:userid/datasets", Authenticate(UsersDatasetsCreate)),
		service.MakeRoute("GET", "/v1/users/:userid/datasets", Authenticate(UsersDatasetsGet)),
	}
}
