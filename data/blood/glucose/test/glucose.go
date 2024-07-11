package test

import (
	"github.com/onsi/gomega"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/test"
)

func RandomUnits() string {
	return test.RandomStringFromArray(dataBloodGlucose.Units())
}

func ExpectRaw(raw *metadata.Metadata, expectedRaw *metadata.Metadata) {
	if expectedRaw != nil {
		gomega.Expect(raw).ToNot(gomega.BeNil())
		if expectedRaw.Get("units") == nil {
			gomega.Expect(raw.Get("units")).To(gomega.BeNil())
		} else {
			gomega.Expect(raw.Get("units")).To(gomega.Equal(expectedRaw.Get("units")))
		}
		if expectedRaw.Get("value") == nil {
			gomega.Expect(raw.Get("value")).To(gomega.BeNil())
		} else {
			gomega.Expect(raw.Get("value")).To(gomega.Equal(expectedRaw.Get("value")))
		}
	} else {
		gomega.Expect(raw).To(gomega.BeNil())
	}
}
