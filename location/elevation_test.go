package location_test

import (
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/location"
	locationTest "github.com/tidepool-org/platform/location/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Elevation", func() {
	It("ElevationUnitsFeet is expected", func() {
		Expect(location.ElevationUnitsFeet).To(Equal("feet"))
	})

	It("ElevationUnitsMeters is expected", func() {
		Expect(location.ElevationUnitsMeters).To(Equal("meters"))
	})

	It("ElevationValueFeetMaximum is expected", func() {
		Expect(location.ElevationValueFeetMaximum).To(Equal(1000000.0 / 0.3048))
	})

	It("ElevationValueFeetMinimum is expected", func() {
		Expect(location.ElevationValueFeetMinimum).To(Equal(-20000.0 / 0.3048))
	})

	It("ElevationValueMetersMaximum is expected", func() {
		Expect(location.ElevationValueMetersMaximum).To(Equal(1000000.0))
	})

	It("ElevationValueMetersMinimum is expected", func() {
		Expect(location.ElevationValueMetersMinimum).To(Equal(-20000.0))
	})

	It("ElevationUnits returns expected", func() {
		Expect(location.ElevationUnits()).To(Equal([]string{"feet", "meters"}))
	})

	Context("Elevation", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *location.Elevation)) {
				datum := locationTest.RandomElevation()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, locationTest.NewObjectFromElevation(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, locationTest.NewObjectFromElevation(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *location.Elevation) {},
			),
			Entry("empty",
				func(datum *location.Elevation) {
					*datum = *location.NewElevation()
				},
			),
			Entry("all",
				func(datum *location.Elevation) {
					datum.Units = pointer.FromString(locationTest.RandomElevationUnits())
					datum.Value = pointer.FromFloat64(locationTest.RandomElevationValue(datum.Units))
				},
			),
		)

		Context("ParseElevation", func() {
			It("returns nil when the object is missing", func() {
				Expect(location.ParseElevation(structureParser.NewObject(logTest.NewLogger(), nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := locationTest.RandomElevation()
				object := locationTest.NewObjectFromElevation(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(logTest.NewLogger(), &object)
				Expect(location.ParseElevation(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewElevation", func() {
			It("returns successfully with default values", func() {
				Expect(location.NewElevation()).To(Equal(&location.Elevation{}))
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *location.Elevation), expectedErrors ...error) {
					expectedDatum := locationTest.RandomElevation()
					object := locationTest.NewObjectFromElevation(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &location.Elevation{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *location.Elevation) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *location.Elevation) {
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
				func(mutator func(datum *location.Elevation), expectedErrors ...error) {
					datum := locationTest.RandomElevation()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *location.Elevation) {},
				),
				Entry("units missing; value missing",
					func(datum *location.Elevation) {
						datum.Units = nil
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units missing; value in range (lower)",
					func(datum *location.Elevation) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(-math.MaxFloat64)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (upper)",
					func(datum *location.Elevation) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(math.MaxFloat64)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; value missing",
					func(datum *location.Elevation) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"feet", "meters"}), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units invalid; value in range (lower)",
					func(datum *location.Elevation) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(-math.MaxFloat64)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"feet", "meters"}), "/units"),
				),
				Entry("units invalid; value in range (upper)",
					func(datum *location.Elevation) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(math.MaxFloat64)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"feet", "meters"}), "/units"),
				),
				Entry("units feet; value missing",
					func(datum *location.Elevation) {
						datum.Units = pointer.FromString("feet")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units feet; value out of range (lower)",
					func(datum *location.Elevation) {
						datum.Units = pointer.FromString("feet")
						datum.Value = pointer.FromFloat64(-20000.1 / 0.3048)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-20000.1/0.3048, -20000.0/0.3048, 1000000.0/0.3048), "/value"),
				),
				Entry("units feet; value in range (lower)",
					func(datum *location.Elevation) {
						datum.Units = pointer.FromString("feet")
						datum.Value = pointer.FromFloat64(-20000.0 / 0.3048)
					},
				),
				Entry("units feet; value in range (upper)",
					func(datum *location.Elevation) {
						datum.Units = pointer.FromString("feet")
						datum.Value = pointer.FromFloat64(1000000.0 / 0.3048)
					},
				),
				Entry("units feet; value out of range (upper)",
					func(datum *location.Elevation) {
						datum.Units = pointer.FromString("feet")
						datum.Value = pointer.FromFloat64(1000000.1 / 0.3048)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000000.1/0.3048, -20000.0/0.3048, 1000000.0/0.3048), "/value"),
				),
				Entry("units meters; value missing",
					func(datum *location.Elevation) {
						datum.Units = pointer.FromString("meters")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units meters; value out of range (lower)",
					func(datum *location.Elevation) {
						datum.Units = pointer.FromString("meters")
						datum.Value = pointer.FromFloat64(-20000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-20000.1, -20000.0, 1000000.0), "/value"),
				),
				Entry("units meters; value in range (lower)",
					func(datum *location.Elevation) {
						datum.Units = pointer.FromString("meters")
						datum.Value = pointer.FromFloat64(-20000.0)
					},
				),
				Entry("units meters; value in range (upper)",
					func(datum *location.Elevation) {
						datum.Units = pointer.FromString("meters")
						datum.Value = pointer.FromFloat64(1000000.0)
					},
				),
				Entry("units meters; value out of range (upper)",
					func(datum *location.Elevation) {
						datum.Units = pointer.FromString("meters")
						datum.Value = pointer.FromFloat64(1000000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000000.1, -20000.0, 1000000.0), "/value"),
				),
				Entry("multiple errors",
					func(datum *location.Elevation) {
						datum.Units = nil
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
			)
		})
	})

	Context("ElevationValueRangeForUnits", func() {
		It("returns expected range for units missing", func() {
			minimum, maximum := location.ElevationValueRangeForUnits(nil)
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units invalid", func() {
			minimum, maximum := location.ElevationValueRangeForUnits(pointer.FromString("invalid"))
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units feet", func() {
			minimum, maximum := location.ElevationValueRangeForUnits(pointer.FromString("feet"))
			Expect(minimum).To(Equal(-20000.0 / 0.3048))
			Expect(maximum).To(Equal(1000000.0 / 0.3048))
		})

		It("returns expected range for units meters", func() {
			minimum, maximum := location.ElevationValueRangeForUnits(pointer.FromString("meters"))
			Expect(minimum).To(Equal(-20000.0))
			Expect(maximum).To(Equal(1000000.0))
		})
	})
})
