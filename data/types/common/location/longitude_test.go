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
)

var _ = Describe("Longitude", func() {
	It("LongitudeUnitsDegrees is expected", func() {
		Expect(location.LongitudeUnitsDegrees).To(Equal("degrees"))
	})

	It("LongitudeValueMaximum is expected", func() {
		Expect(location.LongitudeValueMaximum).To(Equal(180.0))
	})

	It("LongitudeValueMinimum is expected", func() {
		Expect(location.LongitudeValueMinimum).To(Equal(-180.0))
	})

	Context("ParseLongitude", func() {
		// TODO
	})

	Context("NewLongitude", func() {
		It("is successful", func() {
			Expect(location.NewLongitude()).To(Equal(&location.Longitude{}))
		})
	})

	Context("Longitude", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *location.Longitude), expectedErrors ...error) {
					datum := testDataTypesCommonLocation.NewLongitude()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *location.Longitude) {},
				),
				Entry("units missing",
					func(datum *location.Longitude) { datum.Units = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units empty",
					func(datum *location.Longitude) { datum.Units = pointer.FromString("") },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotEqualTo("", "degrees"), "/units"),
				),
				Entry("units invalid",
					func(datum *location.Longitude) { datum.Units = pointer.FromString("invalid") },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", "degrees"), "/units"),
				),
				Entry("units degrees",
					func(datum *location.Longitude) { datum.Units = pointer.FromString("degrees") },
				),
				Entry("value missing",
					func(datum *location.Longitude) { datum.Value = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("value out of range (lower)",
					func(datum *location.Longitude) { datum.Value = pointer.FromFloat64(-180.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-180.1, -180.0, 180.0), "/value"),
				),
				Entry("value in range (lower)",
					func(datum *location.Longitude) { datum.Value = pointer.FromFloat64(-180.0) },
				),
				Entry("value in range (upper)",
					func(datum *location.Longitude) { datum.Value = pointer.FromFloat64(180.0) },
				),
				Entry("value out of range (upper)",
					func(datum *location.Longitude) { datum.Value = pointer.FromFloat64(180.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(180.1, -180.0, 180.0), "/value"),
				),
				Entry("multiple errors",
					func(datum *location.Longitude) {
						datum.Units = nil
						datum.Value = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *location.Longitude)) {
					for _, origin := range structure.Origins() {
						datum := testDataTypesCommonLocation.NewLongitude()
						mutator(datum)
						expectedDatum := testDataTypesCommonLocation.CloneLongitude(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *location.Longitude) {},
				),
				Entry("does not modify the datum; units missing",
					func(datum *location.Longitude) { datum.Units = nil },
				),
				Entry("does not modify the datum; value missing",
					func(datum *location.Longitude) { datum.Value = nil },
				),
				Entry("does not modify the datum; all missing",
					func(datum *location.Longitude) { *datum = location.Longitude{} },
				),
			)
		})
	})
})
