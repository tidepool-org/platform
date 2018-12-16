package location_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/common/location"
	dataTypesCommonLocationTest "github.com/tidepool-org/platform/data/types/common/location/test"
	dataTypesCommonOriginTest "github.com/tidepool-org/platform/data/types/common/origin/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("GPS", func() {
	It("GPSFloorMaximum is expected", func() {
		Expect(location.GPSFloorMaximum).To(Equal(1000))
	})

	It("GPSFloorMinimum is expected", func() {
		Expect(location.GPSFloorMinimum).To(Equal(-1000))
	})

	Context("ParseGPS", func() {
		// TODO
	})

	Context("NewGPS", func() {
		It("is successful", func() {
			Expect(location.NewGPS()).To(Equal(&location.GPS{}))
		})
	})

	Context("GPS", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *location.GPS), expectedErrors ...error) {
					datum := dataTypesCommonLocationTest.NewGPS()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *location.GPS) {},
				),
				Entry("elevation missing",
					func(datum *location.GPS) { datum.Elevation = nil },
				),
				Entry("elevation invalid",
					func(datum *location.GPS) { datum.Elevation.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/elevation/units"),
				),
				Entry("elevation valid",
					func(datum *location.GPS) {
						datum.Elevation = dataTypesCommonLocationTest.NewElevation(pointer.FromString("feet"))
					},
				),
				Entry("floor missing",
					func(datum *location.GPS) { datum.Floor = nil },
				),
				Entry("floor out of range (lower)",
					func(datum *location.GPS) { datum.Floor = pointer.FromInt(-1001) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-1001, -1000, 1000), "/floor"),
				),
				Entry("floor in range (lower)",
					func(datum *location.GPS) { datum.Floor = pointer.FromInt(0) },
				),
				Entry("floor in range (upper)",
					func(datum *location.GPS) { datum.Floor = pointer.FromInt(1000) },
				),
				Entry("floor out of range (upper)",
					func(datum *location.GPS) { datum.Floor = pointer.FromInt(1001) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1001, -1000, 1000), "/floor"),
				),
				Entry("horizontal accuracy missing",
					func(datum *location.GPS) { datum.HorizontalAccuracy = nil },
				),
				Entry("horizontal accuracy invalid",
					func(datum *location.GPS) { datum.HorizontalAccuracy.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/horizontalAccuracy/units"),
				),
				Entry("horizontal accuracy valid",
					func(datum *location.GPS) {
						datum.HorizontalAccuracy = dataTypesCommonLocationTest.NewAccuracy(pointer.FromString("feet"))
					},
				),
				Entry("latitude missing",
					func(datum *location.GPS) { datum.Latitude = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/latitude"),
				),
				Entry("latitude invalid",
					func(datum *location.GPS) { datum.Latitude.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/latitude/units"),
				),
				Entry("latitude valid",
					func(datum *location.GPS) { datum.Latitude = dataTypesCommonLocationTest.NewLatitude() },
				),
				Entry("longitude missing",
					func(datum *location.GPS) { datum.Longitude = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/longitude"),
				),
				Entry("longitude invalid",
					func(datum *location.GPS) { datum.Longitude.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/longitude/units"),
				),
				Entry("longitude valid",
					func(datum *location.GPS) { datum.Longitude = dataTypesCommonLocationTest.NewLongitude() },
				),
				Entry("origin missing",
					func(datum *location.GPS) { datum.Origin = nil },
				),
				Entry("origin invalid",
					func(datum *location.GPS) { datum.Origin.Name = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/origin/name"),
				),
				Entry("origin valid",
					func(datum *location.GPS) { datum.Origin = dataTypesCommonOriginTest.NewOrigin() },
				),
				Entry("vertical accuracy missing",
					func(datum *location.GPS) { datum.VerticalAccuracy = nil },
				),
				Entry("vertical accuracy invalid",
					func(datum *location.GPS) { datum.VerticalAccuracy.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/verticalAccuracy/units"),
				),
				Entry("vertical accuracy valid",
					func(datum *location.GPS) {
						datum.VerticalAccuracy = dataTypesCommonLocationTest.NewAccuracy(pointer.FromString("feet"))
					},
				),
				Entry("multiple errors",
					func(datum *location.GPS) {
						datum.Elevation.Units = nil
						datum.Floor = pointer.FromInt(-1001)
						datum.HorizontalAccuracy.Units = nil
						datum.Latitude.Units = nil
						datum.Longitude.Units = nil
						datum.Origin.Name = pointer.FromString("")
						datum.VerticalAccuracy.Units = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/elevation/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-1001, -1000, 1000), "/floor"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/horizontalAccuracy/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/latitude/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/longitude/units"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/origin/name"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/verticalAccuracy/units"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *location.GPS)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesCommonLocationTest.NewGPS()
						mutator(datum)
						expectedDatum := dataTypesCommonLocationTest.CloneGPS(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *location.GPS) {},
				),
				Entry("does not modify the datum; elevation missing",
					func(datum *location.GPS) { datum.Elevation = nil },
				),
				Entry("does not modify the datum; floor missing",
					func(datum *location.GPS) { datum.Floor = nil },
				),
				Entry("does not modify the datum; horizontal accuracy missing",
					func(datum *location.GPS) { datum.HorizontalAccuracy = nil },
				),
				Entry("does not modify the datum; latitude missing",
					func(datum *location.GPS) { datum.Latitude = nil },
				),
				Entry("does not modify the datum; longitude missing",
					func(datum *location.GPS) { datum.Longitude = nil },
				),
				Entry("does not modify the datum; origin missing",
					func(datum *location.GPS) { datum.Origin = nil },
				),
				Entry("does not modify the datum; vertical accuracy missing",
					func(datum *location.GPS) { datum.VerticalAccuracy = nil },
				),
				Entry("does not modify the datum; all missing",
					func(datum *location.GPS) { *datum = location.GPS{} },
				),
			)
		})
	})
})
