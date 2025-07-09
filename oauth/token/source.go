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
	token       *auth.OAuthToken
	tokenSource oauth2.TokenSource
	httpClient  *http.Client
}

func NewSourceWithToken(tkn *auth.OAuthToken) (*Source, error) {
	if tkn == nil {
		return nil, errors.New("token is missing")
	}

	return &Source{
		token: tkn,
	}, nil
}

func (s *Source) Token() *auth.OAuthToken {
	return s.token
}

func (s *Source) HTTPClient(ctx context.Context, tknSrcSrc oauth.TokenSourceSource) (*http.Client, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if tknSrcSrc == nil {
		return nil, errors.New("token source source is missing")
	}

	if s.tokenSource == nil {
		tknSrc, err := tknSrcSrc.TokenSource(ctx, s.token)
		if err != nil {
			return nil, err
		}

		httpClient := oauth2.NewClient(ctx, tknSrc)
		if httpClient == nil {
			return nil, errors.New("unable to create http client")
		}

		s.tokenSource = tknSrc
		s.httpClient = httpClient
	}

	return s.httpClient, nil
}

func (s *Source) UpdateToken() error {
	if s.tokenSource == nil {
		return nil
	}

	tknSrcTkn, err := s.tokenSource.Token()
	if err != nil {
		if oauth.IsRefreshTokenError(err) {
			err = errors.Wrap(request.ErrorUnauthenticated(), err.Error())
		}
		return errors.Wrap(err, "unable to get token")
	}

	if s.token.MatchesRawToken(tknSrcTkn) {
		return nil
	}

	tkn, err := s.token.Refreshed(tknSrcTkn)
	if err != nil || tkn == nil {
		return errors.Wrap(err, "unable to refresh token")
	}
	s.token = tkn

	return nil
}

func (s *Source) ExpireToken() error {
	s.httpClient = nil
	s.tokenSource = nil
	s.token.Expire()
	return nil
}
