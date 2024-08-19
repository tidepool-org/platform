package alarm_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/data/types/device/alarm"
	dataTypesDeviceStatus "github.com/tidepool-org/platform/data/types/device/status"
	dataTypesDeviceStatusTest "github.com/tidepool-org/platform/data/types/device/status/test"
	dataTypesDeviceTest "github.com/tidepool-org/platform/data/types/device/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() interface{} {
	return &device.Meta{
		Type:    "deviceEvent",
		SubType: "alarm",
	}
}

func NewAlarm() *alarm.Alarm {
	datum := alarm.New()
	datum.Device = *dataTypesDeviceTest.RandomDevice()
	datum.SubType = "alarm"
	datum.AlarmType = pointer.FromString(test.RandomStringFromArray(alarm.AlarmTypes()))
	return datum
}

func NewAlarmWithStatus() *alarm.Alarm {
	var status data.Datum
	status = dataTypesDeviceStatusTest.NewStatus()
	datum := NewAlarm()
	datum.Status = &status
	return datum
}

func NewAlarmWithStatusID() *alarm.Alarm {
	datum := NewAlarm()
	datum.StatusID = pointer.FromString(dataTest.RandomID())
	return datum
}

func CloneAlarm(datum *alarm.Alarm) *alarm.Alarm {
	if datum == nil {
		return nil
	}
	clone := alarm.New()
	clone.Device = *dataTypesDeviceTest.CloneDevice(&datum.Device)
	clone.AlarmType = pointer.CloneString(datum.AlarmType)
	if datum.Status != nil {
		switch status := (*datum.Status).(type) {
		case *dataTypesDeviceStatus.Status:
			clone.Status = data.DatumAsPointer(dataTypesDeviceStatusTest.CloneStatus(status))
		}
	}
	clone.StatusID = pointer.CloneString(datum.StatusID)
	return clone
}

