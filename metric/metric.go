package metric

import (
	"crypto/sha1"
	"encoding/hex"

	"github.com/tidepool-org/platform/errors"
)

func HashFromUserID(userID string, salt string) (string, error) {
	if userID == "" {
		return "", errors.New("user id is missing")
	}
	if salt == "" {
		return "", errors.New("salt is missing")
	}

	sha1Sum := sha1.Sum([]byte(salt + userID))
	return hex.EncodeToString(sha1Sum[:])[0:10], nil
}
