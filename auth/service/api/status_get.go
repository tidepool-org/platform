package api

import (
	"net/http"

	"github.com/mdblp/go-json-rest/rest"

	"github.com/tidepool-org/platform/request"
)

func (r *Router) StatusGet(res rest.ResponseWriter, req *rest.Request) {
	request.MustNewResponder(res, req).Data(http.StatusOK, r.Status())
}
