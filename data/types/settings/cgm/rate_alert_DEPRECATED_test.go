package cgm_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"math"

	testDataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/settings/cgm"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewFallRateAlertDEPRECATED(units *string) *cgm.FallRateAlertDEPRECATED {
	datum := cgm.NewFallRateAlertDEPRECATED()
	datum.Enabled = pointer.FromBool(test.RandomBool())
	if rates := datum.RatesForUnits(units); len(rates) > 0 {
		datum.Rate = pointer.FromFloat64(test.RandomFloat64FromArray(rates))
	} else {
		datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(-math.MaxFloat64, math.MaxFloat64))
	}
	return datum
}

func CloneFallRateAlertDEPRECATED(datum *cgm.FallRateAlertDEPRECATED) *cgm.FallRateAlertDEPRECATED {
	if datum == nil {
		return nil
	}
	clone := cgm.NewFallRateAlertDEPRECATED()
	clone.Enabled = test.CloneBool(datum.Enabled)
	clone.Rate = test.CloneFloat64(datum.Rate)
	return clone
}

func NewRiseRateAlertDEPRECATED(units *string) *cgm.RiseRateAlertDEPRECATED {
	datum := cgm.NewRiseRateAlertDEPRECATED()
	datum.Enabled = pointer.FromBool(test.RandomBool())
	if rates := datum.RatesForUnits(units); len(rates) > 0 {
		datum.Rate = pointer.FromFloat64(test.RandomFloat64FromArray(rates))
	} else {
		datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(-math.MaxFloat64, math.MaxFloat64))
	}
	return datum
}

func CloneRiseRateAlertDEPRECATED(datum *cgm.RiseRateAlertDEPRECATED) *cgm.RiseRateAlertDEPRECATED {
	if datum == nil {
		return nil
	}
	clone := cgm.NewRiseRateAlertDEPRECATED()
	clone.Enabled = test.CloneBool(datum.Enabled)
	clone.Rate = test.CloneFloat64(datum.Rate)
	return clone
}

func NewRateAlertsDEPRECATED(units *string) *cgm.RateAlertsDEPRECATED {
	datum := cgm.NewRateAlertsDEPRECATED()
	datum.FallRateAlert = NewFallRateAlertDEPRECATED(units)
	datum.RiseRateAlert = NewRiseRateAlertDEPRECATED(units)
	return datum
}

func CloneRateAlertsDEPRECATED(datum *cgm.RateAlertsDEPRECATED) *cgm.RateAlertsDEPRECATED {
	if datum == nil {
		return nil
	}
	clone := cgm.NewRateAlertsDEPRECATED()
	clone.FallRateAlert = CloneFallRateAlertDEPRECATED(datum.FallRateAlert)
	clone.RiseRateAlert = CloneRiseRateAlertDEPRECATED(datum.RiseRateAlert)
	return clone
}

