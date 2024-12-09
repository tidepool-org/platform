package fetch_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dexcom/fetch"
)

var _ = Describe("Task", func() {
	Context("NewTaskCreate", func() {
		const providerID = "some-provider-id"
		const sourceID = "some-source-id"

		It("returns an error when provider session id not set", func() {
			tc, err := fetch.NewTaskCreate("", sourceID)
			Expect(err).ToNot(BeNil())
			Expect(tc).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("provider session id is missing"))
		})
		It("returns an error when data source id not set", func() {
			tc, err := fetch.NewTaskCreate(providerID, "")
			Expect(err).ToNot(BeNil())
			Expect(tc).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("data source id is missing"))
		})
		It("returns an initialized task create", func() {
			tc, err := fetch.NewTaskCreate(providerID, sourceID)
			Expect(err).To(BeNil())
			Expect(tc).ToNot(BeNil())
		})

		It("task has data initialized", func() {
			tc, _ := fetch.NewTaskCreate(providerID, sourceID)
			Expect(tc).ToNot(BeNil())
			Expect(tc.Type).To(Equal(fetch.Type))
			Expect(tc.Data["providerSessionId"]).To(Equal(providerID))
			Expect(tc.Data["dataSourceId"]).To(Equal(sourceID))
		})
	})
})
