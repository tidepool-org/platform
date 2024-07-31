package cgm_test

import (
	"sort"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataBloodGlucoseTest "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types"
	dataTypes "github.com/tidepool-org/platform/data/types"
	dataTypesSettingsCgm "github.com/tidepool-org/platform/data/types/settings/cgm"
	dataTypesSettingsCgmTest "github.com/tidepool-org/platform/data/types/settings/cgm/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureTest "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() interface{} {
	return &dataTypes.Meta{
		Type: "cgmSettings",
	}
}

var _ = Describe("CGM", func() {
	It("Type is expected", func() {
		Expect(dataTypesSettingsCgm.Type).To(Equal("cgmSettings"))
	})

	It("FirmwareVersionLengthMaximum is expected", func() {
		Expect(dataTypesSettingsCgm.FirmwareVersionLengthMaximum).To(Equal(100))
	})

	It("HardwareVersionLengthMaximum is expected", func() {
		Expect(dataTypesSettingsCgm.HardwareVersionLengthMaximum).To(Equal(100))
	})

	It("ManufacturerLengthMaximum is expected", func() {
		Expect(dataTypesSettingsCgm.ManufacturerLengthMaximum).To(Equal(100))
	})

	It("ManufacturersLengthMaximum is expected", func() {
		Expect(dataTypesSettingsCgm.ManufacturersLengthMaximum).To(Equal(10))
	})

	It("ModelLengthMaximum is expected", func() {
		Expect(dataTypesSettingsCgm.ModelLengthMaximum).To(Equal(100))
	})

	It("NameLengthMaximum is expected", func() {
		Expect(dataTypesSettingsCgm.NameLengthMaximum).To(Equal(100))
	})

	It("SerialNumberLengthMaximum is expected", func() {
		Expect(dataTypesSettingsCgm.SerialNumberLengthMaximum).To(Equal(100))
	})

	It("SoftwareVersionLengthMaximum is expected", func() {
		Expect(dataTypesSettingsCgm.SoftwareVersionLengthMaximum).To(Equal(100))
	})

	It("TransmitterIDExpressionString is expected", func() {
		Expect(dataTypesSettingsCgm.TransmitterIDExpressionString).To(Equal("^[0-9a-zA-Z]{5,64}$"))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := dataTypesSettingsCgm.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("cgmSettings"))
			Expect(datum.FirmwareVersion).To(BeNil())
			Expect(datum.HardwareVersion).To(BeNil())
			Expect(datum.Manufacturers).To(BeNil())
			Expect(datum.Model).To(BeNil())
			Expect(datum.Name).To(BeNil())
			Expect(datum.SerialNumber).To(BeNil())
			Expect(datum.SoftwareVersion).To(BeNil())
			Expect(datum.TransmitterID).To(BeNil())
			Expect(datum.Units).To(BeNil())
			Expect(datum.DefaultAlerts).To(BeNil())
			Expect(datum.ScheduledAlerts).To(BeNil())
			Expect(datum.HighLevelAlert).To(BeNil())
			Expect(datum.LowLevelAlert).To(BeNil())
			Expect(datum.OutOfRangeAlert).To(BeNil())
			Expect(datum.RateAlerts).To(BeNil())
		})
	})

	Context("CGM", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *dataTypesSettingsCgm.CGM, units *string), expectedErrors ...error) {
					datum := dataTypesSettingsCgmTest.RandomCGM(units)
					mutator(datum, units)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {},
				),
				Entry("type missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &dataTypes.Meta{}),
				),
				Entry("type invalid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "cgmSettings"), "/type", &dataTypes.Meta{Type: "invalidType"}),
				),
				Entry("type cgmSettings",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.Type = "cgmSettings" },
				),
				Entry("firmware version missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.FirmwareVersion = nil },
				),
				Entry("firmware version empty",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.FirmwareVersion = pointer.FromString("") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/firmwareVersion", NewMeta()),
				),
				Entry("firmware version length in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {
						datum.FirmwareVersion = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
				),
				Entry("firmware version length out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {
						datum.FirmwareVersion = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/firmwareVersion", NewMeta()),
				),
				Entry("hardware version missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.HardwareVersion = nil },
				),
				Entry("hardware version empty",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.HardwareVersion = pointer.FromString("") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/hardwareVersion", NewMeta()),
				),
				Entry("hardware version length in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {
						datum.HardwareVersion = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
				),
				Entry("hardware version length out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {
						datum.HardwareVersion = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/hardwareVersion", NewMeta()),
				),
				Entry("manufacturers missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.Manufacturers = nil },
				),
				Entry("manufacturers empty",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {
						datum.Manufacturers = pointer.FromStringArray([]string{})
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/manufacturers", NewMeta()),
				),
				Entry("manufacturers length; in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {
						datum.Manufacturers = pointer.FromStringArray(dataTypesSettingsCgmTest.RandomManufacturersFromRange(10, 10))
					},
				),
				Entry("manufacturers length; out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {
						datum.Manufacturers = pointer.FromStringArray(dataTypesSettingsCgmTest.RandomManufacturersFromRange(11, 11))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(11, 10), "/manufacturers", NewMeta()),
				),
				Entry("manufacturers manufacturer empty",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {
						datum.Manufacturers = pointer.FromStringArray(append([]string{test.RandomStringFromRange(1, 100), "", test.RandomStringFromRange(1, 100), ""}, dataTypesSettingsCgmTest.RandomManufacturersFromRange(0, 6)...))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/manufacturers/1", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/manufacturers/3", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueDuplicate(), "/manufacturers/3", NewMeta()),
				),
				Entry("manufacturers manufacturer length; in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {
						datum.Manufacturers = pointer.FromStringArray(append([]string{test.RandomStringFromRange(100, 100), test.RandomStringFromRange(1, 100), test.RandomStringFromRange(100, 100)}, dataTypesSettingsCgmTest.RandomManufacturersFromRange(0, 7)...))
					},
				),
				Entry("manufacturers manufacturer length; out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {
						datum.Manufacturers = pointer.FromStringArray(append([]string{test.RandomStringFromRange(101, 101), test.RandomStringFromRange(1, 100), test.RandomStringFromRange(101, 101)}, dataTypesSettingsCgmTest.RandomManufacturersFromRange(0, 7)...))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/manufacturers/0", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/manufacturers/2", NewMeta()),
				),
				Entry("model missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.Model = nil },
				),
				Entry("model empty",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.Model = pointer.FromString("") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/model", NewMeta()),
				),
				Entry("model length in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {
						datum.Model = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
				),
				Entry("model length out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {
						datum.Model = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/model", NewMeta()),
				),
				Entry("name missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.Name = nil },
				),
				Entry("name empty",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.Name = pointer.FromString("") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/name", NewMeta()),
				),
				Entry("name length in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {
						datum.Name = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
				),
				Entry("name length out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {
						datum.Name = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/name", NewMeta()),
				),
				Entry("serial number missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.SerialNumber = nil },
				),
				Entry("serial number empty",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.SerialNumber = pointer.FromString("") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/serialNumber", NewMeta()),
				),
				Entry("serial number length in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {
						datum.SerialNumber = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
				),
				Entry("serial number length out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {
						datum.SerialNumber = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/serialNumber", NewMeta()),
				),
				Entry("software version missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.SoftwareVersion = nil },
				),
				Entry("software version empty",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.SoftwareVersion = pointer.FromString("") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/softwareVersion", NewMeta()),
				),
				Entry("software version length in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {
						datum.SoftwareVersion = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
				),
				Entry("software version length out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {
						datum.SoftwareVersion = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/softwareVersion", NewMeta()),
				),
				Entry("transmitted id missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.TransmitterID = nil },
				),
				Entry("transmitted id empty",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.TransmitterID = pointer.FromString("") },
				),
				Entry("transmitted id invalid length",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.TransmitterID = pointer.FromString("ABC") },
					errorsTest.WithPointerSourceAndMeta(dataTypesSettingsCgm.ErrorValueStringAsTransmitterIDNotValid("ABC"), "/transmitterId", NewMeta()),
				),
				Entry("transmitted id invalid characters",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.TransmitterID = pointer.FromString("abc") },
					errorsTest.WithPointerSourceAndMeta(dataTypesSettingsCgm.ErrorValueStringAsTransmitterIDNotValid("abc"), "/transmitterId", NewMeta()),
				),
				Entry("transmitted id valid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {
						datum.TransmitterID = pointer.FromString(test.RandomStringFromRangeAndCharset(5, 6, dataTypesSettingsCgmTest.CharsetTransmitterID))
					},
				),
				Entry("units missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.Units = nil },
				),
				Entry("units invalid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.Units = pointer.FromString("invalid") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units valid; mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.Units = pointer.FromString("mmol/L") },
				),
				Entry("units valid; mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.Units = pointer.FromString("mmol/l") },
				),
				Entry("units valid; mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.Units = pointer.FromString("mg/dL") },
				),
				Entry("units valid; mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.Units = pointer.FromString("mg/dl") },
				),
				Entry("default alerts missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.DefaultAlerts = nil },
				),
				Entry("default alerts invalid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.DefaultAlerts.Enabled = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/defaultAlerts/enabled", NewMeta()),
				),
				Entry("default alerts valid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {
						datum.DefaultAlerts = dataTypesSettingsCgmTest.RandomAlerts()
					},
				),
				Entry("scheduled alerts missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.ScheduledAlerts = nil },
				),
				Entry("scheduled alerts invalid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { (*datum.ScheduledAlerts)[0].Days = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/scheduledAlerts/0/days", NewMeta()),
				),
				Entry("scheduled alerts valid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {
						datum.ScheduledAlerts = dataTypesSettingsCgmTest.RandomScheduledAlerts(1, 3)
					},
				),
				Entry("high level alert missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.HighLevelAlert = nil },
				),
				Entry("high level alert invalid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.HighLevelAlert.Enabled = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/highAlerts/enabled", NewMeta()),
				),
				Entry("high level alert valid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {
						datum.HighLevelAlert = dataTypesSettingsCgmTest.RandomHighLevelAlertDEPRECATED(units)
					},
				),
				Entry("low level alert missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.LowLevelAlert = nil },
				),
				Entry("low level alert invalid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.LowLevelAlert.Enabled = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/lowAlerts/enabled", NewMeta()),
				),
				Entry("low level alert valid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {
						datum.LowLevelAlert = dataTypesSettingsCgmTest.RandomLowLevelAlertDEPRECATED(units)
					},
				),
				Entry("out of range alert missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.OutOfRangeAlert = nil },
				),
				Entry("out of range alert invalid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.OutOfRangeAlert.Enabled = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/outOfRangeAlerts/enabled", NewMeta()),
				),
				Entry("out of range alert valid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {
						datum.OutOfRangeAlert = dataTypesSettingsCgmTest.RandomOutOfRangeAlertDEPRECATED()
					},
				),
				Entry("rate alerts missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.RateAlerts = nil },
				),
				Entry("rate alerts invalid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) { datum.RateAlerts.FallRateAlert = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/rateOfChangeAlerts/fallRate", NewMeta()),
				),
				Entry("rate alerts valid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {
						datum.RateAlerts = dataTypesSettingsCgmTest.RandomRateAlertsDEPRECATED(units)
					},
				),
				Entry("multiple errors",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {
						datum.Type = "invalidType"
						datum.FirmwareVersion = pointer.FromString("")
						datum.HardwareVersion = pointer.FromString("")
						datum.Manufacturers = pointer.FromStringArray([]string{})
						datum.Model = pointer.FromString("")
						datum.Name = pointer.FromString("")
						datum.SerialNumber = pointer.FromString("")
						datum.SoftwareVersion = pointer.FromString("")
						datum.TransmitterID = pointer.FromString("")
						datum.Units = pointer.FromString("invalid")
						datum.DefaultAlerts.Enabled = nil
						(*datum.ScheduledAlerts)[0].Days = nil
						datum.HighLevelAlert.Enabled = nil
						datum.LowLevelAlert.Enabled = nil
						datum.OutOfRangeAlert.Enabled = nil
						datum.RateAlerts.FallRateAlert = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "cgmSettings"), "/type", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/firmwareVersion", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/hardwareVersion", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/manufacturers", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/model", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/name", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/serialNumber", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/softwareVersion", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/defaultAlerts/enabled", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/scheduledAlerts/0/days", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/highAlerts/enabled", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/lowAlerts/enabled", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/outOfRangeAlerts/enabled", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/rateOfChangeAlerts/fallRate", &dataTypes.Meta{Type: "invalidType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum with origin external",
				func(units *string, mutator func(datum *dataTypesSettingsCgm.CGM, units *string), expectator func(datum *dataTypesSettingsCgm.CGM, expectedDatum *dataTypesSettingsCgm.CGM, units *string)) {
					datum := dataTypesSettingsCgmTest.RandomCGM(units)
					mutator(datum, units)
					expectedDatum := dataTypesSettingsCgmTest.CloneCGM(datum)
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
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {},
					func(datum *dataTypesSettingsCgm.CGM, expectedDatum *dataTypesSettingsCgm.CGM, units *string) {
						sort.Strings(*expectedDatum.Manufacturers)
					},
				),
				Entry("modifies the datum; units missing",
					nil,
					func(datum *dataTypesSettingsCgm.CGM, units *string) {},
					func(datum *dataTypesSettingsCgm.CGM, expectedDatum *dataTypesSettingsCgm.CGM, units *string) {
						sort.Strings(*expectedDatum.Manufacturers)
					},
				),
				Entry("modifies the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {},
					func(datum *dataTypesSettingsCgm.CGM, expectedDatum *dataTypesSettingsCgm.CGM, units *string) {
						sort.Strings(*expectedDatum.Manufacturers)
					},
				),
				Entry("modifies the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {},
					func(datum *dataTypesSettingsCgm.CGM, expectedDatum *dataTypesSettingsCgm.CGM, units *string) {
						sort.Strings(*expectedDatum.Manufacturers)
					},
				),
				Entry("modifies the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {},
					func(datum *dataTypesSettingsCgm.CGM, expectedDatum *dataTypesSettingsCgm.CGM, units *string) {
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
						sort.Strings(*expectedDatum.Manufacturers)
					},
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {},
					func(datum *dataTypesSettingsCgm.CGM, expectedDatum *dataTypesSettingsCgm.CGM, units *string) {
						sort.Strings(*expectedDatum.Manufacturers)
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.HighLevelAlert.Level, expectedDatum.HighLevelAlert.Level, units)
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.LowLevelAlert.Level, expectedDatum.LowLevelAlert.Level, units)
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.RateAlerts.FallRateAlert.Rate, expectedDatum.RateAlerts.FallRateAlert.Rate, units)
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.RateAlerts.RiseRateAlert.Rate, expectedDatum.RateAlerts.RiseRateAlert.Rate, units)
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {},
					func(datum *dataTypesSettingsCgm.CGM, expectedDatum *dataTypesSettingsCgm.CGM, units *string) {
						sort.Strings(*expectedDatum.Manufacturers)
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.HighLevelAlert.Level, expectedDatum.HighLevelAlert.Level, units)
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.LowLevelAlert.Level, expectedDatum.LowLevelAlert.Level, units)
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.RateAlerts.FallRateAlert.Rate, expectedDatum.RateAlerts.FallRateAlert.Rate, units)
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.RateAlerts.RiseRateAlert.Rate, expectedDatum.RateAlerts.RiseRateAlert.Rate, units)
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *dataTypesSettingsCgm.CGM, units *string), expectator func(datum *dataTypesSettingsCgm.CGM, expectedDatum *dataTypesSettingsCgm.CGM, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := dataTypesSettingsCgmTest.RandomCGM(units)
						mutator(datum, units)
						expectedDatum := dataTypesSettingsCgmTest.CloneCGM(datum)
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
					pointer.FromString("mmol/L"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesSettingsCgm.CGM, units *string) {},
					nil,
				),
			)
		})

		Context("LegacyIdentityFields", func() {
			It("returns the expected legacy identity fields", func() {
				datum := dataTypesSettingsCgmTest.RandomCGM(pointer.FromString("mmol/l"))
				datum.DeviceID = pointer.FromString("some-cgm-device")
				t, err := time.Parse(types.TimeFormat, "2023-05-13T15:51:58Z")
				Expect(err).ToNot(HaveOccurred())
				datum.Time = pointer.FromTime(t)
				legacyIdentityFields, err := datum.LegacyIdentityFields()
				Expect(err).ToNot(HaveOccurred())
				Expect(legacyIdentityFields).To(Equal([]string{"cgmSettings", "2023-05-13T15:51:58Z", "some-cgm-device"}))
			})
		})
	})

	Context("IsValidTransmitterID, TransmitterIDValidator, ValidateTransmitterID", func() {

		const dexcomStyleHashID = "6f1c584eb070e0e7ec3f8a9af313c34028374eee50928be47d807f333891369f"

		DescribeTable("validates the transmitter id",
			func(value string, expectedErrors ...error) {
				Expect(dataTypesSettingsCgm.IsValidTransmitterID(value)).To(Equal(len(expectedErrors) == 0))
				errorReporter := structureTest.NewErrorReporter()
				dataTypesSettingsCgm.TransmitterIDValidator(value, errorReporter)
				errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
				errorsTest.ExpectEqual(dataTypesSettingsCgm.ValidateTransmitterID(value), expectedErrors...)
			},
			Entry("is empty", ""),
			Entry("is valid", test.RandomStringFromRangeAndCharset(5, 6, dataTypesSettingsCgmTest.CharsetTransmitterID)),
			Entry("is valid when dexcom hash", dexcomStyleHashID),

			Entry("has invalid length; out of range (lower)", "ABCD", dataTypesSettingsCgm.ErrorValueStringAsTransmitterIDNotValid("ABCD")),
			Entry("has invalid length; in range (lower)", test.RandomStringFromRangeAndCharset(5, 5, dataTypesSettingsCgmTest.CharsetTransmitterID)),
			Entry("has invalid length; in range (upper)", dexcomStyleHashID+"m", dataTypesSettingsCgm.ErrorValueStringAsTransmitterIDNotValid(dexcomStyleHashID+"m")),
			Entry("has invalid characters; symbols", "@#$%^&", dataTypesSettingsCgm.ErrorValueStringAsTransmitterIDNotValid("@#$%^&")),
		)
	})

	Context("Errors", func() {
		DescribeTable("have expected details when error",
			errorsTest.ExpectErrorDetails,
			Entry("is ErrorValueStringAsTransmitterIDNotValid with empty string", dataTypesSettingsCgm.ErrorValueStringAsTransmitterIDNotValid(""), "value-not-valid", "value is not valid", `value "" is not valid as transmitter id`),
			Entry("is ErrorValueStringAsTransmitterIDNotValid with non-empty string", dataTypesSettingsCgm.ErrorValueStringAsTransmitterIDNotValid("ABCDEF"), "value-not-valid", "value is not valid", `value "ABCDEF" is not valid as transmitter id`),
		)
	})
})
