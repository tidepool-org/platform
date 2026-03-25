package v1

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/request"
)

func (r *Router) Event(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	responder.Empty(http.StatusOK)
}
