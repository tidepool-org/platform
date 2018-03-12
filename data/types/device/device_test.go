package device_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/parser"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/device"
	testDataTypesDevice "github.com/tidepool-org/platform/data/types/device/test"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewTestDevice(sourceTime interface{}, sourceSubType interface{}) *device.Device {
	datum := &device.Device{}
	datum.Init()
	datum.DeviceID = pointer.String(id.New())
	if val, ok := sourceTime.(string); ok {
		datum.Time = &val
	}
	if val, ok := sourceSubType.(string); ok {
		datum.SubType = val
	}
	return datum
}

var _ = Describe("Device", func() {
	It("Type is expected", func() {
		Expect(device.Type).To(Equal("deviceEvent"))
	})

	Context("with new datum", func() {
		var datum *device.Device

		BeforeEach(func() {
			datum = testDataTypesDevice.NewDevice()
		})

		Context("Init", func() {
			It("initializes the datum", func() {
				datum.Init()
				Expect(datum.Type).To(Equal("deviceEvent"))
				Expect(datum.SubType).To(BeEmpty())
			})
		})

		Context("with initialized", func() {
			BeforeEach(func() {
				datum.Init()
			})

			Context("Meta", func() {
				It("returns the meta with no sub type", func() {
					Expect(datum.Meta()).To(Equal(&device.Meta{Type: "deviceEvent"}))
				})

				It("returns the meta with sub type", func() {
					datum.SubType = testDataTypes.NewType()
					Expect(datum.Meta()).To(Equal(&device.Meta{Type: "deviceEvent", SubType: datum.SubType}))
				})
			})
		})
	})

	Context("Device", func() {
		Context("Parse", func() {
			var datum *device.Device

			BeforeEach(func() {
				datum = &device.Device{}
				datum.Init()
			})

			DescribeTable("parses the datum",
				func(sourceObject *map[string]interface{}, expectedDatum *device.Device, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(null.NewLogger())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testParser, err := parser.NewStandardObject(testContext, sourceObject, parser.AppendErrorNotParsed)
					Expect(err).ToNot(HaveOccurred())
					Expect(testParser).ToNot(BeNil())
					Expect(datum.Parse(testParser)).To(Succeed())
					Expect(datum.Time).To(Equal(expectedDatum.Time))
					Expect(datum.SubType).To(Equal(expectedDatum.SubType))
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("parses object that is nil",
					nil,
					NewTestDevice(nil, nil),
					[]*service.Error{}),
				Entry("parses object that is empty",
					&map[string]interface{}{},
					NewTestDevice(nil, nil),
					[]*service.Error{}),
				Entry("parses object that has valid time",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00"},
					NewTestDevice("2016-09-06T13:45:58-07:00", nil),
					[]*service.Error{}),
				Entry("parses object that has invalid time",
					&map[string]interface{}{"time": 0},
					NewTestDevice(nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", &device.Meta{Type: "deviceEvent"}),
					}),
				Entry("does not parse sub type",
					&map[string]interface{}{"subType": "alarm"},
					NewTestDevice(nil, nil),
					[]*service.Error{}),
				Entry("parses object that has multiple valid fields",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00", "subType": "alarm"},
					NewTestDevice("2016-09-06T13:45:58-07:00", nil),
					[]*service.Error{}),
				Entry("parses object that has multiple invalid fields",
					&map[string]interface{}{"time": 0, "subType": 0},
					NewTestDevice(nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", &device.Meta{Type: "deviceEvent"}),
					}),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *device.Device), expectedErrors ...error) {
					datum := testDataTypesDevice.NewDevice()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *device.Device) {},
				),
				Entry("type missing",
					func(datum *device.Device) { datum.Type = "" },
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/type"),
				),
				Entry("type invalid",
					func(datum *device.Device) { datum.Type = "invalid" },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", "deviceEvent"), "/type"),
				),
				Entry("type deviceEvent",
					func(datum *device.Device) { datum.Type = "deviceEvent" },
				),
				Entry("sub type missing",
					func(datum *device.Device) { datum.SubType = "" },
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/subType"),
				),
				Entry("sub type valid",
					func(datum *device.Device) { datum.SubType = testDataTypes.NewType() },
				),
				Entry("multiple errors",
					func(datum *device.Device) {
						datum.Type = "invalid"
						datum.SubType = ""
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", "deviceEvent"), "/type"),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/subType"),
				),
			)
		})

		Context("IdentityFields", func() {
			var datum *device.Device

			BeforeEach(func() {
				datum = testDataTypesDevice.NewDevice()
			})

			It("returns error if user id is missing", func() {
				datum.UserID = nil
				identityFields, err := datum.IdentityFields()
				Expect(err).To(MatchError("user id is missing"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if user id is empty", func() {
				datum.UserID = pointer.String("")
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
