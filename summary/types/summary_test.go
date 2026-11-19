package types_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/tidepool-org/platform/summary/test"
	. "github.com/tidepool-org/platform/summary/types"
)

var _ = Describe("NewConfig", func() {
	It("stores the schema version and the glucose cut points", func() {
		// Storing cut points is a requirement of BACK-4158.
		Expect(NewConfig()).To(Equal(Config{
			SchemaVersion:            SchemaVersion,
			VeryLowGlucoseThreshold:  VeryLowBloodGlucose,
			LowGlucoseThreshold:      LowBloodGlucose,
			HighGlucoseThreshold:     HighBloodGlucose,
			VeryHighGlucoseThreshold: VeryHighBloodGlucose,
		}))
	})
})
