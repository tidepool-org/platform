package test

import "github.com/tidepool-org/platform/user"

func RandomID() string {
	return user.NewID()
}
