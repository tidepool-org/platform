package test

import (
	"github.com/tidepool-org/platform/crypto"
	"github.com/tidepool-org/platform/test"
)

func RandomBase64EncodedMD5Hash() string {
	return crypto.Base64EncodedMD5Hash(test.RandomBytes())
}

func RandomHexEncodedMD5Hash() string {
	return crypto.HexEncodedMD5Hash(test.RandomString())
}

func RandomBase64EncodedSHA256Hash() string {
	return crypto.Base64EncodedSHA256Hash(test.RandomBytes())
}

func RandomHexEncodedSHA256Hash() string {
	return crypto.HexEncodedSHA256Hash(test.RandomString())
}
