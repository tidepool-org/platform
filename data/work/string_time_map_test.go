package work_test

import (
	"fmt"
	"sort"
	"strings"
	"time"

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

var _ = Describe("string_time_map", func() {
	It("StringTimeMapReferenceLengthMaximum is expected", func() {
		Expect(dataWork.StringTimeMapReferenceLengthMaximum).To(Equal(1000))
	})

	It("StringTimeMapLengthMaximum is expected", func() {
		Expect(dataWork.StringTimeMapLengthMaximum).To(Equal(1000))
	})

	Context("StringTimeMap", func() {
		Context("ParseStringTimeMap", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataWork.ParseStringTimeMap(structureParser.NewObject(logTest.NewLogger(), nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataWorkTest.RandomStringTimeMap()
				object := dataWorkTest.NewObjectFromStringTimeMap(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(logTest.NewLogger(), &object)
				Expect(dataWork.ParseStringTimeMap(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataWork.StringTimeMap)) {
				datum := dataWorkTest.RandomStringTimeMap()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataWorkTest.NewObjectFromStringTimeMap(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataWorkTest.NewObjectFromStringTimeMap(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataWork.StringTimeMap) {},
			),
			Entry("empty",
				func(datum *dataWork.StringTimeMap) {
					*datum = dataWork.StringTimeMap{}
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *dataWork.StringTimeMap), expectedErrors ...error) {
					expectedDatum := dataWorkTest.RandomStringTimeMap()
					object := dataWorkTest.NewObjectFromStringTimeMap(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := pointer.From[dataWork.StringTimeMap](nil)
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *dataWork.StringTimeMap) {},
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataWork.StringTimeMap), expectedErrors ...error) {
					datum := dataWorkTest.RandomStringTimeMap()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataWork.StringTimeMap) {},
				),
				Entry("nil",
					func(datum *dataWork.StringTimeMap) { *datum = nil },
				),
				Entry("empty",
					func(datum *dataWork.StringTimeMap) { *datum = dataWork.StringTimeMap{} },
				),
				Entry("single valid",
					func(datum *dataWork.StringTimeMap) {
						*datum = dataWork.StringTimeMap{}
						(*datum)[dataWorkTest.RandomStringTimeMapReference()] = pointer.From(dataWorkTest.RandomStringTimeMapValue())
					},
				),
				Entry("multiple valid",
					func(datum *dataWork.StringTimeMap) {
						*datum = *dataWorkTest.RandomStringTimeMap()
					},
				),
				Entry("length in range (upper)",
					func(datum *dataWork.StringTimeMap) {
						*datum = dataWork.StringTimeMap{}
						for range dataWork.StringTimeMapLengthMaximum {
							(*datum)[dataWorkTest.RandomStringTimeMapReference()] = pointer.From(dataWorkTest.RandomStringTimeMapValue())
						}
					},
				),
				Entry("length out of range (upper)",
					func(datum *dataWork.StringTimeMap) {
						*datum = dataWork.StringTimeMap{}
						for range dataWork.StringTimeMapLengthMaximum + 1 {
							(*datum)[dataWorkTest.RandomStringTimeMapReference()] = pointer.From(dataWorkTest.RandomStringTimeMapValue())
						}
					},
					structureValidator.ErrorLengthNotLessThanOrEqualTo(dataWork.StringTimeMapLengthMaximum+1, dataWork.StringTimeMapLengthMaximum),
				),
				Entry("reference empty",
					func(datum *dataWork.StringTimeMap) {
						*datum = dataWork.StringTimeMap{}
						(*datum)[""] = pointer.From(dataWorkTest.RandomStringTimeMapValue())
					},
				),
				Entry("reference length in range (upper)",
					func(datum *dataWork.StringTimeMap) {
						*datum = dataWork.StringTimeMap{}
						(*datum)[test.RandomStringFromRange(dataWork.StringTimeMapReferenceLengthMaximum, dataWork.StringTimeMapReferenceLengthMaximum)] = pointer.From(dataWorkTest.RandomStringTimeMapValue())
					},
				),
				Entry("reference length out of range (upper)",
					func(datum *dataWork.StringTimeMap) {
						*datum = dataWork.StringTimeMap{}
						(*datum)[strings.Repeat("X", dataWork.StringTimeMapReferenceLengthMaximum+1)] = pointer.From(dataWorkTest.RandomStringTimeMapValue())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(dataWork.StringTimeMapReferenceLengthMaximum+1, dataWork.StringTimeMapReferenceLengthMaximum), fmt.Sprintf("/%s/#", strings.Repeat("X", dataWork.StringTimeMapReferenceLengthMaximum+1))),
				),
				Entry("value zero",
					func(datum *dataWork.StringTimeMap) {
						*datum = dataWork.StringTimeMap{}
						(*datum)["empty"] = pointer.From(time.Time{})
					},
				),
				Entry("multiple errors",
					func(datum *dataWork.StringTimeMap) {
						*datum = dataWork.StringTimeMap{}
						for range dataWork.StringTimeMapLengthMaximum {
							(*datum)[dataWorkTest.RandomStringTimeMapReference()] = pointer.From(dataWorkTest.RandomStringTimeMapValue())
						}
						(*datum)[strings.Repeat("X", dataWork.StringTimeMapReferenceLengthMaximum+1)] = pointer.From(dataWorkTest.RandomStringTimeMapValue())
					},
					structureValidator.ErrorLengthNotLessThanOrEqualTo(dataWork.StringTimeMapLengthMaximum+1, dataWork.StringTimeMapLengthMaximum),
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(dataWork.StringTimeMapReferenceLengthMaximum+1, dataWork.StringTimeMapReferenceLengthMaximum), fmt.Sprintf("/%s/#", strings.Repeat("X", dataWork.StringTimeMapReferenceLengthMaximum+1))),
				),
			)
		})

		Context("SortedKeys", func() {
			It("returns sorted keys", func() {
				datum := dataWorkTest.RandomStringTimeMap()
				keys := datum.SortedKeys()
				Expect(keys).To(HaveLen(len(*datum)))
				Expect(sort.StringsAreSorted(keys)).To(BeTrue())
			})
		})
	})
})
