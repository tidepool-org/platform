package test

import (
	"time"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

func RandomProviderSessionID() string {
	return auth.NewProviderSessionID()
}

func RandomProviderType() string {
	return test.RandomStringFromArray(auth.ProviderTypes())
}

func RandomProviderName() string {
	return test.RandomStringFromRangeAndCharset(1, auth.ProviderNameLengthMaximum, test.CharsetAlphaNumeric)
}

func RandomProviderExternalID() string {
	return test.RandomStringFromRangeAndCharset(1, auth.ProviderExternalIDLengthMaximum, test.CharsetAlphaNumeric)
}

func RandomProviderSession(options ...test.Option) *auth.ProviderSession {
	datum := &auth.ProviderSession{}
	datum.ID = RandomProviderSessionID()
	datum.UserID = userTest.RandomUserID()
	datum.Type = RandomProviderType()
	datum.Name = RandomProviderName()
	datum.OAuthToken = RandomToken()
	datum.ExternalID = test.RandomOptional(RandomProviderExternalID, options...)
	datum.CreatedTime = test.RandomTimeBeforeNow()
	datum.ModifiedTime = test.RandomOptional(func() time.Time { return test.RandomTimeFromRange(datum.CreatedTime, time.Now()) }, options...)
	return datum
}

func CloneProviderSession(datum *auth.ProviderSession) *auth.ProviderSession {
	if datum == nil {
		return nil
	}
	clone := &auth.ProviderSession{}
	clone.ID = datum.ID
	clone.UserID = datum.UserID
	clone.Type = datum.Type
	clone.Name = datum.Name
	clone.OAuthToken = CloneToken(datum.OAuthToken)
	clone.ExternalID = pointer.CloneString(datum.ExternalID)
	clone.CreatedTime = datum.CreatedTime
	clone.ModifiedTime = pointer.CloneTime(datum.ModifiedTime)
	return clone
}

func NewObjectFromProviderSession(datum *auth.ProviderSession, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	object["id"] = test.NewObjectFromString(datum.ID, objectFormat)
	object["userId"] = test.NewObjectFromString(datum.UserID, objectFormat)
	object["type"] = test.NewObjectFromString(datum.Type, objectFormat)
	object["name"] = test.NewObjectFromString(datum.Name, objectFormat)
	if datum.OAuthToken != nil {
		object["oauthToken"] = NewObjectFromToken(datum.OAuthToken, objectFormat)
	}
	if datum.ExternalID != nil {
		object["externalId"] = test.NewObjectFromString(*datum.ExternalID, objectFormat)
	}
	object["createdTime"] = test.NewObjectFromTime(datum.CreatedTime, objectFormat)
	if datum.ModifiedTime != nil {
		object["modifiedTime"] = test.NewObjectFromTime(*datum.ModifiedTime, objectFormat)
	}
	return object
}

func RandomProviderSessionUpdate(options ...test.Option) *auth.ProviderSessionUpdate {
	datum := &auth.ProviderSessionUpdate{}
	datum.OAuthToken = RandomToken()
	datum.ExternalID = test.RandomOptional(RandomProviderExternalID, options...)
	return datum
}

func CloneProviderSessionUpdate(datum *auth.ProviderSessionUpdate) *auth.ProviderSessionUpdate {
	if datum == nil {
		return nil
	}
	clone := &auth.ProviderSessionUpdate{}
	clone.OAuthToken = CloneToken(datum.OAuthToken)
	clone.ExternalID = pointer.CloneString(datum.ExternalID)
	return clone
}

func NewObjectFromProviderSessionUpdate(datum *auth.ProviderSessionUpdate, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	if datum.OAuthToken != nil {
		object["oauthToken"] = NewObjectFromToken(datum.OAuthToken, objectFormat)
	}
	if datum.ExternalID != nil {
		object["externalId"] = test.NewObjectFromString(*datum.ExternalID, objectFormat)
	}
	return object
}
