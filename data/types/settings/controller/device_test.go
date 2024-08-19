package controller_test

import (
	"sort"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTypesSettingsController "github.com/tidepool-org/platform/data/types/settings/controller"
	dataTypesSettingsControllerTest "github.com/tidepool-org/platform/data/types/settings/controller/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Device", func() {
	It("FirmwareVersionLengthMaximum is expected", func() {
		Expect(dataTypesSettingsController.FirmwareVersionLengthMaximum).To(Equal(100))
	})

	It("HardwareVersionLengthMaximum is expected", func() {
		Expect(dataTypesSettingsController.HardwareVersionLengthMaximum).To(Equal(100))
	})

	It("ManufacturerLengthMaximum is expected", func() {
		Expect(dataTypesSettingsController.ManufacturerLengthMaximum).To(Equal(100))
	})

	It("ManufacturersLengthMaximum is expected", func() {
		Expect(dataTypesSettingsController.ManufacturersLengthMaximum).To(Equal(10))
	})

	It("ModelLengthMaximum is expected", func() {
		Expect(dataTypesSettingsController.ModelLengthMaximum).To(Equal(100))
	})

	It("NameLengthMaximum is expected", func() {
		Expect(dataTypesSettingsController.NameLengthMaximum).To(Equal(100))
	})

	It("SerialNumberLengthMaximum is expected", func() {
		Expect(dataTypesSettingsController.SerialNumberLengthMaximum).To(Equal(100))
	})

	It("SoftwareVersionLengthMaximum is expected", func() {
		Expect(dataTypesSettingsController.SoftwareVersionLengthMaximum).To(Equal(100))
	})

	Context("Device", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesSettingsController.Device)) {
				datum := dataTypesSettingsControllerTest.RandomDevice()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesSettingsControllerTest.NewObjectFromDevice(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesSettingsControllerTest.NewObjectFromDevice(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesSettingsController.Device) {},
			),
			Entry("empty",
				func(datum *dataTypesSettingsController.Device) {
					*datum = *dataTypesSettingsController.NewDevice()
				},
			),
			Entry("all",
				func(datum *dataTypesSettingsController.Device) {
					datum.FirmwareVersion = pointer.FromString(test.RandomStringFromRange(1, dataTypesSettingsController.FirmwareVersionLengthMaximum))
					datum.HardwareVersion = pointer.FromString(test.RandomStringFromRange(1, dataTypesSettingsController.HardwareVersionLengthMaximum))
					datum.Manufacturers = pointer.FromStringArray(dataTypesSettingsControllerTest.RandomManufacturersFromRange(1, dataTypesSettingsController.ManufacturersLengthMaximum))
					datum.Model = pointer.FromString(test.RandomStringFromRange(1, dataTypesSettingsController.ModelLengthMaximum))
					datum.Name = pointer.FromString(test.RandomStringFromRange(1, dataTypesSettingsController.NameLengthMaximum))
					datum.SerialNumber = pointer.FromString(test.RandomStringFromRange(1, dataTypesSettingsController.SerialNumberLengthMaximum))
					datum.SoftwareVersion = pointer.FromString(test.RandomStringFromRange(1, dataTypesSettingsController.SoftwareVersionLengthMaximum))
				},
			),
		)

		Context("ParseDevice", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataTypesSettingsController.ParseDevice(structureParser.NewObject(nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataTypesSettingsControllerTest.RandomDevice()
				object := dataTypesSettingsControllerTest.NewObjectFromDevice(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(&object)
				Expect(dataTypesSettingsController.ParseDevice(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewDevice", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesSettingsController.NewDevice()
				Expect(datum).ToNot(BeNil())
				Expect(datum.FirmwareVersion).To(BeNil())
				Expect(datum.HardwareVersion).To(BeNil())
				Expect(datum.Manufacturers).To(BeNil())
				Expect(datum.Model).To(BeNil())
				Expect(datum.Name).To(BeNil())
				Expect(datum.SerialNumber).To(BeNil())
				Expect(datum.SoftwareVersion).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesSettingsController.Device), expectedErrors ...error) {
					expectedDatum := dataTypesSettingsControllerTest.RandomDevice()
					object := dataTypesSettingsControllerTest.NewObjectFromDevice(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesSettingsController.NewDevice()
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesSettingsController.Device) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesSettingsController.Device) {
						object["firmwareVersion"] = true
						object["hardwareVersion"] = true
						object["manufacturers"] = true
						object["model"] = true
						object["name"] = true
						object["serialNumber"] = true
						object["softwareVersion"] = true
						expectedDatum.FirmwareVersion = nil
						expectedDatum.HardwareVersion = nil
						expectedDatum.Manufacturers = nil
						expectedDatum.Model = nil
						expectedDatum.Name = nil
						expectedDatum.SerialNumber = nil
						expectedDatum.SoftwareVersion = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/firmwareVersion"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/hardwareVersion"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/manufacturers"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/model"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/name"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/serialNumber"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/softwareVersion"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesSettingsController.Device), expectedErrors ...error) {
					datum := dataTypesSettingsControllerTest.RandomDevice()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesSettingsController.Device) {},
				),
				Entry("firmware version missing",
					func(datum *dataTypesSettingsController.Device) { datum.FirmwareVersion = nil },
				),
				Entry("firmware version empty",
					func(datum *dataTypesSettingsController.Device) { datum.FirmwareVersion = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/firmwareVersion"),
				),
				Entry("firmware version length; in range (upper)",
					func(datum *dataTypesSettingsController.Device) {
						datum.FirmwareVersion = pointer.FromString(test.RandomStringFromRange(100, 100))
					},
				),
				Entry("firmware version length; out of range (upper)",
					func(datum *dataTypesSettingsController.Device) {
						datum.FirmwareVersion = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/firmwareVersion"),
				),
				Entry("hardware version missing",
					func(datum *dataTypesSettingsController.Device) { datum.HardwareVersion = nil },
				),
				Entry("hardware version empty",
					func(datum *dataTypesSettingsController.Device) { datum.HardwareVersion = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/hardwareVersion"),
				),
				Entry("hardware version length; in range (upper)",
					func(datum *dataTypesSettingsController.Device) {
						datum.HardwareVersion = pointer.FromString(test.RandomStringFromRange(100, 100))
					},
				),
				Entry("hardware version length; out of range (upper)",
					func(datum *dataTypesSettingsController.Device) {
						datum.HardwareVersion = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/hardwareVersion"),
				),
				Entry("manufacturers missing",
					func(datum *dataTypesSettingsController.Device) { datum.Manufacturers = nil },
				),
				Entry("manufacturers empty",
					func(datum *dataTypesSettingsController.Device) {
						datum.Manufacturers = pointer.FromStringArray([]string{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/manufacturers"),
				),
				Entry("manufacturers length; in range (upper)",
					func(datum *dataTypesSettingsController.Device) {
						datum.Manufacturers = pointer.FromStringArray(dataTypesSettingsControllerTest.RandomManufacturersFromRange(10, 10))
					},
				),
				Entry("manufacturers length; out of range (upper)",
					func(datum *dataTypesSettingsController.Device) {
						datum.Manufacturers = pointer.FromStringArray(dataTypesSettingsControllerTest.RandomManufacturersFromRange(11, 11))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(11, 10), "/manufacturers"),
				),
				Entry("manufacturers manufacturer empty",
					func(datum *dataTypesSettingsController.Device) {
						datum.Manufacturers = pointer.FromStringArray(append([]string{dataTypesSettingsControllerTest.RandomManufacturerFromRange(1, 100), "", dataTypesSettingsControllerTest.RandomManufacturerFromRange(1, 100), ""}, dataTypesSettingsControllerTest.RandomManufacturersFromRange(0, 6)...))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/manufacturers/1"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/manufacturers/3"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/manufacturers/3"),
				),
				Entry("manufacturers manufacturer length; in range (upper)",
					func(datum *dataTypesSettingsController.Device) {
						datum.Manufacturers = pointer.FromStringArray(append([]string{dataTypesSettingsControllerTest.RandomManufacturerFromRange(100, 100), dataTypesSettingsControllerTest.RandomManufacturerFromRange(1, 100), dataTypesSettingsControllerTest.RandomManufacturerFromRange(100, 100)}, dataTypesSettingsControllerTest.RandomManufacturersFromRange(0, 7)...))
					},
				),
				Entry("manufacturers manufacturer length; out of range (upper)",
					func(datum *dataTypesSettingsController.Device) {
						datum.Manufacturers = pointer.FromStringArray(append([]string{dataTypesSettingsControllerTest.RandomManufacturerFromRange(101, 101), dataTypesSettingsControllerTest.RandomManufacturerFromRange(1, 100), dataTypesSettingsControllerTest.RandomManufacturerFromRange(101, 101)}, dataTypesSettingsControllerTest.RandomManufacturersFromRange(0, 7)...))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/manufacturers/0"),
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/manufacturers/2"),
				),
				Entry("manufacturers valid",
					func(datum *dataTypesSettingsController.Device) {
						datum.Manufacturers = pointer.FromStringArray(dataTypesSettingsControllerTest.RandomManufacturersFromRange(1, dataTypesSettingsController.ManufacturersLengthMaximum))
					},
				),
				Entry("model missing",
					func(datum *dataTypesSettingsController.Device) { datum.Model = nil },
				),
				Entry("model empty",
					func(datum *dataTypesSettingsController.Device) { datum.Model = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/model"),
				),
				Entry("model length; in range (upper)",
					func(datum *dataTypesSettingsController.Device) {
						datum.Model = pointer.FromString(test.RandomStringFromRange(100, 100))
					},
				),
				Entry("model length; out of range (upper)",
					func(datum *dataTypesSettingsController.Device) {
						datum.Model = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/model"),
				),
				Entry("name missing",
					func(datum *dataTypesSettingsController.Device) { datum.Name = nil },
				),
				Entry("name empty",
					func(datum *dataTypesSettingsController.Device) { datum.Name = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/name"),
				),
				Entry("name length; in range (upper)",
					func(datum *dataTypesSettingsController.Device) {
						datum.Name = pointer.FromString(test.RandomStringFromRange(100, 100))
					},
				),
				Entry("name length; out of range (upper)",
					func(datum *dataTypesSettingsController.Device) {
						datum.Name = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/name"),
				),
				Entry("serial number missing",
					func(datum *dataTypesSettingsController.Device) { datum.SerialNumber = nil },
				),
				Entry("serial number empty",
					func(datum *dataTypesSettingsController.Device) { datum.SerialNumber = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/serialNumber"),
				),
				Entry("serial number length; in range (upper)",
					func(datum *dataTypesSettingsController.Device) {
						datum.SerialNumber = pointer.FromString(test.RandomStringFromRange(100, 100))
					},
				),
				Entry("serial number length; out of range (upper)",
					func(datum *dataTypesSettingsController.Device) {
						datum.SerialNumber = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/serialNumber"),
				),
				Entry("software version missing",
					func(datum *dataTypesSettingsController.Device) { datum.SoftwareVersion = nil },
				),
				Entry("software version empty",
					func(datum *dataTypesSettingsController.Device) { datum.SoftwareVersion = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/softwareVersion"),
				),
				Entry("software version length; in range (upper)",
					func(datum *dataTypesSettingsController.Device) {
						datum.SoftwareVersion = pointer.FromString(test.RandomStringFromRange(100, 100))
					},
				),
				Entry("software version length; out of range (upper)",
					func(datum *dataTypesSettingsController.Device) {
						datum.SoftwareVersion = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/softwareVersion"),
				),
				Entry("one of required missing",
					func(datum *dataTypesSettingsController.Device) {
						datum.FirmwareVersion = nil
						datum.HardwareVersion = nil
						datum.Manufacturers = nil
						datum.Model = nil
						datum.Name = nil
						datum.SerialNumber = nil
						datum.SoftwareVersion = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValuesNotExistForAny("firmwareVersion", "hardwareVersion", "manufacturers", "model", "name", "serialNumber", "softwareVersion"), ""),
				),
				Entry("multiple errors",
					func(datum *dataTypesSettingsController.Device) {
						datum.FirmwareVersion = pointer.FromString("")
						datum.HardwareVersion = pointer.FromString("")
						datum.Manufacturers = pointer.FromStringArray([]string{})
						datum.Model = pointer.FromString("")
						datum.Name = pointer.FromString("")
						datum.SerialNumber = pointer.FromString("")
						datum.SoftwareVersion = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/firmwareVersion"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/hardwareVersion"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/manufacturers"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/model"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/name"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/serialNumber"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/softwareVersion"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum with origin external",
				func(mutator func(datum *dataTypesSettingsController.Device), expectator func(datum *dataTypesSettingsController.Device, expectedDatum *dataTypesSettingsController.Device)) {
					datum := dataTypesSettingsControllerTest.RandomDevice()
					mutator(datum)
					expectedDatum := dataTypesSettingsControllerTest.CloneDevice(datum)
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
				Entry("modifies the datum",
					func(datum *dataTypesSettingsController.Device) {},
					func(datum *dataTypesSettingsController.Device, expectedDatum *dataTypesSettingsController.Device) {
						sort.Strings(*expectedDatum.Manufacturers)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(mutator func(datum *dataTypesSettingsController.Device), expectator func(datum *dataTypesSettingsController.Device, expectedDatum *dataTypesSettingsController.Device)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := dataTypesSettingsControllerTest.RandomDevice()
						mutator(datum)
						expectedDatum := dataTypesSettingsControllerTest.CloneDevice(datum)
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
				Entry("does not modify the datum",
					func(datum *dataTypesSettingsController.Device) {},
					nil,
				),
			)
		})
	})
})
