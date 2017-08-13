package api

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/task/service/context"
)

func (r *Router) GetStatus(response rest.ResponseWriter, request *rest.Request) {
	ctx := context.MustNew(r, response, request)

	ctx.RespondWithStatusAndData(http.StatusOK, ctx.Status())
}
