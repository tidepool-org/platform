package context

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/oauth"
)

type Source struct {
	token       *oauth.Token
	tokenSource oauth2.TokenSource
	httpClient  *http.Client
	provider    oauth.Provider
}

func NewSource(tkn *oauth.Token) (*Source, error) {
	if tkn == nil {
		return nil, errors.New("token is missing")
	}

	return &Source{
		token: tkn,
	}, nil
}

func (s *Source) HTTPClient(ctx context.Context, prvdr oauth.Provider) (*http.Client, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if prvdr == nil {
		return nil, errors.New("provider is missing")
	}

	if prvdr != s.provider {
		cfg := prvdr.Config()
		if cfg == nil {
			return nil, errors.New("unable to create provider config")
		}

		tknSrc := cfg.TokenSource(ctx, s.token.RawToken())
		if tknSrc == nil {
			return nil, errors.New("unable to create token source")
		}

		httpClient := oauth2.NewClient(ctx, tknSrc)
		if httpClient == nil {
			return nil, errors.New("unable to create http client")
		}

		s.tokenSource = tknSrc
		s.httpClient = httpClient
		s.provider = prvdr
	}

	return s.httpClient, nil
}

func (s *Source) RefreshedToken() (*oauth.Token, error) {
	if s.tokenSource == nil {
		return nil, errors.New("token source is missing")
	}

	tknSrcTkn, err := s.tokenSource.Token()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get token")
	}

	if s.token.MatchesRawToken(tknSrcTkn) {
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
	s.provider = nil
	s.httpClient = nil
	s.tokenSource = nil

	s.token.Expire()
}
