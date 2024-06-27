package utils_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/migrations/20231128_jellyfish_migration/utils"
	"github.com/tidepool-org/platform/migrations/20231128_jellyfish_migration/utils/test"
)

var _ = Describe("CompareDatasets", func() {

	It("will genterate a list of differences between two datasets", func() {
		jfDataset := test.BulkJellyfishData("test-device-88x89", "test-group-id", "test-user-id-123", 2)
		platformDataset := test.BulkJellyfishData("test-device-88x89", "test-group-id_2", "test-user-id-987", 2)
		changes, err := utils.CompareDatasets(jfDataset, platformDataset)
		Expect(err).To(BeNil())
		Expect(changes).ToNot(BeEmpty())
	})

	It("will genterate no differences when the datasets are the same ", func() {
		jfDataset := test.BulkJellyfishData("test-device-88x89", "test-group-id", "test-user-id-123", 100)
		changes, err := utils.CompareDatasets(jfDataset, jfDataset)
		Expect(err).To(BeNil())
		Expect(changes).To(BeEmpty())
	})

	It("will not compare defined fields", func() {
		jfDataset := test.BulkJellyfishData("test-device-88x89", "test-group-id", "test-user-id-123", 10)

		datasetCopy := []map[string]interface{}{}

		for _, datum := range jfDataset {
			datum["_groupId"] = fmt.Sprintf("%v_zz2", datum["_groupId"])
			datum["_userId"] = fmt.Sprintf("%v_99y", datum["_userId"])
			datasetCopy = append(datasetCopy, datum)
		}

		changes, err := utils.CompareDatasets(jfDataset, datasetCopy)
		Expect(err).To(BeNil())
		Expect(changes).To(BeEmpty())
	})

})
