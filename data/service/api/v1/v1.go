package v1

import "github.com/tidepool-org/platform/data/service"

func Routes() []service.Route {
	return []service.Route{
		service.MakeRoute("POST", "/v1/datasets/:dataset_id/data", Authenticate(DatasetsDataCreate)),
		service.MakeRoute("DELETE", "/v1/datasets/:dataset_id", Authenticate(DatasetsDelete)),
		service.MakeRoute("PUT", "/v1/datasets/:dataset_id", Authenticate(DatasetsUpdate)),
		service.MakeRoute("DELETE", "/v1/users/:user_id/data", Authenticate(UsersDataDelete)),
		service.MakeRoute("POST", "/v1/users/:user_id/datasets", Authenticate(UsersDatasetsCreate)),
		service.MakeRoute("GET", "/v1/users/:user_id/datasets", Authenticate(UsersDatasetsGet)),
	}
}
