package location_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/location"
	locationTest "github.com/tidepool-org/platform/location/test"
	originTest "github.com/tidepool-org/platform/origin/test"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("GPS", func() {
	It("GPSFloorMaximum is expected", func() {
		Expect(location.GPSFloorMaximum).To(Equal(1000))
	})

	It("GPSFloorMinimum is expected", func() {
		Expect(location.GPSFloorMinimum).To(Equal(-1000))
	})

	Context("GPS", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *location.GPS)) {
				datum := locationTest.RandomGPS()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, locationTest.NewObjectFromGPS(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, locationTest.NewObjectFromGPS(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *location.GPS) {},
			),
			Entry("empty",
				func(datum *location.GPS) {
					*datum = *location.NewGPS()
				},
			),
			Entry("all",
				func(datum *location.GPS) {
					datum.Elevation = locationTest.RandomElevation()
					datum.Floor = pointer.FromInt(locationTest.RandomFloor())
					datum.HorizontalAccuracy = locationTest.RandomAccuracy()
					datum.Latitude = locationTest.RandomLatitude()
					datum.Longitude = locationTest.RandomLongitude()
					datum.Origin = originTest.RandomOrigin()
					datum.VerticalAccuracy = locationTest.RandomAccuracy()
				},
			),
		)

		Context("ParseGPS", func() {
			It("returns nil when the object is missing", func() {
				Expect(location.ParseGPS(structureParser.NewObject(nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := locationTest.RandomGPS()
				object := locationTest.NewObjectFromGPS(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(&object)
				Expect(location.ParseGPS(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewGPS", func() {
			It("returns successfully with default values", func() {
				Expect(location.NewGPS()).To(Equal(&location.GPS{}))
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *location.GPS), expectedErrors ...error) {
					expectedDatum := locationTest.RandomGPS()
					object := locationTest.NewObjectFromGPS(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &location.GPS{}
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *location.GPS) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *location.GPS) {
						object["elevation"] = true
						object["floor"] = true
						object["horizontalAccuracy"] = true
						object["latitude"] = true
						object["longitude"] = true
						object["origin"] = true
						object["verticalAccuracy"] = true
						expectedDatum.Elevation = nil
						expectedDatum.Floor = nil
						expectedDatum.HorizontalAccuracy = nil
						expectedDatum.Latitude = nil
						expectedDatum.Longitude = nil
						expectedDatum.Origin = nil
						expectedDatum.VerticalAccuracy = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/elevation"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/floor"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/horizontalAccuracy"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/latitude"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/longitude"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/origin"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/verticalAccuracy"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *location.GPS), expectedErrors ...error) {
					datum := locationTest.RandomGPS()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
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
						datum.Elevation = locationTest.RandomElevation()
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
						datum.HorizontalAccuracy = locationTest.RandomAccuracy()
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
					func(datum *location.GPS) { datum.Latitude = locationTest.RandomLatitude() },
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
					func(datum *location.GPS) { datum.Longitude = locationTest.RandomLongitude() },
				),
				Entry("origin missing",
					func(datum *location.GPS) { datum.Origin = nil },
				),
				Entry("origin invalid",
					func(datum *location.GPS) { datum.Origin.Name = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/origin/name"),
				),
				Entry("origin valid",
					func(datum *location.GPS) { datum.Origin = originTest.RandomOrigin() },
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
						datum.VerticalAccuracy = locationTest.RandomAccuracy()
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
	})
})
