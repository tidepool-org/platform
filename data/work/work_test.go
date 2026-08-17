package work_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataWork "github.com/tidepool-org/platform/data/work"
)

var _ = Describe("ingestion_offset", func() {
	It("MetadataKeyDeviceHashes expected value", func() {
		Expect(dataWork.MetadataKeyDeviceHashes).To(Equal("deviceHashes"))
	})

	It("MetadataKeyDataIngestionOffset expected value", func() {
		Expect(dataWork.MetadataKeyIngestionOffset).To(Equal("ingestionOffset"))
	})
})
