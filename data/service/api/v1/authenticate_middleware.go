package v1

import (
	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/service"
)

func Authenticate(handler dataService.HandlerFunc) dataService.HandlerFunc {
	return func(context dataService.Context) {
		if authDetails := context.AuthDetails(); authDetails == nil {
			context.RespondWithError(service.ErrorUnauthenticated())
			return
		}

		handler(context)
	}
}
