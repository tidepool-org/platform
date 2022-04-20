package store_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/store"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
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
					errorsTest.ExpectEqual(parser.Error(), errorsTest.WithPointerSource(structureParser.ErrorTypeNotBool("invalid"), "/deleted"))
				})

				It("parses state missing", func() {
					object := map[string]interface{}{}
					parser := structureParser.NewObject(&object)
					filter.Parse(parser)
					Expect(filter.State).To(BeNil())
					Expect(parser.Error()).ToNot(HaveOccurred())
				})

				It("parses state open", func() {
					object := map[string]interface{}{"state": "open"}
					parser := structureParser.NewObject(&object)
					filter.Parse(parser)
					Expect(filter.State).ToNot(BeNil())
					Expect(*filter.State).To(Equal("open"))
					Expect(parser.Error()).ToNot(HaveOccurred())
				})

				It("parses state closed", func() {
					object := map[string]interface{}{"state": "closed"}
					parser := structureParser.NewObject(&object)
					filter.Parse(parser)
					Expect(filter.State).ToNot(BeNil())
					Expect(*filter.State).To(Equal("closed"))
					Expect(parser.Error()).ToNot(HaveOccurred())
				})

				It("reports an error if state is not a string", func() {
					object := map[string]interface{}{"state": true}
					parser := structureParser.NewObject(&object)
					filter.Parse(parser)
					Expect(filter.State).To(BeNil())
					Expect(parser.Error()).To(HaveOccurred())
				})

				It("parses dataSetType missing", func() {
					object := map[string]interface{}{}
					parser := structureParser.NewObject(&object)
					filter.Parse(parser)
					Expect(filter.DataSetType).To(BeNil())
					Expect(parser.Error()).ToNot(HaveOccurred())
				})

				It("parses dataSetType continuous", func() {
					object := map[string]interface{}{"dataSetType": "continuous"}
					parser := structureParser.NewObject(&object)
					filter.Parse(parser)
					Expect(filter.DataSetType).ToNot(BeNil())
					Expect(*filter.DataSetType).To(Equal("continuous"))
					Expect(parser.Error()).ToNot(HaveOccurred())
				})

				It("parses dataSetType normal", func() {
					object := map[string]interface{}{"dataSetType": "normal"}
					parser := structureParser.NewObject(&object)
					filter.Parse(parser)
					Expect(filter.DataSetType).ToNot(BeNil())
					Expect(*filter.DataSetType).To(Equal("normal"))
					Expect(parser.Error()).ToNot(HaveOccurred())
				})

				It("reports an error if dataSetType is not a string", func() {
					object := map[string]interface{}{"dataSetType": true}
					parser := structureParser.NewObject(&object)
					filter.Parse(parser)
					Expect(filter.DataSetType).To(BeNil())
					Expect(parser.Error()).To(HaveOccurred())
				})
			})

			Context("Validate", func() {
				var validator *structureValidator.Validator

				BeforeEach(func() {
					validator = structureValidator.New()
				})

				It("succeeds with default", func() {
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

				It("succeeds with state open", func() {
					filter.State = pointer.FromString("open")
					filter.Validate(validator)
					Expect(validator.Error()).ToNot(HaveOccurred())
				})

				It("succeeds with state closed", func() {
					filter.State = pointer.FromString("closed")
					filter.Validate(validator)
					Expect(validator.Error()).ToNot(HaveOccurred())
				})

				It("failed with state invalid", func() {
					filter.State = pointer.FromString("invalid")
					filter.Validate(validator)
					Expect(validator.Error()).To(HaveOccurred())
				})

				It("succeeds with dataSetType normal", func() {
					filter.DataSetType = pointer.FromString("normal")
					filter.Validate(validator)
					Expect(validator.Error()).ToNot(HaveOccurred())
				})

				It("succeeds with dataSetType continuous", func() {
					filter.DataSetType = pointer.FromString("continuous")
					filter.Validate(validator)
					Expect(validator.Error()).ToNot(HaveOccurred())
				})

				It("failed with dataSetType invalid", func() {
					filter.DataSetType = pointer.FromString("invalid")
					filter.Validate(validator)
					Expect(validator.Error()).To(HaveOccurred())
				})
			})
		})
	})
})
