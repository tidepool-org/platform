package permission

import (
	"encoding/base64"

	"github.com/tidepool-org/platform/crypto"
	"github.com/tidepool-org/platform/errors"
)

func GroupIDFromUserID(userID string, secret string) (string, error) {
	if userID == "" {
		return "", errors.New("permission", "user id is missing")
	}
	if secret == "" {
		return "", errors.New("permission", "secret is missing")
	}

	groupIDBytes, err := crypto.EncryptWithAES256UsingPassphrase([]byte(userID), []byte(secret))
	if err != nil {
		return "", errors.New("permission", "unable to encrypt with AES-256 using passphrase")
	}

	groupID := base64.StdEncoding.EncodeToString(groupIDBytes)
	return groupID, nil
}

func UserIDFromGroupID(groupID string, secret string) (string, error) {
	if groupID == "" {
		return "", errors.New("permission", "group id is missing")
	}
	if secret == "" {
		return "", errors.New("permission", "secret is missing")
	}

	groupIDBytes, err := base64.StdEncoding.DecodeString(groupID)
	if err != nil {
		return "", errors.New("permission", "unable to decode with Base64")
	}

	userIDBytes, err := crypto.DecryptWithAES256UsingPassphrase(groupIDBytes, []byte(secret))
	if err != nil {
		return "", errors.New("permission", "unable to decrypt with AES-256 using passphrase")
	}

	return string(userIDBytes), nil
}
