package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/tidepool-org/platform/permission"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth"

	"github.com/tidepool-org/platform/blob"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service/api"
	"github.com/tidepool-org/platform/user"
)

type Provider interface {
	AuthClient() auth.Client
	BlobClient() blob.Client
}

type Router struct {
	Provider
}

func NewRouter(provider Provider) (*Router, error) {
	if provider == nil {
		return nil, errors.New("provider is missing")
	}

	return &Router{
		Provider: provider,
	}, nil
}

func (r *Router) Routes() []*rest.Route {
	return []*rest.Route{
		rest.Get("/v1/users/:userId/blobs", r.List),
		rest.Post("/v1/users/:userId/blobs", r.Create),
		rest.Delete("/v1/users/:userId/blobs", r.DeleteAll),

		rest.Post("/v1/users/:userId/device_logs", r.CreateDeviceLogs),
		rest.Get("/v1/users/:userId/device_logs", api.RequireMembership(r.permissionsClient, "userId", r.ListDeviceLogs)),

		rest.Get("/v1/blobs/:id", r.Get),
		rest.Get("/v1/blobs/:id/content", r.GetContent),
		rest.Get("/v1/device_logs/:id", r.GetDeviceLogsBlob),
		rest.Get("/v1/device_logs/:id/content", api.RequireAuth(r.GetDeviceLogsContent)),
		rest.Delete("/v1/blobs/:id", r.Delete),
	}
}

func (r *Router) permissionsClient() permission.Client {
	return r.AuthClient()
}

func (r *Router) List(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

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

	err = r.AuthClient().EnsureAuthorizedService(req.Context())
	if responder.RespondIfError(err) {
		return
	}

	result, err := r.Provider.BlobClient().List(req.Context(), userID, filter, pagination)
	if responder.RespondIfError(err) {
		return
	}

	responder.Data(http.StatusOK, result)
}

func (r *Router) Create(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

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

	content := blob.NewContent()
	content.Body = req.Body
	content.DigestMD5 = digestMD5
	content.MediaType = mediaType

	_, err = r.AuthClient().EnsureAuthorizedUser(req.Context(), userID, permission.Write)
	if responder.RespondIfError(err) {
		return
	}

	result, err := r.Provider.BlobClient().Create(req.Context(), userID, content)
	if err != nil {
		if errors.Code(err) == request.ErrorCodeDigestsNotEqual {
			responder.Error(http.StatusBadRequest, err)
			return
		} else if responder.RespondIfError(err) {
			return
		}
	}

	responder.Data(http.StatusCreated, result)
}

func (r *Router) CreateDeviceLogs(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	userID, err := request.DecodeRequestPathParameter(req, "userId", user.IsValidID)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	digestMD5, err := request.ParseDigestMD5Header(req.Header, "Digest")
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	} else if digestMD5 == nil {
		responder.Error(http.StatusBadRequest, request.ErrorHeaderMissing("Digest"))
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

	startAtTime, err := request.ParseTimeHeader(req.Header, "X-Logs-Start-At-Time", time.RFC3339)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	} else if startAtTime == nil {
		responder.Error(http.StatusBadRequest, request.ErrorHeaderMissing("X-Logs-Start-At-Time"))
		return
	}
	endAtTime, err := request.ParseTimeHeader(req.Header, "X-Logs-End-At-Time", time.RFC3339)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	} else if endAtTime == nil {
		responder.Error(http.StatusBadRequest, request.ErrorHeaderMissing("X-Logs-End-At-Time"))
		return
	}

	content := blob.NewDeviceLogsContent()
	content.Body = req.Body
	content.DigestMD5 = digestMD5
	content.MediaType = mediaType
	content.StartAt = startAtTime
	content.EndAt = endAtTime

	_, err = r.AuthClient().EnsureAuthorizedUser(req.Context(), userID, permission.Write)
	if responder.RespondIfError(err) {
		return
	}

	result, err := r.Provider.BlobClient().CreateDeviceLogs(req.Context(), userID, content)
	if err != nil {
		if errors.Code(err) == request.ErrorCodeDigestsNotEqual {
			responder.Error(http.StatusBadRequest, err)
			return
		} else if responder.RespondIfError(err) {
			return
		}
	}
	responder.Data(http.StatusCreated, result)
}

