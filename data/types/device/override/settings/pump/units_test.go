package pump_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataBloodGlucoseTest "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTypesDeviceOverrideSettingsPump "github.com/tidepool-org/platform/data/types/device/override/settings/pump"
	dataTypesDeviceOverrideSettingsPumpTest "github.com/tidepool-org/platform/data/types/device/override/settings/pump/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Units", func() {
	DescribeTable("serializes the datum as expected",
		func(mutator func(datum *dataTypesDeviceOverrideSettingsPump.Units)) {
			datum := dataTypesDeviceOverrideSettingsPumpTest.RandomUnits()
			mutator(datum)
			test.ExpectSerializedObjectJSON(datum, dataTypesDeviceOverrideSettingsPumpTest.NewObjectFromUnits(datum, test.ObjectFormatJSON))
			test.ExpectSerializedObjectBSON(datum, dataTypesDeviceOverrideSettingsPumpTest.NewObjectFromUnits(datum, test.ObjectFormatBSON))
		},
		Entry("succeeds",
			func(datum *dataTypesDeviceOverrideSettingsPump.Units) {},
		),
		Entry("empty",
			func(datum *dataTypesDeviceOverrideSettingsPump.Units) {
				*datum = *dataTypesDeviceOverrideSettingsPump.NewUnits()
			},
		),
		Entry("all",
			func(datum *dataTypesDeviceOverrideSettingsPump.Units) {
				datum.BloodGlucose = pointer.FromString(dataBloodGlucoseTest.RandomUnits())
			},
		),
	)

	Context("ParseUnits", func() {
		It("returns nil when the object is missing", func() {
			Expect(dataTypesDeviceOverrideSettingsPump.ParseUnits(structureParser.NewObject(nil))).To(BeNil())
		})

		It("returns new datum when the object is valid", func() {
			datum := dataTypesDeviceOverrideSettingsPumpTest.RandomUnits()
			object := dataTypesDeviceOverrideSettingsPumpTest.NewObjectFromUnits(datum, test.ObjectFormatJSON)
			parser := structureParser.NewObject(&object)
			Expect(dataTypesDeviceOverrideSettingsPump.ParseUnits(parser)).To(Equal(datum))
			Expect(parser.Error()).ToNot(HaveOccurred())
		})
	})

	Context("NewUnits", func() {
		It("is successful", func() {
			Expect(dataTypesDeviceOverrideSettingsPump.NewUnits()).To(Equal(&dataTypesDeviceOverrideSettingsPump.Units{}))
		})
	})

	Context("Parse", func() {
		DescribeTable("parses the datum",
			func(mutator func(object map[string]interface{}, expectedDatum *dataTypesDeviceOverrideSettingsPump.Units), expectedErrors ...error) {
				expectedDatum := dataTypesDeviceOverrideSettingsPumpTest.RandomUnits()
				object := dataTypesDeviceOverrideSettingsPumpTest.NewObjectFromUnits(expectedDatum, test.ObjectFormatJSON)
				mutator(object, expectedDatum)
				datum := &dataTypesDeviceOverrideSettingsPump.Units{}
				errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
				Expect(datum).To(Equal(expectedDatum))
			},
			Entry("succeeds",
				func(object map[string]interface{}, expectedDatum *dataTypesDeviceOverrideSettingsPump.Units) {},
			),
			Entry("multiple errors",
				func(object map[string]interface{}, expectedDatum *dataTypesDeviceOverrideSettingsPump.Units) {
					object["bg"] = true
					expectedDatum.BloodGlucose = nil
				},
				errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/bg"),
			),
		)
	})

	Context("Validate", func() {
		DescribeTable("validates the datum",
			func(mutator func(datum *dataTypesDeviceOverrideSettingsPump.Units), expectedErrors ...error) {
				datum := dataTypesDeviceOverrideSettingsPumpTest.RandomUnits()
				mutator(datum)
				dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
			},
			Entry("succeeds",
				func(datum *dataTypesDeviceOverrideSettingsPump.Units) {},
			),
			Entry("blood glucose missing",
				func(datum *dataTypesDeviceOverrideSettingsPump.Units) { datum.BloodGlucose = nil },
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/bg"),
			),
			Entry("blood glucose invalid",
				func(datum *dataTypesDeviceOverrideSettingsPump.Units) {
					datum.BloodGlucose = pointer.FromString("invalid")
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/bg"),
			),
			Entry("blood glucose mmol/L",
				func(datum *dataTypesDeviceOverrideSettingsPump.Units) {
					datum.BloodGlucose = pointer.FromString("mmol/L")
				},
			),
			Entry("blood glucose mmol/l",
				func(datum *dataTypesDeviceOverrideSettingsPump.Units) {
					datum.BloodGlucose = pointer.FromString("mmol/l")
				},
			),
			Entry("blood glucose mg/dL",
				func(datum *dataTypesDeviceOverrideSettingsPump.Units) {
					datum.BloodGlucose = pointer.FromString("mg/dL")
				},
			),
			Entry("blood glucose mg/dl",
				func(datum *dataTypesDeviceOverrideSettingsPump.Units) {
					datum.BloodGlucose = pointer.FromString("mg/dl")
				},
			),
		)
	})

	Context("Normalize", func() {
		DescribeTable("normalizes the datum",
			func(mutator func(datum *dataTypesDeviceOverrideSettingsPump.Units), expectator func(datum *dataTypesDeviceOverrideSettingsPump.Units, expectedDatum *dataTypesDeviceOverrideSettingsPump.Units)) {
				for _, origin := range structure.Origins() {
					datum := dataTypesDeviceOverrideSettingsPumpTest.RandomUnits()
					mutator(datum)
					expectedDatum := dataTypesDeviceOverrideSettingsPumpTest.CloneUnits(datum)
					normalizer := dataNormalizer.New()
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(origin))
					Expect(normalizer.Error()).To(BeNil())
					Expect(normalizer.Data()).To(BeEmpty())
					if expectator != nil {
						expectator(datum, expectedDatum)
					}
					Expect(datum).To(Equal(expectedDatum))
				}
			},
			Entry("does not modify the datum; blood glucose missing",
				func(datum *dataTypesDeviceOverrideSettingsPump.Units) { datum.BloodGlucose = nil },
				nil,
			),
			Entry("does not modify the datum; blood glucose invalid",
				func(datum *dataTypesDeviceOverrideSettingsPump.Units) {
					datum.BloodGlucose = pointer.FromString("invalid")
				},
				nil,
			),
		)

		DescribeTable("normalizes the datum with origin external",
			func(mutator func(datum *dataTypesDeviceOverrideSettingsPump.Units), expectator func(datum *dataTypesDeviceOverrideSettingsPump.Units, expectedDatum *dataTypesDeviceOverrideSettingsPump.Units)) {
				datum := dataTypesDeviceOverrideSettingsPumpTest.RandomUnits()
				mutator(datum)
				expectedDatum := dataTypesDeviceOverrideSettingsPumpTest.CloneUnits(datum)
				normalizer := dataNormalizer.New()
				Expect(normalizer).ToNot(BeNil())
				datum.Normalize(normalizer.WithOrigin(structure.OriginExternal))
				Expect(normalizer.Error()).To(BeNil())
				Expect(normalizer.Data()).To(BeEmpty())
				if expectator != nil {
					expectator(datum, expectedDatum)
				}
				Expect(datum).To(Equal(expectedDatum))
			},
			Entry("does not modify the datum; blood glucose mmol/L",
				func(datum *dataTypesDeviceOverrideSettingsPump.Units) {
					datum.BloodGlucose = pointer.FromString("mmol/L")
				},
				nil,
			),
			Entry("modifies the datum; blood glucose mmol/l",
				func(datum *dataTypesDeviceOverrideSettingsPump.Units) {
					datum.BloodGlucose = pointer.FromString("mmol/l")
				},
				func(datum *dataTypesDeviceOverrideSettingsPump.Units, expectedDatum *dataTypesDeviceOverrideSettingsPump.Units) {
					dataBloodGlucoseTest.ExpectNormalizedUnits(datum.BloodGlucose, expectedDatum.BloodGlucose)
				},
			),
			Entry("modifies the datum; blood glucose mg/dL",
				func(datum *dataTypesDeviceOverrideSettingsPump.Units) {
					datum.BloodGlucose = pointer.FromString("mg/dL")
				},
				func(datum *dataTypesDeviceOverrideSettingsPump.Units, expectedDatum *dataTypesDeviceOverrideSettingsPump.Units) {
					dataBloodGlucoseTest.ExpectNormalizedUnits(datum.BloodGlucose, expectedDatum.BloodGlucose)
				},
			),
			Entry("modifies the datum; blood glucose mg/dl",
				func(datum *dataTypesDeviceOverrideSettingsPump.Units) {
					datum.BloodGlucose = pointer.FromString("mg/dl")
				},
				func(datum *dataTypesDeviceOverrideSettingsPump.Units, expectedDatum *dataTypesDeviceOverrideSettingsPump.Units) {
					dataBloodGlucoseTest.ExpectNormalizedUnits(datum.BloodGlucose, expectedDatum.BloodGlucose)
				},
			),
		)

		DescribeTable("normalizes the datum with origin internal/store",
			func(mutator func(datum *dataTypesDeviceOverrideSettingsPump.Units), expectator func(datum *dataTypesDeviceOverrideSettingsPump.Units, expectedDatum *dataTypesDeviceOverrideSettingsPump.Units)) {
				for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
					datum := dataTypesDeviceOverrideSettingsPumpTest.RandomUnits()
					mutator(datum)
					expectedDatum := dataTypesDeviceOverrideSettingsPumpTest.CloneUnits(datum)
					normalizer := dataNormalizer.New()
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(origin))
					Expect(normalizer.Error()).To(BeNil())
					Expect(normalizer.Data()).To(BeEmpty())
					if expectator != nil {
						expectator(datum, expectedDatum)
					}
					Expect(datum).To(Equal(expectedDatum))
				}
			},
			Entry("does not modify the datum; blood glucose mmol/L",
				func(datum *dataTypesDeviceOverrideSettingsPump.Units) {
					datum.BloodGlucose = pointer.FromString("mmol/L")
				},
				nil,
			),
			Entry("does not modify the datum; blood glucose mmol/l",
				func(datum *dataTypesDeviceOverrideSettingsPump.Units) {
					datum.BloodGlucose = pointer.FromString("mmol/l")
				},
				nil,
			),
			Entry("does not modify the datum; blood glucose mg/dL",
				func(datum *dataTypesDeviceOverrideSettingsPump.Units) {
					datum.BloodGlucose = pointer.FromString("mg/dL")
				},
				nil,
			),
			Entry("does not modify the datum; blood glucose mg/dl",
				func(datum *dataTypesDeviceOverrideSettingsPump.Units) {
					datum.BloodGlucose = pointer.FromString("mg/dl")
				},
				nil,
			),
		)
	})
})
