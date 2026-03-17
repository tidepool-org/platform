package work_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	authTest "github.com/tidepool-org/platform/auth/test"
	ouraDataWork "github.com/tidepool-org/platform/oura/data/work"
)

var _ = Describe("work", func() {
	It("Domain is expected", func() {
		Expect(ouraDataWork.Domain).To(Equal("org.tidepool.oura.data"))
	})

	Context("SerialIDFromProviderSessionID", func() {
		It("returns expected", func() {
			providerSessionID := authTest.RandomProviderSessionID()
			Expect(ouraDataWork.SerialIDFromProviderSessionID(providerSessionID)).To(Equal(ouraDataWork.Domain + ":" + providerSessionID))
		})
	})
})
