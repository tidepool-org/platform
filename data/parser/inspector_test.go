package parser_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/parser"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/service"
)

var _ = Describe("Inspector", func() {
	Context("NewObjectParserInspector", func() {
		It("returns an error when the parser is missing", func() {
			inspector, err := parser.NewObjectParserInspector(nil)
			Expect(err).To(MatchError("parser: parser is missing"))
			Expect(inspector).To(BeNil())
		})
	})

	Context("with an object parser", func() {
		var testObjectParser *TestObjectParser

		BeforeEach(func() {
			testObjectParser = &TestObjectParser{}
		})

		It("successfully returns a new inspector", func() {
			Expect(parser.NewObjectParserInspector(testObjectParser)).ToNot(BeNil())
		})

		Context("with an inspector", func() {
			var inspector *parser.ObjectParserInspector

			BeforeEach(func() {
				var err error
				inspector, err = parser.NewObjectParserInspector(testObjectParser)
				Expect(err).ToNot(HaveOccurred())
				Expect(inspector).ToNot(BeNil())
			})

			Context("GetProperty", func() {
				It("returns nil if the object parser returns nil", func() {
					testObjectParser.ParseStringOutputs = []*string{nil}
					Expect(inspector.GetProperty("test-key")).To(BeNil())
					Expect(testObjectParser.ParseStringInputs).To(ConsistOf("test-key"))
				})

				It("returns the value the object parser returns", func() {
					testObjectParser.ParseStringOutputs = []*string{pointer.String("test-value")}
					value := inspector.GetProperty("test-key")
					Expect(value).ToNot(BeNil())
					Expect(*value).To(Equal("test-value"))
					Expect(testObjectParser.ParseStringInputs).To(ConsistOf("test-key"))
				})
			})

			Context("NewMissingPropertyError", func() {
				It("appends an error to the object parser", func() {
					Expect(inspector.NewMissingPropertyError("test-key")).To(Succeed())
					Expect(testObjectParser.AppendErrorInputs).To(ConsistOf(AppendErrorInput{"test-key", service.ErrorValueNotExists()}))
				})
			})

			Context("NewInvalidPropertyError", func() {
				It("appends an error to the object parser", func() {
					Expect(inspector.NewInvalidPropertyError("test-key", "test-value", []string{"test-value-1", "test-value-2"})).To(Succeed())
					Expect(testObjectParser.AppendErrorInputs).To(ConsistOf(AppendErrorInput{"test-key", service.ErrorValueStringNotOneOf("test-value", []string{"test-value-1", "test-value-2"})}))
				})
			})
		})
	})
})
