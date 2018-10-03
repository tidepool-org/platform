package id

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/tidepool-org/platform/errors"
)

func New(length int) (string, error) {
	if length < 1 {
		return "", errors.New("length is invalid")
	}
	bytes := make([]byte, length)
	n, err := rand.Read(bytes)
	if err != nil {
		return "", errors.Wrap(err, "unable to generate id")
	} else if n != length {
		return "", errors.New("generated id does not have expected length")
	}
	return hex.EncodeToString(bytes), nil
}

func Must(value string, err error) string {
	if err != nil {
		panic(err)
	}
	return value
}
