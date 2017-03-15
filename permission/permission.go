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
		return "", app.ExtError(err, "permission", "unable to encrypt with AES-256 using passphrase")
	}

	groupID := base64.StdEncoding.EncodeToString(groupIDBytes)
	return groupID, nil
}
