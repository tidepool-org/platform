package cgm_test

import (
	"math/rand"
	"sort"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"encoding/json"
	"fmt"

	testDataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/settings/cgm"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/errors"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	testStructure "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

const transmitterIDCharSet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func NewMeta() interface{} {
	return &types.Meta{
		Type: "cgmSettings",
	}
}

func NewManufacturer(minimumLength int, maximumLength int) string {
	return test.NewText(minimumLength, maximumLength)
}

func NewManufacturers(minimumLength int, maximumLength int) []string {
	result := make([]string, minimumLength+rand.Intn(maximumLength-minimumLength+1))
	for index := range result {
		result[index] = NewManufacturer(1, 100)
	}
	return result
}

func NewCGM(units *string) *cgm.CGM {
	datum := cgm.New()
	datum.Base = *testDataTypes.NewBase()
	datum.Type = "cgmSettings"
	datum.HighLevelAlert = NewHighLevelAlert(units)
	datum.LowLevelAlert = NewLowLevelAlert(units)
	datum.Manufacturers = pointer.StringArray(NewManufacturers(1, 10))
	datum.Model = pointer.String(test.NewText(1, 100))
	datum.OutOfRangeAlert = NewOutOfRangeAlert()
	datum.RateAlerts = NewRateAlerts(units)
	datum.SerialNumber = pointer.String(test.NewText(1, 100))
	datum.TransmitterID = pointer.String(test.NewVariableString(5, 6, transmitterIDCharSet))
	datum.Units = units
	return datum
}

func CloneCGM(datum *cgm.CGM) *cgm.CGM {
	if datum == nil {
		return nil
	}
	clone := cgm.New()
	clone.Base = *testDataTypes.CloneBase(&datum.Base)
	clone.HighLevelAlert = CloneHighLevelAlert(datum.HighLevelAlert)
	clone.LowLevelAlert = CloneLowLevelAlert(datum.LowLevelAlert)
	clone.Manufacturers = test.CloneStringArray(datum.Manufacturers)
	clone.Model = test.CloneString(datum.Model)
	clone.OutOfRangeAlert = CloneOutOfRangeAlert(datum.OutOfRangeAlert)
	clone.RateAlerts = CloneRateAlerts(datum.RateAlerts)
	clone.SerialNumber = test.CloneString(datum.SerialNumber)
	clone.TransmitterID = test.CloneString(datum.TransmitterID)
	clone.Units = test.CloneString(datum.Units)
	return clone
}

var _ = Describe("CGM", func() {
	It("Type is expected", func() {
		Expect(cgm.Type).To(Equal("cgmSettings"))
	})

	It("TransmitterIDExpressionString is expected", func() {
		Expect(cgm.TransmitterIDExpressionString).To(Equal("^[0-9A-Z]{5,6}$"))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := cgm.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("cgmSettings"))
			Expect(datum.HighLevelAlert).To(BeNil())
			Expect(datum.LowLevelAlert).To(BeNil())
			Expect(datum.Manufacturers).To(BeNil())
			Expect(datum.Model).To(BeNil())
			Expect(datum.OutOfRangeAlert).To(BeNil())
			Expect(datum.RateAlerts).To(BeNil())
			Expect(datum.SerialNumber).To(BeNil())
			Expect(datum.TransmitterID).To(BeNil())
			Expect(datum.Units).To(BeNil())
		})
	})

	Context("CGM", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *cgm.CGM, units *string), expectedErrors ...error) {
					datum := NewCGM(units)
					mutator(datum, units)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) {},
				),
				Entry("type missing",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) { datum.Type = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &types.Meta{}),
				),
				Entry("type invalid",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) { datum.Type = "invalidType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "cgmSettings"), "/type", &types.Meta{Type: "invalidType"}),
				),
				Entry("type cgmSettings",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) { datum.Type = "cgmSettings" },
				),
				Entry("high level alert missing",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) { datum.HighLevelAlert = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/highAlerts", NewMeta()),
				),
				Entry("high level alert invalid",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) { datum.HighLevelAlert.Enabled = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/highAlerts/enabled", NewMeta()),
				),
				Entry("high level alert valid",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) { datum.HighLevelAlert = NewHighLevelAlert(units) },
				),
				Entry("low level alert missing",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) { datum.LowLevelAlert = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/lowAlerts", NewMeta()),
				),
				Entry("low level alert invalid",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) { datum.LowLevelAlert.Enabled = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/lowAlerts/enabled", NewMeta()),
				),
				Entry("low level alert valid",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) { datum.LowLevelAlert = NewLowLevelAlert(units) },
				),
				Entry("manufacturers missing",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) { datum.Manufacturers = nil },
				),
				Entry("manufacturers empty",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) {
						datum.Manufacturers = pointer.StringArray([]string{})
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/manufacturers", NewMeta()),
				),
				Entry("manufacturers length; in range (upper)",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) {
						datum.Manufacturers = pointer.StringArray(NewManufacturers(10, 10))
					},
				),
				Entry("manufacturers length; out of range (upper)",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) {
						datum.Manufacturers = pointer.StringArray(NewManufacturers(11, 11))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(11, 10), "/manufacturers", NewMeta()),
				),
				Entry("manufacturers manufacturer empty",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) {
						datum.Manufacturers = pointer.StringArray(append([]string{NewManufacturer(1, 100), "", NewManufacturer(1, 100), ""}, NewManufacturers(0, 6)...))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/manufacturers/1", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/manufacturers/3", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueDuplicate(), "/manufacturers/3", NewMeta()),
				),
				Entry("manufacturers manufacturer length; in range (upper)",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) {
						datum.Manufacturers = pointer.StringArray(append([]string{NewManufacturer(100, 100), NewManufacturer(1, 100), NewManufacturer(100, 100)}, NewManufacturers(0, 7)...))
					},
				),
				Entry("manufacturers manufacturer length; out of range (upper)",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) {
						datum.Manufacturers = pointer.StringArray(append([]string{NewManufacturer(101, 101), NewManufacturer(1, 100), NewManufacturer(101, 101)}, NewManufacturers(0, 7)...))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/manufacturers/0", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/manufacturers/2", NewMeta()),
				),
				Entry("model missing",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) { datum.Model = nil },
				),
				Entry("model empty",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) { datum.Model = pointer.String("") },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/model", NewMeta()),
				),
				Entry("model length in range (upper)",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) { datum.Model = pointer.String(test.NewText(1, 100)) },
				),
				Entry("model length out of range (upper)",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) {
						datum.Model = pointer.String(test.NewText(101, 101))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/model", NewMeta()),
				),
				Entry("out of range alert missing",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) { datum.OutOfRangeAlert = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/outOfRangeAlerts", NewMeta()),
				),
				Entry("out of range alert invalid",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) { datum.OutOfRangeAlert.Enabled = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/outOfRangeAlerts/enabled", NewMeta()),
				),
				Entry("out of range alert valid",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) { datum.OutOfRangeAlert = NewOutOfRangeAlert() },
				),
				Entry("rate alerts missing",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) { datum.RateAlerts = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/rateOfChangeAlerts", NewMeta()),
				),
				Entry("rate alerts invalid",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) { datum.RateAlerts.FallRateAlert = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/rateOfChangeAlerts/fallRate", NewMeta()),
				),
				Entry("rate alerts valid",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) { datum.RateAlerts = NewRateAlerts(units) },
				),
				Entry("serial number missing",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) { datum.SerialNumber = nil },
				),
				Entry("serial number empty",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) { datum.SerialNumber = pointer.String("") },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/serialNumber", NewMeta()),
				),
				Entry("serial number length in range (upper)",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) {
						datum.SerialNumber = pointer.String(test.NewText(1, 100))
					},
				),
				Entry("serial number length out of range (upper)",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) {
						datum.SerialNumber = pointer.String(test.NewText(101, 101))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/serialNumber", NewMeta()),
				),
				Entry("transmitted id missing",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) { datum.TransmitterID = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/transmitterId", NewMeta()),
				),
				Entry("transmitted id empty",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) { datum.TransmitterID = pointer.String("") },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/transmitterId", NewMeta()),
				),
				Entry("transmitted id invalid length",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) { datum.TransmitterID = pointer.String("ABC") },
					testErrors.WithPointerSourceAndMeta(cgm.ErrorValueStringAsTransmitterIDNotValid("ABC"), "/transmitterId", NewMeta()),
				),
				Entry("transmitted id invalid characters",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) { datum.TransmitterID = pointer.String("abc") },
					testErrors.WithPointerSourceAndMeta(cgm.ErrorValueStringAsTransmitterIDNotValid("abc"), "/transmitterId", NewMeta()),
				),
				Entry("transmitted id valid",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) {
						datum.TransmitterID = pointer.String(test.NewVariableString(5, 6, transmitterIDCharSet))
					},
				),
				Entry("units missing",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) { datum.Units = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units invalid",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) { datum.Units = pointer.String("invalid") },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units valid; mmol/L",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) { datum.Units = pointer.String("mmol/L") },
				),
				Entry("units valid; mmol/l",
					pointer.String("mmol/l"),
					func(datum *cgm.CGM, units *string) { datum.Units = pointer.String("mmol/l") },
				),
				Entry("units valid; mg/dL",
					pointer.String("mg/dL"),
					func(datum *cgm.CGM, units *string) { datum.Units = pointer.String("mg/dL") },
				),
				Entry("units valid; mg/dl",
					pointer.String("mg/dl"),
					func(datum *cgm.CGM, units *string) { datum.Units = pointer.String("mg/dl") },
				),
				Entry("multiple errors",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) {
						datum.Type = "invalidType"
						datum.HighLevelAlert = nil
						datum.LowLevelAlert = nil
						datum.OutOfRangeAlert = nil
						datum.RateAlerts = nil
						datum.TransmitterID = nil
						datum.Units = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "cgmSettings"), "/type", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/highAlerts", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/lowAlerts", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/outOfRangeAlerts", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/rateOfChangeAlerts", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/transmitterId", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", &types.Meta{Type: "invalidType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum with origin external",
				func(units *string, mutator func(datum *cgm.CGM, units *string), expectator func(datum *cgm.CGM, expectedDatum *cgm.CGM, units *string)) {
					datum := NewCGM(units)
					mutator(datum, units)
					expectedDatum := CloneCGM(datum)
					normalizer := dataNormalizer.New()
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(structure.OriginExternal))
					Expect(normalizer.Error()).To(BeNil())
					Expect(normalizer.Data()).To(BeEmpty())
					if expectator != nil {
						expectator(datum, expectedDatum, units)
					}
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("modifies the datum",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) {},
					func(datum *cgm.CGM, expectedDatum *cgm.CGM, units *string) {
						sort.Strings(*expectedDatum.Manufacturers)
					},
				),
				Entry("modifies the datum; units missing",
					nil,
					func(datum *cgm.CGM, units *string) {},
					func(datum *cgm.CGM, expectedDatum *cgm.CGM, units *string) {
						sort.Strings(*expectedDatum.Manufacturers)
					},
				),
				Entry("modifies the datum; units invalid",
					pointer.String("invalid"),
					func(datum *cgm.CGM, units *string) {},
					func(datum *cgm.CGM, expectedDatum *cgm.CGM, units *string) {
						sort.Strings(*expectedDatum.Manufacturers)
					},
				),
				Entry("modifies the datum; units mmol/L",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) {},
					func(datum *cgm.CGM, expectedDatum *cgm.CGM, units *string) {
						sort.Strings(*expectedDatum.Manufacturers)
					},
				),
				Entry("modifies the datum; units mmol/l",
					pointer.String("mmol/l"),
					func(datum *cgm.CGM, units *string) {},
					func(datum *cgm.CGM, expectedDatum *cgm.CGM, units *string) {
						sort.Strings(*expectedDatum.Manufacturers)
						testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					},
				),
				Entry("modifies the datum; units mg/dL",
					pointer.String("mg/dL"),
					func(datum *cgm.CGM, units *string) {},
					func(datum *cgm.CGM, expectedDatum *cgm.CGM, units *string) {
						testDataBloodGlucose.ExpectNormalizedValue(datum.HighLevelAlert.Level, expectedDatum.HighLevelAlert.Level, units)
						testDataBloodGlucose.ExpectNormalizedValue(datum.LowLevelAlert.Level, expectedDatum.LowLevelAlert.Level, units)
						sort.Strings(*expectedDatum.Manufacturers)
						testDataBloodGlucose.ExpectNormalizedValue(datum.RateAlerts.FallRateAlert.Rate, expectedDatum.RateAlerts.FallRateAlert.Rate, units)
						testDataBloodGlucose.ExpectNormalizedValue(datum.RateAlerts.RiseRateAlert.Rate, expectedDatum.RateAlerts.RiseRateAlert.Rate, units)
						testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.String("mg/dl"),
					func(datum *cgm.CGM, units *string) {},
					func(datum *cgm.CGM, expectedDatum *cgm.CGM, units *string) {
						testDataBloodGlucose.ExpectNormalizedValue(datum.HighLevelAlert.Level, expectedDatum.HighLevelAlert.Level, units)
						testDataBloodGlucose.ExpectNormalizedValue(datum.LowLevelAlert.Level, expectedDatum.LowLevelAlert.Level, units)
						sort.Strings(*expectedDatum.Manufacturers)
						testDataBloodGlucose.ExpectNormalizedValue(datum.RateAlerts.FallRateAlert.Rate, expectedDatum.RateAlerts.FallRateAlert.Rate, units)
						testDataBloodGlucose.ExpectNormalizedValue(datum.RateAlerts.RiseRateAlert.Rate, expectedDatum.RateAlerts.RiseRateAlert.Rate, units)
						testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *cgm.CGM, units *string), expectator func(datum *cgm.CGM, expectedDatum *cgm.CGM, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := NewCGM(units)
						mutator(datum, units)
						expectedDatum := CloneCGM(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum, units)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum; units mmol/L",
					pointer.String("mmol/L"),
					func(datum *cgm.CGM, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.String("mmol/l"),
					func(datum *cgm.CGM, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dL",
					pointer.String("mg/dL"),
					func(datum *cgm.CGM, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dl",
					pointer.String("mg/dl"),
					func(datum *cgm.CGM, units *string) {},
					nil,
				),
			)
		})
	})

	Context("ValidateTransmitterID", func() {
		DescribeTable("validates the transmitter id",
			func(value string, expectedErrors ...error) {
				errorReporter := testStructure.NewErrorReporter()
				Expect(errorReporter).ToNot(BeNil())
				cgm.ValidateTransmitterID(value, errorReporter)
				testErrors.ExpectEqual(errorReporter.Error(), expectedErrors...)
			},
			Entry("is valid", test.NewVariableString(5, 6, transmitterIDCharSet)),
			Entry("is empty string", "", structureValidator.ErrorValueEmpty()),
			Entry("has invalid length; out of range (lower)", "ABCD", cgm.ErrorValueStringAsTransmitterIDNotValid("ABCD")),
			Entry("has invalid length; in range (lower)", test.NewString(5, transmitterIDCharSet)),
			Entry("has invalid length; in range (upper)", test.NewString(6, transmitterIDCharSet)),
			Entry("has invalid length; out of range (upper)", "ABCDEFG", cgm.ErrorValueStringAsTransmitterIDNotValid("ABCDEFG")),
			Entry("has invalid characters; lowercase", "abcdef", cgm.ErrorValueStringAsTransmitterIDNotValid("abcdef")),
			Entry("has invalid characters; symbols", "@#$%^&", cgm.ErrorValueStringAsTransmitterIDNotValid("@#$%^&")),
		)
	})

	Context("ErrorValueStringAsTransmitterIDNotValid", func() {
		DescribeTable("all errors",
			func(err error, code string, title string, detail string) {
				Expect(err).ToNot(BeNil())
				Expect(errors.Code(err)).To(Equal(code))
				Expect(errors.Cause(err)).To(Equal(err))
				bytes, bytesErr := json.Marshal(errors.Sanitize(err))
				Expect(bytesErr).ToNot(HaveOccurred())
				Expect(bytes).To(MatchJSON(fmt.Sprintf(`{"code": %q, "title": %q, "detail": %q}`, code, title, detail)))
			},
			Entry("is ErrorValueStringAsTransmitterIDNotValid with empty string", cgm.ErrorValueStringAsTransmitterIDNotValid(""), "value-not-valid", "value is not valid", `value "" is not valid as transmitter id`),
			Entry("is ErrorValueStringAsTransmitterIDNotValid with non-empty string", cgm.ErrorValueStringAsTransmitterIDNotValid("ABC"), "value-not-valid", "value is not valid", `value "ABC" is not valid as transmitter id`),
		)
	})
})
