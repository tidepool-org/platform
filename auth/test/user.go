package test

import "github.com/tidepool-org/platform/auth"

func RandomUserID() string {
	return auth.NewUserID()
}
