package test

import (
	"context"
	"net/http"

	"github.com/tidepool-org/platform/oauth"
)

type HTTPClientInput struct {
	Context           context.Context
	TokenSourceSource oauth.TokenSourceSource
}

type HTTPClientOutput struct {
	HTTPClient *http.Client
	Error      error
}

type HTTPClientSource struct {
	HTTPClientInvocations int
	HTTPClientInputs      []HTTPClientInput
	HTTPClientStub        func(ctx context.Context, tokenSourceSource oauth.TokenSourceSource) (*http.Client, error)
	HTTPClientOutputs     []HTTPClientOutput
	HTTPClientOutput      *HTTPClientOutput
}

func NewHTTPClientSource() *HTTPClientSource {
	return &HTTPClientSource{}
}

func (h *HTTPClientSource) HTTPClient(ctx context.Context, tokenSourceSource oauth.TokenSourceSource) (*http.Client, error) {
	h.HTTPClientInvocations++
	h.HTTPClientInputs = append(h.HTTPClientInputs, HTTPClientInput{Context: ctx, TokenSourceSource: tokenSourceSource})
	if h.HTTPClientStub != nil {
		return h.HTTPClientStub(ctx, tokenSourceSource)
	}
	if len(h.HTTPClientOutputs) > 0 {
		output := h.HTTPClientOutputs[0]
		h.HTTPClientOutputs = h.HTTPClientOutputs[1:]
		return output.HTTPClient, output.Error
	}
	if h.HTTPClientOutput != nil {
		return h.HTTPClientOutput.HTTPClient, h.HTTPClientOutput.Error
	}
	panic("HTTPClient has no output")
}

func (h *HTTPClientSource) AssertOutputsEmpty() {
	if len(h.HTTPClientOutputs) > 0 {
		panic("HTTPClientOutputs is not empty")
	}
}
