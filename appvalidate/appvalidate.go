// Package appvalidate handles the logic for validating whether an app is a
// valid instance of your app via Apple's App Attest service.
package appvalidate

import (
	"context"
	"regexp"
	"time"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var (
	// base64 regex that supports base64.URLEncoding ("+/" replaced by "-_") or base64.StdEncoding. Used for base64 payloads like the attestation and assertion object.
	base64Chars = regexp.MustCompile("^(?:[A-Za-z0-9+/\\-_]{4})*(?:[A-Za-z0-9+/\\-_]{2}==|[A-Za-z0-9+/\\-_]{3}=)?$")
)

// AppValidation represents the entire state of a person's attestation /
// assertion status that determines if they are using a legitimate instance
// of an iOS app.
type AppValidation struct {
	UserID                  string     `json:"userId" bson:"userId,omitempty"`
	KeyID                   string     `json:"keyId" bson:"keyId,omitempty"`
	PublicKey               string     `json:"-" bson:"publicKey,omitempty"`
	Verified                bool       `json:"verified" bson:"verified"`
	FraudAssessmentReceipt  string     `json:"-" bson:"fraudAssessmentReceipt,omitempty"`
	AttestationChallenge    string     `json:"-" bson:"attestationChallenge,omitempty"`
	AssertionVerifiedTime   *time.Time `json:"-" bson:"assertionVerifiedTime,omitempty"`
	AssertionChallenge      string     `json:"-" bson:"assertionChallenge,omitempty"`
	AttestationVerifiedTime *time.Time `json:"-" bson:"attestationVerifiedTime"`
	AssertionCounter        uint32     `json:"assertionCounter" bson:"assertionCounter"`
}

// NewAppValidation creates a new AppValidation from a ChallengeCreate. Once a
// person starts the attestation process by requesting an attestation
// challenge, a new AppValidation needs to be persisted to keep track of the
// progress and state of the attestation and future assertions.
func NewAppValidation(ctx context.Context, attestChallenge string, create *ChallengeCreate) (*AppValidation, error) {
	if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(create); err != nil {
		return nil, err
	}
	validation := AppValidation{
		UserID:               create.UserID,
		KeyID:                create.KeyID,
		AttestationChallenge: attestChallenge,
	}
	if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(&validation); err != nil {
		return nil, err
	}
	return &validation, nil
}

func (av *AppValidation) Validate(v structure.Validator) {
	v.String("attestationChallenge", &av.AttestationChallenge).NotEmpty()
	v.String("userId", &av.UserID).NotEmpty()
	v.String("keyId", &av.KeyID).NotEmpty()
}
