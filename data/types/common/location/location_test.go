package location_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/common/location"
	testDataTypesCommonLocation "github.com/tidepool-org/platform/data/types/common/location/test"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Location", func() {
	It("NameLengthMaximum is expected", func() {
		Expect(location.NameLengthMaximum).To(Equal(100))
	})

	Context("ParseLocation", func() {
		// TODO
	})

	Context("NewLocation", func() {
		It("is successful", func() {
			Expect(location.NewLocation()).To(Equal(&location.Location{}))
		})
	})

	Context("Location", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *location.Location), expectedErrors ...error) {
					datum := testDataTypesCommonLocation.NewLocation()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
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
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/gps"),
				),
				Entry("gps invalid",
					func(datum *location.Location) { datum.GPS.Latitude = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/gps/latitude"),
				),
				Entry("gps valid",
					func(datum *location.Location) { datum.GPS = testDataTypesCommonLocation.NewGPS() },
				),
				Entry("name missing",
					func(datum *location.Location) { datum.Name = nil },
				),
				Entry("name empty",
					func(datum *location.Location) { datum.Name = pointer.FromString("") },
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/name"),
				),
				Entry("name length; in range (upper)",
					func(datum *location.Location) { datum.Name = pointer.FromString(test.NewText(100, 100)) },
				),
				Entry("name length; out of range (upper)",
					func(datum *location.Location) { datum.Name = pointer.FromString(test.NewText(101, 101)) },
					testErrors.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/name"),
				),
				Entry("multiple errors",
					func(datum *location.Location) {
						datum.GPS.Latitude = nil
						datum.Name = pointer.FromString("")
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/gps/latitude"),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/name"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *location.Location)) {
					for _, origin := range structure.Origins() {
						datum := testDataTypesCommonLocation.NewLocation()
						mutator(datum)
						expectedDatum := testDataTypesCommonLocation.CloneLocation(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *location.Location) {},
				),
				Entry("does not modify the datum; gps missing",
					func(datum *location.Location) { datum.GPS = nil },
				),
				Entry("does not modify the datum; name missing",
					func(datum *location.Location) { datum.Name = nil },
				),
				Entry("does not modify the datum; all missing",
					func(datum *location.Location) { *datum = location.Location{} },
				),
			)
		})
	})
})
