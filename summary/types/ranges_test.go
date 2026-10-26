package types_test

import (
	"fmt"
	"math"
	"math/rand/v2"
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/pointer"
	. "github.com/tidepool-org/platform/summary/types"
)

var _ = Describe("Ranges", func() {
	// stored returns a reading as it's stored after normalization to mmol/L.
	stored := func(value float64, units string) float64 {
		return *glucose.NormalizeValueForUnits(&value, &units)
	}

	// native classifies a reading by the thresholds of its own unit, as a clinician
	// reading the thresholds would: at or above low and up to high is in range.
	native := func(t Thresholds, value float64) Classification {
		switch {
		case t.VeryLow != nil && value < *t.VeryLow:
			return ClassificationVeryLow
		case value < t.Low:
			return ClassificationLow
		case value <= t.High:
			return ClassificationInRange
		case t.VeryHigh == nil || value <= *t.VeryHigh:
			return ClassificationHigh
		default:
			return ClassificationVeryHigh
		}
	}

	// gridReadings returns every supported reading in the unit, at its resolution.
	gridReadings := func(units string) []float64 {
		var readings []float64
		if units == glucose.MgdL {
			for mgdL := 0; mgdL <= int(glucose.MgdLMaximum); mgdL++ {
				readings = append(readings, float64(mgdL))
			}
		} else {
			for tenths := 0; tenths <= int(glucose.MmolLMaximum*10); tenths++ {
				readings = append(readings, float64(tenths)/10)
			}
		}
		return readings
	}

	// classifiesNatively checks that the cut points classify every reading in each of
	// the range set's units by that unit's thresholds.
	classifiesNatively := func(rangeSet RangeSet) {
		GinkgoHelper()
		cutPoints, err := rangeSet.CutPoints()
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("%+v", rangeSet))
		for units, thresholds := range rangeSet.Thresholds {
			for _, reading := range gridReadings(units) {
				Expect(cutPoints.Classify(stored(reading, units))).
					To(Equal(native(thresholds, reading)), func() string {
						return fmt.Sprintf("%v %s with %+v", reading, units, thresholds)
					})
			}
		}
	}

	presets := []string{PresetADAStandard, PresetADAHighRisk, PresetADAPregnancyType1,
		PresetADAPregnancyType2}

	// custom returns a custom range set that must be valid.
	custom := func(units string, thresholds Thresholds) RangeSet {
		GinkgoHelper()
		rangeSet, err := NewCustomRangeSet("custom:test", units, thresholds)
		Expect(err).ToNot(HaveOccurred())
		return rangeSet
	}

	Describe("PresetRangeSet", func() {
		It("defines the presets as the frontend does", func() {
			standard, ok := PresetRangeSet(PresetADAStandard)
			Expect(ok).To(BeTrue())
			Expect(standard.Key).To(Equal("preset:adaStandard"))
			Expect(standard.Thresholds).To(Equal(map[string]Thresholds{
				glucose.MgdL: {VeryLow: pointer.FromFloat64(54), Low: 70, High: 180,
					VeryHigh: pointer.FromFloat64(250)},
				glucose.MmolL: {VeryLow: pointer.FromFloat64(3), Low: 3.9, High: 10,
					VeryHigh: pointer.FromFloat64(13.9)},
			}))

			highRisk, ok := PresetRangeSet(PresetADAHighRisk)
			Expect(ok).To(BeTrue())
			Expect(highRisk.Key).To(Equal("preset:adaHighRisk"))
			Expect(highRisk.Thresholds).To(Equal(map[string]Thresholds{
				glucose.MgdL:  {Low: 70, High: 180, VeryHigh: pointer.FromFloat64(250)},
				glucose.MmolL: {Low: 3.9, High: 10, VeryHigh: pointer.FromFloat64(13.9)},
			}))

			pregnancy := map[string]Thresholds{
				glucose.MgdL:  {VeryLow: pointer.FromFloat64(54), Low: 63, High: 140},
				glucose.MmolL: {VeryLow: pointer.FromFloat64(3), Low: 3.5, High: 7.8},
			}
			type1, ok := PresetRangeSet(PresetADAPregnancyType1)
			Expect(ok).To(BeTrue())
			Expect(type1.Key).To(Equal("preset:adaPregnancyType1"))
			Expect(type1.Thresholds).To(Equal(pregnancy))
			type2, ok := PresetRangeSet(PresetADAPregnancyType2)
			Expect(ok).To(BeTrue())
			Expect(type2.Key).To(Equal("preset:adaPregnancyType2"))
			Expect(type2.Thresholds).To(Equal(pregnancy))
		})

		It("doesn't know other presets", func() {
			_, ok := PresetRangeSet("adaNoSuchPreset")
			Expect(ok).To(BeFalse())
		})

		It("returns a copy that can be modified without affecting the preset", func() {
			standard, _ := PresetRangeSet(PresetADAStandard)
			*standard.Thresholds[glucose.MgdL].VeryLow = 1
			standard.Thresholds[glucose.MgdL] = Thresholds{Low: 1, High: 2}
			delete(standard.Thresholds, glucose.MmolL)

			again, _ := PresetRangeSet(PresetADAStandard)
			Expect(again.Thresholds).To(HaveLen(2))
			Expect(*again.Thresholds[glucose.MgdL].VeryLow).To(Equal(54.0))
			Expect(again.Thresholds[glucose.MgdL].Low).To(Equal(70.0))
		})
	})

	Describe("NewCustomRangeSet", func() {
		DescribeTable("converts the thresholds to the other unit as the presets pair them",
			func(preset string) {
				rangeSet, ok := PresetRangeSet(preset)
				Expect(ok).To(BeTrue())
				for _, units := range []string{glucose.MgdL, glucose.MmolL} {
					converted := custom(units, rangeSet.Thresholds[units])
					Expect(converted.Key).To(Equal("custom:test"))
					Expect(converted.Thresholds).To(Equal(rangeSet.Thresholds),
						fmt.Sprintf("from %s", units))
				}
			},
			Entry(PresetADAStandard, PresetADAStandard),
			Entry(PresetADAHighRisk, PresetADAHighRisk),
			Entry(PresetADAPregnancyType1, PresetADAPregnancyType1),
			Entry(PresetADAPregnancyType2, PresetADAPregnancyType2),
		)

		DescribeTable("converts a threshold",
			func(units string, thresholds Thresholds, other string, expected Thresholds) {
				Expect(custom(units, thresholds).Thresholds[other]).To(Equal(expected))
			},
			Entry("180 mg/dL to 10.0 mmol/L, not 9.99",
				glucose.MgdL, Thresholds{Low: 70, High: 180},
				glucose.MmolL, Thresholds{Low: 3.9, High: 10}),
			Entry("141 mg/dL to 7.8 mmol/L",
				glucose.MgdL, Thresholds{Low: 70, High: 141},
				glucose.MmolL, Thresholds{Low: 3.9, High: 7.8}),
			Entry("7.8 mmol/L to 140 mg/dL, the smaller of the values converting to it",
				glucose.MmolL, Thresholds{Low: 3.9, High: 7.8},
				glucose.MgdL, Thresholds{Low: 70, High: 140}),
			Entry("10.0 mmol/L to 180 mg/dL, not 181",
				glucose.MmolL, Thresholds{Low: 3.9, High: 10},
				glucose.MgdL, Thresholds{Low: 70, High: 180}),
			Entry("0.1 mmol/L to 1 mg/dL",
				glucose.MmolL, Thresholds{Low: 0.1, High: 10},
				glucose.MgdL, Thresholds{Low: 1, High: 180}),
		)

		It("leaves unset optional thresholds unset", func() {
			rangeSet := custom(glucose.MgdL, Thresholds{Low: 70, High: 180})
			Expect(rangeSet.Thresholds[glucose.MmolL]).
				To(Equal(Thresholds{Low: 3.9, High: 10}))
		})

		It("accepts the lowercase spellings of the units", func() {
			rangeSet := custom(glucose.Mgdl, Thresholds{Low: 70, High: 180})
			Expect(rangeSet.Thresholds).To(HaveKey(glucose.MgdL))
			Expect(rangeSet.Thresholds).To(HaveKey(glucose.MmolL))
			rangeSet = custom(glucose.Mmoll, Thresholds{Low: 3.9, High: 10})
			Expect(rangeSet.Thresholds[glucose.MgdL]).
				To(Equal(Thresholds{Low: 70, High: 180}))
		})

		It("doesn't share the thresholds given", func() {
			thresholds := Thresholds{VeryLow: pointer.FromFloat64(54), Low: 70, High: 180}
			rangeSet := custom(glucose.MgdL, thresholds)
			*thresholds.VeryLow = 1
			Expect(*rangeSet.Thresholds[glucose.MgdL].VeryLow).To(Equal(54.0))
		})

		DescribeTable("reports an error",
			func(units string, thresholds Thresholds, message string) {
				_, err := NewCustomRangeSet("custom:test", units, thresholds)
				Expect(err).To(MatchError(ContainSubstring(message)))
			},
			Entry("for unsupported units", "mg", Thresholds{Low: 70, High: 180},
				`unsupported units "mg"`),
			Entry("for a threshold above the unit's range", glucose.MgdL,
				Thresholds{Low: 70, High: 1200}, "high threshold 1200 is outside"),
			Entry("for a threshold above the unit's range in mmol/L", glucose.MmolL,
				Thresholds{Low: 3.9, High: 60}, "high threshold 60 is outside"),
			Entry("for a negative threshold", glucose.MgdL,
				Thresholds{VeryLow: pointer.FromFloat64(-1), Low: 70, High: 180},
				"veryLow threshold -1 is outside"),
			Entry("for thresholds that aren't ascending", glucose.MgdL,
				Thresholds{Low: 180, High: 70}, "aren't ascending"),
			Entry("for equal thresholds", glucose.MgdL,
				Thresholds{Low: 70, High: 180, VeryHigh: pointer.FromFloat64(180)},
				"aren't ascending"),
			Entry("for mg/dL thresholds that convert to the same mmol/L threshold",
				glucose.MgdL, Thresholds{Low: 104, High: 105},
				"converted to mmol/L, thresholds aren't ascending: mmol/L: "+
					"veryLow unset, low 5.8, high 5.8, veryHigh unset"),
			Entry("for a threshold that isn't a whole number of mg/dL", glucose.MgdL,
				Thresholds{Low: 70.5, High: 180},
				"low threshold 70.5 isn't a multiple of 1 mg/dL"),
			Entry("for a threshold that isn't a single decimal of mmol/L", glucose.MmolL,
				Thresholds{Low: 3.85, High: 10},
				"low threshold 3.85 isn't a multiple of 0.1 mmol/L"),
			Entry("for a threshold that isn't a number", glucose.MgdL,
				Thresholds{Low: 70, High: math.NaN()}, "high threshold isn't a number"),
		)
	})

	Describe("CutPoints", func() {
		It("derives cut points for the standard ranges that classify like the constants",
			func() {
				standard, _ := PresetRangeSet(PresetADAStandard)
				cutPoints, err := standard.CutPoints()
				Expect(err).ToNot(HaveOccurred())

				for _, units := range []string{glucose.MgdL, glucose.MmolL} {
					for _, reading := range gridReadings(units) {
						n := stored(reading, units)
						Expect(cutPoints.Classify(n)).
							To(Equal(StandardCutPoints().Classify(n)),
								fmt.Sprintf("%v %s", reading, units))
					}
				}
			})

		It("classifies every reading of each preset by its own unit's thresholds", func() {
			for _, preset := range presets {
				rangeSet, ok := PresetRangeSet(preset)
				Expect(ok).To(BeTrue())
				classifiesNatively(rangeSet)
			}
		})

		DescribeTable("classifies readings at the boundaries as BACK-4158 requires",
			func(preset string, units string, reading float64, expected Classification) {
				rangeSet, _ := PresetRangeSet(preset)
				cutPoints, err := rangeSet.CutPoints()
				Expect(err).ToNot(HaveOccurred())
				Expect(cutPoints.Classify(stored(reading, units))).To(Equal(expected))
			},
			Entry("standard 53 mg/dL", PresetADAStandard, glucose.MgdL, 53.0,
				ClassificationVeryLow),
			Entry("standard 54 mg/dL", PresetADAStandard, glucose.MgdL, 54.0,
				ClassificationLow),
			Entry("standard 69 mg/dL", PresetADAStandard, glucose.MgdL, 69.0,
				ClassificationLow),
			Entry("standard 70 mg/dL", PresetADAStandard, glucose.MgdL, 70.0,
				ClassificationInRange),
			Entry("standard 180 mg/dL", PresetADAStandard, glucose.MgdL, 180.0,
				ClassificationInRange),
			Entry("standard 181 mg/dL", PresetADAStandard, glucose.MgdL, 181.0,
				ClassificationHigh),
			Entry("standard 250 mg/dL", PresetADAStandard, glucose.MgdL, 250.0,
				ClassificationHigh),
			Entry("standard 251 mg/dL", PresetADAStandard, glucose.MgdL, 251.0,
				ClassificationVeryHigh),
			Entry("standard 2.9 mmol/L", PresetADAStandard, glucose.MmolL, 2.9,
				ClassificationVeryLow),
			Entry("standard 3.0 mmol/L", PresetADAStandard, glucose.MmolL, 3.0,
				ClassificationLow),
			Entry("standard 3.8 mmol/L", PresetADAStandard, glucose.MmolL, 3.8,
				ClassificationLow),
			Entry("standard 3.9 mmol/L", PresetADAStandard, glucose.MmolL, 3.9,
				ClassificationInRange),
			Entry("standard 10.0 mmol/L", PresetADAStandard, glucose.MmolL, 10.0,
				ClassificationInRange),
			Entry("standard 10.1 mmol/L", PresetADAStandard, glucose.MmolL, 10.1,
				ClassificationHigh),
			Entry("standard 13.9 mmol/L", PresetADAStandard, glucose.MmolL, 13.9,
				ClassificationHigh),
			Entry("standard 14.0 mmol/L", PresetADAStandard, glucose.MmolL, 14.0,
				ClassificationVeryHigh),
			Entry("pregnancy 62 mg/dL", PresetADAPregnancyType1, glucose.MgdL, 62.0,
				ClassificationLow),
			Entry("pregnancy 63 mg/dL", PresetADAPregnancyType1, glucose.MgdL, 63.0,
				ClassificationInRange),
			Entry("pregnancy 3.4 mmol/L", PresetADAPregnancyType1, glucose.MmolL, 3.4,
				ClassificationLow),
			Entry("pregnancy 3.5 mmol/L", PresetADAPregnancyType1, glucose.MmolL, 3.5,
				ClassificationInRange),
			Entry("pregnancy 140 mg/dL", PresetADAPregnancyType1, glucose.MgdL, 140.0,
				ClassificationInRange),
			Entry("pregnancy 141 mg/dL", PresetADAPregnancyType1, glucose.MgdL, 141.0,
				ClassificationHigh),
			Entry("pregnancy 7.8 mmol/L", PresetADAPregnancyType1, glucose.MmolL, 7.8,
				ClassificationInRange),
			Entry("pregnancy 7.9 mmol/L", PresetADAPregnancyType1, glucose.MmolL, 7.9,
				ClassificationHigh),
			Entry("pregnancy 400 mg/dL is high, having no very high range",
				PresetADAPregnancyType1, glucose.MgdL, 400.0, ClassificationHigh),
			Entry("high risk 40 mg/dL is low, having no very low range",
				PresetADAHighRisk, glucose.MgdL, 40.0, ClassificationLow),
		)

		It("classifies every reading of a custom range set by its unit's thresholds, "+
			"converted from the defining unit", func() {

			random := rand.New(rand.NewPCG(uint64(GinkgoRandomSeed()), 0))

			// randomThresholds picks ascending thresholds at the unit's resolution,
			// randomly leaving the optional ones unset. Thresholds in mg/dL are kept 2
			// apart so that they convert to distinct mmol/L thresholds.
			randomThresholds := func(units string) Thresholds {
				steps, scale, apart := 300, 1.0, 2
				if units == glucose.MmolL {
					steps, scale, apart = 160, 10, 1
				}
				picks := random.Perm(steps / apart)[:4]
				slices.Sort(picks)
				values := make([]float64, len(picks))
				for i, pick := range picks {
					values[i] = float64(apart*(1+pick)) / scale
				}
				thresholds := Thresholds{Low: values[1], High: values[2]}
				if random.IntN(2) == 0 {
					thresholds.VeryLow = pointer.FromFloat64(values[0])
				}
				if random.IntN(2) == 0 {
					thresholds.VeryHigh = pointer.FromFloat64(values[3])
				}
				return thresholds
			}

			for i := 0; i < 50; i++ {
				for _, units := range []string{glucose.MgdL, glucose.MmolL} {
					rangeSet := custom(units, randomThresholds(units))
					Expect(rangeSet.Thresholds).To(HaveLen(2))
					classifiesNatively(rangeSet)
				}
			}
		})

		It("classifies a custom range set like the preset it matches", func() {
			for _, preset := range presets {
				rangeSet, _ := PresetRangeSet(preset)
				presetCutPoints, err := rangeSet.CutPoints()
				Expect(err).ToNot(HaveOccurred())
				for _, units := range []string{glucose.MgdL, glucose.MmolL} {
					customCutPoints, err := custom(units, rangeSet.Thresholds[units]).
						CutPoints()
					Expect(err).ToNot(HaveOccurred())
					Expect(customCutPoints).To(Equal(presetCutPoints),
						fmt.Sprintf("%s from %s", preset, units))
				}
			}
		})

		DescribeTable("classifies readings in the other unit by the converted threshold",
			func(units string, thresholds Thresholds, readingUnits string, reading float64,
				expected Classification) {

				cutPoints, err := custom(units, thresholds).CutPoints()
				Expect(err).ToNot(HaveOccurred())
				Expect(cutPoints.Classify(stored(reading, readingUnits))).
					To(Equal(expected))
			},
			Entry("10.0 mmol/L is in range for a high of 180 mg/dL, though it's 180.16",
				glucose.MgdL, Thresholds{Low: 70, High: 180}, glucose.MmolL, 10.0,
				ClassificationInRange),
			Entry("10.1 mmol/L is high for a high of 180 mg/dL",
				glucose.MgdL, Thresholds{Low: 70, High: 180}, glucose.MmolL, 10.1,
				ClassificationHigh),
			Entry("141 mg/dL is high for a high of 7.8 mmol/L, though it's 7.83",
				glucose.MmolL, Thresholds{Low: 3.5, High: 7.8}, glucose.MgdL, 141.0,
				ClassificationHigh),
			Entry("140 mg/dL is in range for a high of 7.8 mmol/L",
				glucose.MmolL, Thresholds{Low: 3.5, High: 7.8}, glucose.MgdL, 140.0,
				ClassificationInRange),
			Entry("69 mg/dL is low for a low of 3.9 mmol/L",
				glucose.MmolL, Thresholds{Low: 3.9, High: 10}, glucose.MgdL, 69.0,
				ClassificationLow),
			Entry("70 mg/dL is in range for a low of 3.9 mmol/L, though it's 3.89",
				glucose.MmolL, Thresholds{Low: 3.9, High: 10}, glucose.MgdL, 70.0,
				ClassificationInRange),
		)

		It("leaves the cut points of unset thresholds infinite", func() {
			cutPoints, err := custom(glucose.MmolL, Thresholds{Low: 3.9, High: 10}).
				CutPoints()
			Expect(err).ToNot(HaveOccurred())
			Expect(math.IsInf(cutPoints.VeryLow, -1)).To(BeTrue())
			Expect(math.IsInf(cutPoints.VeryHigh, 1)).To(BeTrue())
			Expect(cutPoints.Classify(0)).To(Equal(ClassificationLow))
			Expect(cutPoints.Classify(50)).To(Equal(ClassificationHigh))
		})

		DescribeTable("reports an error",
			func(rangeSet RangeSet, message string) {
				_, err := rangeSet.CutPoints()
				Expect(err).To(MatchError(ContainSubstring(message)))
			},
			Entry("without thresholds", RangeSet{}, "no thresholds"),
			Entry("with unsupported units", RangeSet{
				Thresholds: map[string]Thresholds{"mg": {Low: 70, High: 180}},
			}, `unsupported units "mg"`),
			Entry("with thresholds that aren't ascending", RangeSet{
				Thresholds: map[string]Thresholds{glucose.MgdL: {Low: 180, High: 70}},
			}, "aren't ascending"),
			Entry("with a threshold outside the unit's range", RangeSet{
				Thresholds: map[string]Thresholds{glucose.MmolL: {Low: 3.9, High: 60}},
			}, "high threshold 60 is outside"),
			Entry("with units spelled other than the glucose constants", RangeSet{
				Thresholds: map[string]Thresholds{
					glucose.Mgdl:  {Low: 70, High: 180},
					glucose.MmolL: {Low: 3.9, High: 10},
				},
			}, `unsupported units "mg/dl": must be "mg/dL" or "mmol/L"`),
			Entry("with a threshold that isn't a whole number of mg/dL", RangeSet{
				Thresholds: map[string]Thresholds{
					glucose.MgdL:  {Low: 70.5, High: 180},
					glucose.MmolL: {Low: 3.9, High: 10},
				},
			}, "low threshold 70.5 isn't a multiple of 1 mg/dL"),
			Entry("with a threshold that isn't a number", RangeSet{
				Thresholds: map[string]Thresholds{
					glucose.MgdL:  {Low: 70, High: math.NaN()},
					glucose.MmolL: {Low: 3.9, High: 10},
				},
			}, "high threshold isn't a number"),
			Entry("with a threshold set in only some units", RangeSet{
				Key: "preset:test",
				Thresholds: map[string]Thresholds{
					glucose.MgdL:  {VeryLow: pointer.FromFloat64(54), Low: 70, High: 180},
					glucose.MmolL: {Low: 3.9, High: 10},
				},
			}, "veryLow in only some of its units"),
			Entry("with unit thresholds that leave no gap", RangeSet{
				Key: "preset:test",
				Thresholds: map[string]Thresholds{
					glucose.MgdL:  {Low: 70, High: 180},
					glucose.MmolL: {Low: 3.0, High: 10},
				},
			}, "low thresholds leave no gap"),
		)
	})
})
