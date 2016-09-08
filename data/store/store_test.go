package store_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/store"
)

var _ = Describe("Store", func() {
	Context("Filter", func() {
		Context("NewFilter", func() {
			It("successfully returns a new filter", func() {
				Expect(store.NewFilter()).ToNot(BeNil())
			})
		})

		Context("with a new filter", func() {
			var filter *store.Filter

			BeforeEach(func() {
				filter = store.NewFilter()
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

	Context("Pagination", func() {
		Context("NewPagination", func() {
			It("successfully returns a new pagination", func() {
				Expect(store.NewPagination()).ToNot(BeNil())
			})
		})

		Context("with a new pagination", func() {
			var pagination *store.Pagination

			BeforeEach(func() {
				pagination = store.NewPagination()
				Expect(pagination).ToNot(BeNil())
			})

			Context("Validate", func() {
				It("succeeds with defaults", func() {
					Expect(pagination.Validate()).To(Succeed())
				})

				Context("Page", func() {
					It("succeeds if Page is 0", func() {
						pagination.Page = 0
						Expect(pagination.Validate()).To(Succeed())
					})

					It("succeeds if Page is 1000000", func() {
						pagination.Page = 1000000
						Expect(pagination.Validate()).To(Succeed())
					})

					It("returns an error if Page is -1", func() {
						pagination.Page = -1
						Expect(pagination.Validate()).To(MatchError("store: page is invalid"))
					})
				})

				Context("Size", func() {
					It("succeeds if Size is 1", func() {
						pagination.Size = 1
						Expect(pagination.Validate()).To(Succeed())
					})

					It("succeeds if Size is 100", func() {
						pagination.Size = 100
						Expect(pagination.Validate()).To(Succeed())
					})

					It("returns an error if Size is 0", func() {
						pagination.Size = 0
						Expect(pagination.Validate()).To(MatchError("store: size is invalid"))
					})

					It("returns an error if Size is 101", func() {
						pagination.Size = 101
						Expect(pagination.Validate()).To(MatchError("store: size is invalid"))
					})
				})
			})
		})
	})
})
