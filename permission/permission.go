package permission

import (
	"encoding/base64"

	"github.com/tidepool-org/platform/app"
)

func GroupIDFromUserID(userID string, secret string) (string, error) {
	if userID == "" {
		return "", app.Error("permission", "user id is missing")
	}
	if secret == "" {
		return "", app.Error("permission", "secret is missing")
	}

	groupIDBytes, err := app.EncryptWithAES256UsingPassphrase([]byte(userID), []byte(secret))
	if err != nil {
		return "", app.Error("permission", "unable to encrypt with AES-256 using passphrase")
	}

	groupID := base64.StdEncoding.EncodeToString(groupIDBytes)
	return groupID, nil
}

func UserIDFromGroupID(groupID string, secret string) (string, error) {
	if groupID == "" {
		return "", app.Error("permission", "group id is missing")
	}
	if secret == "" {
		return "", app.Error("permission", "secret is missing")
	}

	groupIDBytes, err := base64.StdEncoding.DecodeString(groupID)
	if err != nil {
		return "", app.Error("permission", "unable to decode with Base64")
	}

	userIDBytes, err := app.DecryptWithAES256UsingPassphrase(groupIDBytes, []byte(secret))
	if err != nil {
		return "", app.Error("permission", "unable to decrypt with AES-256 using passphrase")
	}

	return string(userIDBytes), nil
}