var _ = Describe("Change", func() {
	It("SubType is expected", func() {
		Expect(alarm.SubType).To(Equal("alarm"))
	})

	It("AlarmTypeAutoOff is expected", func() {
		Expect(alarm.AlarmTypeAutoOff).To(Equal("auto_off"))
	})

	It("AlarmTypeLowInsulin is expected", func() {
		Expect(alarm.AlarmTypeLowInsulin).To(Equal("low_insulin"))
	})

	It("AlarmTypeLowPower is expected", func() {
		Expect(alarm.AlarmTypeLowPower).To(Equal("low_power"))
	})

	It("AlarmTypeNoDelivery is expected", func() {
		Expect(alarm.AlarmTypeNoDelivery).To(Equal("no_delivery"))
	})

	It("AlarmTypeNoInsulin is expected", func() {
		Expect(alarm.AlarmTypeNoInsulin).To(Equal("no_insulin"))
	})

	It("AlarmTypeNoPower is expected", func() {
		Expect(alarm.AlarmTypeNoPower).To(Equal("no_power"))
	})

	It("AlarmTypeOcclusion is expected", func() {
		Expect(alarm.AlarmTypeOcclusion).To(Equal("occlusion"))
	})

	It("AlarmTypeOther is expected", func() {
		Expect(alarm.AlarmTypeOther).To(Equal("other"))
	})

	It("AlarmTypeOverLimit is expected", func() {
		Expect(alarm.AlarmTypeOverLimit).To(Equal("over_limit"))
	})

	It("AlarmTypes returns expected", func() {
		Expect(alarm.AlarmTypes()).To(Equal([]string{"auto_off", "low_insulin", "low_power", "no_delivery", "no_insulin", "no_power", "occlusion", "other", "over_limit"}))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := alarm.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("deviceEvent"))
			Expect(datum.SubType).To(Equal("alarm"))
			Expect(datum.AlarmType).To(BeNil())
			Expect(datum.Status).To(BeNil())
			Expect(datum.StatusID).To(BeNil())
		})
	})

	Context("Alarm", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *alarm.Alarm), expectedErrors ...error) {
					datum := NewAlarm()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *alarm.Alarm) {},
				),
				Entry("type missing",
					func(datum *alarm.Alarm) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &device.Meta{SubType: "alarm"}),
				),
				Entry("type invalid",
					func(datum *alarm.Alarm) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "alarm"}),
				),
				Entry("type device",
					func(datum *alarm.Alarm) { datum.Type = "deviceEvent" },
				),
				Entry("sub type missing",
					func(datum *alarm.Alarm) { datum.SubType = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &device.Meta{Type: "deviceEvent"}),
				),
				Entry("sub type invalid",
					func(datum *alarm.Alarm) { datum.SubType = "invalidSubType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "alarm"), "/subType", &device.Meta{Type: "deviceEvent", SubType: "invalidSubType"}),
				),
				Entry("sub type alarm",
					func(datum *alarm.Alarm) { datum.SubType = "alarm" },
				),
				Entry("alarm type missing",
					func(datum *alarm.Alarm) { datum.AlarmType = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/alarmType", NewMeta()),
				),
				Entry("alarm type invalid",
					func(datum *alarm.Alarm) { datum.AlarmType = pointer.FromString("invalid") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"auto_off", "low_insulin", "low_power", "no_delivery", "no_insulin", "no_power", "occlusion", "other", "over_limit"}), "/alarmType", NewMeta()),
				),
				Entry("alarm type auto_off",
					func(datum *alarm.Alarm) { datum.AlarmType = pointer.FromString("auto_off") },
				),
				Entry("alarm type low_insulin",
					func(datum *alarm.Alarm) { datum.AlarmType = pointer.FromString("low_insulin") },
				),
				Entry("alarm type low_power",
					func(datum *alarm.Alarm) { datum.AlarmType = pointer.FromString("low_power") },
				),
				Entry("alarm type no_delivery",
					func(datum *alarm.Alarm) { datum.AlarmType = pointer.FromString("no_delivery") },
				),
				Entry("alarm type no_insulin",
					func(datum *alarm.Alarm) { datum.AlarmType = pointer.FromString("no_insulin") },
				),
				Entry("alarm type no_power",
					func(datum *alarm.Alarm) { datum.AlarmType = pointer.FromString("no_power") },
				),
				Entry("alarm type occlusion",
					func(datum *alarm.Alarm) { datum.AlarmType = pointer.FromString("occlusion") },
				),
				Entry("alarm type other",
					func(datum *alarm.Alarm) { datum.AlarmType = pointer.FromString("other") },
				),
				Entry("alarm type over_limit",
					func(datum *alarm.Alarm) { datum.AlarmType = pointer.FromString("over_limit") },
				),
				Entry("multiple errors",
					func(datum *alarm.Alarm) {
						datum.Type = "invalidType"
						datum.SubType = "invalidSubType"
						datum.AlarmType = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "alarm"), "/subType", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"auto_off", "low_insulin", "low_power", "no_delivery", "no_insulin", "no_power", "occlusion", "other", "over_limit"}), "/alarmType", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
				),
			)

			DescribeTable("validates the datum with origin external",
				func(mutator func(datum *alarm.Alarm), expectedErrors ...error) {
					datum := NewAlarmWithStatus()
					mutator(datum)
					dataTypesTest.ValidateWithOrigin(datum, structure.OriginExternal, expectedErrors...)
				},
				Entry("succeeds",
					func(datum *alarm.Alarm) {},
				),
				Entry("status missing",
					func(datum *alarm.Alarm) { datum.Status = nil },
				),
				Entry("status valid",
					func(datum *alarm.Alarm) {
						datum.Status = data.DatumAsPointer(dataTypesDeviceStatusTest.NewStatus())
					},
				),
				Entry("status id missing",
					func(datum *alarm.Alarm) { datum.StatusID = nil },
				),
				Entry("status id exists",
					func(datum *alarm.Alarm) { datum.StatusID = pointer.FromString(dataTest.RandomID()) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/statusId", NewMeta()),
				),
				Entry("multiple errors",
					func(datum *alarm.Alarm) {
						datum.StatusID = pointer.FromString(dataTest.RandomID())
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/statusId", NewMeta()),
				),
			)

			DescribeTable("validates the datum with origin internal/store",
				func(mutator func(datum *alarm.Alarm), expectedErrors ...error) {
					datum := NewAlarmWithStatusID()
					mutator(datum)
					dataTypesTest.ValidateWithOrigin(datum, structure.OriginInternal, expectedErrors...)
					dataTypesTest.ValidateWithOrigin(datum, structure.OriginStore, expectedErrors...)
				},
				Entry("succeeds",
					func(datum *alarm.Alarm) {},
				),
				Entry("status missing",
					func(datum *alarm.Alarm) { datum.Status = nil },
				),
				Entry("status exists",
					func(datum *alarm.Alarm) {
						datum.Status = data.DatumAsPointer(dataTypesDeviceStatusTest.NewStatus())
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/status", NewMeta()),
				),
				Entry("status id missing",
					func(datum *alarm.Alarm) { datum.StatusID = nil },
				),
				Entry("status id invalid",
					func(datum *alarm.Alarm) { datum.StatusID = pointer.FromString("invalid") },
					errorsTest.WithPointerSourceAndMeta(data.ErrorValueStringAsIDNotValid("invalid"), "/statusId", NewMeta()),
				),
				Entry("status id valid",
					func(datum *alarm.Alarm) { datum.StatusID = pointer.FromString(dataTest.RandomID()) },
				),
				Entry("multiple errors",
					func(datum *alarm.Alarm) {
						datum.Status = data.DatumAsPointer(dataTypesDeviceStatusTest.NewStatus())
						datum.StatusID = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/status", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(data.ErrorValueStringAsIDNotValid("invalid"), "/statusId", NewMeta()),
				),
			)
		})

		Context("Normalize", func() {
			It("does not modify datum if status is missing", func() {
				datum := NewAlarmWithStatusID()
				expectedDatum := CloneAlarm(datum)
				normalizer := dataNormalizer.New()
				Expect(normalizer).ToNot(BeNil())
				datum.Normalize(normalizer)
				Expect(normalizer.Error()).To(BeNil())
				Expect(normalizer.Data()).To(BeEmpty())
				Expect(datum).To(Equal(expectedDatum))
			})

			It("normalizes the datum and replaces status with status id", func() {
				datumStatus := dataTypesDeviceStatusTest.NewStatus()
				datum := NewAlarmWithStatusID()
				datum.Status = data.DatumAsPointer(datumStatus)
				expectedDatum := CloneAlarm(datum)
				normalizer := dataNormalizer.New()
				Expect(normalizer).ToNot(BeNil())
				datum.Normalize(normalizer)
				Expect(normalizer.Error()).To(BeNil())
				Expect(normalizer.Data()).To(Equal([]data.Datum{datumStatus}))
				expectedDatum.Status = nil
				expectedDatum.StatusID = pointer.FromString(*datumStatus.ID)
				Expect(datum).To(Equal(expectedDatum))
			})
		})
	})
})
