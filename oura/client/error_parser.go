package client

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/request"
)

type errorResponseParser struct{}

func NewErrorResponseParser() client.ErrorResponseParser {
	return &errorResponseParser{}
}

func (e *errorResponseParser) ParseErrorResponse(ctx context.Context, res *http.Response, req *http.Request) error {

	// Capture full response body for 422 Unprocessable Entity which indicates a request body validation error
	// Equivalent to our use of 400 Bad Request
	if res.StatusCode == http.StatusUnprocessableEntity {
		var errorResponse validationErrorResponse
		if err := json.NewDecoder(res.Body).Decode(&errorResponse); err == nil {
			return errors.WithMeta(request.ErrorBadRequest(), errorResponse)
		}
	}

	// Let caller handle
	return nil
}

type validationErrorResponse struct {
	Detail []struct {
		Location []string `json:"loc,omitempty" bson:"loc,omitempty"`
		Message  string   `json:"msg,omitempty" bson:"msg,omitempty"`
		Type     string   `json:"type,omitempty" bson:"type,omitempty"`
	} `json:"detail,omitempty" bson:"detail,omitempty"`
}
