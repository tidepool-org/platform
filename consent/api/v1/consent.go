package v1

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/consent"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	structValidator "github.com/tidepool-org/platform/structure/validator"
)

func (r *Router) ListConsents(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	filter := consent.NewConsentFilter()
	pagination := page.NewPagination()
	if err := request.DecodeRequestQuery(req.Request, filter, pagination); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	consents, err := r.service.ListConsents(req.Context(), filter, pagination)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, consents)
}

func (r *Router) GetConsentByType(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	typ := req.PathParam("type")
	if typ == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("type"))
		return
	}

	filter := consent.NewConsentFilter()
	filter.Type = consent.NewConsentType(&typ)
	filter.Latest = pointer.FromAny(true)

	consents, err := r.service.ListConsents(req.Context(), filter, nil)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, consents)
}

func (r *Router) GetConsentVersions(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	typ := req.PathParam("type")
	if typ == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("type"))
		return
	}

	filter := consent.NewConsentFilter()
	filter.Type = consent.NewConsentType(&typ)
	if err := structValidator.New(log.LoggerFromContext(req.Context())).Validate(filter); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	pagination := page.NewPagination()
	if err := request.DecodeRequestQuery(req.Request, pagination); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	consents, err := r.service.ListConsents(req.Context(), filter, pagination)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, consents)
}
