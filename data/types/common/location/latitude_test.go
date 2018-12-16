package location_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/common/location"
	dataTypesCommonLocationTest "github.com/tidepool-org/platform/data/types/common/location/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Latitude", func() {
	It("LatitudeUnitsDegrees is expected", func() {
		Expect(location.LatitudeUnitsDegrees).To(Equal("degrees"))
	})

	It("LatitudeValueMaximum is expected", func() {
		Expect(location.LatitudeValueMaximum).To(Equal(90.0))
	})

	It("LatitudeValueMinimum is expected", func() {
		Expect(location.LatitudeValueMinimum).To(Equal(-90.0))
	})

	Context("ParseLatitude", func() {
		// TODO
	})

	Context("NewLatitude", func() {
		It("is successful", func() {
			Expect(location.NewLatitude()).To(Equal(&location.Latitude{}))
		})
	})

	Context("Latitude", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *location.Latitude), expectedErrors ...error) {
					datum := dataTypesCommonLocationTest.NewLatitude()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *location.Latitude) {},
				),
				Entry("units missing",
					func(datum *location.Latitude) { datum.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units empty",
					func(datum *location.Latitude) { datum.Units = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("", "degrees"), "/units"),
				),
				Entry("units invalid",
					func(datum *location.Latitude) { datum.Units = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", "degrees"), "/units"),
				),
				Entry("units degrees",
					func(datum *location.Latitude) { datum.Units = pointer.FromString("degrees") },
				),
				Entry("value missing",
					func(datum *location.Latitude) { datum.Value = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("value out of range (lower)",
					func(datum *location.Latitude) { datum.Value = pointer.FromFloat64(-90.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-90.1, -90.0, 90.0), "/value"),
				),
				Entry("value in range (lower)",
					func(datum *location.Latitude) { datum.Value = pointer.FromFloat64(-90.0) },
				),
				Entry("value in range (upper)",
					func(datum *location.Latitude) { datum.Value = pointer.FromFloat64(90.0) },
				),
				Entry("value out of range (upper)",
					func(datum *location.Latitude) { datum.Value = pointer.FromFloat64(90.1) },
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

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *location.Latitude)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesCommonLocationTest.NewLatitude()
						mutator(datum)
						expectedDatum := dataTypesCommonLocationTest.CloneLatitude(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *location.Latitude) {},
				),
				Entry("does not modify the datum; units missing",
					func(datum *location.Latitude) { datum.Units = nil },
				),
				Entry("does not modify the datum; value missing",
					func(datum *location.Latitude) { datum.Value = nil },
				),
				Entry("does not modify the datum; all missing",
					func(datum *location.Latitude) { *datum = location.Latitude{} },
				),
			)
		})
	})
})
