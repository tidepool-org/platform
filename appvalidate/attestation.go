package appvalidate

import (
	"encoding/base64"
	"time"

	"github.com/tidepool-org/platform/structure"

	appAttest "github.com/bas-d/appattest/attestation"
	appUtils "github.com/bas-d/appattest/utils"
)

// AttestationVerify is the request body used to validate an app's
// attestation. It is decoded from a JSON object.
// https://developer.apple.com/documentation/devicecheck/establishing_your_app_s_integrity#3561588
// KeyID and AttestationObject is data returned by the iOS APIs.
// AttestationObject will be returned in CBOR format and should be base64
// encoded before sending.
type AttestationVerify struct {
	AttestationObject string `json:"attestationObject"`
	Challenge         string `json:"challenge"`
	KeyID             string `json:"keyId"`
	UserID            string `json:"-"`
}

func NewAttestationVerify(userID string) *AttestationVerify {
	return &AttestationVerify{
		UserID: userID,
	}
}

func (av *AttestationVerify) Validate(v structure.Validator) {
	v.String("challenge", &av.Challenge).NotEmpty()
	v.String("attestationObject", &av.AttestationObject).Matches(base64Chars)
	v.String("userId", &av.UserID).NotEmpty()
	v.String("keyId", &av.KeyID).NotEmpty()
}

type AttestationUpdate struct {
	PublicKey              string    `bson:"publicKey,omitempty"`
	Verified               bool      `bson:"verified"`
	FraudAssessmentReceipt string    `bson:"fraudAssessmentReceipt,omitempty"`
	VerifiedTime           time.Time `bson:"attestationVerifiedTime"`
}

func (au *AttestationUpdate) Validate(v structure.Validator) {
	v.String("publicKey", &au.PublicKey).NotEmpty()
	v.String("fraudAssessmentReceipt", &au.FraudAssessmentReceipt).NotEmpty()
	v.Time("assertionVerifiedTime", &au.VerifiedTime).NotZero()
}

func transformAttestation(av *AttestationVerify) (*appAttest.AuthenticatorAttestationResponse, error) {
	// The appattest library expects all the data to be base64 encoded but for
	// convenience, the AttestationVerify struct only expects the
	// AttestationObject to be base64 encoded. So this converts to the
	// expected format.

	clientDataRaw := make([]byte, base64.RawURLEncoding.EncodedLen(len([]byte(av.Challenge))))
	base64.RawURLEncoding.Encode(clientDataRaw, []byte(av.Challenge))
	var clientData appUtils.URLEncodedBase64
	if err := clientData.UnmarshalJSON(clientDataRaw); err != nil {
		return nil, err
	}

	var attestationObject appUtils.URLEncodedBase64
	if err := attestationObject.UnmarshalJSON([]byte(av.AttestationObject)); err != nil {
		return nil, err
	}

	keyID := base64.StdEncoding.EncodeToString([]byte(av.KeyID))

	return &appAttest.AuthenticatorAttestationResponse{
		ClientData:        clientData,
		KeyID:             keyID,
		AttestationObject: attestationObject,
	}, nil
}
