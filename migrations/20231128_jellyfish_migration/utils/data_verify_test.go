package utils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/r3labs/diff/v3"

	"github.com/tidepool-org/platform/migrations/20231128_jellyfish_migration/utils"
)

var _ = Describe("CompareDatasets", func() {

	datasetOne := []map[string]interface{}{
		{
			"one":   1,
			"value": 2,
		},
		{
			"three": 3,
			"more":  true,
		},
	}
	datasetTwo := []map[string]interface{}{
		{
			"one":   "one",
			"value": 2,
		},
		{
			"three": 3,
			"more":  false,
		},
	}

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
