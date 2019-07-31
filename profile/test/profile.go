package test

import (
	"encoding/json"
	"time"

	metadataTest "github.com/tidepool-org/platform/metadata/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/profile"
	requestTest "github.com/tidepool-org/platform/request/test"
	"github.com/tidepool-org/platform/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

func RandomProfile() *profile.Profile {
	fullName := RandomFullName()
	value := *metadataTest.RandomMetadata()
	value["profile"] = map[string]interface{}{
		"fullName": fullName,
	}
	datum := &profile.Profile{}
	datum.UserID = pointer.FromString(userTest.RandomID())
	datum.Value = pointer.FromString(string(test.MustBytes(json.Marshal(value))))
	datum.CreatedTime = pointer.FromTime(test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second))
	datum.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*datum.CreatedTime, time.Now()).Truncate(time.Second))
	datum.Revision = pointer.FromInt(requestTest.RandomRevision())
	datum.FullName = pointer.FromString(fullName)
	return datum
}

func RandomProfileArray(minimumLength int, maximumLength int) profile.ProfileArray {
	datum := make(profile.ProfileArray, test.RandomIntFromRange(minimumLength, maximumLength))
	for index := range datum {
		datum[index] = RandomProfile()
	}
	return datum
}

func RandomFullName() string {
	return test.RandomString()
}
