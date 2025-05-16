package test

import (
	"context"
	"net/http"

	"github.com/tidepool-org/platform/auth"
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

type RefreshedTokenOutput struct {
	Token *auth.OAuthToken
	Error error
}

type TokenSource struct {
	HTTPClientInvocations     int
	HTTPClientInputs          []HTTPClientInput
	HTTPClientStub            func(ctx context.Context, tokenSourceSource oauth.TokenSourceSource) (*http.Client, error)
	HTTPClientOutputs         []HTTPClientOutput
	HTTPClientOutput          *HTTPClientOutput
	RefreshedTokenInvocations int
	RefreshedTokenStub        func() (*auth.OAuthToken, error)
	RefreshedTokenOutputs     []RefreshedTokenOutput
	RefreshedTokenOutput      *RefreshedTokenOutput
	ExpireTokenInvocations    int
	ExpireTokenStub           func()
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

func (t *TokenSource) RefreshedToken() (*auth.OAuthToken, error) {
	t.RefreshedTokenInvocations++
	if t.RefreshedTokenStub != nil {
		return t.RefreshedTokenStub()
	}
	if len(t.RefreshedTokenOutputs) > 0 {
		output := t.RefreshedTokenOutputs[0]
		t.RefreshedTokenOutputs = t.RefreshedTokenOutputs[1:]
		return output.Token, output.Error
	}
	if t.RefreshedTokenOutput != nil {
		return t.RefreshedTokenOutput.Token, t.RefreshedTokenOutput.Error
	}
	panic("RefreshedToken has no output")
}

func (t *TokenSource) ExpireToken() {
	t.ExpireTokenInvocations++
	if t.ExpireTokenStub != nil {
		t.ExpireTokenStub()
	}
}

func (t *TokenSource) AssertOutputsEmpty() {
	if len(t.HTTPClientOutputs) > 0 {
		panic("HTTPClientOutputs is not empty")
	}
	if len(t.RefreshedTokenOutputs) > 0 {
		panic("RefreshedTokenOutputs is not empty")
	}
}
