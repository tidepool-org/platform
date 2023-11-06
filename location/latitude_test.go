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

var _ = Describe("Latitude", func() {
	It("LatitudeUnitsDegrees is expected", func() {
		Expect(location.LatitudeUnitsDegrees).To(Equal("degrees"))
	})

	It("LatitudeValueDegreesMaximum is expected", func() {
		Expect(location.LatitudeValueDegreesMaximum).To(Equal(90.0))
	})

	It("LatitudeValueDegreesMinimum is expected", func() {
		Expect(location.LatitudeValueDegreesMinimum).To(Equal(-90.0))
	})

	It("LatitudeUnits returns expected", func() {
		Expect(location.LatitudeUnits()).To(Equal([]string{"degrees"}))
	})

	Context("Latitude", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *location.Latitude)) {
				datum := locationTest.RandomLatitude()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, locationTest.NewObjectFromLatitude(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, locationTest.NewObjectFromLatitude(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *location.Latitude) {},
			),
			Entry("empty",
				func(datum *location.Latitude) {
					*datum = *location.NewLatitude()
				},
			),
			Entry("all",
				func(datum *location.Latitude) {
					datum.Units = pointer.FromString(locationTest.RandomLatitudeUnits())
					datum.Value = pointer.FromFloat64(locationTest.RandomLatitudeValue(datum.Units))
				},
			),
		)

		Context("ParseLatitude", func() {
			It("returns nil when the object is missing", func() {
				Expect(location.ParseLatitude(structureParser.NewObject(nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := locationTest.RandomLatitude()
				object := locationTest.NewObjectFromLatitude(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(&object)
				Expect(location.ParseLatitude(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewLatitude", func() {
			It("returns successfully with default values", func() {
				Expect(location.NewLatitude()).To(Equal(&location.Latitude{}))
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *location.Latitude), expectedErrors ...error) {
					expectedDatum := locationTest.RandomLatitude()
					object := locationTest.NewObjectFromLatitude(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &location.Latitude{}
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *location.Latitude) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *location.Latitude) {
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
				func(mutator func(datum *location.Latitude), expectedErrors ...error) {
					datum := locationTest.RandomLatitude()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *location.Latitude) {},
				),
				Entry("units missing; value missing",
					func(datum *location.Latitude) {
						datum.Units = nil
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units missing; value in range (lower)",
					func(datum *location.Latitude) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(-math.MaxFloat64)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units missing; value in range (upper)",
					func(datum *location.Latitude) {
						datum.Units = nil
						datum.Value = pointer.FromFloat64(math.MaxFloat64)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid; value missing",
					func(datum *location.Latitude) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"degrees"}), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units invalid; value in range (lower)",
					func(datum *location.Latitude) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(-math.MaxFloat64)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"degrees"}), "/units"),
				),
				Entry("units invalid; value in range (upper)",
					func(datum *location.Latitude) {
						datum.Units = pointer.FromString("invalid")
						datum.Value = pointer.FromFloat64(math.MaxFloat64)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"degrees"}), "/units"),
				),
				Entry("units degrees; value missing",
					func(datum *location.Latitude) {
						datum.Units = pointer.FromString("degrees")
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("units degrees; value out of range (lower)",
					func(datum *location.Latitude) {
						datum.Units = pointer.FromString("degrees")
						datum.Value = pointer.FromFloat64(-90.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-90.1, -90.0, 90.0), "/value"),
				),
				Entry("units degrees; value in range (lower)",
					func(datum *location.Latitude) {
						datum.Units = pointer.FromString("degrees")
						datum.Value = pointer.FromFloat64(-90.0)
					},
				),
				Entry("units degrees; value in range (upper)",
					func(datum *location.Latitude) {
						datum.Units = pointer.FromString("degrees")
						datum.Value = pointer.FromFloat64(90.0)
					},
				),
				Entry("units degrees; value out of range (upper)",
					func(datum *location.Latitude) {
						datum.Units = pointer.FromString("degrees")
						datum.Value = pointer.FromFloat64(90.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(90.1, -90.0, 90.0), "/value"),
				),
				Entry("multiple errors",
					func(datum *location.Latitude) {
						datum.Units = nil
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
			)
		})
	})

	Context("LatitudeValueRangeForUnits", func() {
		It("returns expected range for units missing", func() {
			minimum, maximum := location.LatitudeValueRangeForUnits(nil)
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units invalid", func() {
			minimum, maximum := location.LatitudeValueRangeForUnits(pointer.FromString("invalid"))
			Expect(minimum).To(Equal(-math.MaxFloat64))
			Expect(maximum).To(Equal(math.MaxFloat64))
		})

		It("returns expected range for units degrees", func() {
			minimum, maximum := location.LatitudeValueRangeForUnits(pointer.FromString("degrees"))
			Expect(minimum).To(Equal(-90.0))
			Expect(maximum).To(Equal(90.0))
		})
	})
})
