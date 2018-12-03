package test

import (
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/time/zone"
)

func RandomName() string {
	return test.RandomStringFromArray(zone.Names())
}
