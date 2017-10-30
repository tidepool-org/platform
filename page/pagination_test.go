package page_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"net/http"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/page"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	testHTTP "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Pagination", func() {
	Context("NewPagination", func() {
		It("successfully returns a new pagination", func() {
			Expect(page.NewPagination()).ToNot(BeNil())
		})

		It("returns the defaults", func() {
			pagination := page.NewPagination()
			Expect(pagination).ToNot(BeNil())
			Expect(pagination.Page).To(Equal(0))
			Expect(pagination.Size).To(Equal(100))
		})
	})

	Context("with a new pagination", func() {
		var pagination *page.Pagination

		BeforeEach(func() {
			pagination = page.NewPagination()
			Expect(pagination).ToNot(BeNil())
		})

		Context("Parse", func() {
			It("parses the page", func() {
				object := map[string]interface{}{"page": 2}
				parser := structureParser.NewObject(&object)
				pagination.Parse(parser)
				Expect(pagination.Page).To(Equal(2))
				Expect(pagination.Size).To(Equal(100))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})

			It("parses the size", func() {
				object := map[string]interface{}{"size": 10}
				parser := structureParser.NewObject(&object)
				pagination.Parse(parser)
				Expect(pagination.Page).To(Equal(0))
				Expect(pagination.Size).To(Equal(10))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})

			It("parses the page and size", func() {
				object := map[string]interface{}{"page": 2, "size": 10}
				parser := structureParser.NewObject(&object)
				pagination.Parse(parser)
				Expect(pagination.Page).To(Equal(2))
				Expect(pagination.Size).To(Equal(10))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})

			It("reports an error if page is not an int", func() {
				object := map[string]interface{}{"page": false, "size": 10}
				parser := structureParser.NewObject(&object)
				pagination.Parse(parser)
				Expect(pagination.Page).To(Equal(0))
				Expect(pagination.Size).To(Equal(10))
				Expect(parser.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(parser.Error())).To(Equal(errors.Sanitize(structureParser.ErrorTypeNotInt(false))))
			})

			It("reports an error if size is not an int", func() {
				object := map[string]interface{}{"page": 2, "size": false}
				parser := structureParser.NewObject(&object)
				pagination.Parse(parser)
				Expect(pagination.Page).To(Equal(2))
				Expect(pagination.Size).To(Equal(100))
				Expect(parser.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(parser.Error())).To(Equal(errors.Sanitize(structureParser.ErrorTypeNotInt(false))))
			})

			It("reports an error if page and size are not ints", func() {
				object := map[string]interface{}{"page": false, "size": false}
				parser := structureParser.NewObject(&object)
				pagination.Parse(parser)
				Expect(pagination.Page).To(Equal(0))
				Expect(pagination.Size).To(Equal(100))
				Expect(parser.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(parser.Error())).To(Equal(errors.Sanitize(errors.Append(
					structureParser.ErrorTypeNotInt(false),
					structureParser.ErrorTypeNotInt(false),
				))))
			})
		})

		Context("Validate", func() {
			var validator *structureValidator.Validator

			BeforeEach(func() {
				validator = structureValidator.New()
			})

			It("succeeds with defaults", func() {
				pagination.Validate(validator)
				Expect(validator.Error()).ToNot(HaveOccurred())
			})

			It("reports an error if the page is less than zero", func() {
				pagination.Page = -1
				pagination.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(validator.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0))))
			})

			It("reports an error if the size is less than 1", func() {
				pagination.Size = 0
				pagination.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(validator.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueNotInRange(0, 1, 100))))
			})

			It("reports an error if the size is greater than 100", func() {
				pagination.Size = 101
				pagination.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(validator.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueNotInRange(101, 1, 100))))
			})

			It("reports an error if the page and size are less than minimum", func() {
				pagination.Page = -1
				pagination.Size = 0
				pagination.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(validator.Error())).To(Equal(errors.Sanitize(errors.Append(
					structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0),
					structureValidator.ErrorValueNotInRange(0, 1, 100),
				))))
			})
		})

		Context("Mutate", func() {
			var req *http.Request

			BeforeEach(func() {
				req = testHTTP.NewRequest()
			})

			It("returns an error if the request is missing", func() {
				Expect(pagination.Mutate(nil)).To(MatchError("request is missing"))
			})

			It("adds default page and size to the request as query parameters", func() {
				Expect(pagination.Mutate(req)).To(Succeed())
				Expect(req.URL.Query()).To(And(HaveKeyWithValue("page", []string{"0"}), HaveKeyWithValue("size", []string{"100"})))
			})

			It("adds custom page and size to the request as query parameters", func() {
				pagination.Page = 2
				pagination.Size = 10
				Expect(pagination.Mutate(req)).To(Succeed())
				Expect(req.URL.Query()).To(And(HaveKeyWithValue("page", []string{"2"}), HaveKeyWithValue("size", []string{"10"})))
			})
		})
	})
})
