package page_test

import (
	"net/http"
	"net/url"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/page"
	pageTest "github.com/tidepool-org/platform/page/test"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Pagination", func() {
	It("PaginationPageDefault is expected", func() {
		Expect(page.PaginationPageDefault).To(Equal(0))
	})

	It("PaginationPageMinimum is expected", func() {
		Expect(page.PaginationPageMinimum).To(Equal(0))
	})

	It("PaginationSizeDefault is expected", func() {
		Expect(page.PaginationSizeDefault).To(Equal(100))
	})

	It("PaginationSizeMinimum is expected", func() {
		Expect(page.PaginationSizeMinimum).To(Equal(1))
	})

	It("PaginationSizeMaximum is expected", func() {
		Expect(page.PaginationSizeMaximum).To(Equal(1000))
	})

	Context("Pagination", func() {
		Context("NewPagination", func() {
			It("returns successfully with default values", func() {
				datum := page.NewPagination()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Page).To(Equal(0))
				Expect(datum.Size).To(Equal(100))
			})
		})

		Context("with a new pagination", func() {
			var datum *page.Pagination
			var original *page.Pagination

			BeforeEach(func() {
				datum = pageTest.RandomPagination()
				original = pageTest.ClonePagination(datum)
			})

			// TODO: Use current Parse test mechanism

			Context("Parse", func() {
				It("parses the page", func() {
					object := map[string]interface{}{"page": 2}
					parser := structureParser.NewObject(&object)
					datum.Parse(parser)
					Expect(datum.Page).To(Equal(2))
					Expect(datum.Size).To(Equal(original.Size))
					Expect(parser.Error()).ToNot(HaveOccurred())
				})

				It("parses the size", func() {
					object := map[string]interface{}{"size": 10}
					parser := structureParser.NewObject(&object)
					datum.Parse(parser)
					Expect(datum.Page).To(Equal(original.Page))
					Expect(datum.Size).To(Equal(10))
					Expect(parser.Error()).ToNot(HaveOccurred())
				})

				It("parses the page and size", func() {
					object := map[string]interface{}{"page": 2, "size": 10}
					parser := structureParser.NewObject(&object)
					datum.Parse(parser)
					Expect(datum.Page).To(Equal(2))
					Expect(datum.Size).To(Equal(10))
					Expect(parser.Error()).ToNot(HaveOccurred())
				})

				It("reports an error if page is not an int", func() {
					object := map[string]interface{}{"page": false, "size": 10}
					parser := structureParser.NewObject(&object)
					datum.Parse(parser)
					Expect(datum.Page).To(Equal(original.Page))
					Expect(datum.Size).To(Equal(10))
					Expect(parser.Error()).To(HaveOccurred())
					errorsTest.ExpectEqual(parser.Error(), errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(false), "/page"))
				})

				It("reports an error if size is not an int", func() {
					object := map[string]interface{}{"page": 2, "size": false}
					parser := structureParser.NewObject(&object)
					datum.Parse(parser)
					Expect(datum.Page).To(Equal(2))
					Expect(datum.Size).To(Equal(original.Size))
					Expect(parser.Error()).To(HaveOccurred())
					errorsTest.ExpectEqual(parser.Error(), errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(false), "/size"))
				})

				It("reports an error if page and size are not ints", func() {
					object := map[string]interface{}{"page": false, "size": false}
					parser := structureParser.NewObject(&object)
					datum.Parse(parser)
					Expect(datum.Page).To(Equal(original.Page))
					Expect(datum.Size).To(Equal(original.Size))
					Expect(parser.Error()).To(HaveOccurred())
					errorsTest.ExpectEqual(parser.Error(), errors.Append(
						errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(false), "/page"),
						errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(false), "/size"),
					))
				})
			})

			// TODO: Use current Validate test mechanism

			Context("Validate", func() {
				var validator *structureValidator.Validator

				BeforeEach(func() {
					validator = structureValidator.New()
				})

				It("succeeds with defaults", func() {
					datum.Validate(validator)
					Expect(validator.Error()).ToNot(HaveOccurred())
				})

				It("reports an error if the page is less than zero", func() {
					datum.Page = -1
					datum.Validate(validator)
					Expect(validator.Error()).To(HaveOccurred())
					errorsTest.ExpectEqual(validator.Error(), errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/page"))
				})

				It("reports an error if the size is less than 1", func() {
					datum.Size = 0
					datum.Validate(validator)
					Expect(validator.Error()).To(HaveOccurred())
					errorsTest.ExpectEqual(validator.Error(), errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(0, 1, 1000), "/size"))
				})

				It("reports an error if the size is greater than 1000", func() {
					datum.Size = 1001
					datum.Validate(validator)
					Expect(validator.Error()).To(HaveOccurred())
					errorsTest.ExpectEqual(validator.Error(), errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1001, 1, 1000), "/size"))
				})

				It("reports an error if the page and size are less than minimum", func() {
					datum.Page = -1
					datum.Size = 0
					datum.Validate(validator)
					Expect(validator.Error()).To(HaveOccurred())
					errorsTest.ExpectEqual(validator.Error(), errors.Append(
						errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/page"),
						errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(0, 1, 1000), "/size"),
					))
				})
			})

			Context("MutateRequest", func() {
				var req *http.Request

				BeforeEach(func() {
					req = testHttp.NewRequest()
				})

				It("returns an error if the request is missing", func() {
					Expect(datum.MutateRequest(nil)).To(MatchError("request is missing"))
				})

				It("does not adds default page and size to the request as query parameters", func() {
					datum = page.NewPagination()
					Expect(datum.MutateRequest(req)).To(Succeed())
					Expect(req.URL.Query()).To(BeEmpty())
				})

				It("adds custom page and size to the request as query parameters", func() {
					Expect(datum.MutateRequest(req)).To(Succeed())
					Expect(req.URL.Query()).To(Equal(url.Values{"page": []string{strconv.Itoa(datum.Page)}, "size": []string{strconv.Itoa(datum.Size)}}))
				})
			})
		})
	})
})
