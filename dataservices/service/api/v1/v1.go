package v1

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import "github.com/tidepool-org/platform/dataservices/service"

func Routes() []service.Route {
	return []service.Route{
		service.MakeRoute("POST", "/api/v1/users/:userid/datasets", Authenticate(UsersDatasetsCreate)),
		service.MakeRoute("PUT", "/api/v1/datasets/:datasetid", Authenticate(DatasetsUpdate)),
		service.MakeRoute("POST", "/api/v1/datasets/:datasetid/data", Authenticate(DatasetsDataCreate)),
	}
}
