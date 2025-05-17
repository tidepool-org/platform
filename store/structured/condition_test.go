package structured_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	requestTest "github.com/tidepool-org/platform/request/test"
	storeStructured "github.com/tidepool-org/platform/store/structured"
	storeStructuredTest "github.com/tidepool-org/platform/store/structured/test"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Condition", func() {
	Context("NewCondition", func() {
		It("returns successfully with default values", func() {
			datum := request.NewCondition()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Revision).To(BeNil())
		})
	})

	Context("MapCondition", func() {
		It("returns nil if the condition is nil", func() {
			Expect(storeStructured.MapCondition(nil)).To(BeNil())
		})

		It("returns a condition with the expected values", func() {
			requestCondition := requestTest.RandomCondition()
			condition := storeStructured.MapCondition(requestCondition)
			Expect(condition).ToNot(BeNil())
			Expect(condition.Revision).To(Equal(requestCondition.Revision))
		})
	})

	Context("Condition", func() {
		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *storeStructured.Condition), expectedErrors ...error) {
					expectedDatum := storeStructuredTest.RandomCondition()
					object := storeStructuredTest.NewObjectFromCondition(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &storeStructured.Condition{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *storeStructured.Condition) {},
				),
				Entry("revision invalid type",
					func(object map[string]interface{}, expectedDatum *storeStructured.Condition) {
						object["revision"] = true
						expectedDatum.Revision = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/revision"),
				),
				Entry("revision valid",
					func(object map[string]interface{}, expectedDatum *storeStructured.Condition) {
						valid := storeStructuredTest.RandomRevision()
						object["revision"] = valid
						expectedDatum.Revision = pointer.FromInt(valid)
					},
				),
				Entry("multiple",
					func(object map[string]interface{}, expectedDatum *storeStructured.Condition) {
						object["revision"] = true
						expectedDatum.Revision = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/revision"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *storeStructured.Condition), expectedErrors ...error) {
					datum := storeStructuredTest.RandomCondition()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *storeStructured.Condition) {},
				),
				Entry("revision missing",
					func(datum *storeStructured.Condition) { datum.Revision = nil },
				),
				Entry("revision out of range (lower)",
					func(datum *storeStructured.Condition) {
						datum.Revision = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/revision"),
				),
				Entry("revision in range (lower)",
					func(datum *storeStructured.Condition) {
						datum.Revision = pointer.FromInt(0)
					},
				),
				Entry("multiple errors",
					func(datum *storeStructured.Condition) {
						datum.Revision = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/revision"),
				),
			)
		})
	})
})
