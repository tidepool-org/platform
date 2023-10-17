package v1

import (
	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
)

// EnforceAuthentication responds with an error if AuthDetails are absent.
//
// In essence, this function indicates that a request must be authenticated to
// be processed. Any unauthenticated requests will trigger an error response.
//
// EnforceAuthentication works by checking for the existence of an AuthDetails
// sentinel value, which implies an important assumption:
//
//	The existence of AuthDetails in the request's Context indicates that the
//	request has already been properly authenticated.
//
// The function that performs the actual authentication is in the
// service/middleware package. As long as no other code adds an AuthDetails
// value to the request's Context (when the request isn't properly
// authenticated) then things should be good.
func EnforceAuthentication(handler dataService.HandlerFunc) dataService.HandlerFunc {
	return func(context dataService.Context) {
		if authDetails := request.GetAuthDetails(context.Request().Context()); authDetails == nil {
			context.RespondWithError(service.ErrorUnauthenticated())
			return
		}

		handler(context)
	}
}
