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
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	MetadataKeyOAuthToken = "oauthToken"
)

type Mixin struct {
	*workBase.Processor
	providerSessionMixin *providerSessionWork.Mixin
	tokenSource          *oauthToken.Source
}

func NewMixin(processor *workBase.Processor, providerSessionMixin *providerSessionWork.Mixin) (*Mixin, error) {
	if processor == nil {
		return nil, errors.New("processor is missing")
	}
	if providerSessionMixin == nil {
		return nil, errors.New("provider session mixin is missing")
	}

	return &Mixin{
		Processor:            processor,
		providerSessionMixin: providerSessionMixin,
	}, nil
}

func (m *Mixin) TokenSource() oauth.TokenSource {
	return m // Encapsulate to persist updated token
}

func (m *Mixin) FetchTokenSource() *work.ProcessResult {
	if m.providerSessionMixin.ProviderSession == nil {
		return m.Failed(errors.New("provider session is missing"))
	}

	tokenSource, err := oauthToken.NewSourceWithToken(m.providerSessionMixin.ProviderSession.OAuthToken)
	if err != nil {
		return m.Failed(errors.Wrap(err, "unable to create token source"))
	}
	m.tokenSource = tokenSource

	return nil
}

func (m *Mixin) HTTPClient(ctx context.Context, tokenSourceSource oauth.TokenSourceSource) (*http.Client, error) {
	if m.tokenSource == nil {
		return nil, errors.New("token source is missing")
	} else {
		return m.tokenSource.HTTPClient(ctx, tokenSourceSource)
	}
}

func (m *Mixin) UpdateToken(ctx context.Context) (bool, error) {
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

func (m *Mixin) ExpireToken(ctx context.Context) (bool, error) {
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

func (m *Mixin) updateProviderSessionFromTokenSource() error {
	if m.providerSessionMixin.ProviderSession == nil {
		return errors.New("provider session is missing")
	}
	if token := m.tokenSource.Token(); token == m.providerSessionMixin.ProviderSession.OAuthToken {
		return nil
	} else if result := m.providerSessionMixin.UpdateProviderSession(auth.ProviderSessionUpdate{OAuthToken: token, ExternalID: m.providerSessionMixin.ProviderSession.ExternalID}); result != nil {
		return result.Error()
	} else {
		return nil
	}
}

type OAuthTokenMixin struct {
	*workBase.Processor
}

func NewOAuthTokenMixin(processor *workBase.Processor) (*OAuthTokenMixin, error) {
	if processor == nil {
		return nil, errors.New("processor is missing")
	}
	return &OAuthTokenMixin{
		Processor: processor,
	}, nil
}

func (o *OAuthTokenMixin) OAuthTokenFromMetadata() (*auth.OAuthToken, error) {
	parser := o.MetadataParser()
	oauthToken := auth.ParseOAuthToken(parser.WithReferenceObjectParser(MetadataKeyOAuthToken))
	if err := parser.Error(); err != nil {
		return nil, errors.Wrap(err, "unable to parse oauth token from metadata")
	}
	return oauthToken, nil
}
