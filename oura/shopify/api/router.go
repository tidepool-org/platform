package api

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/oura/shopify"
	"github.com/tidepool-org/platform/request"
)

type Router struct {
	fulfillmentCreatedEventProcessor *shopify.FulfillmentCreatedEventProcessor
}

func NewRouter() (*Router, error) {
	return &Router{}, nil
}

func (r *Router) Routes() []*rest.Route {
	return []*rest.Route{
		rest.Post("/v1/partners/shopify/fulfillment_event/created", r.HandleFulfillmentEventCreated),
	}
}

func (r *Router) HandleFulfillmentEventCreated(res rest.ResponseWriter, req *rest.Request) {
	ctx := req.Context()
	responder := request.MustNewResponder(res, req)

	event := shopify.FulfillmentEventCreated{}
	if err := request.DecodeRequestBody(req.Request, event); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	if err := r.fulfillmentCreatedEventProcessor.Process(ctx, event); err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Empty(http.StatusOK)
	return
}
