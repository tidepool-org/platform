package test

import (
	"time"

	"github.com/onsi/gomega"
	gomegaGstruct "github.com/onsi/gomega/gstruct"
	gomegaTypes "github.com/onsi/gomega/types"

	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/user"
)

func RandomDelete() *user.Delete {
	datum := &user.Delete{}
	datum.Password = pointer.FromString(RandomPassword())
	return datum
}

func CloneDelete(datum *user.Delete) *user.Delete {
	if datum == nil {
		return nil
	}
	clone := &user.Delete{}
	clone.Password = pointer.CloneString(datum.Password)
	return clone
}

func NewObjectFromDelete(datum *user.Delete, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.Password != nil {
		object["password"] = test.NewObjectFromString(*datum.Password, objectFormat)
	}
	return object
}

func RandomPassword() string {
	return test.RandomString()
}

func RandomUser() *user.User {
	datum := &user.User{}
	datum.UserID = pointer.FromString(RandomID())
	datum.Username = pointer.FromString(RandomUsername())
	datum.EmailVerified = pointer.FromBool(test.RandomBool())
	datum.TermsAccepted = pointer.FromString(test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Format(time.RFC3339Nano))
	datum.Roles = nil
	datum.CreatedTime = pointer.FromTime(test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second))
	datum.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*datum.CreatedTime, time.Now()).Truncate(time.Second))
	return datum
}

func CloneUser(datum *user.User) *user.User {
	if datum == nil {
		return nil
	}
	clone := &user.User{}
	clone.UserID = pointer.CloneString(datum.UserID)
	clone.Username = pointer.CloneString(datum.Username)
	clone.PasswordHash = pointer.CloneString(datum.PasswordHash)
	clone.EmailVerified = pointer.CloneBool(datum.EmailVerified)
	clone.TermsAccepted = pointer.CloneString(datum.TermsAccepted)
	clone.Roles = pointer.CloneStringArray(datum.Roles)
	clone.CreatedTime = pointer.CloneTime(datum.CreatedTime)
	clone.ModifiedTime = pointer.CloneTime(datum.ModifiedTime)
	clone.DeletedTime = pointer.CloneTime(datum.DeletedTime)
	return clone
}

func NewObjectFromUser(datum *user.User, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.UserID != nil {
		object["userid"] = test.NewObjectFromString(*datum.UserID, objectFormat)
	}
	if datum.Username != nil {
		object["username"] = test.NewObjectFromString(*datum.Username, objectFormat)
	}
	if datum.EmailVerified != nil {
		object["emailVerified"] = test.NewObjectFromBool(*datum.EmailVerified, objectFormat)
	}
	if datum.TermsAccepted != nil {
		object["termsAccepted"] = test.NewObjectFromString(*datum.TermsAccepted, objectFormat)
	}
	if datum.Roles != nil {
		object["roles"] = test.NewObjectFromStringArray(*datum.Roles, objectFormat)
	}
	if datum.CreatedTime != nil {
		object["createdTime"] = test.NewObjectFromTime(*datum.CreatedTime, objectFormat)
	}
	if datum.ModifiedTime != nil {
		object["modifiedTime"] = test.NewObjectFromTime(*datum.ModifiedTime, objectFormat)
	}
	if datum.DeletedTime != nil {
		object["deletedTime"] = test.NewObjectFromTime(*datum.DeletedTime, objectFormat)
	}
	return object
}

func MatchUser(datum *user.User) gomegaTypes.GomegaMatcher {
	if datum == nil {
		return gomega.BeNil()
	}
	return gomegaGstruct.PointTo(gomegaGstruct.MatchAllFields(gomegaGstruct.Fields{
		"UserID":        gomega.Equal(datum.UserID),
		"Username":      gomega.Equal(datum.Username),
		"PasswordHash":  gomega.Equal(datum.PasswordHash),
		"EmailVerified": gomega.Equal(datum.EmailVerified),
		"TermsAccepted": gomega.Equal(datum.TermsAccepted),
		"Roles":         gomega.Equal(datum.Roles),
		"CreatedTime":   test.MatchTime(datum.CreatedTime),
		"ModifiedTime":  test.MatchTime(datum.ModifiedTime),
		"DeletedTime":   test.MatchTime(datum.DeletedTime),
	}))
}

func RandomUsername() string {
	return netTest.RandomEmail()
}

func RandomUserArray(minimumLength int, maximumLength int) user.UserArray {
	datum := make(user.UserArray, test.RandomIntFromRange(minimumLength, maximumLength))
	for index := range datum {
		datum[index] = RandomUser()
	}
	return datum
}

func CloneUserArray(datum user.UserArray) user.UserArray {
	if datum == nil {
		return nil
	}
	clone := make(user.UserArray, len(datum))
	for index := range datum {
		clone[index] = CloneUser(datum[index])
	}
	return clone
}

func MatchUserArray(datum user.UserArray) gomegaTypes.GomegaMatcher {
	matchers := []gomegaTypes.GomegaMatcher{}
	for _, d := range datum {
		matchers = append(matchers, MatchUser(d))
	}
	return test.MatchArray(matchers)
}

func RandomID() string {
	return user.NewID()
}
