package location_test

import (
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

var _ = Describe("Location", func() {
	It("NameLengthMaximum is expected", func() {
		Expect(location.NameLengthMaximum).To(Equal(100))
	})

	Context("Location", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *location.Location)) {
				datum := locationTest.RandomLocation()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, locationTest.NewObjectFromLocation(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, locationTest.NewObjectFromLocation(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *location.Location) {},
			),
			Entry("empty",
				func(datum *location.Location) {
					*datum = *location.NewLocation()
				},
			),
			Entry("all",
				func(datum *location.Location) {
					datum.GPS = locationTest.RandomGPS()
					datum.Name = pointer.FromString(locationTest.RandomName())
				},
			),
		)

		Context("ParseLocation", func() {
			It("returns nil when the object is missing", func() {
				Expect(location.ParseLocation(structureParser.NewObject(logTest.NewLogger(), nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := locationTest.RandomLocation()
				object := locationTest.NewObjectFromLocation(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(logTest.NewLogger(), &object)
				Expect(location.ParseLocation(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewLocation", func() {
			It("returns successfully with default values", func() {
				Expect(location.NewLocation()).To(Equal(&location.Location{}))
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *location.Location), expectedErrors ...error) {
					expectedDatum := locationTest.RandomLocation()
					object := locationTest.NewObjectFromLocation(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &location.Location{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *location.Location) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *location.Location) {
						object["gps"] = true
						object["name"] = true
						expectedDatum.GPS = nil
						expectedDatum.Name = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/gps"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/name"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *location.Location), expectedErrors ...error) {
					datum := locationTest.RandomLocation()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *location.Location) {},
				),
				Entry("gps missing",
					func(datum *location.Location) { datum.GPS = nil },
				),
				Entry("gps missing; name missing",
					func(datum *location.Location) {
						datum.GPS = nil
						datum.Name = nil
					},
					structureValidator.ErrorValuesNotExistForAny("gps", "name"),
				),
				Entry("gps invalid",
					func(datum *location.Location) { datum.GPS.Latitude = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/gps/latitude"),
				),
				Entry("gps valid",
					func(datum *location.Location) { datum.GPS = locationTest.RandomGPS() },
				),
				Entry("name missing",
					func(datum *location.Location) { datum.Name = nil },
				),
				Entry("name empty",
					func(datum *location.Location) { datum.Name = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/name"),
				),
				Entry("name length; in range (upper)",
					func(datum *location.Location) {
						datum.Name = pointer.FromString(test.RandomStringFromRange(100, 100))
					},
				),
				Entry("name length; out of range (upper)",
					func(datum *location.Location) {
						datum.Name = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/name"),
				),
				Entry("multiple errors",
					func(datum *location.Location) {
						datum.GPS.Latitude = nil
						datum.Name = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/gps/latitude"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/name"),
				),
			)
		})
	})
})
