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

func NewFallRateAlert(units *string) *cgm.FallRateAlert {
	datum := cgm.NewFallRateAlert()
	datum.Enabled = pointer.FromBool(test.RandomBool())
	if rates := datum.RatesForUnits(units); len(rates) > 0 {
		datum.Rate = pointer.FromFloat64(test.RandomFloat64FromArray(rates))
	} else {
		datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(-math.MaxFloat64, math.MaxFloat64))
	}
	return datum
}

func CloneFallRateAlert(datum *cgm.FallRateAlert) *cgm.FallRateAlert {
	if datum == nil {
		return nil
	}
	clone := cgm.NewFallRateAlert()
	clone.Enabled = test.CloneBool(datum.Enabled)
	clone.Rate = test.CloneFloat64(datum.Rate)
	return clone
}

func NewRiseRateAlert(units *string) *cgm.RiseRateAlert {
	datum := cgm.NewRiseRateAlert()
	datum.Enabled = pointer.FromBool(test.RandomBool())
	if rates := datum.RatesForUnits(units); len(rates) > 0 {
		datum.Rate = pointer.FromFloat64(test.RandomFloat64FromArray(rates))
	} else {
		datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(-math.MaxFloat64, math.MaxFloat64))
	}
	return datum
}

func CloneRiseRateAlert(datum *cgm.RiseRateAlert) *cgm.RiseRateAlert {
	if datum == nil {
		return nil
	}
	clone := cgm.NewRiseRateAlert()
	clone.Enabled = test.CloneBool(datum.Enabled)
	clone.Rate = test.CloneFloat64(datum.Rate)
	return clone
}

func NewRateAlerts(units *string) *cgm.RateAlerts {
	datum := cgm.NewRateAlerts()
	datum.FallRateAlert = NewFallRateAlert(units)
	datum.RiseRateAlert = NewRiseRateAlert(units)
	return datum
}

func CloneRateAlerts(datum *cgm.RateAlerts) *cgm.RateAlerts {
	if datum == nil {
		return nil
	}
	clone := cgm.NewRateAlerts()
	clone.FallRateAlert = CloneFallRateAlert(datum.FallRateAlert)
	clone.RiseRateAlert = CloneRiseRateAlert(datum.RiseRateAlert)
	return clone
}

