package work_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	authTest "github.com/tidepool-org/platform/auth/test"
	ouraWork "github.com/tidepool-org/platform/oura/work"
)

var _ = Describe("work", func() {
	It("Domain is expected", func() {
		Expect(ouraWork.Domain).To(Equal("org.tidepool.oura"))
	})

	Context("GroupIDFromProviderSessionID", func() {
		It("returns expected", func() {
			providerSessionID := authTest.RandomProviderSessionID()
			Expect(ouraWork.GroupIDFromProviderSessionID(providerSessionID)).To(Equal(ouraWork.Domain + ":" + providerSessionID))
		})
	})
})
