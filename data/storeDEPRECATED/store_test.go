package storeDEPRECATED_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/storeDEPRECATED"
	testErrors "github.com/tidepool-org/platform/errors/test"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
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

			It("use expected defaults", func() {
				Expect(filter.Deleted).To(BeFalse())
			})

			Context("Parse", func() {
				It("parses deleted missing", func() {
					object := map[string]interface{}{}
					parser := structureParser.NewObject(&object)
					filter.Parse(parser)
					Expect(filter.Deleted).To(BeFalse())
					Expect(parser.Error()).ToNot(HaveOccurred())
				})

				It("parses deleted false", func() {
					object := map[string]interface{}{"deleted": false}
					parser := structureParser.NewObject(&object)
					filter.Parse(parser)
					Expect(filter.Deleted).To(BeFalse())
					Expect(parser.Error()).ToNot(HaveOccurred())
				})

				It("parses deleted true", func() {
					object := map[string]interface{}{"deleted": true}
					parser := structureParser.NewObject(&object)
					filter.Parse(parser)
					Expect(filter.Deleted).To(BeTrue())
					Expect(parser.Error()).ToNot(HaveOccurred())
				})

				It("reports an error if page is not a bool", func() {
					object := map[string]interface{}{"deleted": "invalid"}
					parser := structureParser.NewObject(&object)
					filter.Parse(parser)
					Expect(filter.Deleted).To(BeFalse())
					Expect(parser.Error()).To(HaveOccurred())
					testErrors.ExpectEqual(parser.Error(), testErrors.WithPointerSource(structureParser.ErrorTypeNotBool("invalid"), "/deleted"))
				})
			})

			Context("Validate", func() {
				var validator *structureValidator.Validator

				BeforeEach(func() {
					validator = structureValidator.New()
				})

				It("succeeds with deleted default", func() {
					filter.Validate(validator)
					Expect(validator.Error()).ToNot(HaveOccurred())
				})

				It("succeeds with deleted false", func() {
					filter.Deleted = false
					filter.Validate(validator)
					Expect(validator.Error()).ToNot(HaveOccurred())
				})

				It("succeeds with deleted true", func() {
					filter.Deleted = true
					filter.Validate(validator)
					Expect(validator.Error()).ToNot(HaveOccurred())
				})
			})
		})
	})
})
