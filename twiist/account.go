package twiist

import (
	"slices"

	"github.com/kelseyhightower/envconfig"
)

type ServiceAccountAuthorizer interface {
	IsAuthorized(userID string) bool
}

func NewServiceAccountAuthorizer() (ServiceAccountAuthorizer, error) {
	authorizer := &serviceAccountAuthorizer{}
	err := envconfig.Process("", authorizer)
	if err != nil {
		return nil, err
	}

	return authorizer, nil
}

type serviceAccountAuthorizer struct {
	ServiceAccountIDs []string `envconfig:"TIDEPOOL_TWIIST_SERVICE_ACCOUNT_IDS"`
}

func (s *serviceAccountAuthorizer) IsAuthorized(userID string) bool {
	return slices.Contains(s.ServiceAccountIDs, userID)
}
