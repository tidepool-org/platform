package device_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/factory"
	"github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/service"
)

func NewMeta(subType string) interface{} {
	return &device.Meta{
		Type:    "deviceEvent",
		SubType: subType,
	}
}

func NewTestDevice(sourceTime interface{}, sourceSubType interface{}) *device.Device {
	testDevice := &device.Device{}
	testDevice.Init()
	testDevice.DeviceID = pointer.String(id.New())
	if value, ok := sourceTime.(string); ok {
		testDevice.Time = pointer.String(value)
	}
	if value, ok := sourceSubType.(string); ok {
		testDevice.SubType = value
	}
	return testDevice
}

var _ = Describe("Device", func() {
	Context("Type", func() {
		It("returns the expected type", func() {
			Expect(device.Type()).To(Equal("deviceEvent"))
		})
	})

	Context("with new device", func() {
		var testDevice *device.Device

		BeforeEach(func() {
			testDevice = &device.Device{}
		})

		Context("Init", func() {
			It("initializes the device", func() {
				testDevice.Init()
				Expect(testDevice.ID).ToNot(BeEmpty())
				Expect(testDevice.Type).To(Equal("deviceEvent"))
				Expect(testDevice.SubType).To(BeEmpty())
			})
		})

		Context("with initialized", func() {
			BeforeEach(func() {
				testDevice.Init()
			})

			Context("Meta", func() {
				It("returns the meta with no sub type", func() {
					testDevice.Init()
					Expect(testDevice.Meta()).To(Equal(NewMeta("")))
				})

				It("returns the meta with sub type", func() {
					testDevice.Init()
					testDevice.SubType = "alarm"
					Expect(testDevice.Meta()).To(Equal(NewMeta("alarm")))
				})
			})

			DescribeTable("Parse",
				func(sourceObject *map[string]interface{}, expectedDevice *device.Device, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(null.NewLogger())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testFactory, err := factory.NewStandard()
					Expect(err).ToNot(HaveOccurred())
					Expect(testFactory).ToNot(BeNil())
					testParser, err := parser.NewStandardObject(testContext, testFactory, sourceObject, parser.AppendErrorNotParsed)
					Expect(err).ToNot(HaveOccurred())
					Expect(testParser).ToNot(BeNil())
					Expect(testDevice.Parse(testParser)).To(Succeed())
					Expect(testDevice.Time).To(Equal(expectedDevice.Time))
					Expect(testDevice.SubType).To(Equal(expectedDevice.SubType))
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
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta("")),
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
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta("")),
					}),
			)

			DescribeTable("Validate",
				func(sourceDevice *device.Device, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(null.NewLogger())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testValidator, err := validator.NewStandard(testContext)
					Expect(err).ToNot(HaveOccurred())
					Expect(testValidator).ToNot(BeNil())
					Expect(sourceDevice.Validate(testValidator)).To(Succeed())
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("all valid",
					NewTestDevice("2016-09-06T13:45:58-07:00", "alarm"),
					[]*service.Error{}),
				Entry("missing time",
					NewTestDevice(nil, "alarm"),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotExists(), "/time", NewMeta("alarm")),
					}),
				Entry("missing sub type",
					NewTestDevice("2016-09-06T13:45:58-07:00", nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueEmpty(), "/subType", NewMeta("")),
					}),
				Entry("specified sub type",
					NewTestDevice("2016-09-06T13:45:58-07:00", "specified"),
					[]*service.Error{}),
				Entry("multiple",
					NewTestDevice(nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotExists(), "/time", NewMeta("")),
						testData.ComposeError(service.ErrorValueEmpty(), "/subType", NewMeta("")),
					}),
			)

			Context("Normalize", func() {
				It("succeeds", func() {
					testContext, err := context.NewStandard(null.NewLogger())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testNormalizer, err := normalizer.NewStandard(testContext)
					Expect(err).ToNot(HaveOccurred())
					Expect(testNormalizer).ToNot(BeNil())
					Expect(testDevice.Normalize(testNormalizer)).To(Succeed())
				})
			})

			Context("IdentityFields", func() {
				var userID string
				var deviceID string

				BeforeEach(func() {
					userID = id.New()
					deviceID = id.New()
					testDevice.UserID = userID
					testDevice.DeviceID = &deviceID
					testDevice.Time = pointer.String("2016-09-06T13:45:58-07:00")
					testDevice.SubType = "alarm"
				})

				It("returns error if user id is empty", func() {
					testDevice.UserID = ""
					identityFields, err := testDevice.IdentityFields()
					Expect(err).To(MatchError("user id is empty"))
					Expect(identityFields).To(BeEmpty())
				})

				It("returns error if sub type is empty", func() {
					testDevice.SubType = ""
					identityFields, err := testDevice.IdentityFields()
					Expect(err).To(MatchError("sub type is empty"))
					Expect(identityFields).To(BeEmpty())
				})

				It("returns the expected identity fields", func() {
					identityFields, err := testDevice.IdentityFields()
					Expect(err).ToNot(HaveOccurred())
					Expect(identityFields).To(Equal([]string{userID, deviceID, "2016-09-06T13:45:58-07:00", "deviceEvent", "alarm"}))
				})
			})
		})
	})
})
