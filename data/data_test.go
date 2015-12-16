package data_test

import (
	. "github.com/tidepool-org/platform/data"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Data", func() {
	Context("with no parameters", func() {
		It("should return data", func() {
			Expect(GetData()).To(Equal("data"))
		})
	})
})

var _ = Describe("Builder", func() {

	var (
		builder Builder

		jsonBasalData       = []byte(`{"deviceTime": "2014-06-11T06:00:00", "time": "2014-06-11T06:00:00.000Z","timezoneOffset": 0, "conversionOffset": 0, "type": "basal", "deliveryType": "scheduled", "scheduleName": "Standard", "rate": 2, "duration": 21600000, "deviceId": "tools"}`)
		jsonBasalDataExtras = []byte(`{"deviceTime": "2014-06-11T06:00:00", "time": "2014-06-11T06:00:00.000Z","timezoneOffset": 0, "conversionOffset": 0, "type": "basal", "deliveryType": "scheduled", "scheduleName": "Standard", "rate": 2, "duration": 21600000, "deviceId": "tools", "stuff": "feed me", "moar": 0}`)

		jsonDeviceEventData       = []byte(`{"deviceTime": "2014-06-11T06:00:00", "time": "2014-06-11T06:00:00.000Z","timezoneOffset": 0, "conversionOffset": 0, "type": "deviceEvent", "subType": "alarm", "deviceId": "platform-tests"}`)
		jsonDeviceEventDataExtras = []byte(`{"deviceTime": "2014-06-11T06:00:00", "time": "2014-06-11T06:00:00.000Z","timezoneOffset": 0, "conversionOffset": 0, "type": "deviceEvent", "subType": "alarm", "deviceId": "platform-tests", "stuff": "feed me", "moar": 0}`)
	)

	BeforeEach(func() {
		builder = NewTypeBuilder()
	})

	Context("for unkown json", func() {
		It("should return an error", func() {
			_, errs := builder.Build([]byte(`{"Stuff": "2014-06-11T06:00:00"}`))
			Expect(errs).To(Not(BeNil()))
		})
		It("should tell user what is invalid in error", func() {
			_, errs := builder.Build([]byte(`{"Stuff": "2014-06-11T06:00:00"}`))
			Expect(errs.Error()).To(Equal("processing map[Stuff:2014-06-11T06:00:00] found: there is no match for that type"))
		})
	})

	Context("for basal json", func() {
		It("should return a basal when there is a match", func() {
			event, _ := builder.Build(jsonBasalData)
			var basalType *Basal
			Expect(event).To(BeAssignableToTypeOf(basalType))
		})

		It("should return return a basal even when there are extra feilds", func() {
			event, _ := builder.Build(jsonBasalDataExtras)
			var basalType *Basal
			Expect(event).To(BeAssignableToTypeOf(basalType))
		})

	})

	Context("for deviceEvent json", func() {
		It("should return deviceEvent when there is a match", func() {
			event, _ := builder.Build(jsonDeviceEventData)
			var deviceEventType *DeviceEvent
			Expect(event).To(BeAssignableToTypeOf(deviceEventType))
		})

		It("should return return a basal even when there are extra feilds", func() {
			event, _ := builder.Build(jsonDeviceEventDataExtras)
			var deviceEventType *DeviceEvent
			Expect(event).To(BeAssignableToTypeOf(deviceEventType))
		})
	})
})

var _ = Describe("Basal", func() {

	var (
		basalObj = map[string]interface{}{
			"deviceTime":       "2014-06-11T06:00:00",
			"time":             "2014-06-11T06:00:00.000Z",
			"timezoneOffset":   0,
			"conversionOffset": 0,
			"type":             "basal",
			"deliveryType":     "scheduled",
			"scheduleName":     "Standard",
			"rate":             2,
			"duration":         21600000,
			"deviceId":         "tools",
		}
	)

	Context("can be built from obj", func() {
		It("should return a basal if the obj is valid", func() {
			basal, _ := BuildBasal(basalObj)
			var basalType *Basal
			Expect(basal).To(BeAssignableToTypeOf(basalType))
		})
	})
})

var _ = Describe("DeviceEvent", func() {

	var (
		deviceEventObj = map[string]interface{}{
			"deviceTime":       "2014-06-11T06:00:00",
			"time":             "2014-06-11T06:00:00.000Z",
			"timezoneOffset":   0,
			"conversionOffset": 0,
			"type":             "deviceEvent",
			"subType":          "alarm",
			"deviceId":         "platform-tests",
		}
	)

	Context("can be built from obj", func() {
		It("should return a basal if the obj is valid", func() {
			deviceEvent, _ := BuildDeviceEvent(deviceEventObj)
			var deviceEventType *DeviceEvent
			Expect(deviceEvent).To(BeAssignableToTypeOf(deviceEventType))
		})
	})
})
