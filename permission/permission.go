package permission

import (
	"context"
	"encoding/base64"

	"github.com/tidepool-org/platform/crypto"
	"github.com/tidepool-org/platform/errors"
)

type Permission map[string]interface{}
type Permissions map[string]Permission

const Owner = "root"
const Custodian = "custodian"
const Write = "upload"
const Read = "view"

type Client interface {
	GetUserPermissions(ctx context.Context, requestUserID string, targetUserID string) (Permissions, error)
}

func FixOwnerPermissions(permissions Permissions) Permissions {
	if ownerPermission, ok := permissions[Owner]; ok {
		if _, ok = permissions[Write]; !ok {
			permissions[Write] = ownerPermission
		}
		if _, ok = permissions[Read]; !ok {
			permissions[Read] = ownerPermission
		}
	}
	return permissions
}

func GroupIDFromUserID(userID string, secret string) (string, error) {
	if userID == "" {
		return "", errors.New("user id is missing")
	}
	if secret == "" {
		return "", errors.New("secret is missing")
	}

	groupIDBytes, err := crypto.EncryptWithAES256UsingPassphrase([]byte(userID), []byte(secret))
	if err != nil {
		return "", errors.New("unable to encrypt with AES-256 using passphrase")
	}

	groupID := base64.StdEncoding.EncodeToString(groupIDBytes)
	return groupID, nil
}

func UserIDFromGroupID(groupID string, secret string) (string, error) {
	if groupID == "" {
		return "", errors.New("group id is missing")
	}
	if secret == "" {
		return "", errors.New("secret is missing")
	}

	groupIDBytes, err := base64.StdEncoding.DecodeString(groupID)
	if err != nil {
		return "", errors.New("unable to decode with Base64")
	}

	userIDBytes, err := crypto.DecryptWithAES256UsingPassphrase(groupIDBytes, []byte(secret))
	if err != nil {
		return "", errors.New("unable to decrypt with AES-256 using passphrase")
	}

	return string(userIDBytes), nil
}
