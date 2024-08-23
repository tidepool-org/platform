package appvalidate

import (
	"encoding/json"
	"time"

	"github.com/tidepool-org/platform/structure"

	appAssert "github.com/bas-d/appattest/assertion"
	appUtils "github.com/bas-d/appattest/utils"
)

// AssertionVerify is the expected request body used by clients to complete
// the assertion process. Assertion can only be done after attestation is
// completed. The Assertion should be the base64 encoding of the binary CBOR
// data returned from the iOS APIs.
type AssertionVerify struct {
	UserID     string              `json:"-"`
	KeyID      string              `json:"keyId"`
	ClientData AssertionClientData `json:"clientData"`
	Assertion  string              `json:"assertion"`
}

// AssertionUpdate contains the assertion fields to update in an AppValidation
// to pass to a repository.
type AssertionUpdate struct {
	Challenge        string    `bson:"assertionChallenge,omitempty"`
	VerifiedTime     time.Time `bson:"assertionVerifiedTime,omitempty"`
	AssertionCounter uint32    `bson:"assertionCounter,omitempty"`
}

type AssertionResponse struct {
	Data any `json:"data"`
}

type AssertionClientData struct {
	Challenge   string          `json:"challenge"`
	Partner     string          `json:"partner"`     // Which partner are we requesting a secret from
	PartnerData json.RawMessage `json:"partnerData"` // Data to send to partner - This is a RawMessage because it is partner specific. The validation of this is delayed until later.
}

func NewAssertionVerify(userID string) *AssertionVerify {
	return &AssertionVerify{
		UserID: userID,
	}
}

func (av *AssertionVerify) Validate(v structure.Validator) {
	v.String("assertion", &av.Assertion).NotEmpty().Matches(base64Chars)
	v.String("clientData.challenge", &av.ClientData.Challenge).NotEmpty()
	v.String("clientData.partner", &av.ClientData.Partner).OneOf(partners...)

	v.String("userId", &av.UserID).NotEmpty()
	v.String("keyId", &av.KeyID).NotEmpty()
}

func transformAssertion(av *AssertionVerify) (*appAssert.AuthenticatorAssertionResponse, error) {
	clientDataRaw, err := json.Marshal(av.ClientData)
	if err != nil {
		return nil, err
	}

	var assertion appUtils.URLEncodedBase64
	assertionRaw := b64StdEncodingToURLEncoding(av.Assertion)
	if err := assertion.UnmarshalJSON([]byte(assertionRaw)); err != nil {
		return nil, err
	}

	return &appAssert.AuthenticatorAssertionResponse{
		RawClientData: appUtils.URLEncodedBase64(clientDataRaw),
		Assertion:     assertion,
	}, nil
}
