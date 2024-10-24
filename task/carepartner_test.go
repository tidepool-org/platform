package task

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NewCarePartnerTaskCreate", func() {
	It("succeeds", func() {
		Expect(func() {
			Expect(NewCarePartnerTaskCreate()).ToNot(Equal(nil))
		}).ToNot(Panic())
	})
})