var _ = Describe("RateAlert", func() {
	It("RateMgdLThree is expected", func() {
		Expect(cgm.RateMgdLThree).To(Equal(3.0))
	})

	It("RateMgdLTwo is expected", func() {
		Expect(cgm.RateMgdLTwo).To(Equal(2.0))
	})

	It("RateMmolLThree is expected", func() {
		Expect(cgm.RateMmolLThree).To(Equal(0.16652243973136602))
	})

	It("RateMmolLTwo is expected", func() {
		Expect(cgm.RateMmolLTwo).To(Equal(0.11101495982091067))
	})

	Context("ParseFallRateAlert", func() {
		// TODO
	})

	Context("NewFallRateAlert", func() {
		It("is successful", func() {
			Expect(cgm.NewFallRateAlert()).To(Equal(&cgm.FallRateAlert{}))
		})
	})

	Context("FallRateAlert", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *cgm.FallRateAlert, units *string), expectedErrors ...error) {
					datum := NewFallRateAlert(units)
					mutator(datum, units)
					testDataTypes.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithStringAdapter(datum, units), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *cgm.FallRateAlert, units *string) {},
				),
				Entry("enabled missing",
					nil,
					func(datum *cgm.FallRateAlert, units *string) { datum.Enabled = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),
				Entry("enabled true",
					nil,
					func(datum *cgm.FallRateAlert, units *string) { datum.Enabled = pointer.FromBool(true) },
				),
				Entry("enabled false",
					nil,
					func(datum *cgm.FallRateAlert, units *string) { datum.Enabled = pointer.FromBool(false) },
				),
				Entry("units missing; rate missing",
					nil,
					func(datum *cgm.FallRateAlert, units *string) { datum.Rate = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units missing; rate valid",
					nil,
					func(datum *cgm.FallRateAlert, units *string) {
						datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(-math.MaxFloat64, math.MaxFloat64))
					},
				),
				Entry("units invalid; rate missing",
					pointer.FromString("invalid"),
					func(datum *cgm.FallRateAlert, units *string) { datum.Rate = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units invalid; rate valid",
					pointer.FromString("invalid"),
					func(datum *cgm.FallRateAlert, units *string) {
						datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(-math.MaxFloat64, math.MaxFloat64))
					},
				),
				Entry("units mmol/L; rate missing",
					pointer.FromString("mmol/L"),
					func(datum *cgm.FallRateAlert, units *string) { datum.Rate = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units mmol/L; rate invalid",
					pointer.FromString("mmol/L"),
					func(datum *cgm.FallRateAlert, units *string) {
						datum.Rate = pointer.FromFloat64(0.0)
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(0.0, []float64{-0.16652243973136602, -0.11101495982091067}), "/rate"),
				),
				Entry("units mmol/L; rate valid -3 mg/dL",
					pointer.FromString("mmol/L"),
					func(datum *cgm.FallRateAlert, units *string) {
						datum.Rate = pointer.FromFloat64(-0.16652243973136602)
					},
				),
				Entry("units mmol/L; rate valid -2 mg/dL",
					pointer.FromString("mmol/L"),
					func(datum *cgm.FallRateAlert, units *string) {
						datum.Rate = pointer.FromFloat64(-0.11101495982091067)
					},
				),
				Entry("units mmol/l; rate missing",
					pointer.FromString("mmol/l"),
					func(datum *cgm.FallRateAlert, units *string) { datum.Rate = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units mmol/l; rate invalid",
					pointer.FromString("mmol/l"),
					func(datum *cgm.FallRateAlert, units *string) {
						datum.Rate = pointer.FromFloat64(0.0)
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(0.0, []float64{-0.16652243973136602, -0.11101495982091067}), "/rate"),
				),
				Entry("units mmol/l; rate valid -3 mg/dL",
					pointer.FromString("mmol/l"),
					func(datum *cgm.FallRateAlert, units *string) {
						datum.Rate = pointer.FromFloat64(-0.16652243973136602)
					},
				),
				Entry("units mmol/l; rate valid -2 mg/dL",
					pointer.FromString("mmol/l"),
					func(datum *cgm.FallRateAlert, units *string) {
						datum.Rate = pointer.FromFloat64(-0.11101495982091067)
					},
				),
				Entry("units mg/dL; rate missing",
					pointer.FromString("mg/dL"),
					func(datum *cgm.FallRateAlert, units *string) { datum.Rate = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units mg/dL; rate invalid",
					pointer.FromString("mg/dL"),
					func(datum *cgm.FallRateAlert, units *string) {
						datum.Rate = pointer.FromFloat64(0.0)
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(0.0, []float64{-3.0, -2.0}), "/rate"),
				),
				Entry("units mg/dL; rate valid -3 mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *cgm.FallRateAlert, units *string) {
						datum.Rate = pointer.FromFloat64(-3.0)
					},
				),
				Entry("units mg/dL; rate valid -2 mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *cgm.FallRateAlert, units *string) {
						datum.Rate = pointer.FromFloat64(-2.0)
					},
				),
				Entry("units mg/dl; rate missing",
					pointer.FromString("mg/dl"),
					func(datum *cgm.FallRateAlert, units *string) { datum.Rate = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units mg/dl; rate invalid",
					pointer.FromString("mg/dl"),
					func(datum *cgm.FallRateAlert, units *string) {
						datum.Rate = pointer.FromFloat64(0.0)
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(0.0, []float64{-3.0, -2.0}), "/rate"),
				),
				Entry("units mg/dl; rate valid -3 mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *cgm.FallRateAlert, units *string) {
						datum.Rate = pointer.FromFloat64(-3.0)
					},
				),
				Entry("units mg/dl; rate valid -2 mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *cgm.FallRateAlert, units *string) {
						datum.Rate = pointer.FromFloat64(-2.0)
					},
				),
				Entry("multiple errors",
					pointer.FromString("mmol/L"),
					func(datum *cgm.FallRateAlert, units *string) {
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
				func(units *string, mutator func(datum *cgm.FallRateAlert, units *string), expectator func(datum *cgm.FallRateAlert, expectedDatum *cgm.FallRateAlert, units *string)) {
					for _, origin := range structure.Origins() {
						datum := NewFallRateAlert(units)
						mutator(datum, units)
						expectedDatum := CloneFallRateAlert(datum)
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
					func(datum *cgm.FallRateAlert, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units missing",
					nil,
					func(datum *cgm.FallRateAlert, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *cgm.FallRateAlert, units *string) {},
					nil,
				),
				Entry("does not modify the datum; enabled missing",
					pointer.FromString("mmol/L"),
					func(datum *cgm.FallRateAlert, units *string) { datum.Enabled = nil },
					nil,
				),
			)

			DescribeTable("normalizes the datum with origin external",
				func(units *string, mutator func(datum *cgm.FallRateAlert, units *string), expectator func(datum *cgm.FallRateAlert, expectedDatum *cgm.FallRateAlert, units *string)) {
					datum := NewFallRateAlert(units)
					mutator(datum, units)
					expectedDatum := CloneFallRateAlert(datum)
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
					func(datum *cgm.FallRateAlert, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *cgm.FallRateAlert, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *cgm.FallRateAlert, units *string) {},
					func(datum *cgm.FallRateAlert, expectedDatum *cgm.FallRateAlert, units *string) {
						testDataBloodGlucose.ExpectNormalizedValue(datum.Rate, expectedDatum.Rate, units)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *cgm.FallRateAlert, units *string) {},
					func(datum *cgm.FallRateAlert, expectedDatum *cgm.FallRateAlert, units *string) {
						testDataBloodGlucose.ExpectNormalizedValue(datum.Rate, expectedDatum.Rate, units)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *cgm.FallRateAlert, units *string), expectator func(datum *cgm.FallRateAlert, expectedDatum *cgm.FallRateAlert, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := NewFallRateAlert(units)
						mutator(datum, units)
						expectedDatum := CloneFallRateAlert(datum)
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
					func(datum *cgm.FallRateAlert, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *cgm.FallRateAlert, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *cgm.FallRateAlert, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *cgm.FallRateAlert, units *string) {},
					nil,
				),
			)
		})
	})

	Context("ParseRiseRateAlert", func() {
		// TODO
	})

	Context("NewRiseRateAlert", func() {
		It("is successful", func() {
			Expect(cgm.NewRiseRateAlert()).To(Equal(&cgm.RiseRateAlert{}))
		})
	})

	Context("RiseRateAlert", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *cgm.RiseRateAlert, units *string), expectedErrors ...error) {
					datum := NewRiseRateAlert(units)
					mutator(datum, units)
					testDataTypes.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithStringAdapter(datum, units), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *cgm.RiseRateAlert, units *string) {},
				),
				Entry("enabled missing",
					nil,
					func(datum *cgm.RiseRateAlert, units *string) { datum.Enabled = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),
				Entry("enabled true",
					nil,
					func(datum *cgm.RiseRateAlert, units *string) { datum.Enabled = pointer.FromBool(true) },
				),
				Entry("enabled false",
					nil,
					func(datum *cgm.RiseRateAlert, units *string) { datum.Enabled = pointer.FromBool(false) },
				),
				Entry("units missing; rate missing",
					nil,
					func(datum *cgm.RiseRateAlert, units *string) { datum.Rate = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units missing; rate valid",
					nil,
					func(datum *cgm.RiseRateAlert, units *string) {
						datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(-math.MaxFloat64, math.MaxFloat64))
					},
				),
				Entry("units invalid; rate missing",
					pointer.FromString("invalid"),
					func(datum *cgm.RiseRateAlert, units *string) { datum.Rate = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units invalid; rate valid",
					pointer.FromString("invalid"),
					func(datum *cgm.RiseRateAlert, units *string) {
						datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(-math.MaxFloat64, math.MaxFloat64))
					},
				),
				Entry("units mmol/L; rate missing",
					pointer.FromString("mmol/L"),
					func(datum *cgm.RiseRateAlert, units *string) { datum.Rate = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units mmol/L; rate invalid",
					pointer.FromString("mmol/L"),
					func(datum *cgm.RiseRateAlert, units *string) {
						datum.Rate = pointer.FromFloat64(0.0)
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(0.0, []float64{0.11101495982091067, 0.16652243973136602}), "/rate"),
				),
				Entry("units mmol/L; rate valid 2 mg/dL",
					pointer.FromString("mmol/L"),
					func(datum *cgm.RiseRateAlert, units *string) {
						datum.Rate = pointer.FromFloat64(0.11101495982091067)
					},
				),
				Entry("units mmol/L; rate valid 3 mg/dL",
					pointer.FromString("mmol/L"),
					func(datum *cgm.RiseRateAlert, units *string) {
						datum.Rate = pointer.FromFloat64(0.16652243973136602)
					},
				),
				Entry("units mmol/l; rate missing",
					pointer.FromString("mmol/l"),
					func(datum *cgm.RiseRateAlert, units *string) { datum.Rate = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units mmol/l; rate invalid",
					pointer.FromString("mmol/l"),
					func(datum *cgm.RiseRateAlert, units *string) {
						datum.Rate = pointer.FromFloat64(0.0)
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(0.0, []float64{0.11101495982091067, 0.16652243973136602}), "/rate"),
				),
				Entry("units mmol/l; rate valid 2 mg/dL",
					pointer.FromString("mmol/l"),
					func(datum *cgm.RiseRateAlert, units *string) {
						datum.Rate = pointer.FromFloat64(0.11101495982091067)
					},
				),
				Entry("units mmol/l; rate valid 3 mg/dL",
					pointer.FromString("mmol/l"),
					func(datum *cgm.RiseRateAlert, units *string) {
						datum.Rate = pointer.FromFloat64(0.16652243973136602)
					},
				),
				Entry("units mg/dL; rate missing",
					pointer.FromString("mg/dL"),
					func(datum *cgm.RiseRateAlert, units *string) { datum.Rate = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units mg/dL; rate invalid",
					pointer.FromString("mg/dL"),
					func(datum *cgm.RiseRateAlert, units *string) {
						datum.Rate = pointer.FromFloat64(0.0)
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(0.0, []float64{2.0, 3.0}), "/rate"),
				),
				Entry("units mg/dL; rate valid 2 mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *cgm.RiseRateAlert, units *string) {
						datum.Rate = pointer.FromFloat64(2.0)
					},
				),
				Entry("units mg/dL; rate valid 3 mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *cgm.RiseRateAlert, units *string) {
						datum.Rate = pointer.FromFloat64(3.0)
					},
				),
				Entry("units mg/dl; rate missing",
					pointer.FromString("mg/dl"),
					func(datum *cgm.RiseRateAlert, units *string) { datum.Rate = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units mg/dl; rate invalid",
					pointer.FromString("mg/dl"),
					func(datum *cgm.RiseRateAlert, units *string) {
						datum.Rate = pointer.FromFloat64(0.0)
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(0.0, []float64{2.0, 3.0}), "/rate"),
				),
				Entry("units mg/dl; rate valid 2 mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *cgm.RiseRateAlert, units *string) {
						datum.Rate = pointer.FromFloat64(2.0)
					},
				),
				Entry("units mg/dl; rate valid 3 mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *cgm.RiseRateAlert, units *string) {
						datum.Rate = pointer.FromFloat64(3.0)
					},
				),
				Entry("multiple errors",
					pointer.FromString("mmol/L"),
					func(datum *cgm.RiseRateAlert, units *string) {
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
				func(units *string, mutator func(datum *cgm.RiseRateAlert, units *string), expectator func(datum *cgm.RiseRateAlert, expectedDatum *cgm.RiseRateAlert, units *string)) {
					for _, origin := range structure.Origins() {
						datum := NewRiseRateAlert(units)
						mutator(datum, units)
						expectedDatum := CloneRiseRateAlert(datum)
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
					func(datum *cgm.RiseRateAlert, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units missing",
					nil,
					func(datum *cgm.RiseRateAlert, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *cgm.RiseRateAlert, units *string) {},
					nil,
				),
				Entry("does not modify the datum; enabled missing",
					pointer.FromString("mmol/L"),
					func(datum *cgm.RiseRateAlert, units *string) { datum.Enabled = nil },
					nil,
				),
			)

			DescribeTable("normalizes the datum with origin external",
				func(units *string, mutator func(datum *cgm.RiseRateAlert, units *string), expectator func(datum *cgm.RiseRateAlert, expectedDatum *cgm.RiseRateAlert, units *string)) {
					datum := NewRiseRateAlert(units)
					mutator(datum, units)
					expectedDatum := CloneRiseRateAlert(datum)
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
					func(datum *cgm.RiseRateAlert, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *cgm.RiseRateAlert, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *cgm.RiseRateAlert, units *string) {},
					func(datum *cgm.RiseRateAlert, expectedDatum *cgm.RiseRateAlert, units *string) {
						testDataBloodGlucose.ExpectNormalizedValue(datum.Rate, expectedDatum.Rate, units)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *cgm.RiseRateAlert, units *string) {},
					func(datum *cgm.RiseRateAlert, expectedDatum *cgm.RiseRateAlert, units *string) {
						testDataBloodGlucose.ExpectNormalizedValue(datum.Rate, expectedDatum.Rate, units)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *cgm.RiseRateAlert, units *string), expectator func(datum *cgm.RiseRateAlert, expectedDatum *cgm.RiseRateAlert, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := NewRiseRateAlert(units)
						mutator(datum, units)
						expectedDatum := CloneRiseRateAlert(datum)
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
					func(datum *cgm.RiseRateAlert, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *cgm.RiseRateAlert, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *cgm.RiseRateAlert, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *cgm.RiseRateAlert, units *string) {},
					nil,
				),
			)
		})
	})

	Context("ParseRateAlerts", func() {
		// TODO
	})

	Context("NewRateAlerts", func() {
		It("is successful", func() {
			Expect(cgm.NewRateAlerts()).To(Equal(&cgm.RateAlerts{}))
		})
	})

	Context("RateAlerts", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *cgm.RateAlerts, units *string), expectedErrors ...error) {
					datum := NewRateAlerts(units)
					mutator(datum, units)
					testDataTypes.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithStringAdapter(datum, units), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *cgm.RateAlerts, units *string) {},
				),
				Entry("fall rate alert missing",
					nil,
					func(datum *cgm.RateAlerts, units *string) { datum.FallRateAlert = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/fallRate"),
				),
				Entry("fall rate alert invalid",
					nil,
					func(datum *cgm.RateAlerts, units *string) { datum.FallRateAlert.Enabled = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/fallRate/enabled"),
				),
				Entry("fall rate alert valid",
					nil,
					func(datum *cgm.RateAlerts, units *string) { datum.FallRateAlert = NewFallRateAlert(units) },
				),
				Entry("rise rate alert missing",
					nil,
					func(datum *cgm.RateAlerts, units *string) { datum.RiseRateAlert = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/riseRate"),
				),
				Entry("rise rate alert invalid",
					nil,
					func(datum *cgm.RateAlerts, units *string) { datum.RiseRateAlert.Enabled = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/riseRate/enabled"),
				),
				Entry("rise rate alert valid",
					nil,
					func(datum *cgm.RateAlerts, units *string) { datum.RiseRateAlert = NewRiseRateAlert(units) },
				),
				Entry("multiple errors",
					pointer.FromString("mmol/L"),
					func(datum *cgm.RateAlerts, units *string) {
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
				func(units *string, mutator func(datum *cgm.RateAlerts, units *string), expectator func(datum *cgm.RateAlerts, expectedDatum *cgm.RateAlerts, units *string)) {
					for _, origin := range structure.Origins() {
						datum := NewRateAlerts(units)
						mutator(datum, units)
						expectedDatum := CloneRateAlerts(datum)
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
					func(datum *cgm.RateAlerts, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units missing",
					nil,
					func(datum *cgm.RateAlerts, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *cgm.RateAlerts, units *string) {},
					nil,
				),
				Entry("does not modify the datum; fall rate alert missing",
					pointer.FromString("mmol/L"),
					func(datum *cgm.RateAlerts, units *string) { datum.FallRateAlert = nil },
					nil,
				),
				Entry("does not modify the datum; rise rate alert missing",
					pointer.FromString("mmol/L"),
					func(datum *cgm.RateAlerts, units *string) { datum.RiseRateAlert = nil },
					nil,
				),
			)

			DescribeTable("normalizes the datum with origin external",
				func(units *string, mutator func(datum *cgm.RateAlerts, units *string), expectator func(datum *cgm.RateAlerts, expectedDatum *cgm.RateAlerts, units *string)) {
					datum := NewRateAlerts(units)
					mutator(datum, units)
					expectedDatum := CloneRateAlerts(datum)
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
					func(datum *cgm.RateAlerts, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *cgm.RateAlerts, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *cgm.RateAlerts, units *string) {},
					func(datum *cgm.RateAlerts, expectedDatum *cgm.RateAlerts, units *string) {
						testDataBloodGlucose.ExpectNormalizedValue(datum.FallRateAlert.Rate, expectedDatum.FallRateAlert.Rate, units)
						testDataBloodGlucose.ExpectNormalizedValue(datum.RiseRateAlert.Rate, expectedDatum.RiseRateAlert.Rate, units)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *cgm.RateAlerts, units *string) {},
					func(datum *cgm.RateAlerts, expectedDatum *cgm.RateAlerts, units *string) {
						testDataBloodGlucose.ExpectNormalizedValue(datum.FallRateAlert.Rate, expectedDatum.FallRateAlert.Rate, units)
						testDataBloodGlucose.ExpectNormalizedValue(datum.RiseRateAlert.Rate, expectedDatum.RiseRateAlert.Rate, units)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *cgm.RateAlerts, units *string), expectator func(datum *cgm.RateAlerts, expectedDatum *cgm.RateAlerts, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := NewRateAlerts(units)
						mutator(datum, units)
						expectedDatum := CloneRateAlerts(datum)
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
					func(datum *cgm.RateAlerts, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *cgm.RateAlerts, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *cgm.RateAlerts, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *cgm.RateAlerts, units *string) {},
					nil,
				),
			)
		})
	})
})
