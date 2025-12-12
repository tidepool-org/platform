package client

import (
	"net/http"

	"golang.org/x/oauth2"
)

type authedTransport struct {
	tokenSource oauth2.TokenSource
	wrapped     http.RoundTripper
}

func NewAuthedTransport(tokenSource oauth2.TokenSource, wrapped http.RoundTripper) *authedTransport {
	return &authedTransport{tokenSource, wrapped}
}

func (t *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	token, err := t.tokenSource.Token()
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Shopify-Access-Token", token.AccessToken)
	return t.wrapped.RoundTrip(req)
}
