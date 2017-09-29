package page_test

// import (
// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"

// 	"github.com/tidepool-org/platform/page"
// )

// var _ = Describe("Pagination", func() {
// 	Context("NewPagination", func() {
// 		It("successfully returns a new pagination", func() {
// 			Expect(page.NewPagination()).ToNot(BeNil())
// 		})

// 		It("returns the defaults", func() {
// 			pagination := page.NewPagination()
// 			Expect(pagination).ToNot(BeNil())
// 			Expect(pagination.Page).To(Equal(0))
// 			Expect(pagination.Size).To(Equal(100))
// 		})
// 	})

// 	Context("with a new pagination", func() {
// 		var pagination *page.Pagination

// 		BeforeEach(func() {
// 			pagination = page.NewPagination()
// 			Expect(pagination).ToNot(BeNil())
// 		})

// 		Context("Validate", func() {
// 			It("succeeds with defaults", func() {
// 				Expect(pagination.Validate()).To(Succeed())
// 			})

// 			Context("Page", func() {
// 				It("succeeds if Page is 0", func() {
// 					pagination.Page = 0
// 					Expect(pagination.Validate()).To(Succeed())
// 				})

// 				It("succeeds if Page is 1000000", func() {
// 					pagination.Page = 1000000
// 					Expect(pagination.Validate()).To(Succeed())
// 				})

// 				It("returns an error if Page is -1", func() {
// 					pagination.Page = -1
// 					Expect(pagination.Validate()).To(MatchError("page is invalid"))
// 				})
// 			})

// 			Context("Size", func() {
// 				It("succeeds if Size is 1", func() {
// 					pagination.Size = 1
// 					Expect(pagination.Validate()).To(Succeed())
// 				})

// 				It("succeeds if Size is 100", func() {
// 					pagination.Size = 100
// 					Expect(pagination.Validate()).To(Succeed())
// 				})

// 				It("returns an error if Size is 0", func() {
// 					pagination.Size = 0
// 					Expect(pagination.Validate()).To(MatchError("size is invalid"))
// 				})

// 				It("returns an error if Size is 101", func() {
// 					pagination.Size = 101
// 					Expect(pagination.Validate()).To(MatchError("size is invalid"))
// 				})
// 			})
// 		})
// 	})
// })
