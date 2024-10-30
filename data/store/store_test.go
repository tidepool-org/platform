package store_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/store"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
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

			It("use expected defaults", func() {
				Expect(filter.Deleted).To(BeFalse())
			})

			Context("Parse", func() {
				It("parses deleted missing", func() {
					object := map[string]interface{}{}
					parser := structureParser.NewObject(logTest.NewLogger(), &object)
					filter.Parse(parser)
					Expect(filter.Deleted).To(BeFalse())
					Expect(parser.Error()).ToNot(HaveOccurred())
				})

				It("parses deleted false", func() {
					object := map[string]interface{}{"deleted": false}
					parser := structureParser.NewObject(logTest.NewLogger(), &object)
					filter.Parse(parser)
					Expect(filter.Deleted).To(BeFalse())
					Expect(parser.Error()).ToNot(HaveOccurred())
				})

				It("parses deleted true", func() {
					object := map[string]interface{}{"deleted": true}
					parser := structureParser.NewObject(logTest.NewLogger(), &object)
					filter.Parse(parser)
					Expect(filter.Deleted).To(BeTrue())
					Expect(parser.Error()).ToNot(HaveOccurred())
				})

				It("reports an error if page is not a bool", func() {
					object := map[string]interface{}{"deleted": "invalid"}
					parser := structureParser.NewObject(logTest.NewLogger(), &object)
					filter.Parse(parser)
					Expect(filter.Deleted).To(BeFalse())
					Expect(parser.Error()).To(HaveOccurred())
					errorsTest.ExpectEqual(parser.Error(), errorsTest.WithPointerSource(structureParser.ErrorTypeNotBool("invalid"), "/deleted"))
				})
			})

			Context("Validate", func() {
				var validator *structureValidator.Validator

				BeforeEach(func() {
					validator = structureValidator.New(logTest.NewLogger())
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
