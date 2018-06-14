package test

import (
	"github.com/tidepool-org/platform/crypto"
	"github.com/tidepool-org/platform/test"
)

func RandomBase64EncodedMD5Hash() string {
	return crypto.Base64EncodedMD5Hash(test.RandomBytes())
}
