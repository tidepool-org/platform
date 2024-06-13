package context

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/oauth"
	"github.com/tidepool-org/platform/request"
)

type Source struct {
	token             *oauth.Token
	tokenSourceSource oauth.TokenSourceSource
	tokenSource       oauth2.TokenSource
	httpClient        *http.Client
}

func NewSource() (*Source, error) {
	return &Source{}, nil
}

func NewSourceWithToken(tkn *oauth.Token) (*Source, error) {
	if tkn == nil {
		return nil, errors.New("token is missing")
	}

	return &Source{
		token: tkn,
	}, nil
}

func (s *Source) HTTPClient(ctx context.Context, tknSrcSrc oauth.TokenSourceSource) (*http.Client, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if tknSrcSrc == nil {
		return nil, errors.New("token source source is missing")
	}

	if tknSrcSrc != s.tokenSourceSource {
		tknSrc, err := tknSrcSrc.TokenSource(ctx, s.token)
		if err != nil {
			return nil, err
		}

		// TODO: DO NOT COMMIT!!!
		proxyUrl, _ := url.Parse("http://192.168.0.102:9090")
		baseHTTPClient := &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyUrl),
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}

		ctx = context.WithValue(ctx, oauth2.HTTPClient, baseHTTPClient)

		httpClient := oauth2.NewClient(ctx, tknSrc)
		if httpClient == nil {
			return nil, errors.New("unable to create http client")
		}

		s.tokenSourceSource = tknSrcSrc
		s.tokenSource = tknSrc
		s.httpClient = httpClient
	}

	return s.httpClient, nil
}

func (s *Source) RefreshedToken() (*oauth.Token, error) {
	if s.tokenSource == nil {
		return nil, errors.New("token source is missing")
	}

	tknSrcTkn, err := s.tokenSource.Token()
	if err != nil {
		if oauth.IsRefreshTokenError(err) {
			err = errors.Wrap(request.ErrorUnauthenticated(), err.Error())
		}
		return nil, errors.Wrap(err, "unable to get token")
	}

	if s.token == nil || s.token.MatchesRawToken(tknSrcTkn) {
		return nil, nil
	}

	tkn, err := oauth.NewTokenFromRawToken(tknSrcTkn)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create token")
	}

	s.token = tkn
	return s.token, nil
}

func (s *Source) ExpireToken() {
	s.httpClient = nil
	s.tokenSource = nil
	s.tokenSourceSource = nil

	if s.token != nil {
		s.token.Expire()
	}
}
