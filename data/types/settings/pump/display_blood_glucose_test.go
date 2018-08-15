package pump_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewDisplayBloodGlucose() *pump.DisplayBloodGlucose {
	datum := pump.NewDisplayBloodGlucose()
	datum.Units = pointer.FromString(test.RandomStringFromArray(pump.DisplayBloodGlucoseUnits()))
	return datum
}

func CloneDisplayBloodGlucose(datum *pump.DisplayBloodGlucose) *pump.DisplayBloodGlucose {
	if datum == nil {
		return nil
	}
	clone := pump.NewDisplayBloodGlucose()
	clone.Units = test.CloneString(datum.Units)
	return clone
}

var _ = Describe("DisplayBloodGlucose", func() {
	It("DisplayBloodGlucoseUnitsMgPerDL is expected", func() {
		Expect(pump.DisplayBloodGlucoseUnitsMgPerDL).To(Equal("mg/dL"))
	})

	It("DisplayBloodGlucoseUnitsMmolPerL is expected", func() {
		Expect(pump.DisplayBloodGlucoseUnitsMmolPerL).To(Equal("mmol/L"))
	})

	It("DisplayBloodGlucoseUnits returns expected", func() {
		Expect(pump.DisplayBloodGlucoseUnits()).To(Equal([]string{"mg/dL", "mmol/L"}))
	})

	Context("ParseDisplayBloodGlucose", func() {
		// TODO
	})

	Context("NewDisplayBloodGlucose", func() {
		It("is successful", func() {
			Expect(pump.NewDisplayBloodGlucose()).To(Equal(&pump.DisplayBloodGlucose{}))
		})
	})

	Context("DisplayBloodGlucose", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *pump.DisplayBloodGlucose), expectedErrors ...error) {
					datum := NewDisplayBloodGlucose()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *pump.DisplayBloodGlucose) {},
				),
				Entry("units missing",
					func(datum *pump.DisplayBloodGlucose) { datum.Units = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid",
					func(datum *pump.DisplayBloodGlucose) { datum.Units = pointer.FromString("invalid") },
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mg/dL", "mmol/L"}), "/units"),
				),
				Entry("units mg/dL",
					func(datum *pump.DisplayBloodGlucose) { datum.Units = pointer.FromString("mg/dL") },
				),
				Entry("units mmol/L",
					func(datum *pump.DisplayBloodGlucose) { datum.Units = pointer.FromString("mmol/L") },
				),
				Entry("multiple errors",
					func(datum *pump.DisplayBloodGlucose) {
						datum.Units = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *pump.DisplayBloodGlucose)) {
					for _, origin := range structure.Origins() {
						datum := NewDisplayBloodGlucose()
						mutator(datum)
						expectedDatum := CloneDisplayBloodGlucose(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *pump.DisplayBloodGlucose) {},
				),
				Entry("does not modify the datum; units missing",
					func(datum *pump.DisplayBloodGlucose) { datum.Units = nil },
				),
			)
		})
	})
})
