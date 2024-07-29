package main_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/r3labs/diff/v3"

	"github.com/tidepool-org/platform/migrations/20231128_jellyfish_migration/verify/test"

	verify "github.com/tidepool-org/platform/migrations/20231128_jellyfish_migration/verify"
)

var _ = Describe("DataVerify", func() {

	var _ = Describe("CompareDatasetDatums", func() {

		var datasetOne = []map[string]interface{}{}
		var datasetTwo = []map[string]interface{}{}

		BeforeEach(func() {

			datasetOne = []map[string]interface{}{
				{
					"one":   1,
					"value": 2,
				},
				{
					"three": 3,
					"more":  true,
				},
			}

			datasetTwo = []map[string]interface{}{
				{
					"one":   "one",
					"value": 2,
				},
				{
					"three": 3,
					"more":  false,
				},
			}

		})

		It("will genterate a list of differences between two datasets", func() {
			changes, err := verify.CompareDatasetDatums(datasetOne, datasetTwo)
			Expect(err).To(BeNil())
			Expect(changes).ToNot(BeEmpty())
		})

		It("will genterate no differences when the datasets are the same ", func() {
			changes, err := verify.CompareDatasetDatums(datasetOne, datasetOne)
			Expect(err).To(BeNil())
			Expect(changes).To(BeEmpty())
		})

		It("changes will contain each diff", func() {
			changes, err := verify.CompareDatasetDatums(datasetOne, datasetTwo)
			Expect(err).To(BeNil())
			Expect(changes).To(Equal(map[string]interface{}{
				"platform_0": diff.Changelog{{Type: diff.UPDATE, Path: []string{"one"}, From: 1, To: "one"}},
				"platform_1": diff.Changelog{{Type: diff.UPDATE, Path: []string{"more"}, From: true, To: false}},
			}))
		})

		It("can filter based on path", func() {
			changes, err := verify.CompareDatasetDatums(datasetOne, datasetTwo, "more")
			Expect(err).To(BeNil())
			Expect(changes).To(Equal(map[string]interface{}{
				"platform_0": diff.Changelog{{Type: diff.UPDATE, Path: []string{"one"}, From: 1, To: "one"}},
			}))
		})

		It("can filter multiple based on path", func() {
			changes, err := verify.CompareDatasetDatums(datasetOne, datasetTwo, "more", "one")
			Expect(err).To(BeNil())
			Expect(changes).To(BeEmpty())
		})

	})
	var _ = Describe("CompareDatasets", func() {

		It("will have no differences when that same and no dups", func() {
			dSetDifference := verify.CompareDatasets(test.JFBolusSet, test.JFBolusSet)
			Expect(len(dSetDifference[verify.PlatformDuplicate])).To(Equal(0))
			Expect(len(dSetDifference[verify.PlatformExtra])).To(Equal(0))
			Expect(len(dSetDifference[verify.PlatformMissing])).To(Equal(0))
		})

		It("will find duplicates in the platform dataset", func() {
			dSetDifference := verify.CompareDatasets(test.PlatformBolusSet, test.JFBolusSet)
			Expect(len(dSetDifference[verify.PlatformDuplicate])).To(Equal(395))
			Expect(len(dSetDifference[verify.PlatformExtra])).To(Equal(0))
			Expect(len(dSetDifference[verify.PlatformMissing])).To(Equal(0))
		})

		It("will find extras in the platform dataset that have duplicate timestamp but not data", func() {
			duplicateTimeStamp := map[string]interface{}{
				"extra":      true,
				"deviceTime": "2018-01-03T13:07:10",
			}

			dSetDifference := verify.CompareDatasets(append(test.PlatformBolusSet, duplicateTimeStamp), test.JFBolusSet)
			Expect(len(dSetDifference[verify.PlatformDuplicate])).To(Equal(395))
			Expect(len(dSetDifference[verify.PlatformExtra])).To(Equal(1))
			Expect(len(dSetDifference[verify.PlatformMissing])).To(Equal(0))
		})

		It("will find extras in the platform dataset", func() {
			expectedExtra := map[string]interface{}{
				"extra":      3,
				"deviceTime": "2023-01-18T12:00:00",
			}

			dSetDifference := verify.CompareDatasets(append(test.PlatformBolusSet, expectedExtra), test.JFBolusSet)
			Expect(len(dSetDifference[verify.PlatformDuplicate])).To(Equal(395))
			Expect(len(dSetDifference[verify.PlatformExtra])).To(Equal(1))
			Expect(dSetDifference[verify.PlatformExtra][0]).To(Equal(expectedExtra))
			Expect(len(dSetDifference[verify.PlatformMissing])).To(Equal(0))
		})

		It("will find datums that are missing in the platform dataset", func() {
			platformBasals := test.GetPlatformBasalData()
			jellyfishBasals := test.GetJFBasalData()

			Expect(len(platformBasals)).To(Equal(3123))
			Expect(len(jellyfishBasals)).To(Equal(3386))

			dSetDifference := verify.CompareDatasets(platformBasals, jellyfishBasals)
			Expect(len(dSetDifference[verify.PlatformDuplicate])).To(Equal(5))
			Expect(len(dSetDifference[verify.PlatformExtra])).To(Equal(4))
			Expect(len(dSetDifference[verify.PlatformMissing])).To(Equal(263))
		})

	})
})
