package test

import "github.com/tidepool-org/platform/oauth"

type RefreshedTokenOutput struct {
	Token *oauth.Token
	Error error
}

type TokenSource struct {
	*HTTPClientSource
	RefreshedTokenInvocations int
	RefreshedTokenStub        func() (*oauth.Token, error)
	RefreshedTokenOutputs     []RefreshedTokenOutput
	RefreshedTokenOutput      *RefreshedTokenOutput
	ExpireTokenInvocations    int
	ExpireTokenStub           func()
}

func NewTokenSource() *TokenSource {
	return &TokenSource{
		HTTPClientSource: NewHTTPClientSource(),
	}
}

func (t *TokenSource) RefreshedToken() (*oauth.Token, error) {
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
	if len(t.RefreshedTokenOutputs) > 0 {
		panic("RefreshedTokenOutputs is not empty")
	}
}
