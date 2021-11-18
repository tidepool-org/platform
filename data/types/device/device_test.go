package device_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/types/device"
	dataTypesDeviceTest "github.com/tidepool-org/platform/data/types/device/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

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

			BeforeEach(func() {
				datum = dataTypesDeviceTest.RandomDevice()
			})

			It("returns error if user id is missing", func() {
				datum.UserID = nil
				identityFields, err := datum.IdentityFields()
				Expect(err).To(MatchError("user id is missing"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if user id is empty", func() {
				datum.UserID = pointer.FromString("")
				identityFields, err := datum.IdentityFields()
				Expect(err).To(MatchError("user id is empty"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if sub type is empty", func() {
				datum.SubType = ""
				identityFields, err := datum.IdentityFields()
				Expect(err).To(MatchError("sub type is empty"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns the expected identity fields", func() {
				identityFields, err := datum.IdentityFields()
				Expect(err).ToNot(HaveOccurred())
				Expect(identityFields).To(Equal([]string{*datum.UserID, *datum.DeviceID, *datum.Time, datum.Type, datum.SubType}))
			})
		})
	})
})
