package dosingdecision_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataTypesDosingDecision "github.com/tidepool-org/platform/data/types/dosingdecision"
	dataTypesDosingDecisionTest "github.com/tidepool-org/platform/data/types/dosingdecision/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/metadata"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Issue", func() {
	It("IssueArrayLengthMaximum is expected", func() {
		Expect(dataTypesDosingDecision.IssueArrayLengthMaximum).To(Equal(100))
	})

	It("IssueIDLengthMaximum is expected", func() {
		Expect(dataTypesDosingDecision.IssueIDLengthMaximum).To(Equal(100))
	})

	Context("Issue", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesDosingDecision.Issue)) {
				datum := dataTypesDosingDecisionTest.RandomIssue()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesDosingDecisionTest.NewObjectFromIssue(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesDosingDecisionTest.NewObjectFromIssue(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesDosingDecision.Issue) {},
			),
			Entry("empty",
				func(datum *dataTypesDosingDecision.Issue) {
					*datum = *dataTypesDosingDecision.NewIssue()
				},
			),
			Entry("all",
				func(datum *dataTypesDosingDecision.Issue) {
					datum.ID = pointer.FromString(test.RandomStringFromRange(1, dataTypesDosingDecision.IssueIDLengthMaximum))
					datum.Metadata = metadataTest.RandomMetadata()
				},
			),
		)

		Context("ParseIssue", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataTypesDosingDecision.ParseIssue(structureParser.NewObject(nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataTypesDosingDecisionTest.RandomIssue()
				object := dataTypesDosingDecisionTest.NewObjectFromIssue(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(&object)
				Expect(dataTypesDosingDecision.ParseIssue(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewIssue", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesDosingDecision.NewIssue()
				Expect(datum).ToNot(BeNil())
				Expect(datum.ID).To(BeNil())
				Expect(datum.Metadata).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesDosingDecision.Issue), expectedErrors ...error) {
					expectedDatum := dataTypesDosingDecisionTest.RandomIssue()
					object := dataTypesDosingDecisionTest.NewObjectFromIssue(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesDosingDecision.NewIssue()
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesDosingDecision.Issue) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesDosingDecision.Issue) {
						object["id"] = true
						object["metadata"] = true
						expectedDatum.ID = nil
						expectedDatum.Metadata = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/id"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/metadata"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",
				func(mutator func(datum *dataTypesDosingDecision.Issue), expectedErrors ...error) {
					datum := dataTypesDosingDecisionTest.RandomIssue()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesDosingDecision.Issue) {},
				),
				Entry("id missing",
					func(datum *dataTypesDosingDecision.Issue) { datum.ID = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/id"),
				),
				Entry("id empty",
					func(datum *dataTypesDosingDecision.Issue) { datum.ID = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/id"),
				),
				Entry("id length; in range (upper)",
					func(datum *dataTypesDosingDecision.Issue) {
						datum.ID = pointer.FromString(test.RandomStringFromRange(100, 100))
					},
				),
				Entry("id length; out of range (upper)",
					func(datum *dataTypesDosingDecision.Issue) {
						datum.ID = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/id"),
				),
				Entry("metadata missing",
					func(datum *dataTypesDosingDecision.Issue) { datum.Metadata = nil },
				),
				Entry("metadata invalid",
					func(datum *dataTypesDosingDecision.Issue) { datum.Metadata = metadata.NewMetadata() },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/metadata"),
				),
				Entry("metadata valid",
					func(datum *dataTypesDosingDecision.Issue) { datum.Metadata = metadataTest.RandomMetadata() },
				),
				Entry("multiple errors",
					func(datum *dataTypesDosingDecision.Issue) {
						datum.ID = pointer.FromString("")
						datum.Metadata = metadata.NewMetadata()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/id"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/metadata"),
				),
			)
		})
	})

	Context("IssueArray", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesDosingDecision.IssueArray)) {
				datum := dataTypesDosingDecisionTest.RandomIssueArray()
				mutator(datum)
				test.ExpectSerializedArrayJSON(dataTypesDosingDecisionTest.AnonymizeIssueArray(datum), dataTypesDosingDecisionTest.NewArrayFromIssueArray(datum, test.ObjectFormatJSON))
				test.ExpectSerializedArrayBSON(dataTypesDosingDecisionTest.AnonymizeIssueArray(datum), dataTypesDosingDecisionTest.NewArrayFromIssueArray(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesDosingDecision.IssueArray) {},
			),
			Entry("empty",
				func(datum *dataTypesDosingDecision.IssueArray) {
					*datum = *dataTypesDosingDecision.NewIssueArray()
				},
			),
		)

		Context("ParseIssueArray", func() {
			It("returns nil when the array is missing", func() {
				Expect(dataTypesDosingDecision.ParseIssueArray(structureParser.NewArray(nil))).To(BeNil())
			})

			It("returns new datum when the array is valid", func() {
				datum := dataTypesDosingDecisionTest.RandomIssueArray()
				array := dataTypesDosingDecisionTest.NewArrayFromIssueArray(datum, test.ObjectFormatJSON)
				parser := structureParser.NewArray(&array)
				Expect(dataTypesDosingDecision.ParseIssueArray(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewIssueArray", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesDosingDecision.NewIssueArray()
				Expect(datum).ToNot(BeNil())
				Expect(*datum).To(BeEmpty())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object []interface{}, expectedDatum *dataTypesDosingDecision.IssueArray), expectedErrors ...error) {
					expectedDatum := dataTypesDosingDecisionTest.RandomIssueArray()
					array := dataTypesDosingDecisionTest.NewArrayFromIssueArray(expectedDatum, test.ObjectFormatJSON)
					mutator(array, expectedDatum)
					datum := dataTypesDosingDecision.NewIssueArray()
					errorsTest.ExpectEqual(structureParser.NewArray(&array).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object []interface{}, expectedDatum *dataTypesDosingDecision.IssueArray) {},
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesDosingDecision.IssueArray), expectedErrors ...error) {
					datum := dataTypesDosingDecision.NewIssueArray()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesDosingDecision.IssueArray) {},
					structureValidator.ErrorValueEmpty(),
				),
				Entry("empty",
					func(datum *dataTypesDosingDecision.IssueArray) { *datum = *dataTypesDosingDecision.NewIssueArray() },
					structureValidator.ErrorValueEmpty(),
				),
				Entry("nil",
					func(datum *dataTypesDosingDecision.IssueArray) { *datum = append(*datum, nil) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
				),
				Entry("single invalid",
					func(datum *dataTypesDosingDecision.IssueArray) {
						invalid := dataTypesDosingDecisionTest.RandomIssue()
						invalid.ID = nil
						*datum = append(*datum, invalid)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0/id"),
				),
				Entry("single valid",
					func(datum *dataTypesDosingDecision.IssueArray) {
						*datum = append(*datum, dataTypesDosingDecisionTest.RandomIssue())
					},
				),
				Entry("multiple invalid",
					func(datum *dataTypesDosingDecision.IssueArray) {
						invalid := dataTypesDosingDecisionTest.RandomIssue()
						invalid.ID = nil
						*datum = append(*datum, dataTypesDosingDecisionTest.RandomIssue(), invalid, dataTypesDosingDecisionTest.RandomIssue())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/id"),
				),
				Entry("multiple valid",
					func(datum *dataTypesDosingDecision.IssueArray) {
						*datum = append(*datum, dataTypesDosingDecisionTest.RandomIssue(), dataTypesDosingDecisionTest.RandomIssue(), dataTypesDosingDecisionTest.RandomIssue())
					},
				),
				Entry("multiple; length in range (upper)",
					func(datum *dataTypesDosingDecision.IssueArray) {
						for len(*datum) < 100 {
							*datum = append(*datum, dataTypesDosingDecisionTest.RandomIssue())
						}
					},
				),
				Entry("multiple; length out of range (upper)",
					func(datum *dataTypesDosingDecision.IssueArray) {
						for len(*datum) < 101 {
							*datum = append(*datum, dataTypesDosingDecisionTest.RandomIssue())
						}
					},
					structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100),
				),
				Entry("multiple errors",
					func(datum *dataTypesDosingDecision.IssueArray) {
						invalid := dataTypesDosingDecisionTest.RandomIssue()
						invalid.ID = nil
						*datum = append(*datum, nil, invalid, dataTypesDosingDecisionTest.RandomIssue())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/id"),
				),
			)
		})
	})
})
