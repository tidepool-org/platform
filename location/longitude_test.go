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

var _ = Describe("Longitude", func() {
	It("LongitudeUnitsDegrees is expected", func() {
		Expect(location.LongitudeUnitsDegrees).To(Equal("degrees"))
	})

	It("LongitudeValueDegreesMaximum is expected", func() {
		Expect(location.LongitudeValueDegreesMaximum).To(Equal(180.0))
	})

	It("LongitudeValueDegreesMinimum is expected", func() {
		Expect(location.LongitudeValueDegreesMinimum).To(Equal(-180.0))
	})

	It("LongitudeUnits returns expected", func() {
		Expect(location.LongitudeUnits()).To(Equal([]string{"degrees"}))
	})

	Context("Longitude", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *location.Longitude)) {
				datum := locationTest.RandomLongitude()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, locationTest.NewObjectFromLongitude(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, locationTest.NewObjectFromLongitude(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *location.Longitude) {},
			),
			Entry("empty",
				func(datum *location.Longitude) {
					*datum = *location.NewLongitude()
				},
			),
			Entry("all",
				func(datum *location.Longitude) {
					datum.Units = pointer.FromString(locationTest.RandomLongitudeUnits())
					datum.Value = pointer.FromFloat64(locationTest.RandomLongitudeValue(datum.Units))
				},
			),
		)

		Context("ParseLongitude", func() {
			It("returns nil when the object is missing", func() {
				Expect(location.ParseLongitude(structureParser.NewObject(nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := locationTest.RandomLongitude()
				object := locationTest.NewObjectFromLongitude(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(&object)
				Expect(location.ParseLongitude(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewLongitude", func() {
			It("returns successfully with default values", func() {
				Expect(location.NewLongitude()).To(Equal(&location.Longitude{}))
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *location.Longitude), expectedErrors ...error) {
					expectedDatum := locationTest.RandomLongitude()
					object := locationTest.NewObjectFromLongitude(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &location.Longitude{}
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *location.Longitude) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *location.Longitude) {
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
				func(mutator func(datum *location.Longitude), expectedErrors ...error) {
					datum := locationTest.RandomLongitude()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *location.Longitude) {},
				),
				Entry("units missing; value missing",
					func(datum *location.Longitude) {
						datum.Units = nil
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units missing; value in range (lower)",
					func(datum *location.Longitude) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(-math.MaxFloat64)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (upper)",
					func(datum *location.Longitude) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(math.MaxFloat64)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; value missing",
					func(datum *location.Longitude) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"degrees"}), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units invalid; value in range (lower)",
					func(datum *location.Longitude) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(-math.MaxFloat64)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"degrees"}), "/units"),
				),
				Entry("units invalid; value in range (upper)",
					func(datum *location.Longitude) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(math.MaxFloat64)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"degrees"}), "/units"),
				),
				Entry("units degrees; value missing",
					func(datum *location.Longitude) {
						datum.Units = pointer.FromString("degrees")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units degrees; value out of range (lower)",
					func(datum *location.Longitude) {
						datum.Units = pointer.FromString("degrees")
						datum.Value = pointer.FromFloat64(-180.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-180.1, -180.0, 180.0), "/value"),
				),
				Entry("units degrees; value in range (lower)",
					func(datum *location.Longitude) {
						datum.Units = pointer.FromString("degrees")
						datum.Value = pointer.FromFloat64(-180.0)
					},
				),
				Entry("units degrees; value in range (upper)",
					func(datum *location.Longitude) {
						datum.Units = pointer.FromString("degrees")
						datum.Value = pointer.FromFloat64(180.0)
					},
				),
				Entry("units degrees; value out of range (upper)",
					func(datum *location.Longitude) {
						datum.Units = pointer.FromString("degrees")
						datum.Value = pointer.FromFloat64(180.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(180.1, -180.0, 180.0), "/value"),
				),
				Entry("multiple errors",
					func(datum *location.Longitude) {
						datum.Units = nil
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
			)
		})
	})

	Context("LongitudeValueRangeForUnits", func() {
		It("returns expected range for units missing", func() {
			minimum, maximum := location.LongitudeValueRangeForUnits(nil)
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units invalid", func() {
			minimum, maximum := location.LongitudeValueRangeForUnits(pointer.FromString("invalid"))
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units degrees", func() {
			minimum, maximum := location.LongitudeValueRangeForUnits(pointer.FromString("degrees"))
			Expect(minimum).To(Equal(-180.0))
			Expect(maximum).To(Equal(180.0))
		})
	})
})
