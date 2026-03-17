package work

import (
	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/structure"
)

const MetadataKeyOAuthToken = "oauthToken"

type TokenMetadata struct {
	OAuthToken *auth.OAuthToken `json:"oauthToken,omitempty" bson:"oauthToken,omitempty"`
}

func (t *TokenMetadata) Parse(parser structure.ObjectParser) {
	t.OAuthToken = auth.ParseOAuthToken(parser.WithReferenceObjectParser(MetadataKeyOAuthToken))
}

func (t *TokenMetadata) Validate(validator structure.Validator) {
	if t.OAuthToken != nil {
		t.OAuthToken.Validate(validator.WithReference(MetadataKeyOAuthToken))
	}
}
