package keycloak

import (
	"github.com/Nerzal/gocloak/v13/pkg/jwx"
)

const serverRole = "backend_service"

type TokenIntrospectionResult struct {
	Active           bool        `json:"active"`
	Subject          string      `json:"sub"`
	EmailVerified    bool        `json:"email_verified"`
	ExpiresAt        int64       `json:"eat"`
	RealmAccess      RealmAccess `json:"realm_access"`
	IdentityProvider string      `json:"identityProvider"`
}

type AccessTokenCustomClaims struct {
	jwx.Claims
	IdentityProvider string `json:"identity_provider,omitempty"`
}

type RealmAccess struct {
	Roles []string `json:"roles"`
}

func (t *TokenIntrospectionResult) IsServerToken() bool {
	if len(t.RealmAccess.Roles) > 0 {
		for _, role := range t.RealmAccess.Roles {
			if role == serverRole {
				return true
			}
		}
	}

	return false
}
