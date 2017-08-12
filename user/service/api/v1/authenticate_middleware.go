package v1

import (
	"github.com/tidepool-org/platform/service"
	userService "github.com/tidepool-org/platform/user/service"
)

func Authenticate(handler userService.HandlerFunc) userService.HandlerFunc {
	return func(context userService.Context) {
		if authDetails := context.AuthDetails(); authDetails == nil {
			context.RespondWithError(service.ErrorUnauthenticated())
			return
		}

		handler(context)
	}
}
