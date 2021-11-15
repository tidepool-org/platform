package api

import (
	"net/http"

	"github.com/mdblp/go-json-rest/rest"

	"github.com/tidepool-org/platform/notification/service/context"
)

func (r *Router) StatusGet(response rest.ResponseWriter, request *rest.Request) {
	ctx := context.MustNew(r, response, request)

	ctx.RespondWithStatusAndData(http.StatusOK, ctx.Status())
}
