package storeDEPRECATED_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/storeDEPRECATED"
)

var _ = Describe("Store", func() {
	Context("Filter", func() {
		Context("NewFilter", func() {
			It("successfully returns a new filter", func() {
				Expect(storeDEPRECATED.NewFilter()).ToNot(BeNil())
			})
		})

		Context("with a new filter", func() {
			var filter *storeDEPRECATED.Filter

			BeforeEach(func() {
				filter = storeDEPRECATED.NewFilter()
				Expect(filter).ToNot(BeNil())
			})

			Context("Validate", func() {
				It("succeeds with defaults", func() {
					Expect(filter.Validate()).To(Succeed())
				})

				Context("Deleted", func() {
					It("succeeds if Deleted is true", func() {
						filter.Deleted = true
						Expect(filter.Validate()).To(Succeed())
					})

					It("succeeds if Deleted is false", func() {
						filter.Deleted = false
						Expect(filter.Validate()).To(Succeed())
					})
				})
			})
		})
	})
})
