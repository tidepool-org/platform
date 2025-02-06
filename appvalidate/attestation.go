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
// Attestation will be returned in CBOR format and should be base64
// encoded before sending.
type AttestationVerify struct {
	Attestation string `json:"attestation"`
	Challenge   string `json:"challenge"`
	KeyID       string `json:"keyId"`
	UserID      string `json:"-"`
}

func NewAttestationVerify(userID string) *AttestationVerify {
	return &AttestationVerify{
		UserID: userID,
	}
}

func (av *AttestationVerify) Validate(v structure.Validator) {
	v.String("challenge", &av.Challenge).NotEmpty()
	v.String("attestation", &av.Attestation).Matches(base64Chars)
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
	// The appattest library expects all the data to use base64 URLEncoding when the data from IOS is base64 StdEncoding so convert first.

	clientDataRaw := make([]byte, base64.RawURLEncoding.EncodedLen(len([]byte(av.Challenge))))
	base64.RawURLEncoding.Encode(clientDataRaw, []byte(av.Challenge))
	var clientData appUtils.URLEncodedBase64
	if err := clientData.UnmarshalJSON(clientDataRaw); err != nil {
		return nil, err
	}

	attestationRaw := b64StdEncodingToURLEncoding(av.Attestation)
	var attestation appUtils.URLEncodedBase64
	if err := attestation.UnmarshalJSON([]byte(attestationRaw)); err != nil {
		return nil, err
	}

	return &appAttest.AuthenticatorAttestationResponse{
		ClientData:        clientData,
		KeyID:             av.KeyID,
		AttestationObject: attestation,
	}, nil
}
