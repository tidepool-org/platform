package devicetokens

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	AppleEnvProduction = "production"
	AppleEnvSandbox    = "sandbox"

	// MaxTokenLen for an opaque token blob sent by Apple.
	//
	// Apple's docs indicate that the length should not be hard-coded, but
	// we've decided this is an appropriate maximum limit. Assuming the blob
	// is just a randomly generated identifier, there's no forseeable reason
	// it should require anywhere near this much. There are only 128 bits
	// (that's a mere 16 bytes!) in a UUID afterall.
	MaxTokenLen = 8192
)

type Document struct {
	// UserID of the user that owns the DeviceToken.
	UserID string `json:"userId" bson:"userId"`
	// TokenKey is string that uniquely identifies the DeviceToken.
	//
	// It's useful for generating unique indexes.
	TokenKey string `json:"tokenKey" bson:"tokenKey"`
	// DeviceToken wraps the device-specific token.
	DeviceToken DeviceToken `json:"deviceToken" bson:"deviceToken"`
}

func NewDocument(userID string, deviceToken DeviceToken) *Document {
	return &Document{
		UserID:      userID,
		TokenKey:    deviceToken.key(),
		DeviceToken: deviceToken,
	}
}

// DeviceToken is received from a Tidepool client application.
//
// It contains the information necessary for a service to send a push
// notification to the device.
type DeviceToken struct {
	// Apple devices should provide this information.
	Apple *AppleDeviceToken `json:"apple,omitempty" bson:"apple,omitempty"`
}

// key provides a unique string value to identify this device token.
//
// Intended to be used as part of a unique index for database indexes.
func (t DeviceToken) key() string {
	if t.Apple != nil {
		return t.Apple.key()
	}
	return ""
}

func (t DeviceToken) Validate(validator structure.Validator) {
	appleValidator := validator.WithReference("apple")
	if t.Apple != nil {
		t.Apple.Validate(appleValidator)
	} else {
		// There's no other kind of token, so if there's no Apple, this is invalid.
		appleValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
}

type AppleDeviceToken struct {
	// Token from Apple that identifies this specific device.
	Token AppleBlob `json:"token" bson:"token"`
	// Environment is either sandbox or production.
	Environment string `json:"environment" bson:"environment"`
}

func (t AppleDeviceToken) key() string {
	if len(t.Token) == 0 || t.Environment == "" {
		return ""
	}
	l := sha256.Sum256(fmt.Append(t.Token, t.Environment))
	return hex.EncodeToString(l[:])
}

func (t AppleDeviceToken) Validate(validator structure.Validator) {
	validator.Bytes("token", t.Token).NotEmpty().
		LengthLessThanOrEqualTo(MaxTokenLen)
	validator.String("environment", &t.Environment).
		NotEmpty().
		OneOf(AppleEnvProduction, AppleEnvSandbox)
}

// AppleBlob is an opaque blob to identify the device.
type AppleBlob []byte

// Repository abstracts persistent storage for Token data.
type Repository interface {
	GetAllByUserID(ctx context.Context, userID string) ([]*Document, error)
	Upsert(ctx context.Context, doc *Document) error

	EnsureIndexes() error
}
