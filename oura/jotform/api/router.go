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
	submissionProcessor *jotform.SubmissionProcessor
}

func NewRouter(submissionProcessor *jotform.SubmissionProcessor) (*Router, error) {
	return &Router{
		submissionProcessor: submissionProcessor,
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

	submissionID := req.PostFormValue("submissionID")
	if len(submissionID) == 0 {
		responder.Error(http.StatusBadRequest, fmt.Errorf("missing submission ID"))
		return
	}

	err := r.submissionProcessor.ProcessSubmission(req.Context(), submissionID)
	if err != nil {
		log.LoggerFromContext(ctx).WithError(err).Error("unable to process submission")
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Empty(http.StatusOK)
}
