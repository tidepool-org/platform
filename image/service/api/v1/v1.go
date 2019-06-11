package v1

import (
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/image"
	imageMultipart "github.com/tidepool-org/platform/image/multipart"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/user"
)

type Provider interface {
	ImageClient() image.Client
	ImageMultipartFormDecoder() imageMultipart.FormDecoder
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
		rest.Get("/v1/users/:userId/images", r.List),
		rest.Post("/v1/users/:userId/images", r.Create),
		rest.Post("/v1/users/:userId/images/metadata", r.CreateWithMetadata),
		rest.Post("/v1/users/:userId/images/content/:contentIntent", r.CreateWithContent),
		rest.Delete("/v1/users/:userId/images", r.DeleteAll),
		rest.Get("/v1/images/:id", r.Get),
		rest.Get("/v1/images/:id/metadata", r.GetMetadata),
		rest.Get("/v1/images/:id/content", r.GetContent),
		rest.Get("/v1/images/:id/content/*suffix", r.GetContent),
		rest.Get("/v1/images/:id/rendition/*suffix", r.GetRenditionContent),
		rest.Put("/v1/images/:id/metadata", r.PutMetadata),
		rest.Put("/v1/images/:id/content/:contentIntent", r.PutContent),
		rest.Delete("/v1/images/:id", r.Delete),
	}
}

func (r *Router) List(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	userID, err := request.DecodeRequestPathParameter(req, "userId", user.IsValidID)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	filter := image.NewFilter()
	pagination := page.NewPagination()
	if err = request.DecodeRequestQuery(req.Request, filter, pagination); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	result, err := r.provider.ImageClient().List(req.Context(), userID, filter, pagination)
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

	header, headerErr := request.ParseSingletonHeader(req.Header, "Content-Type")
	if headerErr != nil {
		responder.Error(http.StatusBadRequest, headerErr)
		return
	} else if header == nil {
		responder.Error(http.StatusBadRequest, request.ErrorHeaderMissing("Content-Type"))
		return
	} else if mediaType, parameters, mediaTypeErr := mime.ParseMediaType(*header); mediaTypeErr != nil {
		responder.Error(http.StatusBadRequest, request.ErrorHeaderInvalid("Content-Type"))
		return
	} else if !strings.HasPrefix(mediaType, "multipart/") {
		responder.Error(http.StatusBadRequest, request.ErrorMediaTypeNotSupported(*header))
		return
	} else if boundary := parameters["boundary"]; boundary == "" {
		responder.Error(http.StatusBadRequest, request.ErrorHeaderInvalid("Content-Type"))
		return
	}

	metadata, contentIntent, content, err := r.provider.ImageMultipartFormDecoder().DecodeForm(req.Body, *header)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	result, err := r.provider.ImageClient().Create(req.Context(), userID, metadata, contentIntent, content)
	if err != nil {
		switch errors.Code(err) {
		case request.ErrorCodeDigestsNotEqual, image.ErrorCodeImageContentIntentUnexpected, image.ErrorCodeImageMalformed:
			responder.Error(http.StatusBadRequest, err)
			return
		}
		if responder.RespondIfError(err) {
			return
		}
	}

	responder.Data(http.StatusCreated, result)
}

func (r *Router) CreateWithMetadata(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	userID, err := request.DecodeRequestPathParameter(req, "userId", user.IsValidID)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	metadata := image.NewMetadata()
	if err = request.DecodeRequestBody(req.Request, metadata); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	result, err := r.provider.ImageClient().CreateWithMetadata(req.Context(), userID, metadata)
	if responder.RespondIfError(err) {
		return
	}

	responder.Data(http.StatusCreated, result)
}

func (r *Router) CreateWithContent(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	userID, err := request.DecodeRequestPathParameter(req, "userId", user.IsValidID)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}
	contentIntent, err := request.DecodeRequestPathParameter(req, "contentIntent", image.IsValidContentIntent)
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
	} else if err = image.ValidateMediaType(*mediaType); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	content := image.NewContent()
	content.Body = req.Body
	content.DigestMD5 = digestMD5
	content.MediaType = mediaType

	result, err := r.provider.ImageClient().CreateWithContent(req.Context(), userID, contentIntent, content)
	if err != nil {
		switch errors.Code(err) {
		case request.ErrorCodeDigestsNotEqual, image.ErrorCodeImageContentIntentUnexpected, image.ErrorCodeImageMalformed:
			responder.Error(http.StatusBadRequest, err)
			return
		}
		if responder.RespondIfError(err) {
			return
		}
	}

	responder.Data(http.StatusCreated, result)
}

