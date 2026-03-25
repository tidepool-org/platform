package work

import (
	"context"
	"net/http"

	"github.com/tidepool-org/platform/auth"
	providerSessionWork "github.com/tidepool-org/platform/auth/providersession/work"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/oauth"
	oauthToken "github.com/tidepool-org/platform/oauth/token"
	"github.com/tidepool-org/platform/work"
)

//go:generate mockgen -source=mixin.go -destination=test/mixin_mocks.go -package=test -typed

type Mixin interface {
	oauth.TokenSource

	TokenSource() oauth.TokenSource
	FetchTokenSource() *work.ProcessResult
}

func NewMixin(provider work.Provider, providerSessionMixin providerSessionWork.Mixin) (Mixin, error) {
	if provider == nil {
		return nil, errors.New("provider is missing")
	}
	if providerSessionMixin == nil {
		return nil, errors.New("provider session mixin is missing")
	}
	return &mixin{
		Provider:             provider,
		providerSessionMixin: providerSessionMixin,
	}, nil
}

type providerSessionMixin = providerSessionWork.Mixin

type mixin struct {
	work.Provider
	providerSessionMixin
	tokenSource *oauthToken.Source
}

func (m *mixin) TokenSource() oauth.TokenSource {
	return m // Encapsulate to persist updated token
}

func (m *mixin) FetchTokenSource() *work.ProcessResult {
	if !m.HasProviderSession() {
		return m.Failed(errors.New("provider session is missing"))
	}
	tokenSource, err := oauthToken.NewSourceWithToken(m.ProviderSession().OAuthToken)
	if err != nil {
		return m.Failed(errors.Wrap(err, "unable to create token source"))
	}
	m.tokenSource = tokenSource
	return nil
}

func (m *mixin) HTTPClient(ctx context.Context, tokenSourceSource oauth.TokenSourceSource) (*http.Client, error) {
	if m.tokenSource == nil {
		return nil, errors.New("token source is missing")
	} else {
		return m.tokenSource.HTTPClient(ctx, tokenSourceSource)
	}
}

func (m *mixin) UpdateToken(ctx context.Context) (bool, error) {
	if m.tokenSource == nil {
		return false, errors.New("token source is missing")
	} else if updated, err := m.tokenSource.UpdateToken(ctx); err != nil {
		return false, err
	} else if !updated {
		return false, nil
	} else if err = m.updateProviderSessionFromTokenSource(); err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func (m *mixin) ExpireToken(ctx context.Context) (bool, error) {
	if m.tokenSource == nil {
		return false, errors.New("token source is missing")
	} else if expired, err := m.tokenSource.ExpireToken(ctx); err != nil {
		return false, err
	} else if !expired {
		return false, nil
	} else if err = m.updateProviderSessionFromTokenSource(); err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func (m *mixin) updateProviderSessionFromTokenSource() error {
	if result := m.UpdateProviderSession(&auth.ProviderSessionUpdate{OAuthToken: m.tokenSource.Token(), ExternalID: m.ProviderSession().ExternalID}); result != nil {
		return result.Error()
	} else {
		return nil
	}
}
