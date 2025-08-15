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

type TokenSource struct {
	HTTPClientInvocations  int
	HTTPClientInputs       []HTTPClientInput
	HTTPClientStub         func(ctx context.Context, tokenSourceSource oauth.TokenSourceSource) (*http.Client, error)
	HTTPClientOutputs      []HTTPClientOutput
	HTTPClientOutput       *HTTPClientOutput
	UpdateTokenInvocations int
	UpdateTokenStub        func() error
	UpdateTokenOutputs     []error
	UpdateTokenOutput      error
	ExpireTokenInvocations int
	ExpireTokenStub        func() error
	ExpireTokenOutputs     []error
	ExpireTokenOutput      error
}

func NewTokenSource() *TokenSource {
	return &TokenSource{}
}

func (t *TokenSource) HTTPClient(ctx context.Context, tokenSourceSource oauth.TokenSourceSource) (*http.Client, error) {
	t.HTTPClientInvocations++
	t.HTTPClientInputs = append(t.HTTPClientInputs, HTTPClientInput{Context: ctx, TokenSourceSource: tokenSourceSource})
	if t.HTTPClientStub != nil {
		return t.HTTPClientStub(ctx, tokenSourceSource)
	}
	if len(t.HTTPClientOutputs) > 0 {
		output := t.HTTPClientOutputs[0]
		t.HTTPClientOutputs = t.HTTPClientOutputs[1:]
		return output.HTTPClient, output.Error
	}
	if t.HTTPClientOutput != nil {
		return t.HTTPClientOutput.HTTPClient, t.HTTPClientOutput.Error
	}
	panic("HTTPClient has no output")
}

func (t *TokenSource) UpdateToken() error {
	t.UpdateTokenInvocations++
	if t.UpdateTokenStub != nil {
		return t.UpdateTokenStub()
	}
	if len(t.UpdateTokenOutputs) > 0 {
		output := t.UpdateTokenOutputs[0]
		t.UpdateTokenOutputs = t.UpdateTokenOutputs[1:]
		return output
	}
	if t.UpdateTokenOutput != nil {
		return t.UpdateTokenOutput
	}
	panic("UpdateToken has no output")
}

func (t *TokenSource) ExpireToken() error {
	t.ExpireTokenInvocations++
	if t.ExpireTokenStub != nil {
		return t.ExpireTokenStub()
	}
	if len(t.ExpireTokenOutputs) > 0 {
		output := t.ExpireTokenOutputs[0]
		t.ExpireTokenOutputs = t.ExpireTokenOutputs[1:]
		return output
	}
	if t.ExpireTokenOutput != nil {
		return t.ExpireTokenOutput
	}
	panic("ExpireToken has no output")
}

func (t *TokenSource) AssertOutputsEmpty() {
	if len(t.HTTPClientOutputs) > 0 {
		panic("HTTPClientOutputs is not empty")
	}
	if len(t.UpdateTokenOutputs) > 0 {
		panic("UpdateTokenOutputs is not empty")
	}
	if len(t.ExpireTokenOutputs) > 0 {
		panic("ExpireTokenOutputs is not empty")
	}
}
