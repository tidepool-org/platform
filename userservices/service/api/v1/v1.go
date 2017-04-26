package v1

import "github.com/tidepool-org/platform/userservices/service"

func Routes() []service.Route {
	return []service.Route{
		service.MakeRoute("DELETE", "/v1/users/:userid", Authenticate(UsersDelete)),
	}
}
