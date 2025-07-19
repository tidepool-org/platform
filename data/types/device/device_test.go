package device_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
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
			var datumDevice *device.Device
			var datum data.Datum
			var version string

			BeforeEach(func() {
				datumDevice = dataTypesDeviceTest.RandomDevice()
				datum = datumDevice
			})

			identityFieldsAssertions := func() {
				It("returns error if user id is missing", func() {
					datumDevice.UserID = nil
					identityFields, err := datum.IdentityFields(version)
					Expect(err).To(MatchError("user id is missing"))
					Expect(identityFields).To(BeEmpty())
				})

				It("returns error if user id is empty", func() {
					datumDevice.UserID = pointer.FromString("")
					identityFields, err := datum.IdentityFields(version)
					Expect(err).To(MatchError("user id is empty"))
					Expect(identityFields).To(BeEmpty())
				})

				It("returns error if sub type is empty", func() {
					datumDevice.SubType = ""
					identityFields, err := datum.IdentityFields(version)
					Expect(err).To(MatchError("sub type is empty"))
					Expect(identityFields).To(BeEmpty())
				})
			}

			When("version is IdentityFieldsVersionDefault", func() {
				BeforeEach(func() {
					version = dataTypes.IdentityFieldsVersionDeviceID
				})

				identityFieldsAssertions()

				It("returns the expected identity fields", func() {
					identityFields, err := datum.IdentityFields(version)
					Expect(err).ToNot(HaveOccurred())
					Expect(identityFields).To(Equal([]string{*datumDevice.UserID, *datumDevice.DeviceID, (*datumDevice.Time).Format(ExpectedTimeFormat), datumDevice.Type, datumDevice.SubType}))
				})
			})

			When("version is IdentityFieldsVersionDataSetID", func() {
				BeforeEach(func() {
					version = dataTypes.IdentityFieldsVersionDataSetID
				})

				identityFieldsAssertions()

				It("returns the expected identity fields", func() {
					identityFields, err := datum.IdentityFields(version)
					Expect(err).ToNot(HaveOccurred())
					Expect(identityFields).To(Equal([]string{*datumDevice.UserID, *datumDevice.UploadID, (*datumDevice.Time).Format(ExpectedTimeFormat), datumDevice.Type, datumDevice.SubType}))
				})
			})

			When("version is invalid", func() {
				It("returns an error", func() {
					identityFields, err := datum.IdentityFields("invalid")
					Expect(err).To(MatchError("version is invalid"))
					Expect(identityFields).To(BeEmpty())
				})
			})
		})
	})
})
