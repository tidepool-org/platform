package controller_test

import (
	"sort"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTypes "github.com/tidepool-org/platform/data/types"
	dataTypesSettingsController "github.com/tidepool-org/platform/data/types/settings/controller"
	dataTypesSettingsControllerTest "github.com/tidepool-org/platform/data/types/settings/controller/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() interface{} {
	return &dataTypes.Meta{
		Type: "controllerSettings",
	}
}

var _ = Describe("Controller", func() {
	It("Type is expected", func() {
		Expect(dataTypesSettingsController.Type).To(Equal("controllerSettings"))
	})

	Context("Controller", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesSettingsController.Controller)) {
				datum := dataTypesSettingsControllerTest.RandomController()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesSettingsControllerTest.NewObjectFromController(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesSettingsControllerTest.NewObjectFromController(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesSettingsController.Controller) {},
			),
			Entry("empty",
				func(datum *dataTypesSettingsController.Controller) {
					*datum = *dataTypesSettingsController.New()
				},
			),
			Entry("all",
				func(datum *dataTypesSettingsController.Controller) {
					datum.Device = dataTypesSettingsControllerTest.RandomDevice()
					datum.Notifications = dataTypesSettingsControllerTest.RandomNotifications()
				},
			),
		)

		Context("New", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesSettingsController.New()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Type).To(Equal("controllerSettings"))
				Expect(datum.Device).To(BeNil())
				Expect(datum.Notifications).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesSettingsController.Controller), expectedErrors ...error) {
					expectedDatum := dataTypesSettingsControllerTest.RandomControllerForParser()
					object := dataTypesSettingsControllerTest.NewObjectFromController(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesSettingsController.New()
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesSettingsController.Controller) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesSettingsController.Controller) {
						object["device"] = true
						object["notifications"] = true
						expectedDatum.Device = nil
						expectedDatum.Notifications = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotObject(true), "/device", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotObject(true), "/notifications", NewMeta()),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesSettingsController.Controller), expectedErrors ...error) {
					datum := dataTypesSettingsControllerTest.RandomController()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesSettingsController.Controller) {},
				),
				Entry("type missing",
					func(datum *dataTypesSettingsController.Controller) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &dataTypes.Meta{Type: ""}),
				),
				Entry("type invalid",
					func(datum *dataTypesSettingsController.Controller) {
						datum.Type = "invalidType"
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "controllerSettings"), "/type", &dataTypes.Meta{Type: "invalidType"}),
				),
				Entry("type controllerSettings",
					func(datum *dataTypesSettingsController.Controller) {
						datum.Type = "controllerSettings"
					},
				),
				Entry("device missing",
					func(datum *dataTypesSettingsController.Controller) { datum.Device = nil },
				),
				Entry("device invalid",
					func(datum *dataTypesSettingsController.Controller) {
						datum.Device.FirmwareVersion = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/device/firmwareVersion", NewMeta()),
				),
				Entry("device valid",
					func(datum *dataTypesSettingsController.Controller) {
						datum.Device = dataTypesSettingsControllerTest.RandomDevice()
					},
				),
				Entry("notifications missing",
					func(datum *dataTypesSettingsController.Controller) { datum.Notifications = nil },
				),
				Entry("notifications invalid",
					func(datum *dataTypesSettingsController.Controller) {
						datum.Notifications.Authorization = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"authorized", "denied", "ephemeral", "notDetermined", "provisional"}), "/notifications/authorization", NewMeta()),
				),
				Entry("notifications valid",
					func(datum *dataTypesSettingsController.Controller) {
						datum.Notifications = dataTypesSettingsControllerTest.RandomNotifications()
					},
				),
				Entry("one of required missing",
					func(datum *dataTypesSettingsController.Controller) {
						datum.Device = nil
						datum.Notifications = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValuesNotExistForAny("device", "notifications"), "", NewMeta()),
				),
				Entry("multiple errors",
					func(datum *dataTypesSettingsController.Controller) {
						datum.Type = "invalidType"
						datum.Device.FirmwareVersion = pointer.FromString("")
						datum.Notifications.Authorization = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "controllerSettings"), "/type", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/device/firmwareVersion", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"authorized", "denied", "ephemeral", "notDetermined", "provisional"}), "/notifications/authorization", &dataTypes.Meta{Type: "invalidType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum with origin external",
				func(mutator func(datum *dataTypesSettingsController.Controller), expectator func(datum *dataTypesSettingsController.Controller, expectedDatum *dataTypesSettingsController.Controller)) {
					datum := dataTypesSettingsControllerTest.RandomController()
					mutator(datum)
					expectedDatum := dataTypesSettingsControllerTest.CloneController(datum)
					normalizer := dataNormalizer.New(logTest.NewLogger())
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
					func(datum *dataTypesSettingsController.Controller) {},
					func(datum *dataTypesSettingsController.Controller, expectedDatum *dataTypesSettingsController.Controller) {
						sort.Strings(*expectedDatum.Device.Manufacturers)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(mutator func(datum *dataTypesSettingsController.Controller), expectator func(datum *dataTypesSettingsController.Controller, expectedDatum *dataTypesSettingsController.Controller)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := dataTypesSettingsControllerTest.RandomController()
						mutator(datum)
						expectedDatum := dataTypesSettingsControllerTest.CloneController(datum)
						normalizer := dataNormalizer.New(logTest.NewLogger())
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
					func(datum *dataTypesSettingsController.Controller) {},
					nil,
				),
			)
		})
	})
})
