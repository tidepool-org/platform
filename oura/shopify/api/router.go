package api

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/oura/shopify"
	"github.com/tidepool-org/platform/request"
)

type Router struct {
	processor *shopify.OrderProcessor
}

func NewRouter(processor *shopify.OrderProcessor) (*Router, error) {
	return &Router{
		processor: processor,
	}, nil
}

func (r *Router) Routes() []*rest.Route {
	return []*rest.Route{
		rest.Post("/v1/partners/shopify/fulfillment", r.HandleFulfillmentEvent),
		rest.Post("/v1/partners/shopify/orders/create", r.HandleOrdersCreateEvent),
	}
}

func (r *Router) HandleFulfillmentEvent(res rest.ResponseWriter, req *rest.Request) {
	ctx := req.Context()
	responder := request.MustNewResponder(res, req)

	event := shopify.FulfillmentEvent{}
	if err := request.DecodeRequestBody(req.Request, &event); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	if err := r.processor.ProcessFulfillment(ctx, event); err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Empty(http.StatusOK)
	return
}

func (r *Router) HandleOrdersCreateEvent(res rest.ResponseWriter, req *rest.Request) {
	ctx := req.Context()
	responder := request.MustNewResponder(res, req)

	event := shopify.OrdersCreateEvent{}
	if err := request.DecodeRequestBody(req.Request, &event); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	if err := r.processor.ProcessOrderCreate(ctx, event); err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Empty(http.StatusOK)
	return
}
