package devicecheck

import (
	"crypto/ecdsa"
	"encoding/json"
	"time"

	jose "github.com/dvsekhvalnov/jose2go"
)

type jwt struct {
	issuer string
	keyID  string
}

func newJWT(issuer, keyID string) jwt {
	return jwt{
		issuer: issuer,
		keyID:  keyID,
	}
}

func (jwt jwt) generate(key *ecdsa.PrivateKey) (string, error) {
	claims := map[string]interface{}{
		"iss": jwt.issuer,
		"iat": time.Now().UTC().Unix(),
	}

	// Ignoring error, because json.Marshal never fails.
	claimsJSON, _ := json.Marshal(claims)

	headers := map[string]interface{}{
		"alg": jose.ES256,
		"kid": jwt.keyID,
	}

	return jose.Sign(string(claimsJSON), jose.ES256, key, jose.Headers(headers))
}
