package api

import (
	"fmt"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/oura/jotform"
	"github.com/tidepool-org/platform/request"
)

const (
	multipartMaxMemory = 1000000 // 1MB
)

type Router struct {
	webhookProcessor *jotform.WebhookProcessor
}

func NewRouter(webhookProcessor *jotform.WebhookProcessor) (*Router, error) {
	return &Router{
		webhookProcessor: webhookProcessor,
	}, nil
}

func (r *Router) Routes() []*rest.Route {
	return []*rest.Route{
		rest.Post("/v1/partners/jotform/submission", r.HandleJotformSubmission),
	}
}

func (r *Router) HandleJotformSubmission(res rest.ResponseWriter, req *rest.Request) {
	ctx := req.Context()
	responder := request.MustNewResponder(res, req)

	if err := req.ParseMultipartForm(multipartMaxMemory); err != nil {
		responder.Error(http.StatusInternalServerError, fmt.Errorf("unable to parse form data"))
		return
	}

	values, ok := req.MultipartForm.Value["submissionID"]
	if !ok || len(values) == 0 || len(values[0]) == 0 {
		responder.Error(http.StatusBadRequest, fmt.Errorf("missing submission ID"))
		return
	}

	err := r.webhookProcessor.ProcessSubmission(req.Context(), values[0])
	if err != nil {
		log.LoggerFromContext(ctx).WithError(err).Error("unable to process submission")
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Empty(http.StatusOK)
}
