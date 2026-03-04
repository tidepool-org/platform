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

func RandomProviderSessionIDs() []string {
	return test.RandomStringArrayFromRangeAndGeneratorWithoutDuplicates(1, 3, RandomProviderSessionID)
}

func RandomProviderType() string {
	return test.RandomStringFromArray(auth.ProviderTypes())
}

func RandomProviderTypes() []string {
	return test.RandomStringArrayFromRangeAndArrayWithoutDuplicates(1, len(auth.ProviderTypes()), auth.ProviderTypes())
}

func RandomProviderName() string {
	return test.RandomStringFromRangeAndCharset(1, auth.ProviderNameLengthMaximum, test.CharsetAlphaNumeric)
}

func RandomProviderNames() []string {
	return test.RandomStringArrayFromRangeAndGeneratorWithoutDuplicates(1, 2, RandomProviderName)
}

func RandomProviderExternalID() string {
	return test.RandomStringFromRangeAndCharset(1, auth.ProviderExternalIDLengthMaximum, test.CharsetAlphaNumeric)
}

func RandomProviderExternalIDs() []string {
	return test.RandomStringArrayFromRangeAndGeneratorWithoutDuplicates(1, 2, RandomProviderExternalID)
}

func RandomProviderSession() *auth.ProviderSession {
	datum := &auth.ProviderSession{}
	datum.ID = RandomProviderSessionID()
	datum.UserID = userTest.RandomUserID()
	datum.Type = RandomProviderType()
	datum.Name = RandomProviderName()
	datum.OAuthToken = RandomToken()
	datum.ExternalID = pointer.FromString(RandomProviderExternalID())
	datum.CreatedTime = test.RandomTimeBeforeNow()
	datum.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(datum.CreatedTime, time.Now()))
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

func RandomProviderSessionUpdate() *auth.ProviderSessionUpdate {
	datum := &auth.ProviderSessionUpdate{}
	datum.OAuthToken = RandomToken()
	datum.ExternalID = pointer.FromString(RandomProviderExternalID())
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
