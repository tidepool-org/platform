package devicetokens

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/tidepool-org/platform/structure"
)

type Document struct {
	// UserID of the user that owns the DeviceToken.
	UserID string `json:"userId" bson:"userId"`
	// TokenID is string that uniquely identifies the DeviceToken.
	//
	// It's useful for generating unique indexes.
	TokenID string `json:"tokenId" bson:"tokenId"`
	// DeviceToken wraps the device-specific token.
	DeviceToken DeviceToken `json:"deviceToken" bson:"deviceToken"`
}

func NewDocument(userID string, deviceToken DeviceToken) *Document {
	return &Document{
		UserID:      userID,
		TokenID:     deviceToken.key(),
		DeviceToken: deviceToken,
	}
}

// DeviceToken is received from a Tidepool client application.
//
// It contains the information necessary for a service to send a push
// notification to the device.
type DeviceToken struct {
	// Apple devices should provide this information.
	Apple AppleDeviceToken `json:"apple,omitempty" bson:"apple,omitempty"`
}

// key provides a unique string value to identify this device token.
//
// Intended to be used as part of a unique index for database indexes.
func (d DeviceToken) key() string {
	if appleKey := d.Apple.key(); appleKey != "" {
		return appleKey
	}
	return ""
}

func (d DeviceToken) Validate(validator structure.Validator) {
	d.Apple.Validate(validator)
}

type AppleDeviceToken struct {
	// Token from Apple that identifies this specific device.
	Token AppleBlob
	// Environment is either sandbox or production.
	Environment string
}

func (b AppleDeviceToken) key() string {
	if len(b.Token) == 0 || b.Environment == "" {
		return ""
	}
	l := sha256.Sum256(fmt.Append(b.Token, b.Environment))
	return hex.EncodeToString(l[:])
}

func (b AppleDeviceToken) Validate(validator structure.Validator) {
	validator.Bytes("Token", b.Token).NotEmpty()
	validator.String("Environment", &b.Environment).
		NotEmpty().
		OneOf("production", "sandbox")
}

// AppleBlob is an opaque blob to identify the device.
type AppleBlob []byte

// Repository abstracts persistent storage for Token data.
type Repository interface {
	Upsert(ctx context.Context, doc *Document) error

	EnsureIndexes() error
}
