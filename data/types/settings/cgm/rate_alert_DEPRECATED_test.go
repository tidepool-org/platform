package cgm_test

import (
	"math"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataBloodGlucoseTest "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTypesSettingsCgm "github.com/tidepool-org/platform/data/types/settings/cgm"
	dataTypesSettingsCgmTest "github.com/tidepool-org/platform/data/types/settings/cgm/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("RateAlertDEPRECATED", func() {
	It("RateDEPRECATEDMgdLThree is expected", func() {
		Expect(dataTypesSettingsCgm.RateDEPRECATEDMgdLThree).To(Equal(3.0))
	})

	It("RateDEPRECATEDMgdLTwo is expected", func() {
		Expect(dataTypesSettingsCgm.RateDEPRECATEDMgdLTwo).To(Equal(2.0))
	})

	It("RateDEPRECATEDMmolLThree is expected", func() {
		Expect(dataTypesSettingsCgm.RateDEPRECATEDMmolLThree).To(Equal(0.16652243973136602))
	})

	It("RateDEPRECATEDMmolLTwo is expected", func() {
		Expect(dataTypesSettingsCgm.RateDEPRECATEDMmolLTwo).To(Equal(0.11101495982091067))
	})

	Context("ParseFallRateAlertDEPRECATED", func() {
		// TODO
	})

	Context("NewFallRateAlertDEPRECATED", func() {
		It("is successful", func() {
			Expect(dataTypesSettingsCgm.NewFallRateAlertDEPRECATED()).To(Equal(&dataTypesSettingsCgm.FallRateAlertDEPRECATED{}))
		})
	})

	Context("FallRateAlertDEPRECATED", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string), expectedErrors ...error) {
					datum := dataTypesSettingsCgmTest.RandomFallRateAlertDEPRECATED(units)
					mutator(datum, units)
					dataTypesTest.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithStringAdapter(datum, units), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) {},
				),
				Entry("enabled missing",
					nil,
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) { datum.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),
				Entry("enabled false",
					nil,
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) {
						datum.Enabled = pointer.FromBool(false)
					},
				),
				Entry("enabled true",
					nil,
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) {
						datum.Enabled = pointer.FromBool(true)
					},
				),
				Entry("units missing; rate missing",
					nil,
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) { datum.Rate = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units missing; rate valid",
					nil,
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(-math.MaxFloat64, math.MaxFloat64))
					},
				),
				Entry("units invalid; rate missing",
					pointer.FromString("invalid"),
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) { datum.Rate = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units invalid; rate valid",
					pointer.FromString("invalid"),
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(-math.MaxFloat64, math.MaxFloat64))
					},
				),
				Entry("units mmol/L; rate missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) { datum.Rate = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units mmol/L; rate invalid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(0.0, []float64{-0.16652243973136602, -0.11101495982091067}), "/rate"),
				),
				Entry("units mmol/L; rate valid -3 mg/dL",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(-0.16652243973136602)
					},
				),
				Entry("units mmol/L; rate valid -2 mg/dL",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(-0.11101495982091067)
					},
				),
				Entry("units mmol/l; rate missing",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) { datum.Rate = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units mmol/l; rate invalid",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(0.0, []float64{-0.16652243973136602, -0.11101495982091067}), "/rate"),
				),
				Entry("units mmol/l; rate valid -3 mg/dL",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(-0.16652243973136602)
					},
				),
				Entry("units mmol/l; rate valid -2 mg/dL",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(-0.11101495982091067)
					},
				),
				Entry("units mg/dL; rate missing",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) { datum.Rate = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units mg/dL; rate invalid",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(0.0, []float64{-3.0, -2.0}), "/rate"),
				),
				Entry("units mg/dL; rate valid -3 mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(-3.0)
					},
				),
				Entry("units mg/dL; rate valid -2 mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(-2.0)
					},
				),
				Entry("units mg/dl; rate missing",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) { datum.Rate = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units mg/dl; rate invalid",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(0.0, []float64{-3.0, -2.0}), "/rate"),
				),
				Entry("units mg/dl; rate valid -3 mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(-3.0)
					},
				),
				Entry("units mg/dl; rate valid -2 mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(-2.0)
					},
				),
				Entry("multiple errors",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) {
						datum.Enabled = nil
						datum.Rate = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(units *string, mutator func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string), expectator func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, expectedDatum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesSettingsCgmTest.RandomFallRateAlertDEPRECATED(units)
						mutator(datum, units)
						expectedDatum := dataTypesSettingsCgmTest.CloneFallRateAlertDEPRECATED(datum)
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
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units missing",
					nil,
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; enabled missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) { datum.Enabled = nil },
					nil,
				),
			)

			DescribeTable("normalizes the datum with origin external",
				func(units *string, mutator func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string), expectator func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, expectedDatum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string)) {
					datum := dataTypesSettingsCgmTest.RandomFallRateAlertDEPRECATED(units)
					mutator(datum, units)
					expectedDatum := dataTypesSettingsCgmTest.CloneFallRateAlertDEPRECATED(datum)
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
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) {},
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, expectedDatum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) {
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.Rate, expectedDatum.Rate, units)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) {},
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, expectedDatum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) {
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.Rate, expectedDatum.Rate, units)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string), expectator func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, expectedDatum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := dataTypesSettingsCgmTest.RandomFallRateAlertDEPRECATED(units)
						mutator(datum, units)
						expectedDatum := dataTypesSettingsCgmTest.CloneFallRateAlertDEPRECATED(datum)
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
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsCgm.FallRateAlertDEPRECATED, units *string) {},
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
			Expect(dataTypesSettingsCgm.NewRiseRateAlertDEPRECATED()).To(Equal(&dataTypesSettingsCgm.RiseRateAlertDEPRECATED{}))
		})
	})

	Context("RiseRateAlertDEPRECATED", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string), expectedErrors ...error) {
					datum := dataTypesSettingsCgmTest.RandomRiseRateAlertDEPRECATED(units)
					mutator(datum, units)
					dataTypesTest.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithStringAdapter(datum, units), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) {},
				),
				Entry("enabled missing",
					nil,
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) { datum.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),
				Entry("enabled false",
					nil,
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Enabled = pointer.FromBool(false)
					},
				),
				Entry("enabled true",
					nil,
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Enabled = pointer.FromBool(true)
					},
				),
				Entry("units missing; rate missing",
					nil,
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) { datum.Rate = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units missing; rate valid",
					nil,
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(-math.MaxFloat64, math.MaxFloat64))
					},
				),
				Entry("units invalid; rate missing",
					pointer.FromString("invalid"),
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) { datum.Rate = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units invalid; rate valid",
					pointer.FromString("invalid"),
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(-math.MaxFloat64, math.MaxFloat64))
					},
				),
				Entry("units mmol/L; rate missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) { datum.Rate = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units mmol/L; rate invalid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(0.0, []float64{0.11101495982091067, 0.16652243973136602}), "/rate"),
				),
				Entry("units mmol/L; rate valid 2 mg/dL",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(0.11101495982091067)
					},
				),
				Entry("units mmol/L; rate valid 3 mg/dL",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(0.16652243973136602)
					},
				),
				Entry("units mmol/l; rate missing",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) { datum.Rate = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units mmol/l; rate invalid",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(0.0, []float64{0.11101495982091067, 0.16652243973136602}), "/rate"),
				),
				Entry("units mmol/l; rate valid 2 mg/dL",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(0.11101495982091067)
					},
				),
				Entry("units mmol/l; rate valid 3 mg/dL",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(0.16652243973136602)
					},
				),
				Entry("units mg/dL; rate missing",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) { datum.Rate = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units mg/dL; rate invalid",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(0.0, []float64{2.0, 3.0}), "/rate"),
				),
				Entry("units mg/dL; rate valid 2 mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(2.0)
					},
				),
				Entry("units mg/dL; rate valid 3 mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(3.0)
					},
				),
				Entry("units mg/dl; rate missing",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) { datum.Rate = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("units mg/dl; rate invalid",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(0.0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(0.0, []float64{2.0, 3.0}), "/rate"),
				),
				Entry("units mg/dl; rate valid 2 mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(2.0)
					},
				),
				Entry("units mg/dl; rate valid 3 mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Rate = pointer.FromFloat64(3.0)
					},
				),
				Entry("multiple errors",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) {
						datum.Enabled = nil
						datum.Rate = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(units *string, mutator func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string), expectator func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, expectedDatum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesSettingsCgmTest.RandomRiseRateAlertDEPRECATED(units)
						mutator(datum, units)
						expectedDatum := dataTypesSettingsCgmTest.CloneRiseRateAlertDEPRECATED(datum)
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
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units missing",
					nil,
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; enabled missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) { datum.Enabled = nil },
					nil,
				),
			)

			DescribeTable("normalizes the datum with origin external",
				func(units *string, mutator func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string), expectator func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, expectedDatum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string)) {
					datum := dataTypesSettingsCgmTest.RandomRiseRateAlertDEPRECATED(units)
					mutator(datum, units)
					expectedDatum := dataTypesSettingsCgmTest.CloneRiseRateAlertDEPRECATED(datum)
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
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) {},
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, expectedDatum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) {
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.Rate, expectedDatum.Rate, units)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) {},
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, expectedDatum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) {
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.Rate, expectedDatum.Rate, units)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string), expectator func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, expectedDatum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := dataTypesSettingsCgmTest.RandomRiseRateAlertDEPRECATED(units)
						mutator(datum, units)
						expectedDatum := dataTypesSettingsCgmTest.CloneRiseRateAlertDEPRECATED(datum)
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
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsCgm.RiseRateAlertDEPRECATED, units *string) {},
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
			Expect(dataTypesSettingsCgm.NewRateAlertsDEPRECATED()).To(Equal(&dataTypesSettingsCgm.RateAlertsDEPRECATED{}))
		})
	})

	Context("RateAlertsDEPRECATED", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *dataTypesSettingsCgm.RateAlertsDEPRECATED, units *string), expectedErrors ...error) {
					datum := dataTypesSettingsCgmTest.RandomRateAlertsDEPRECATED(units)
					mutator(datum, units)
					dataTypesTest.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithStringAdapter(datum, units), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.RateAlertsDEPRECATED, units *string) {},
				),
				Entry("fall rate alert missing",
					nil,
					func(datum *dataTypesSettingsCgm.RateAlertsDEPRECATED, units *string) { datum.FallRateAlert = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/fallRate"),
				),
				Entry("fall rate alert invalid",
					nil,
					func(datum *dataTypesSettingsCgm.RateAlertsDEPRECATED, units *string) {
						datum.FallRateAlert.Enabled = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/fallRate/enabled"),
				),
				Entry("fall rate alert valid",
					nil,
					func(datum *dataTypesSettingsCgm.RateAlertsDEPRECATED, units *string) {
						datum.FallRateAlert = dataTypesSettingsCgmTest.RandomFallRateAlertDEPRECATED(units)
					},
				),
				Entry("rise rate alert missing",
					nil,
					func(datum *dataTypesSettingsCgm.RateAlertsDEPRECATED, units *string) { datum.RiseRateAlert = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/riseRate"),
				),
				Entry("rise rate alert invalid",
					nil,
					func(datum *dataTypesSettingsCgm.RateAlertsDEPRECATED, units *string) {
						datum.RiseRateAlert.Enabled = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/riseRate/enabled"),
				),
				Entry("rise rate alert valid",
					nil,
					func(datum *dataTypesSettingsCgm.RateAlertsDEPRECATED, units *string) {
						datum.RiseRateAlert = dataTypesSettingsCgmTest.RandomRiseRateAlertDEPRECATED(units)
					},
				),
				Entry("multiple errors",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.RateAlertsDEPRECATED, units *string) {
						datum.FallRateAlert = nil
						datum.RiseRateAlert = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/fallRate"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/riseRate"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(units *string, mutator func(datum *dataTypesSettingsCgm.RateAlertsDEPRECATED, units *string), expectator func(datum *dataTypesSettingsCgm.RateAlertsDEPRECATED, expectedDatum *dataTypesSettingsCgm.RateAlertsDEPRECATED, units *string)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesSettingsCgmTest.RandomRateAlertsDEPRECATED(units)
						mutator(datum, units)
						expectedDatum := dataTypesSettingsCgmTest.CloneRateAlertsDEPRECATED(datum)
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
					func(datum *dataTypesSettingsCgm.RateAlertsDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units missing",
					nil,
					func(datum *dataTypesSettingsCgm.RateAlertsDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *dataTypesSettingsCgm.RateAlertsDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; fall rate alert missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.RateAlertsDEPRECATED, units *string) { datum.FallRateAlert = nil },
					nil,
				),
				Entry("does not modify the datum; rise rate alert missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.RateAlertsDEPRECATED, units *string) { datum.RiseRateAlert = nil },
					nil,
				),
			)

			DescribeTable("normalizes the datum with origin external",
				func(units *string, mutator func(datum *dataTypesSettingsCgm.RateAlertsDEPRECATED, units *string), expectator func(datum *dataTypesSettingsCgm.RateAlertsDEPRECATED, expectedDatum *dataTypesSettingsCgm.RateAlertsDEPRECATED, units *string)) {
					datum := dataTypesSettingsCgmTest.RandomRateAlertsDEPRECATED(units)
					mutator(datum, units)
					expectedDatum := dataTypesSettingsCgmTest.CloneRateAlertsDEPRECATED(datum)
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
					func(datum *dataTypesSettingsCgm.RateAlertsDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsCgm.RateAlertsDEPRECATED, units *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsCgm.RateAlertsDEPRECATED, units *string) {},
					func(datum *dataTypesSettingsCgm.RateAlertsDEPRECATED, expectedDatum *dataTypesSettingsCgm.RateAlertsDEPRECATED, units *string) {
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.FallRateAlert.Rate, expectedDatum.FallRateAlert.Rate, units)
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.RiseRateAlert.Rate, expectedDatum.RiseRateAlert.Rate, units)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsCgm.RateAlertsDEPRECATED, units *string) {},
					func(datum *dataTypesSettingsCgm.RateAlertsDEPRECATED, expectedDatum *dataTypesSettingsCgm.RateAlertsDEPRECATED, units *string) {
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.FallRateAlert.Rate, expectedDatum.FallRateAlert.Rate, units)
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.RiseRateAlert.Rate, expectedDatum.RiseRateAlert.Rate, units)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *dataTypesSettingsCgm.RateAlertsDEPRECATED, units *string), expectator func(datum *dataTypesSettingsCgm.RateAlertsDEPRECATED, expectedDatum *dataTypesSettingsCgm.RateAlertsDEPRECATED, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := dataTypesSettingsCgmTest.RandomRateAlertsDEPRECATED(units)
						mutator(datum, units)
						expectedDatum := dataTypesSettingsCgmTest.CloneRateAlertsDEPRECATED(datum)
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
					func(datum *dataTypesSettingsCgm.RateAlertsDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsCgm.RateAlertsDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsCgm.RateAlertsDEPRECATED, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsCgm.RateAlertsDEPRECATED, units *string) {},
					nil,
				),
			)
		})
	})
})
