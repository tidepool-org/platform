package appvalidate

import (
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/structure"
)

// ChallengeCreate is the expected request body used to create an attestation
// or assertion challenge.
type ChallengeCreate struct {
	UserID string `json:"-"` // json ignored because taken from request.Details and not from user supplied input.
	KeyID  string `json:"keyId"`
}

// ChallengeResult is the response to a successful request with
// ChallengeCreate
type ChallengeResult struct {
	Challenge string `json:"challenge"`
}

func NewChallengeCreate(userID string) *ChallengeCreate {
	return &ChallengeCreate{
		UserID: userID,
	}
}

func (c *ChallengeCreate) Validate(v structure.Validator) {
	v.String("userId", &c.UserID).NotEmpty()
	v.String("keyId", &c.KeyID).NotEmpty()
}

type ChallengeGenerator interface {
	GenerateChallenge(size int) (string, error)
}

type challengeGenerator struct{}

func NewChallengeGenerator() ChallengeGenerator {
	return challengeGenerator{}
}

func (c challengeGenerator) GenerateChallenge(size int) (string, error) {
	return id.New(size)
}
