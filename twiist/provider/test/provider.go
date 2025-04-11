package test

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/hashicorp/go-uuid"
	"github.com/lestrrat-go/jwx/v2/jwk"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/twiist/provider"
)

const JWKSRaw = `{
  "keys": [
    {
      "p": "1AOb8jPCv2xHXUFgp7-VLMLWGs0tVMkXzyBwhF0ARKyjPeBPzFukNDZ33vkv-8vHOlBohbeXsxJaZVCUpWN3zw",
      "kty": "RSA",
      "q": "pzNLWlzqeR7kV6k68VOu3Gvb0qLM2zqaHvFAQgFqYAVhJDHafnPKSMq-e-5O0T-lvojDbpTv9QiGImnbVmtsrQ",
      "d": "gRyoMnCJf2TePpV4f_1eLQ3KJdBlm8hnYYCj35aqbIZMR0fRKsjMKtaA7yowuNfDOotG649BJD_1NmzhjrWts5xz20N1jN0Qp14hMT-HgDeoN12Ygdber_IpWgcTgH-oEfldEwalQ26PJ_SeyRUPEfCKXhRLIAjgHCORnjteSrk",
      "e": "AQAB",
      "use": "sig",
      "kid": "test",
      "qi": "qzw6SmuFyPgc3oSt6B9SRYhZC_crl2FlTewn23YwVIxGRSk0qg5XDhYc1I8ecsAU5fn_-qvG0B4_8kHHoHwZmQ",
      "dp": "EyNwRGDfx5_ioUxxiTMGKFA-O5Uh7nFosM3g2lH64DglVESXb38mR4BTOdGMv1IZ3e28QbXc_9E8T8ECahuciQ",
      "alg": "RS256",
      "dq": "OQkSZ1zSz0ZudkjQRopZV--jKRNH9nDjKjL5zIpXEzJClOo8sm4lTvd6SyRb1p1zmK9mm05LHLcvqoWZwL0ccQ",
      "n": "injV2vXsnqLl5vAfZJUERU8QDD5edFBU9-1azBd8VT-V4aTqoe1vyrjtJGof7Wa2Df6fpvFrc344J-zUUNF0IMbNFCoV-KUM8UsahfNGjQ6rCSQcxRFhN8NVpaqcvGQoAl6sb7E0NiLkz38mTz0E0aueV8pcKFtX_eAaIVhISuM"
    }
  ]
}
`

func RandomSubjectID() string {
	id, err := uuid.GenerateUUID()
	if err != nil {
		panic(err)
	}

	return id
}

func RandomTidepoolLinkID() string {
	return fmt.Sprintf("twiist-%s", RandomSubjectID())
}

func GenerateIDToken(subjectID, tidepoolLinkID string, jwks jwk.Set) (string, error) {
	key, exists := jwks.Key(0)
	if !exists {
		return "", errors.New("unable to obtain rsa private key from jwks")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, provider.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "https://cognito-idp.us-east-1.amazonaws.com/us-east-1_12345678",
			Subject:   subjectID,
			Audience:  jwt.ClaimStrings{"k4kt11ukctj9u6lvfbjhkdn72"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Minute * 5)),
			ID:        "1141576b-0bee-4526-8985-7688d41c141a",
		},
		TidepoolLinkID: tidepoolLinkID,
	})
	token.Header["kid"] = key.KeyID()

	var v interface{}
	if err := key.Raw(&v); err != nil {
		return "", err
	}

	priv, ok := v.(*rsa.PrivateKey)
	if !ok {
		return "", errors.New("unable to obtain rsa private key from jwks")
	}

	return token.SignedString(priv)
}
