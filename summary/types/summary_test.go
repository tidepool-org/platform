package types_test

import (
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/blood/glucose"
	. "github.com/tidepool-org/platform/summary/test"
	. "github.com/tidepool-org/platform/summary/types"
)

var _ = Describe("Config", func() {
	It("NewConfig stores the schema version, the standard ranges, and their cut points",
		func() {
			// Storing cut points is a requirement of BACK-4158.
			Expect(NewConfig()).To(Equal(Config{
				SchemaVersion:            SchemaVersion,
				GlycemicRangesKey:        "preset:adaStandard",
				VeryLowGlucoseThreshold:  VeryLowBloodGlucose,
				LowGlucoseThreshold:      LowBloodGlucose,
				HighGlucoseThreshold:     HighBloodGlucose,
				VeryHighGlucoseThreshold: VeryHighBloodGlucose,
			}))
		})

	It("NewConfigFor stores the range set's key and its cut points", func() {
		rangeSet, err := NewCustomRangeSet("custom:mg/dL:54,70,180,250", glucose.MgdL,
			Thresholds{Low: 70, High: 180})
		Expect(err).ToNot(HaveOccurred())
		cutPoints, err := rangeSet.CutPoints()
		Expect(err).ToNot(HaveOccurred())

		config := NewConfigFor(rangeSet.Key, cutPoints)
		Expect(config.SchemaVersion).To(Equal(SchemaVersion))
		Expect(config.GlycemicRangesKey).To(Equal("custom:mg/dL:54,70,180,250"))
		Expect(config.LowGlucoseThreshold).To(Equal(cutPoints.Low))
		Expect(config.HighGlucoseThreshold).To(Equal(cutPoints.High))
		Expect(math.IsInf(config.VeryLowGlucoseThreshold, -1)).To(BeTrue())
		Expect(math.IsInf(config.VeryHighGlucoseThreshold, 1)).To(BeTrue())
	})
})
