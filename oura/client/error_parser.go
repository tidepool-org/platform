package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/request"
)

const ErrorResponseBodyLimit = 1024 * 1024

type ErrorResponseParser struct{}

func (e *ErrorResponseParser) ParseErrorResponse(ctx context.Context, res *http.Response, req *http.Request) error {

	// Capture full response body for 422 Unprocessable Entity which indicates a request body validation error
	// Equivalent to our use of 400 Bad Request
	if res.StatusCode == http.StatusUnprocessableEntity {
		var errorResponse ErrorResponse

		if bites, err := io.ReadAll(io.LimitReader(res.Body, ErrorResponseBodyLimit+1)); err != nil || len(bites) > ErrorResponseBodyLimit {
			log.LoggerFromContext(ctx).WithError(err).Error("unable to read error response body")
			return request.ErrorBadRequest()
		} else if err = json.NewDecoder(bytes.NewReader(bites)).Decode(&errorResponse); err != nil {
			log.LoggerFromContext(ctx).WithError(err).Error("unable to decode error response body")
			return request.ErrorBadRequest()
		}

		return errors.WithMeta(request.ErrorBadRequest(), errorResponse)
	}

	// Let caller handle
	return nil
}

type ErrorResponse struct {
	Detail []ErrorResponseDetail `json:"detail,omitempty" bson:"detail,omitempty"`
}

type ErrorResponseDetail struct {
	Location []string `json:"loc,omitempty" bson:"loc,omitempty"`
	Message  string   `json:"msg,omitempty" bson:"msg,omitempty"`
	Type     string   `json:"type,omitempty" bson:"type,omitempty"`
}
