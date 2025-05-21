package twiist

import (
	"slices"

	"github.com/kelseyhightower/envconfig"
)

func NewServiceAccountAuthorizer() (*ServiceAccountAuthorizer, error) {
	serviceAccountAuthorizer := &ServiceAccountAuthorizer{}
	if err := envconfig.Process("", serviceAccountAuthorizer); err != nil {
		return nil, err
	}
	return serviceAccountAuthorizer, nil
}

type ServiceAccountAuthorizer struct {
	ServiceAccountIDs []string `envconfig:"TIDEPOOL_TWIIST_SERVICE_ACCOUNT_IDS"`
}

func (s *ServiceAccountAuthorizer) IsServiceAccountAuthorized(userID string) bool {
	return slices.Contains(s.ServiceAccountIDs, userID)
}
