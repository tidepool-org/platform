package utils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/r3labs/diff/v3"

	"github.com/tidepool-org/platform/migrations/20231128_jellyfish_migration/utils"
	"github.com/tidepool-org/platform/migrations/20231128_jellyfish_migration/utils/test"
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
			changes, err := utils.CompareDatasetDatums(datasetOne, datasetTwo)
			Expect(err).To(BeNil())
			Expect(changes).ToNot(BeEmpty())
		})

		It("will genterate no differences when the datasets are the same ", func() {
			changes, err := utils.CompareDatasetDatums(datasetOne, datasetOne)
			Expect(err).To(BeNil())
			Expect(changes).To(BeEmpty())
		})

		It("changes will contain each diff", func() {
			changes, err := utils.CompareDatasetDatums(datasetOne, datasetTwo)
			Expect(err).To(BeNil())
			Expect(changes).To(Equal(map[string]interface{}{
				"platform_0": diff.Changelog{{Type: diff.UPDATE, Path: []string{"one"}, From: 1, To: "one"}},
				"platform_1": diff.Changelog{{Type: diff.UPDATE, Path: []string{"more"}, From: true, To: false}},
			}))
		})

		It("can filter based on path", func() {
			changes, err := utils.CompareDatasetDatums(datasetOne, datasetTwo, "more")
			Expect(err).To(BeNil())
			Expect(changes).To(Equal(map[string]interface{}{
				"platform_0": diff.Changelog{{Type: diff.UPDATE, Path: []string{"one"}, From: 1, To: "one"}},
			}))
		})

		It("can filter multiple based on path", func() {
			changes, err := utils.CompareDatasetDatums(datasetOne, datasetTwo, "more", "one")
			Expect(err).To(BeNil())
			Expect(changes).To(BeEmpty())
		})

	})
	var _ = Describe("CompareDatasets", func() {

		It("will have no differences when that same and no dups", func() {
			missing, duplicates, extras := utils.CompareDatasets(test.JFBolusSet, test.JFBolusSet)
			Expect(len(duplicates)).To(Equal(0))
			Expect(len(extras)).To(Equal(0))
			Expect(len(missing)).To(Equal(0))
		})

		It("will find duplicates in the platform dataset", func() {
			missing, duplicates, extras := utils.CompareDatasets(test.PlatformBolusSet, test.JFBolusSet)
			Expect(len(duplicates)).To(Equal(395))
			Expect(len(extras)).To(Equal(0))
			Expect(len(missing)).To(Equal(0))
		})

		It("will find extras in the platform dataset", func() {

			expectedExtra := map[string]interface{}{
				"extra":      3,
				"deviceTime": "2023-01-18T12:00:00",
			}

			missing, duplicates, extras := utils.CompareDatasets(append(test.PlatformBolusSet, expectedExtra), test.JFBolusSet)
			Expect(len(duplicates)).To(Equal(395))
			Expect(len(extras)).To(Equal(1))
			Expect(extras[0]).To(Equal(expectedExtra))
			Expect(len(missing)).To(Equal(0))
		})

		It("will find missing in the platform dataset", func() {

			expectedMissing := map[string]interface{}{
				"missing":    3,
				"deviceTime": "2023-01-18T12:00:00",
			}

			missing, duplicates, extras := utils.CompareDatasets(test.PlatformBolusSet, append(test.JFBolusSet, expectedMissing))
			Expect(len(duplicates)).To(Equal(395))
			Expect(len(extras)).To(Equal(0))
			Expect(len(missing)).To(Equal(1))
			Expect(missing[0]).To(Equal(expectedMissing))
		})

	})
})
