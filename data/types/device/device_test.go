package device_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataTypes "github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/device"
	dataTypesDeviceTest "github.com/tidepool-org/platform/data/types/device/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const ExpectedTimeFormat = time.RFC3339Nano

var _ = Describe("Device", func() {
	It("Type is expected", func() {
		Expect(device.Type).To(Equal("deviceEvent"))
	})

	Context("New", func() {
		It("creates a new datum with all values initialized", func() {
			subType := dataTypesTest.NewType()
			datum := device.New(subType)
			Expect(datum.Type).To(Equal("deviceEvent"))
			Expect(datum.SubType).To(Equal(subType))
		})
	})

	Context("with new datum", func() {
		var subType string
		var datum device.Device

		BeforeEach(func() {
			subType = dataTypesTest.NewType()
			datum = device.New(subType)
		})

		Context("Meta", func() {
			It("returns the meta with delivery type", func() {
				Expect(datum.Meta()).To(Equal(&device.Meta{Type: "deviceEvent", SubType: subType}))
			})
		})
	})

	Context("Device", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *device.Device), expectedErrors ...error) {
					datum := dataTypesDeviceTest.RandomDevice()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *device.Device) {},
				),
				Entry("type missing",
					func(datum *device.Device) { datum.Type = "" },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/type"),
				),
				Entry("type invalid",
					func(datum *device.Device) { datum.Type = "invalid" },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", "deviceEvent"), "/type"),
				),
				Entry("type deviceEvent",
					func(datum *device.Device) { datum.Type = "deviceEvent" },
				),
				Entry("sub type missing",
					func(datum *device.Device) { datum.SubType = "" },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/subType"),
				),
				Entry("sub type valid",
					func(datum *device.Device) { datum.SubType = dataTypesTest.NewType() },
				),
				Entry("multiple errors",
					func(datum *device.Device) {
						datum.Type = "invalid"
						datum.SubType = ""
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", "deviceEvent"), "/type"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/subType"),
				),
			)
		})

		Context("IdentityFields", func() {
			var datum *device.Device
			const currentVersion = "1.1.0"

			BeforeEach(func() {
				datum = dataTypesDeviceTest.RandomDevice()
			})

			It("returns error if user id is missing", func() {
				datum.UserID = nil
				identityFields, err := datum.IdentityFields(currentVersion)
				Expect(err).To(MatchError("user id is missing"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if user id is empty", func() {
				datum.UserID = pointer.FromString("")
				identityFields, err := datum.IdentityFields(currentVersion)
				Expect(err).To(MatchError("user id is empty"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if sub type is empty", func() {
				datum.SubType = ""
				identityFields, err := datum.IdentityFields(currentVersion)
				Expect(err).To(MatchError("sub type is empty"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns the expected identity fields", func() {
				identityFields, err := datum.IdentityFields(currentVersion)
				Expect(err).ToNot(HaveOccurred())
				Expect(identityFields).To(Equal([]string{*datum.UserID, *datum.DeviceID, (*datum.Time).Format(ExpectedTimeFormat), datum.Type, datum.SubType}))
			})
		})

		Context("Legacy IdentityFields", func() {
			var datum *device.Device

			BeforeEach(func() {
				datum = dataTypesDeviceTest.RandomDevice()
			})

			It("returns error if sub type is empty", func() {
				datum.SubType = ""
				identityFields, err := datum.IdentityFields(dataTypes.LegacyIdentityFieldsVersion)
				Expect(err).To(MatchError("sub type is empty"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns the expected legacy identity fields", func() {
				datum.DeviceID = pointer.FromString("some-device")
				t, err := time.Parse(dataTypes.TimeFormat, "2023-05-13T15:51:58Z")
				Expect(err).ToNot(HaveOccurred())
				datum.Time = pointer.FromTime(t)
				datum.SubType = "some-sub-type"
				legacyIdentityFields, err := datum.IdentityFields(dataTypes.LegacyIdentityFieldsVersion)
				Expect(err).ToNot(HaveOccurred())
				Expect(legacyIdentityFields).To(Equal([]string{"deviceEvent", "some-sub-type", "2023-05-13T15:51:58.000Z", "some-device"}))
			})
		})
	})
})
