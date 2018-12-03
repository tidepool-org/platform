package v1

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/blob"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/user"
)

type Provider interface {
	BlobClient() blob.Client
}

type Router struct {
	provider Provider
}

func NewRouter(provider Provider) (*Router, error) {
	if provider == nil {
		return nil, errors.New("provider is missing")
	}

	return &Router{
		provider: provider,
	}, nil
}

func (r *Router) Routes() []*rest.Route {
	return []*rest.Route{
		rest.Get("/v1/users/:userId/blobs", r.List),
		rest.Post("/v1/users/:userId/blobs", r.Create),
		rest.Get("/v1/blobs/:id", r.Get),
		rest.Get("/v1/blobs/:id/content", r.GetContent),
		rest.Delete("/v1/blobs/:id", r.Delete),
	}
}

func (r *Router) List(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	// FUTURE: Validate supplemental request headers

	userID, err := request.DecodeRequestPathParameter(req, "userId", user.IsValidID)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	filter := blob.NewFilter()
	pagination := page.NewPagination()
	if err = request.DecodeRequestQuery(req.Request, filter, pagination); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	result, err := r.provider.BlobClient().List(req.Context(), userID, filter, pagination)
	if responder.RespondIfError(err) {
		return
	}

	responder.Data(http.StatusOK, result)
}

func (r *Router) Create(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	// FUTURE: Validate supplemental request headers

	userID, err := request.DecodeRequestPathParameter(req, "userId", user.IsValidID)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	digestMD5, err := request.ParseDigestMD5Header(req.Header, "Digest")
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}
	mediaType, err := request.ParseMediaTypeHeader(req.Header, "Content-Type")
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	} else if mediaType == nil {
		responder.Error(http.StatusBadRequest, request.ErrorHeaderMissing("Content-Type"))
		return
	}

	create := blob.NewCreate()
	create.Body = req.Body
	create.DigestMD5 = digestMD5
	create.MediaType = mediaType

	result, err := r.provider.BlobClient().Create(req.Context(), userID, create)
	if err != nil {
		if errors.Code(err) == blob.ErrorCodeDigestsNotEqual {
			responder.Error(http.StatusBadRequest, err)
			return
		} else if responder.RespondIfError(err) {
			return
		}
	}

	responder.Data(http.StatusCreated, result)
}

func (r *Router) Get(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	// FUTURE: Validate supplemental request headers

	id, err := request.DecodeRequestPathParameter(req, "id", blob.IsValidID)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	result, err := r.provider.BlobClient().Get(req.Context(), id)
	if responder.RespondIfError(err) {
		return
	} else if result == nil {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithID(id))
		return
	}

	responder.Data(http.StatusOK, result)
}

func (r *Router) GetContent(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	// FUTURE: Validate supplemental request headers
	// FUTURE: Support range request headers, add range response headers

	id, err := request.DecodeRequestPathParameter(req, "id", blob.IsValidID)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	content, err := r.provider.BlobClient().GetContent(req.Context(), id)
	if responder.RespondIfError(err) {
		return
	} else if content == nil {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithID(id))
		return
	}

	defer content.Body.Close()

	mutators := []request.ResponseMutator{}
	if content.DigestMD5 != nil {
		mutators = append(mutators, request.NewHeaderMutator("Digest", fmt.Sprintf("MD5=%s", *content.DigestMD5)))
	}
	if content.MediaType != nil {
		mutators = append(mutators, request.NewHeaderMutator("Content-Type", *content.MediaType))
	}
	if content.Size != nil {
		mutators = append(mutators, request.NewHeaderMutator("Content-Length", strconv.Itoa(*content.Size)))
	}

	responder.Reader(http.StatusOK, content.Body, mutators...)
}

func (r *Router) Delete(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	// FUTURE: Validate supplemental request headers

	id, err := request.DecodeRequestPathParameter(req, "id", blob.IsValidID)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	condition := request.NewCondition()
	if err = request.DecodeRequestQuery(req.Request, condition); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	deleted, err := r.provider.BlobClient().Delete(req.Context(), id, condition)
	if responder.RespondIfError(err) {
		return
	} else if !deleted {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithIDAndOptionalRevision(id, condition.Revision))
		return
	}

	responder.Empty(http.StatusNoContent)
}
