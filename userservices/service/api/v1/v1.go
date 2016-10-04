package v1

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import "github.com/tidepool-org/platform/userservices/service"

func Routes() []service.Route {
	return []service.Route{
		service.MakeRoute("DELETE", "/v1/users/:userid", Authenticate(UsersDelete)),
	}
}
