package api

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/oura/shopify"
	"github.com/tidepool-org/platform/request"
)

type Router struct {
	fulfillmentEventProcessor *shopify.FulfillmentEventProcessor
}

func NewRouter(fulfillmentEventProcessor *shopify.FulfillmentEventProcessor) (*Router, error) {
	return &Router{
		fulfillmentEventProcessor: fulfillmentEventProcessor,
	}, nil
}

func (r *Router) Routes() []*rest.Route {
	return []*rest.Route{
		rest.Post("/v1/partners/shopify/fulfillment", r.HandleFulfillmentEvent),
	}
}

func (r *Router) HandleFulfillmentEvent(res rest.ResponseWriter, req *rest.Request) {
	ctx := req.Context()
	responder := request.MustNewResponder(res, req)

	event := shopify.FulfillmentEvent{}
	if err := request.DecodeRequestBody(req.Request, event); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	if err := r.fulfillmentEventProcessor.Process(ctx, event); err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Empty(http.StatusOK)
	return
}