func (r *Router) ListDeviceLogs(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	userID, err := request.DecodeRequestPathParameter(req, "userId", user.IsValidID)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}
	filter := blob.NewDeviceLogsFilter()
	pagination := page.NewPagination()
	if err := request.DecodeRequestQuery(req.Request, filter, pagination); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}
	logs, err := r.Provider.BlobClient().ListDeviceLogs(req.Context(), userID, filter, pagination)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}
	responder.Data(http.StatusOK, logs)
}

func (r *Router) GetDeviceLogsBlob(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	blobClient := r.Provider.BlobClient()

	deviceLogID, err := request.DecodeRequestPathParameter(req, "id", blob.IsValidID)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	deviceLogMetadata, err := blobClient.GetDeviceLogsBlob(req.Context(), deviceLogID)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}
	if deviceLogMetadata == nil {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithID(deviceLogID))
		return
	}

	allowed, err := api.CheckMembership(req, r.AuthClient(), *deviceLogMetadata.UserID)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}
	if !allowed {
		request.MustNewResponder(res, req).Error(http.StatusForbidden, request.ErrorUnauthorized())
		return
	}
	responder.Data(http.StatusOK, deviceLogMetadata)
}

func (r *Router) GetDeviceLogsContent(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	blobClient := r.Provider.BlobClient()

	deviceLogID, err := request.DecodeRequestPathParameter(req, "id", blob.IsValidID)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	deviceLogMetadata, err := blobClient.GetDeviceLogsBlob(req.Context(), deviceLogID)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}
	if deviceLogMetadata == nil {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithID(deviceLogID))
		return
	}

	allowed, err := api.CheckMembership(req, r.AuthClient(), *deviceLogMetadata.UserID)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}
	if !allowed {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithID(deviceLogID))
		return
	}

	content, err := blobClient.GetDeviceLogsContent(req.Context(), *deviceLogMetadata.ID)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}
	if content == nil || content.Body == nil {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithID(deviceLogID))
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
	if content.StartAt != nil {
		mutators = append(mutators, request.NewHeaderMutator("Start-At", content.StartAt.Format(time.RFC3339Nano)))
	}
	if content.EndAt != nil {
		mutators = append(mutators, request.NewHeaderMutator("End-At", content.EndAt.Format(time.RFC3339Nano)))
	}

	responder.Reader(http.StatusOK, content.Body, mutators...)
}

func (r *Router) DeleteAll(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	userID, err := request.DecodeRequestPathParameter(req, "userId", user.IsValidID)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	err = r.AuthClient().EnsureAuthorizedService(req.Context())
	if responder.RespondIfError(err) {
		return
	}

	if responder.RespondIfError(r.Provider.BlobClient().DeleteAll(req.Context(), userID)) {
		return
	}

	responder.Empty(http.StatusNoContent)
}

func (r *Router) Get(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	id, err := request.DecodeRequestPathParameter(req, "id", blob.IsValidID)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	err = r.AuthClient().EnsureAuthorizedService(req.Context())
	if responder.RespondIfError(err) {
		return
	}

	result, err := r.Provider.BlobClient().Get(req.Context(), id)
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
	// FUTURE: Support range request headers, add range response headers

	id, err := request.DecodeRequestPathParameter(req, "id", blob.IsValidID)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	err = r.AuthClient().EnsureAuthorizedService(req.Context())
	if responder.RespondIfError(err) {
		return
	}

	content, err := r.Provider.BlobClient().GetContent(req.Context(), id)
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

	responder.Reader(http.StatusOK, content.Body, mutators...)
}

func (r *Router) Delete(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

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

	err = r.AuthClient().EnsureAuthorizedService(req.Context())
	if responder.RespondIfError(err) {
		return
	}

	deleted, err := r.Provider.BlobClient().Delete(req.Context(), id, condition)
	if responder.RespondIfError(err) {
		return
	} else if !deleted {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithIDAndOptionalRevision(id, condition.Revision))
		return
	}

	responder.Empty(http.StatusNoContent)
}
