package dexcom_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/dexcom/test"
	"github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("DataRange", func() {

	Describe("Validate", func() {
		DescribeTable("errors when",
			func(setupDataRangeFunc func() *dexcom.DataRange) {
				testDataRange := setupDataRangeFunc()
				validator := validator.New()
				testDataRange.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
			},
			Entry("required end is not set", func() *dexcom.DataRange {
				dataRange := test.RandomDataRange()
				dataRange.End = nil
				return dataRange
			}),
			Entry("required start is not set", func() *dexcom.DataRange {
				dataRange := test.RandomDataRange()
				dataRange.Start = nil
				return dataRange
			}),
			Entry("required end display time is not set", func() *dexcom.DataRange {
				dataRange := test.RandomDataRange()
				dataRange.End.DisplayTime = nil
				return dataRange
			}),
			Entry("required start system time is not set", func() *dexcom.DataRange {
				dataRange := test.RandomDataRange()
				dataRange.Start.SystemTime = nil
				return dataRange
			}),
			Entry("required start display time is not set", func() *dexcom.DataRange {
				dataRange := test.RandomDataRange()
				dataRange.Start.DisplayTime = nil
				return dataRange
			}),
		)
	})
})

var _ = Describe("DataRangeResponse", func() {

	Describe("Validate", func() {
		DescribeTable("errors when",
			func(setupDataRangeFunc func() *dexcom.DataRangeResponse) {
				testDataRange := setupDataRangeFunc()
				validator := validator.New()
				testDataRange.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
			},
			Entry("required end is not set", func() *dexcom.DataRange {
				dataRange := test.RandomDataRange()
				dataRange.End = nil
				return dataRange
			}),
			Entry("required start is not set", func() *dexcom.DataRange {
				dataRange := test.RandomDataRange()
				dataRange.Start = nil
				return dataRange
			}),
			Entry("required end display time is not set", func() *dexcom.DataRange {
				dataRange := test.RandomDataRange()
				dataRange.End.DisplayTime = nil
				return dataRange
			}),
			Entry("required start system time is not set", func() *dexcom.DataRange {
				dataRange := test.RandomDataRange()
				dataRange.Start.SystemTime = nil
				return dataRange
			}),
			Entry("required start display time is not set", func() *dexcom.DataRange {
				dataRange := test.RandomDataRange()
				dataRange.Start.DisplayTime = nil
				return dataRange
			}),
		)
	})
})