var _ = Describe("RateAlertDEPRECATED", func() {
	It("RateDEPRECATEDMgdLThree is expected", func() {
		Expect(cgm.RateDEPRECATEDMgdLThree).To(Equal(3.0))
	})

	It("RateDEPRECATEDMgdLTwo is expected", func() {
		Expect(cgm.RateDEPRECATEDMgdLTwo).To(Equal(2.0))
	})

	It("RateDEPRECATEDMmolLThree is expected", func() {
		Expect(cgm.RateDEPRECATEDMmolLThree).To(Equal(0.16652243973136602))
	})

	It("RateDEPRECATEDMmolLTwo is expected", func() {
		Expect(cgm.RateDEPRECATEDMmolLTwo).To(Equal(0.11101495982091067))
	})

	Context("ParseFallRateAlertDEPRECATED", func() {
		// TODO
	})

	Context("NewFallRateAlertDEPRECATED", func() {
		It("is successful", func() {
			Expect(cgm.NewFallRateAlertDEPRECATED()).To(Equal(&cgm.FallRateAlertDEPRECATED{}))
		})
	})

	Context("FallRateAlertDEPRECATED", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *cgm.FallRateAlertDEPRECATED, units *string), expectedErrors ...error) {
					datum := NewFallRateAlertDEPRECATED(units)
					mutator(datum, units)
					testDataTypes.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithStringAdapter(datum, units), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) {},
				),
				Entry("enabled missing",
					nil,
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) { datum.Enabled = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),
				Entry("enabled true",
					nil,
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) { datum.Enabled = pointer.FromBool(true) },
				),
				Entry("enabled false",
					nil,
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) { datum.Enabled = pointer.FromBool(false) },
				),
				Entry("units missing; rate missing",
					nil,
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) { datum.Rate = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units missing; rate valid",
					nil,
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(-math.MaxFloat64, math.MaxFloat64))
					},
				),
				Entry("units invalid; rate missing",
					pointer.FromString("invalid"),
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) { datum.Rate = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units invalid; rate valid",
					pointer.FromString("invalid"),
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(-math.MaxFloat64, math.MaxFloat64))
					},
				),
				Entry("units mmol/L; rate missing",
					pointer.FromString("mmol/L"),
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) { datum.Rate = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units mmol/L; rate invalid",
					pointer.FromString("mmol/L"),
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(0.0)
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(0.0, []float64{-0.16652243973136602, -0.11101495982091067}), "/rate"),
				),
				Entry("units mmol/L; rate valid -3 mg/dL",
					pointer.FromString("mmol/L"),
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(-0.16652243973136602)
					},
				),
				Entry("units mmol/L; rate valid -2 mg/dL",
					pointer.FromString("mmol/L"),
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(-0.11101495982091067)
					},
				),
				Entry("units mmol/l; rate missing",
					pointer.FromString("mmol/l"),
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) { datum.Rate = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units mmol/l; rate invalid",
					pointer.FromString("mmol/l"),
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(0.0)
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(0.0, []float64{-0.16652243973136602, -0.11101495982091067}), "/rate"),
				),
				Entry("units mmol/l; rate valid -3 mg/dL",
					pointer.FromString("mmol/l"),
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(-0.16652243973136602)
					},
				),
				Entry("units mmol/l; rate valid -2 mg/dL",
					pointer.FromString("mmol/l"),
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(-0.11101495982091067)
					},
				),
				Entry("units mg/dL; rate missing",
					pointer.FromString("mg/dL"),
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) { datum.Rate = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units mg/dL; rate invalid",
					pointer.FromString("mg/dL"),
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(0.0)
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(0.0, []float64{-3.0, -2.0}), "/rate"),
				),
				Entry("units mg/dL; rate valid -3 mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(-3.0)
					},
				),
				Entry("units mg/dL; rate valid -2 mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(-2.0)
					},
				),
				Entry("units mg/dl; rate missing",
					pointer.FromString("mg/dl"),
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) { datum.Rate = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units mg/dl; rate invalid",
					pointer.FromString("mg/dl"),
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(0.0)
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(0.0, []float64{-3.0, -2.0}), "/rate"),
				),
				Entry("units mg/dl; rate valid -3 mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(-3.0)
					},
				),
				Entry("units mg/dl; rate valid -2 mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(-2.0)
					},
				),
				Entry("multiple errors",
					pointer.FromString("mmol/L"),
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) {
						datum.Enabled = nil
						datum.Rate = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(units *string, mutator func(datum *cgm.FallRateAlertDEPRECATED, units *string), expectator func(datum *cgm.FallRateAlertDEPRECATED, expectedDatum *cgm.FallRateAlertDEPRECATED, units *string)) {
					for _, origin := range structure.Origins() {
						datum := NewFallRateAlertDEPRECATED(units)
						mutator(datum, units)
						expectedDatum := CloneFallRateAlertDEPRECATED(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin), units)
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum, units)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					pointer.FromString("mmol/L"),
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units missing",
					nil,
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; enabled missing",
					pointer.FromString("mmol/L"),
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) { datum.Enabled = nil },
					nil,
				),
			)

			DescribeTable("normalizes the datum with origin external",
				func(units *string, mutator func(datum *cgm.FallRateAlertDEPRECATED, units *string), expectator func(datum *cgm.FallRateAlertDEPRECATED, expectedDatum *cgm.FallRateAlertDEPRECATED, units *string)) {
					datum := NewFallRateAlertDEPRECATED(units)
					mutator(datum, units)
					expectedDatum := CloneFallRateAlertDEPRECATED(datum)
					normalizer := dataNormalizer.New()
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(structure.OriginExternal), units)
					Expect(normalizer.Error()).To(BeNil())
					Expect(normalizer.Data()).To(BeEmpty())
					if expectator != nil {
						expectator(datum, expectedDatum, units)
					}
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) {},
					func(datum *cgm.FallRateAlertDEPRECATED, expectedDatum *cgm.FallRateAlertDEPRECATED, units *string) {
						testDataBloodGlucose.ExpectNormalizedValue(datum.Rate, expectedDatum.Rate, units)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) {},
					func(datum *cgm.FallRateAlertDEPRECATED, expectedDatum *cgm.FallRateAlertDEPRECATED, units *string) {
						testDataBloodGlucose.ExpectNormalizedValue(datum.Rate, expectedDatum.Rate, units)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *cgm.FallRateAlertDEPRECATED, units *string), expectator func(datum *cgm.FallRateAlertDEPRECATED, expectedDatum *cgm.FallRateAlertDEPRECATED, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := NewFallRateAlertDEPRECATED(units)
						mutator(datum, units)
						expectedDatum := CloneFallRateAlertDEPRECATED(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin), units)
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum, units)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *cgm.FallRateAlertDEPRECATED, units *string) {},
					nil,
				),
			)
		})
	})

	Context("ParseRiseRateAlertDEPRECATED", func() {
		// TODO
	})

	Context("NewRiseRateAlertDEPRECATED", func() {
		It("is successful", func() {
			Expect(cgm.NewRiseRateAlertDEPRECATED()).To(Equal(&cgm.RiseRateAlertDEPRECATED{}))
		})
	})

	Context("RiseRateAlertDEPRECATED", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *cgm.RiseRateAlertDEPRECATED, units *string), expectedErrors ...error) {
					datum := NewRiseRateAlertDEPRECATED(units)
					mutator(datum, units)
					testDataTypes.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithStringAdapter(datum, units), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) {},
				),
				Entry("enabled missing",
					nil,
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) { datum.Enabled = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),
				Entry("enabled true",
					nil,
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) { datum.Enabled = pointer.FromBool(true) },
				),
				Entry("enabled false",
					nil,
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) { datum.Enabled = pointer.FromBool(false) },
				),
				Entry("units missing; rate missing",
					nil,
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) { datum.Rate = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units missing; rate valid",
					nil,
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(-math.MaxFloat64, math.MaxFloat64))
					},
				),
				Entry("units invalid; rate missing",
					pointer.FromString("invalid"),
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) { datum.Rate = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units invalid; rate valid",
					pointer.FromString("invalid"),
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(-math.MaxFloat64, math.MaxFloat64))
					},
				),
				Entry("units mmol/L; rate missing",
					pointer.FromString("mmol/L"),
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) { datum.Rate = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units mmol/L; rate invalid",
					pointer.FromString("mmol/L"),
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(0.0)
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(0.0, []float64{0.11101495982091067, 0.16652243973136602}), "/rate"),
				),
				Entry("units mmol/L; rate valid 2 mg/dL",
					pointer.FromString("mmol/L"),
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(0.11101495982091067)
					},
				),
				Entry("units mmol/L; rate valid 3 mg/dL",
					pointer.FromString("mmol/L"),
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(0.16652243973136602)
					},
				),
				Entry("units mmol/l; rate missing",
					pointer.FromString("mmol/l"),
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) { datum.Rate = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units mmol/l; rate invalid",
					pointer.FromString("mmol/l"),
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(0.0)
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(0.0, []float64{0.11101495982091067, 0.16652243973136602}), "/rate"),
				),
				Entry("units mmol/l; rate valid 2 mg/dL",
					pointer.FromString("mmol/l"),
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(0.11101495982091067)
					},
				),
				Entry("units mmol/l; rate valid 3 mg/dL",
					pointer.FromString("mmol/l"),
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(0.16652243973136602)
					},
				),
				Entry("units mg/dL; rate missing",
					pointer.FromString("mg/dL"),
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) { datum.Rate = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units mg/dL; rate invalid",
					pointer.FromString("mg/dL"),
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(0.0)
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(0.0, []float64{2.0, 3.0}), "/rate"),
				),
				Entry("units mg/dL; rate valid 2 mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(2.0)
					},
				),
				Entry("units mg/dL; rate valid 3 mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(3.0)
					},
				),
				Entry("units mg/dl; rate missing",
					pointer.FromString("mg/dl"),
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) { datum.Rate = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units mg/dl; rate invalid",
					pointer.FromString("mg/dl"),
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(0.0)
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(0.0, []float64{2.0, 3.0}), "/rate"),
				),
				Entry("units mg/dl; rate valid 2 mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(2.0)
					},
				),
				Entry("units mg/dl; rate valid 3 mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(3.0)
					},
				),
				Entry("multiple errors",
					pointer.FromString("mmol/L"),
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Enabled = nil
						datum.Rate = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(units *string, mutator func(datum *cgm.RiseRateAlertDEPRECATED, units *string), expectator func(datum *cgm.RiseRateAlertDEPRECATED, expectedDatum *cgm.RiseRateAlertDEPRECATED, units *string)) {
					for _, origin := range structure.Origins() {
						datum := NewRiseRateAlertDEPRECATED(units)
						mutator(datum, units)
						expectedDatum := CloneRiseRateAlertDEPRECATED(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin), units)
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum, units)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					pointer.FromString("mmol/L"),
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units missing",
					nil,
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; enabled missing",
					pointer.FromString("mmol/L"),
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) { datum.Enabled = nil },
					nil,
				),
			)

			DescribeTable("normalizes the datum with origin external",
				func(units *string, mutator func(datum *cgm.RiseRateAlertDEPRECATED, units *string), expectator func(datum *cgm.RiseRateAlertDEPRECATED, expectedDatum *cgm.RiseRateAlertDEPRECATED, units *string)) {
					datum := NewRiseRateAlertDEPRECATED(units)
					mutator(datum, units)
					expectedDatum := CloneRiseRateAlertDEPRECATED(datum)
					normalizer := dataNormalizer.New()
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(structure.OriginExternal), units)
					Expect(normalizer.Error()).To(BeNil())
					Expect(normalizer.Data()).To(BeEmpty())
					if expectator != nil {
						expectator(datum, expectedDatum, units)
					}
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) {},
					func(datum *cgm.RiseRateAlertDEPRECATED, expectedDatum *cgm.RiseRateAlertDEPRECATED, units *string) {
						testDataBloodGlucose.ExpectNormalizedValue(datum.Rate, expectedDatum.Rate, units)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) {},
					func(datum *cgm.RiseRateAlertDEPRECATED, expectedDatum *cgm.RiseRateAlertDEPRECATED, units *string) {
						testDataBloodGlucose.ExpectNormalizedValue(datum.Rate, expectedDatum.Rate, units)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *cgm.RiseRateAlertDEPRECATED, units *string), expectator func(datum *cgm.RiseRateAlertDEPRECATED, expectedDatum *cgm.RiseRateAlertDEPRECATED, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := NewRiseRateAlertDEPRECATED(units)
						mutator(datum, units)
						expectedDatum := CloneRiseRateAlertDEPRECATED(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin), units)
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum, units)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *cgm.RiseRateAlertDEPRECATED, units *string) {},
					nil,
				),
			)
		})
	})

	Context("ParseRateAlertsDEPRECATED", func() {
		// TODO
	})

	Context("NewRateAlertsDEPRECATED", func() {
		It("is successful", func() {
			Expect(cgm.NewRateAlertsDEPRECATED()).To(Equal(&cgm.RateAlertsDEPRECATED{}))
		})
	})

	Context("RateAlertsDEPRECATED", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *cgm.RateAlertsDEPRECATED, units *string), expectedErrors ...error) {
					datum := NewRateAlertsDEPRECATED(units)
					mutator(datum, units)
					testDataTypes.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithStringAdapter(datum, units), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *cgm.RateAlertsDEPRECATED, units *string) {},
				),
				Entry("fall rate alert missing",
					nil,
					func(datum *cgm.RateAlertsDEPRECATED, units *string) { datum.FallRateAlert = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/fallRate"),
				),
				Entry("fall rate alert invalid",
					nil,
					func(datum *cgm.RateAlertsDEPRECATED, units *string) { datum.FallRateAlert.Enabled = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/fallRate/enabled"),
				),
				Entry("fall rate alert valid",
					nil,
					func(datum *cgm.RateAlertsDEPRECATED, units *string) {
						datum.FallRateAlert = NewFallRateAlertDEPRECATED(units)
					},
				),
				Entry("rise rate alert missing",
					nil,
					func(datum *cgm.RateAlertsDEPRECATED, units *string) { datum.RiseRateAlert = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/riseRate"),
				),
				Entry("rise rate alert invalid",
					nil,
					func(datum *cgm.RateAlertsDEPRECATED, units *string) { datum.RiseRateAlert.Enabled = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/riseRate/enabled"),
				),
				Entry("rise rate alert valid",
					nil,
					func(datum *cgm.RateAlertsDEPRECATED, units *string) {
						datum.RiseRateAlert = NewRiseRateAlertDEPRECATED(units)
					},
				),
				Entry("multiple errors",
					pointer.FromString("mmol/L"),
					func(datum *cgm.RateAlertsDEPRECATED, units *string) {
						datum.FallRateAlert = nil
						datum.RiseRateAlert = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/fallRate"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/riseRate"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(units *string, mutator func(datum *cgm.RateAlertsDEPRECATED, units *string), expectator func(datum *cgm.RateAlertsDEPRECATED, expectedDatum *cgm.RateAlertsDEPRECATED, units *string)) {
					for _, origin := range structure.Origins() {
						datum := NewRateAlertsDEPRECATED(units)
						mutator(datum, units)
						expectedDatum := CloneRateAlertsDEPRECATED(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin), units)
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum, units)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					pointer.FromString("mmol/L"),
					func(datum *cgm.RateAlertsDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units missing",
					nil,
					func(datum *cgm.RateAlertsDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *cgm.RateAlertsDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; fall rate alert missing",
					pointer.FromString("mmol/L"),
					func(datum *cgm.RateAlertsDEPRECATED, units *string) { datum.FallRateAlert = nil },
					nil,
				),
				Entry("does not modify the datum; rise rate alert missing",
					pointer.FromString("mmol/L"),
					func(datum *cgm.RateAlertsDEPRECATED, units *string) { datum.RiseRateAlert = nil },
					nil,
				),
			)

			DescribeTable("normalizes the datum with origin external",
				func(units *string, mutator func(datum *cgm.RateAlertsDEPRECATED, units *string), expectator func(datum *cgm.RateAlertsDEPRECATED, expectedDatum *cgm.RateAlertsDEPRECATED, units *string)) {
					datum := NewRateAlertsDEPRECATED(units)
					mutator(datum, units)
					expectedDatum := CloneRateAlertsDEPRECATED(datum)
					normalizer := dataNormalizer.New()
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(structure.OriginExternal), units)
					Expect(normalizer.Error()).To(BeNil())
					Expect(normalizer.Data()).To(BeEmpty())
					if expectator != nil {
						expectator(datum, expectedDatum, units)
					}
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *cgm.RateAlertsDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *cgm.RateAlertsDEPRECATED, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *cgm.RateAlertsDEPRECATED, units *string) {},
					func(datum *cgm.RateAlertsDEPRECATED, expectedDatum *cgm.RateAlertsDEPRECATED, units *string) {
						testDataBloodGlucose.ExpectNormalizedValue(datum.FallRateAlert.Rate, expectedDatum.FallRateAlert.Rate, units)
						testDataBloodGlucose.ExpectNormalizedValue(datum.RiseRateAlert.Rate, expectedDatum.RiseRateAlert.Rate, units)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *cgm.RateAlertsDEPRECATED, units *string) {},
					func(datum *cgm.RateAlertsDEPRECATED, expectedDatum *cgm.RateAlertsDEPRECATED, units *string) {
						testDataBloodGlucose.ExpectNormalizedValue(datum.FallRateAlert.Rate, expectedDatum.FallRateAlert.Rate, units)
						testDataBloodGlucose.ExpectNormalizedValue(datum.RiseRateAlert.Rate, expectedDatum.RiseRateAlert.Rate, units)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *cgm.RateAlertsDEPRECATED, units *string), expectator func(datum *cgm.RateAlertsDEPRECATED, expectedDatum *cgm.RateAlertsDEPRECATED, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := NewRateAlertsDEPRECATED(units)
						mutator(datum, units)
						expectedDatum := CloneRateAlertsDEPRECATED(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin), units)
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum, units)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *cgm.RateAlertsDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *cgm.RateAlertsDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *cgm.RateAlertsDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *cgm.RateAlertsDEPRECATED, units *string) {},
					nil,
				),
			)
		})
	})
})
