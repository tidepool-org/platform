package test

import (
	time2 "time"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/test"
)

func RandomDevicePushToken() string {
	return test.RandomStringFromRangeAndCharset(64, 64, test.CharsetAlphaNumeric)
}

func RandomDeviceCheckToken() string {
	return test.RandomStringFromRangeAndCharset(64, 64, test.CharsetAlphaNumeric)
}

func RandomDeviceAuthorizationID() string {
	return test.RandomStringFromRangeAndCharset(16, 16, test.CharsetHexidecimalLowercase)
}

func RandomDeviceAuthorizationToken() string {
	return test.RandomStringFromRangeAndCharset(32, 32, test.CharsetHexidecimalLowercase)
}

func RandomDeviceAuthorizationVerificationCode() string {
	return test.RandomStringFromRangeAndCharset(6, 6, test.CharsetHexidecimalLowercase)
}

func RandomDeviceAuthorization() *auth.DeviceAuthorization {
	time := time2.Now()
	return &auth.DeviceAuthorization{
		ID:               RandomDeviceAuthorizationID(),
		UserID:           RandomUserID(),
		Token:            RandomDeviceAuthorizationToken(),
		DevicePushToken:  RandomDevicePushToken(),
		Status:           test.RandomStringFromArray([]string{auth.DeviceAuthorizationPending, auth.DeviceAuthorizationSuccessful, auth.DeviceAuthorizationFailed}),
		BundleID:         test.RandomStringFromArray([]string{auth.LoopBundleID, auth.LoopBundleIDWithTeamPrefix}),
		VerificationCode: RandomDeviceAuthorizationVerificationCode(),
		CreatedTime:      time,
		ExpirationTime:   time.Add(auth.DeviceAuthorizationExpirationDuration),
		ModifiedTime:     &time,
	}
}

func RandomDeviceAuthorizationCreate() *auth.DeviceAuthorizationCreate {
	return &auth.DeviceAuthorizationCreate{
		DevicePushToken: RandomDevicePushToken(),
	}
}

func RandomDeviceAuthorizationUpdate() *auth.DeviceAuthorizationUpdate {
	return &auth.DeviceAuthorizationUpdate{
		BundleID:         test.RandomStringFromArray([]string{auth.LoopBundleID, auth.LoopBundleIDWithTeamPrefix}),
		VerificationCode: RandomDeviceAuthorizationVerificationCode(),
		DeviceCheckToken: RandomDeviceCheckToken(),
	}
}
