package v1

import "github.com/tidepool-org/platform/user/service"

func Routes() []service.Route {
	return []service.Route{
		service.MakeRoute("DELETE", "/v1/users/:userId", Authenticate(UsersDelete)),
	}
}
