package v1

import (
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
	userService "github.com/tidepool-org/platform/user/service"
)

func Authenticate(handler userService.HandlerFunc) userService.HandlerFunc {
	return func(context userService.Context) {
		if details := request.DetailsFromContext(context.Request().Context()); details == nil {
			context.RespondWithError(service.ErrorUnauthenticated())
			return
		}

		handler(context)
	}
}
