package v1

import (
	"net/http"

	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/request"
)

func Subscription(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()
	responder := request.MustNewResponder(res, req)

	responder.String(http.StatusOK, "OK")
}
