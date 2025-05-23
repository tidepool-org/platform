package context

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/oauth"
	"github.com/tidepool-org/platform/request"
)

type Source struct {
	token             *auth.OAuthToken
	tokenSourceSource oauth.TokenSourceSource
	tokenSource       oauth2.TokenSource
	httpClient        *http.Client
}

func NewSource() (*Source, error) {
	return &Source{}, nil
}

func NewSourceWithToken(tkn *auth.OAuthToken) (*Source, error) {
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

func (s *Source) RefreshedToken() (*auth.OAuthToken, error) {
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

	tkn, err := s.token.Refreshed(tknSrcTkn)
	if err != nil {
		return nil, errors.Wrap(err, "unable to refresh token")
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
