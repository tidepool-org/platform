package platform

import (
	"context"
	"net/http"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/request"
)

type SessionTokenHeaderMutator struct {
	*request.HeaderMutator
}

func NewSessionTokenHeaderMutator(sessionToken string) *SessionTokenHeaderMutator {
	return &SessionTokenHeaderMutator{
		HeaderMutator: request.NewHeaderMutator(auth.TidepoolSessionTokenHeaderKey, sessionToken),
	}
}

func (s *SessionTokenHeaderMutator) Mutate(req *http.Request) error {
	if s.HeaderMutator.Value == "" {
		return errors.New("session token is missing")
	}

	return s.HeaderMutator.Mutate(req)
}

type RestrictedTokenParameterMutator struct {
	*request.ParameterMutator
}

func NewRestrictedTokenParameterMutator(restrictedToken string) *RestrictedTokenParameterMutator {
	return &RestrictedTokenParameterMutator{
		ParameterMutator: request.NewParameterMutator(auth.TidepoolRestrictedTokenParameterKey, restrictedToken),
	}
}

func (r *RestrictedTokenParameterMutator) Mutate(req *http.Request) error {
	if r.ParameterMutator.Value == "" {
		return errors.New("restricted token is missing")
	}

	return r.ParameterMutator.Mutate(req)
}

type ServiceSecretHeaderMutator struct {
	*request.HeaderMutator
}

func NewServiceSecretHeaderMutator(serviceSecret string) *ServiceSecretHeaderMutator {
	return &ServiceSecretHeaderMutator{
		HeaderMutator: request.NewHeaderMutator(auth.TidepoolServiceSecretHeaderKey, serviceSecret),
	}
}

func (s *ServiceSecretHeaderMutator) Mutate(req *http.Request) error {
	if s.HeaderMutator.Value == "" {
		return errors.New("service secret is missing")
	}

	return s.HeaderMutator.Mutate(req)
}

type TraceMutator struct {
	Context context.Context
}

func NewTraceMutator(ctx context.Context) *TraceMutator {
	return &TraceMutator{
		Context: ctx,
	}
}

func (t *TraceMutator) Mutate(req *http.Request) error {
	if req == nil {
		return errors.New("request is missing")
	}

	if t.Context != nil {
		if err := request.CopyTrace(t.Context, req); err != nil {
			return errors.Wrapf(err, "unable to copy trace")
		}
	}

	return nil
}
