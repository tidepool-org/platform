package appvalidate

import (
	"time"
)

type AssertionUpdate struct {
	Challenge        string    `bson:"assertionChallenge,omitempty"`
	VerifiedTime     time.Time `bson:"assertionVerifiedTime,omitempty"`
	AssertionCounter uint32    `bson:"assertionCounter,omitempty"`
}
