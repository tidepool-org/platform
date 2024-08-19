package location_test

import (
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/location"
	locationTest "github.com/tidepool-org/platform/location/test"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Accuracy", func() {
	It("AccuracyUnitsFeet is expected", func() {
		Expect(location.AccuracyUnitsFeet).To(Equal("feet"))
	})

	It("AccuracyUnitsMeters is expected", func() {
		Expect(location.AccuracyUnitsMeters).To(Equal("meters"))
	})

	It("AccuracyValueFeetMaximum is expected", func() {
		Expect(location.AccuracyValueFeetMaximum).To(Equal(1000.0 / 0.3048))
	})

	It("AccuracyValueFeetMinimum is expected", func() {
		Expect(location.AccuracyValueFeetMinimum).To(Equal(0.0))
	})

	It("AccuracyValueMetersMaximum is expected", func() {
		Expect(location.AccuracyValueMetersMaximum).To(Equal(1000.0))
	})

	It("AccuracyValueMetersMinimum is expected", func() {
		Expect(location.AccuracyValueMetersMinimum).To(Equal(0.0))
	})

	It("AccuracyUnits returns expected", func() {
		Expect(location.AccuracyUnits()).To(Equal([]string{"feet", "meters"}))
	})

	Context("Accuracy", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *location.Accuracy)) {
				datum := locationTest.RandomAccuracy()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, locationTest.NewObjectFromAccuracy(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, locationTest.NewObjectFromAccuracy(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *location.Accuracy) {},
			),
			Entry("empty",
				func(datum *location.Accuracy) {
					*datum = *location.NewAccuracy()
				},
			),
			Entry("all",
				func(datum *location.Accuracy) {
					datum.Units = pointer.FromString(locationTest.RandomAccuracyUnits())
					datum.Value = pointer.FromFloat64(locationTest.RandomAccuracyValue(datum.Units))
				},
			),
		)

		Context("ParseAccuracy", func() {
			It("returns nil when the object is missing", func() {
				Expect(location.ParseAccuracy(structureParser.NewObject(nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := locationTest.RandomAccuracy()
				object := locationTest.NewObjectFromAccuracy(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(&object)
				Expect(location.ParseAccuracy(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewAccuracy", func() {
			It("returns successfully with default values", func() {
				Expect(location.NewAccuracy()).To(Equal(&location.Accuracy{}))
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *location.Accuracy), expectedErrors ...error) {
					expectedDatum := locationTest.RandomAccuracy()
					object := locationTest.NewObjectFromAccuracy(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &location.Accuracy{}
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *location.Accuracy) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *location.Accuracy) {
						object["units"] = true
						object["value"] = true
						expectedDatum.Units = nil
						expectedDatum.Value = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/units"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotFloat64(true), "/value"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *location.Accuracy), expectedErrors ...error) {
					datum := locationTest.RandomAccuracy()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *location.Accuracy) {},
				),
				Entry("units missing; value missing",
					func(datum *location.Accuracy) {
						datum.Units = nil
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units missing; value in range (lower)",
					func(datum *location.Accuracy) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(-math.MaxFloat64)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (upper)",
					func(datum *location.Accuracy) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(math.MaxFloat64)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; value missing",
					func(datum *location.Accuracy) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"feet", "meters"}), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units invalid; value in range (lower)",
					func(datum *location.Accuracy) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(-math.MaxFloat64)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"feet", "meters"}), "/units"),
				),
				Entry("units invalid; value in range (upper)",
					func(datum *location.Accuracy) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(math.MaxFloat64)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"feet", "meters"}), "/units"),
				),
				Entry("units feet; value missing",
					func(datum *location.Accuracy) {
						datum.Units = pointer.FromString("feet")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units feet; value out of range (lower)",
					func(datum *location.Accuracy) {
						datum.Units = pointer.FromString("feet")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0/0.3048), "/value"),
				),
				Entry("units feet; value in range (lower)",
					func(datum *location.Accuracy) {
						datum.Units = pointer.FromString("feet")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units feet; value in range (upper)",
					func(datum *location.Accuracy) {
						datum.Units = pointer.FromString("feet")
						datum.Value = pointer.FromFloat64(1000.0 / 0.3048)
					},
				),
				Entry("units feet; value out of range (upper)",
					func(datum *location.Accuracy) {
						datum.Units = pointer.FromString("feet")
						datum.Value = pointer.FromFloat64(1000.1 / 0.3048)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1/0.3048, 0.0, 1000.0/0.3048), "/value"),
				),
				Entry("units meters; value missing",
					func(datum *location.Accuracy) {
						datum.Units = pointer.FromString("meters")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units meters; value out of range (lower)",
					func(datum *location.Accuracy) {
						datum.Units = pointer.FromString("meters")
						datum.Value = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/value"),
				),
				Entry("units meters; value in range (lower)",
					func(datum *location.Accuracy) {
						datum.Units = pointer.FromString("meters")
						datum.Value = pointer.FromFloat64(0.0)
					},
				),
				Entry("units meters; value in range (upper)",
					func(datum *location.Accuracy) {
						datum.Units = pointer.FromString("meters")
						datum.Value = pointer.FromFloat64(1000.0)
					},
				),
				Entry("units meters; value out of range (upper)",
					func(datum *location.Accuracy) {
						datum.Units = pointer.FromString("meters")
						datum.Value = pointer.FromFloat64(1000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/value"),
				),
				Entry("multiple errors",
					func(datum *location.Accuracy) {
						datum.Units = nil
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
			)
		})
	})

	Context("AccuracyValueRangeForUnits", func() {
		It("returns expected range for units missing", func() {
			minimum, maximum := location.AccuracyValueRangeForUnits(nil)
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units invalid", func() {
			minimum, maximum := location.AccuracyValueRangeForUnits(pointer.FromString("invalid"))
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units feet", func() {
			minimum, maximum := location.AccuracyValueRangeForUnits(pointer.FromString("feet"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(1000.0 / 0.3048))
		})

		It("returns expected range for units meters", func() {
			minimum, maximum := location.AccuracyValueRangeForUnits(pointer.FromString("meters"))
			Expect(minimum).To(Equal(0.0))
			Expect(maximum).To(Equal(1000.0))
		})
	})
})
