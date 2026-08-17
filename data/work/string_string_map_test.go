package work_test

import (
	"fmt"
	"sort"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataWork "github.com/tidepool-org/platform/data/work"
	dataWorkTest "github.com/tidepool-org/platform/data/work/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("string_string_map", func() {
	It("StringStringMapReferenceLengthMaximum is expected", func() {
		Expect(dataWork.StringStringMapReferenceLengthMaximum).To(Equal(1000))
	})

	It("StringStringMapValueLengthMaximum is expected", func() {
		Expect(dataWork.StringStringMapValueLengthMaximum).To(Equal(1000))
	})

	It("StringStringMapLengthMaximum is expected", func() {
		Expect(dataWork.StringStringMapLengthMaximum).To(Equal(1000))
	})

	Context("StringStringMap", func() {
		Context("ParseStringStringMap", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataWork.ParseStringStringMap(structureParser.NewObject(logTest.NewLogger(), nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataWorkTest.RandomStringStringMap()
				object := dataWorkTest.NewObjectFromStringStringMap(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(logTest.NewLogger(), &object)
				Expect(dataWork.ParseStringStringMap(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataWork.StringStringMap)) {
				datum := dataWorkTest.RandomStringStringMap()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataWorkTest.NewObjectFromStringStringMap(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataWorkTest.NewObjectFromStringStringMap(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataWork.StringStringMap) {},
			),
			Entry("empty",
				func(datum *dataWork.StringStringMap) {
					*datum = dataWork.StringStringMap{}
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *dataWork.StringStringMap), expectedErrors ...error) {
					expectedDatum := dataWorkTest.RandomStringStringMap()
					object := dataWorkTest.NewObjectFromStringStringMap(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := pointer.From[dataWork.StringStringMap](nil)
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *dataWork.StringStringMap) {},
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataWork.StringStringMap), expectedErrors ...error) {
					datum := dataWorkTest.RandomStringStringMap()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataWork.StringStringMap) {},
				),
				Entry("nil",
					func(datum *dataWork.StringStringMap) { *datum = nil },
				),
				Entry("empty",
					func(datum *dataWork.StringStringMap) { *datum = dataWork.StringStringMap{} },
				),
				Entry("single valid",
					func(datum *dataWork.StringStringMap) {
						*datum = dataWork.StringStringMap{}
						(*datum)[dataWorkTest.RandomStringStringMapReference()] = pointer.From(dataWorkTest.RandomStringStringMapValue())
					},
				),
				Entry("multiple valid",
					func(datum *dataWork.StringStringMap) {
						*datum = *dataWorkTest.RandomStringStringMap()
					},
				),
				Entry("length in range (upper)",
					func(datum *dataWork.StringStringMap) {
						*datum = dataWork.StringStringMap{}
						for len(*datum) < dataWork.StringStringMapLengthMaximum {
							(*datum)[dataWorkTest.RandomStringStringMapReference()] = pointer.From(dataWorkTest.RandomStringStringMapValue())
						}
					},
				),
				Entry("length out of range (upper)",
					func(datum *dataWork.StringStringMap) {
						*datum = dataWork.StringStringMap{}
						for len(*datum) < dataWork.StringStringMapLengthMaximum+1 {
							(*datum)[dataWorkTest.RandomStringStringMapReference()] = pointer.From(dataWorkTest.RandomStringStringMapValue())
						}
					},
					structureValidator.ErrorLengthNotLessThanOrEqualTo(dataWork.StringStringMapLengthMaximum+1, dataWork.StringStringMapLengthMaximum),
				),
				Entry("reference empty",
					func(datum *dataWork.StringStringMap) {
						*datum = dataWork.StringStringMap{}
						(*datum)[""] = pointer.From(dataWorkTest.RandomStringStringMapValue())
					},
				),
				Entry("reference length in range (upper)",
					func(datum *dataWork.StringStringMap) {
						*datum = dataWork.StringStringMap{}
						(*datum)[test.RandomStringFromRange(dataWork.StringStringMapReferenceLengthMaximum, dataWork.StringStringMapReferenceLengthMaximum)] = pointer.From(dataWorkTest.RandomStringStringMapValue())
					},
				),
				Entry("reference length out of range (upper)",
					func(datum *dataWork.StringStringMap) {
						*datum = dataWork.StringStringMap{}
						(*datum)[strings.Repeat("X", dataWork.StringStringMapReferenceLengthMaximum+1)] = pointer.From(dataWorkTest.RandomStringStringMapValue())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(dataWork.StringStringMapReferenceLengthMaximum+1, dataWork.StringStringMapReferenceLengthMaximum), fmt.Sprintf("/%s/#", strings.Repeat("X", dataWork.StringStringMapReferenceLengthMaximum+1))),
				),
				Entry("value empty",
					func(datum *dataWork.StringStringMap) {
						*datum = dataWork.StringStringMap{}
						(*datum)["empty"] = pointer.From("")
					},
				),
				Entry("value length in range (upper)",
					func(datum *dataWork.StringStringMap) {
						*datum = dataWork.StringStringMap{}
						(*datum)[dataWorkTest.RandomStringStringMapReference()] = pointer.From(test.RandomStringFromRange(dataWork.StringStringMapValueLengthMaximum, dataWork.StringStringMapValueLengthMaximum))
					},
				),
				Entry("value length out of range (upper)",
					func(datum *dataWork.StringStringMap) {
						*datum = dataWork.StringStringMap{}
						(*datum)["length"] = pointer.From(test.RandomStringFromRange(dataWork.StringStringMapValueLengthMaximum+1, dataWork.StringStringMapValueLengthMaximum+1))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(dataWork.StringStringMapValueLengthMaximum+1, dataWork.StringStringMapValueLengthMaximum), "/length"),
				),
				Entry("multiple errors",
					func(datum *dataWork.StringStringMap) {
						*datum = dataWork.StringStringMap{}
						for len(*datum) < dataWork.StringStringMapLengthMaximum {
							(*datum)[dataWorkTest.RandomStringStringMapReference()] = pointer.From(dataWorkTest.RandomStringStringMapValue())
						}
						(*datum)[strings.Repeat("X", dataWork.StringStringMapReferenceLengthMaximum+1)] = pointer.From(dataWorkTest.RandomStringStringMapValue())
						(*datum)["length"] = pointer.From(test.RandomStringFromRange(dataWork.StringStringMapValueLengthMaximum+1, dataWork.StringStringMapValueLengthMaximum+1))
					},
					structureValidator.ErrorLengthNotLessThanOrEqualTo(dataWork.StringStringMapLengthMaximum+2, dataWork.StringStringMapLengthMaximum),
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(dataWork.StringStringMapReferenceLengthMaximum+1, dataWork.StringStringMapReferenceLengthMaximum), fmt.Sprintf("/%s/#", strings.Repeat("X", dataWork.StringStringMapReferenceLengthMaximum+1))),
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(dataWork.StringStringMapValueLengthMaximum+1, dataWork.StringStringMapValueLengthMaximum), "/length"),
				),
			)
		})

		Context("SortedKeys", func() {
			It("returns sorted keys", func() {
				datum := dataWorkTest.RandomStringStringMap()
				keys := datum.SortedKeys()
				Expect(keys).To(HaveLen(len(*datum)))
				Expect(sort.StringsAreSorted(keys)).To(BeTrue())
			})
		})
	})
})