func (r *Router) DeleteAll(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	userID, err := request.DecodeRequestPathParameter(req, "userId", user.IsValidID)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	if responder.RespondIfError(r.provider.ImageClient().DeleteAll(req.Context(), userID)) {
		return
	}

	responder.Empty(http.StatusNoContent)
}

func (r *Router) Get(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	id, err := request.DecodeRequestPathParameter(req, "id", image.IsValidID)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	result, err := r.provider.ImageClient().Get(req.Context(), id)
	if responder.RespondIfError(err) {
		return
	} else if result == nil {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithID(id))
		return
	}

	responder.Data(http.StatusOK, result)
}

func (r *Router) GetMetadata(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	id, err := request.DecodeRequestPathParameter(req, "id", image.IsValidID)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	result, err := r.provider.ImageClient().GetMetadata(req.Context(), id)
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

	id, err := request.DecodeRequestPathParameter(req, "id", image.IsValidID)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}
	suffix, _ := request.DecodeOptionalRequestPathParameter(req, "suffix", nil)

	var mediaType *string
	if suffix != nil {
		suffixParts := strings.Split(*suffix, "/")
		extensionParts := strings.Split(suffixParts[len(suffixParts)-1], ".")
		if len(extensionParts) > 1 {
			extension := extensionParts[len(extensionParts)-1]
			if err = image.ValidateExtension(extension); err != nil {
				responder.Error(http.StatusBadRequest, err)
				return
			}
			value, _ := image.MediaTypeFromExtension(extension)
			mediaType = pointer.FromString(value)
		}
	}

	content, err := r.provider.ImageClient().GetContent(req.Context(), id, mediaType)
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

func (r *Router) GetRenditionContent(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	id, err := request.DecodeRequestPathParameter(req, "id", image.IsValidID)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}
	suffix, _ := request.DecodeRequestPathParameter(req, "suffix", nil)

	suffixParts := strings.Split(suffix, "/")

	renditionString, err := url.PathUnescape(suffixParts[0])
	if err != nil {
		renditionString = suffixParts[0]
	}

	if len(suffixParts) > 1 {
		extensionParts := strings.Split(suffixParts[len(suffixParts)-1], ".")
		extension := extensionParts[len(extensionParts)-1]
		renditionString = fmt.Sprintf("%s.%s", renditionString, extension)
	}

	rendition, err := image.ParseRenditionFromString(renditionString)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	content, err := r.provider.ImageClient().GetRenditionContent(req.Context(), id, rendition)
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

func (r *Router) PutMetadata(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	id, err := request.DecodeRequestPathParameter(req, "id", image.IsValidID)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	condition := request.NewCondition()
	if err = request.DecodeRequestQuery(req.Request, condition); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	metadata := image.NewMetadata()
	if err = request.DecodeRequestBody(req.Request, metadata); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	result, err := r.provider.ImageClient().PutMetadata(req.Context(), id, condition, metadata)
	if responder.RespondIfError(err) {
		return
	} else if result == nil {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithIDAndOptionalRevision(id, condition.Revision))
		return
	}

	responder.Data(http.StatusOK, result)
}

func (r *Router) PutContent(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	id, err := request.DecodeRequestPathParameter(req, "id", image.IsValidID)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}
	contentIntent, err := request.DecodeRequestPathParameter(req, "contentIntent", image.IsValidContentIntent)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	condition := request.NewCondition()
	if err = request.DecodeRequestQuery(req.Request, condition); err != nil {
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
	} else if err = image.ValidateMediaType(*mediaType); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	content := image.NewContent()
	content.Body = req.Body
	content.DigestMD5 = digestMD5
	content.MediaType = mediaType

	result, err := r.provider.ImageClient().PutContent(req.Context(), id, condition, contentIntent, content)
	if err != nil {
		switch errors.Code(err) {
		case request.ErrorCodeDigestsNotEqual, image.ErrorCodeImageContentIntentUnexpected, image.ErrorCodeImageMalformed:
			responder.Error(http.StatusBadRequest, err)
			return
		}
		if responder.RespondIfError(err) {
			return
		}
	} else if result == nil {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithIDAndOptionalRevision(id, condition.Revision))
		return
	}

	responder.Data(http.StatusOK, result)
}

func (r *Router) Delete(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	id, err := request.DecodeRequestPathParameter(req, "id", image.IsValidID)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	condition := request.NewCondition()
	if err = request.DecodeRequestQuery(req.Request, condition); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	deleted, err := r.provider.ImageClient().Delete(req.Context(), id, condition)
	if responder.RespondIfError(err) {
		return
	} else if !deleted {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithIDAndOptionalRevision(id, condition.Revision))
		return
	}

	responder.Empty(http.StatusNoContent)
}
