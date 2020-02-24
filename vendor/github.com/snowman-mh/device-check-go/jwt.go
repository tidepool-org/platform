package devicecheck

import (
	"crypto/ecdsa"
	"encoding/json"
	"time"
	"unsafe"

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
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	headers := map[string]interface{}{
		"alg": jose.ES256,
		"kid": jwt.keyID,
	}

	return jose.Sign(*(*string)(unsafe.Pointer(&claimsJSON)), jose.ES256, key, jose.Headers(headers))
}
