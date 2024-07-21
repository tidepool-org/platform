package utils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/r3labs/diff/v3"

	"github.com/tidepool-org/platform/migrations/20231128_jellyfish_migration/utils"
	"github.com/tidepool-org/platform/migrations/20231128_jellyfish_migration/utils/test"
)

var _ = Describe("DataVerify", func() {

	var _ = Describe("CompareDatasets", func() {

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
			changes, err := utils.CompareDatasets(datasetOne, datasetTwo)
			Expect(err).To(BeNil())
			Expect(changes).ToNot(BeEmpty())
		})

		It("will genterate no differences when the datasets are the same ", func() {
			changes, err := utils.CompareDatasets(datasetOne, datasetOne)
			Expect(err).To(BeNil())
			Expect(changes).To(BeEmpty())
		})

		It("changes will contain each diff", func() {
			changes, err := utils.CompareDatasets(datasetOne, datasetTwo)
			Expect(err).To(BeNil())
			Expect(changes).To(Equal(map[string]interface{}{
				"platform_0": diff.Changelog{{Type: diff.UPDATE, Path: []string{"one"}, From: 1, To: "one"}},
				"platform_1": diff.Changelog{{Type: diff.UPDATE, Path: []string{"more"}, From: true, To: false}},
			}))
		})

		It("can filter based on path", func() {
			changes, err := utils.CompareDatasets(datasetOne, datasetTwo, "more")
			Expect(err).To(BeNil())
			Expect(changes).To(Equal(map[string]interface{}{
				"platform_0": diff.Changelog{{Type: diff.UPDATE, Path: []string{"one"}, From: 1, To: "one"}},
			}))
		})

		It("can filter multiple based on path", func() {
			changes, err := utils.CompareDatasets(datasetOne, datasetTwo, "more", "one")
			Expect(err).To(BeNil())
			Expect(changes).To(BeEmpty())
		})

	})
	var _ = Describe("GetMissing", func() {

		var dOne = []map[string]interface{}{}
		var dLarge = []map[string]interface{}{}
		var dLargeTwo = []map[string]interface{}{}

		BeforeEach(func() {

			dOne = []map[string]interface{}{
				{
					"one":        1,
					"value":      2,
					"deviceTime": "2023-01-18T00:00:00",
				},
				{
					"three":      3,
					"more":       true,
					"deviceTime": "2023-01-18T01:00:00",
				},
			}

			dLargeTwo = []map[string]interface{}{
				{
					"one":        1,
					"value":      2,
					"deviceTime": "2023-01-18T00:00:00",
				},
				{
					"three":      3,
					"more":       true,
					"deviceTime": "2023-01-18T01:00:00",
				},
				{
					"four":       44,
					"more":       true,
					"deviceTime": "2023-01-18T02:00:00",
				},
			}

			dLarge = test.BulkJellyfishUploadData("test-device-id", "group-id", "user-id", 2112)

		})

		It("will be empty when the two datasets match for large amount of data ", func() {
			missing := utils.GetMissing(dLarge, dLarge)
			Expect(missing).To(BeEmpty())
		})

		It("will return the missing datum when no match", func() {
			missing := utils.GetMissing(dOne, dLargeTwo)
			Expect(missing).To(Equal([]map[string]interface{}{{
				"four":       44,
				"more":       true,
				"deviceTime": "2023-01-18T02:00:00",
			}}))
		})

		var _ = Describe("order of datasets", func() {

			It("shows missing if largest set first", func() {
				dLargeTwo = []map[string]interface{}{}
				for i := 10; i < len(dLarge); i++ {
					dLargeTwo = append(dLargeTwo, dLarge[i])
				}
				missing := utils.GetMissing(dLarge, dLargeTwo)
				Expect(len(missing)).To(Equal(10))
			})

			It("shows missing if largest set second", func() {
				dLargeTwo = []map[string]interface{}{}
				for i := 10; i < len(dLarge); i++ {
					dLargeTwo = append(dLargeTwo, dLarge[i])
				}
				missing := utils.GetMissing(dLargeTwo, dLarge)
				Expect(len(missing)).To(Equal(10))

			})
		})

	})
})
