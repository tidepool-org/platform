package test

import (
	"context"

	"golang.org/x/oauth2"

	"github.com/tidepool-org/platform/oauth"
)

type TokenSourceInput struct {
	Context context.Context
	Token   *oauth.Token
}

type TokenSourceOutput struct {
	TokenSource oauth2.TokenSource
	Error       error
}

type TokenSourceSource struct {
	TokenSourceInvocations int
	TokenSourceInputs      []TokenSourceInput
	TokenSourceStub        func(ctx context.Context, token *oauth.Token) (oauth2.TokenSource, error)
	TokenSourceOutputs     []TokenSourceOutput
	TokenSourceOutput      *TokenSourceOutput
}

func NewTokenSourceSource() *TokenSourceSource {
	return &TokenSourceSource{}
}

func (t *TokenSourceSource) TokenSource(ctx context.Context, token *oauth.Token) (oauth2.TokenSource, error) {
	t.TokenSourceInvocations++
	t.TokenSourceInputs = append(t.TokenSourceInputs, TokenSourceInput{Context: ctx, Token: token})
	if t.TokenSourceStub != nil {
		return t.TokenSourceStub(ctx, token)
	}
	if len(t.TokenSourceOutputs) > 0 {
		output := t.TokenSourceOutputs[0]
		t.TokenSourceOutputs = t.TokenSourceOutputs[1:]
		return output.TokenSource, output.Error
	}
	if t.TokenSourceOutput != nil {
		return t.TokenSourceOutput.TokenSource, t.TokenSourceOutput.Error
	}
	panic("TokenSource has no output")
}

func (t *TokenSourceSource) AssertOutputsEmpty() {
	if len(t.TokenSourceOutputs) > 0 {
		panic("TokenSourceOutputs is not empty")
	}
}
