package controller_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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

var _ = Describe("Notifications", func() {
	It("AlertStyleAlert is expected", func() {
		Expect(dataTypesSettingsController.AlertStyleAlert).To(Equal("alert"))
	})

	It("AlertStyleBanner is expected", func() {
		Expect(dataTypesSettingsController.AlertStyleBanner).To(Equal("banner"))
	})

	It("AlertStyleNone is expected", func() {
		Expect(dataTypesSettingsController.AlertStyleNone).To(Equal("none"))
	})

	It("AuthorizationAuthorized is expected", func() {
		Expect(dataTypesSettingsController.AuthorizationAuthorized).To(Equal("authorized"))
	})

	It("AuthorizationDenied is expected", func() {
		Expect(dataTypesSettingsController.AuthorizationDenied).To(Equal("denied"))
	})

	It("AuthorizationEphemeral is expected", func() {
		Expect(dataTypesSettingsController.AuthorizationEphemeral).To(Equal("ephemeral"))
	})

	It("AuthorizationNotDetermined is expected", func() {
		Expect(dataTypesSettingsController.AuthorizationNotDetermined).To(Equal("notDetermined"))
	})

	It("AuthorizationProvisional is expected", func() {
		Expect(dataTypesSettingsController.AuthorizationProvisional).To(Equal("provisional"))
	})

	It("AlertStyles returns expected", func() {
		Expect(dataTypesSettingsController.AlertStyles()).To(Equal([]string{"alert", "banner", "none"}))
	})

	It("Authorizations returns expected", func() {
		Expect(dataTypesSettingsController.Authorizations()).To(Equal([]string{"authorized", "denied", "ephemeral", "notDetermined", "provisional"}))
	})

	Context("Notifications", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesSettingsController.Notifications)) {
				datum := dataTypesSettingsControllerTest.RandomNotifications()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesSettingsControllerTest.NewObjectFromNotifications(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesSettingsControllerTest.NewObjectFromNotifications(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesSettingsController.Notifications) {},
			),
			Entry("empty",
				func(datum *dataTypesSettingsController.Notifications) {
					*datum = *dataTypesSettingsController.NewNotifications()
				},
			),
			Entry("all",
				func(datum *dataTypesSettingsController.Notifications) {
					datum.Authorization = pointer.FromString(dataTypesSettingsControllerTest.RandomAuthorization())
					datum.Alert = pointer.FromBool(test.RandomBool())
					datum.CriticalAlert = pointer.FromBool(test.RandomBool())
					datum.Badge = pointer.FromBool(test.RandomBool())
					datum.Sound = pointer.FromBool(test.RandomBool())
					datum.Announcement = pointer.FromBool(test.RandomBool())
					datum.NotificationCenter = pointer.FromBool(test.RandomBool())
					datum.LockScreen = pointer.FromBool(test.RandomBool())
					datum.AlertStyle = pointer.FromString(dataTypesSettingsControllerTest.RandomAlertStyle())
				},
			),
		)

		Context("ParseNotifications", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataTypesSettingsController.ParseNotifications(structureParser.NewObject(logTest.NewLogger(), nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataTypesSettingsControllerTest.RandomNotifications()
				object := dataTypesSettingsControllerTest.NewObjectFromNotifications(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(logTest.NewLogger(), &object)
				Expect(dataTypesSettingsController.ParseNotifications(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewNotifications", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesSettingsController.NewNotifications()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Authorization).To(BeNil())
				Expect(datum.Alert).To(BeNil())
				Expect(datum.CriticalAlert).To(BeNil())
				Expect(datum.Badge).To(BeNil())
				Expect(datum.Sound).To(BeNil())
				Expect(datum.Announcement).To(BeNil())
				Expect(datum.NotificationCenter).To(BeNil())
				Expect(datum.LockScreen).To(BeNil())
				Expect(datum.AlertStyle).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesSettingsController.Notifications), expectedErrors ...error) {
					expectedDatum := dataTypesSettingsControllerTest.RandomNotifications()
					object := dataTypesSettingsControllerTest.NewObjectFromNotifications(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesSettingsController.NewNotifications()
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesSettingsController.Notifications) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesSettingsController.Notifications) {
						object["authorization"] = 0
						object["alert"] = 0
						object["criticalAlert"] = 0
						object["badge"] = 0
						object["sound"] = 0
						object["announcement"] = 0
						object["notificationCenter"] = 0
						object["lockScreen"] = 0
						object["alertStyle"] = 0
						expectedDatum.Authorization = nil
						expectedDatum.Alert = nil
						expectedDatum.CriticalAlert = nil
						expectedDatum.Badge = nil
						expectedDatum.Sound = nil
						expectedDatum.Announcement = nil
						expectedDatum.NotificationCenter = nil
						expectedDatum.LockScreen = nil
						expectedDatum.AlertStyle = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(0), "/authorization"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotBool(0), "/alert"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotBool(0), "/criticalAlert"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotBool(0), "/badge"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotBool(0), "/sound"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotBool(0), "/announcement"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotBool(0), "/notificationCenter"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotBool(0), "/lockScreen"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(0), "/alertStyle"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesSettingsController.Notifications), expectedErrors ...error) {
					datum := dataTypesSettingsControllerTest.RandomNotifications()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesSettingsController.Notifications) {},
				),
				Entry("authorization missing",
					func(datum *dataTypesSettingsController.Notifications) { datum.Authorization = nil },
				),
				Entry("authorization invalid",
					func(datum *dataTypesSettingsController.Notifications) {
						datum.Authorization = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"authorized", "denied", "ephemeral", "notDetermined", "provisional"}), "/authorization"),
				),
				Entry("authorization authorized",
					func(datum *dataTypesSettingsController.Notifications) {
						datum.Authorization = pointer.FromString(dataTypesSettingsController.AuthorizationAuthorized)
					},
				),
				Entry("authorization denied",
					func(datum *dataTypesSettingsController.Notifications) {
						datum.Authorization = pointer.FromString(dataTypesSettingsController.AuthorizationDenied)
					},
				),
				Entry("authorization ephemeral",
					func(datum *dataTypesSettingsController.Notifications) {
						datum.Authorization = pointer.FromString(dataTypesSettingsController.AuthorizationEphemeral)
					},
				),
				Entry("authorization notDetermined",
					func(datum *dataTypesSettingsController.Notifications) {
						datum.Authorization = pointer.FromString(dataTypesSettingsController.AuthorizationNotDetermined)
					},
				),
				Entry("authorization provisional",
					func(datum *dataTypesSettingsController.Notifications) {
						datum.Authorization = pointer.FromString(dataTypesSettingsController.AuthorizationProvisional)
					},
				),
				Entry("alert missing",
					func(datum *dataTypesSettingsController.Notifications) { datum.Alert = nil },
				),
				Entry("alert false",
					func(datum *dataTypesSettingsController.Notifications) { datum.Alert = pointer.FromBool(false) },
				),
				Entry("alert true",
					func(datum *dataTypesSettingsController.Notifications) { datum.Alert = pointer.FromBool(true) },
				),
				Entry("critical alert missing",
					func(datum *dataTypesSettingsController.Notifications) { datum.CriticalAlert = nil },
				),
				Entry("critical alert false",
					func(datum *dataTypesSettingsController.Notifications) { datum.CriticalAlert = pointer.FromBool(false) },
				),
				Entry("critical alert true",
					func(datum *dataTypesSettingsController.Notifications) { datum.CriticalAlert = pointer.FromBool(true) },
				),
				Entry("badge missing",
					func(datum *dataTypesSettingsController.Notifications) { datum.Badge = nil },
				),
				Entry("badge false",
					func(datum *dataTypesSettingsController.Notifications) { datum.Badge = pointer.FromBool(false) },
				),
				Entry("badge true",
					func(datum *dataTypesSettingsController.Notifications) { datum.Badge = pointer.FromBool(true) },
				),
				Entry("sound missing",
					func(datum *dataTypesSettingsController.Notifications) { datum.Sound = nil },
				),
				Entry("sound false",
					func(datum *dataTypesSettingsController.Notifications) { datum.Sound = pointer.FromBool(false) },
				),
				Entry("sound true",
					func(datum *dataTypesSettingsController.Notifications) { datum.Sound = pointer.FromBool(true) },
				),
				Entry("announcement missing",
					func(datum *dataTypesSettingsController.Notifications) { datum.Announcement = nil },
				),
				Entry("announcement false",
					func(datum *dataTypesSettingsController.Notifications) { datum.Announcement = pointer.FromBool(false) },
				),
				Entry("announcement true",
					func(datum *dataTypesSettingsController.Notifications) { datum.Announcement = pointer.FromBool(true) },
				),
				Entry("notification center missing",
					func(datum *dataTypesSettingsController.Notifications) { datum.NotificationCenter = nil },
				),
				Entry("notification center false",
					func(datum *dataTypesSettingsController.Notifications) {
						datum.NotificationCenter = pointer.FromBool(false)
					},
				),
				Entry("notification center true",
					func(datum *dataTypesSettingsController.Notifications) {
						datum.NotificationCenter = pointer.FromBool(true)
					},
				),
				Entry("lock screen missing",
					func(datum *dataTypesSettingsController.Notifications) { datum.LockScreen = nil },
				),
				Entry("lock screen false",
					func(datum *dataTypesSettingsController.Notifications) { datum.LockScreen = pointer.FromBool(false) },
				),
				Entry("lock screen true",
					func(datum *dataTypesSettingsController.Notifications) { datum.LockScreen = pointer.FromBool(true) },
				),
				Entry("alert style missing",
					func(datum *dataTypesSettingsController.Notifications) { datum.AlertStyle = nil },
				),
				Entry("alert style invalid",
					func(datum *dataTypesSettingsController.Notifications) {
						datum.AlertStyle = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"alert", "banner", "none"}), "/alertStyle"),
				),
				Entry("alert style alert",
					func(datum *dataTypesSettingsController.Notifications) {
						datum.AlertStyle = pointer.FromString(dataTypesSettingsController.AlertStyleAlert)
					},
				),
				Entry("alert style banner",
					func(datum *dataTypesSettingsController.Notifications) {
						datum.AlertStyle = pointer.FromString(dataTypesSettingsController.AlertStyleBanner)
					},
				),
				Entry("alert style none",
					func(datum *dataTypesSettingsController.Notifications) {
						datum.AlertStyle = pointer.FromString(dataTypesSettingsController.AlertStyleNone)
					},
				),
				Entry("one of required missing",
					func(datum *dataTypesSettingsController.Notifications) {
						datum.Authorization = nil
						datum.Alert = nil
						datum.CriticalAlert = nil
						datum.Badge = nil
						datum.Sound = nil
						datum.Announcement = nil
						datum.NotificationCenter = nil
						datum.LockScreen = nil
						datum.AlertStyle = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValuesNotExistForAny("authorization", "alert", "criticalAlert", "badge", "sound", "announcement", "notificationCenter", "lockScreen", "alertStyle"), ""),
				),
				Entry("multiple errors",
					func(datum *dataTypesSettingsController.Notifications) {
						datum.Authorization = pointer.FromString("invalid")
						datum.AlertStyle = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"authorized", "denied", "ephemeral", "notDetermined", "provisional"}), "/authorization"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"alert", "banner", "none"}), "/alertStyle"),
				),
			)
		})
	})
})
