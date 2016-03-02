package data_test

import (
	"encoding/json"
	"io/ioutil"

	. "github.com/tidepool-org/platform/data"

	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
)

var _ = Describe("Builder", func() {

	var (
		builder Builder

		injectedFields = map[string]interface{}{"userId": "b676436f60"}

		jsonBasalData       = []byte(`{ "deviceTime": "2014-06-11T06:00:00.000Z", "time": "2014-06-11T06:00:00.000Z","timezoneOffset": 0, "conversionOffset": 0, "type": "basal", "deliveryType": "scheduled", "scheduleName": "Standard", "rate": 2, "duration": 21600000, "deviceId": "tools"}`)
		jsonBasalDataExtras = []byte(`{ "deviceTime": "2014-06-11T06:00:00.000Z", "time": "2014-06-11T06:00:00.000Z","timezoneOffset": 0, "conversionOffset": 0, "type": "basal", "deliveryType": "scheduled", "scheduleName": "Standard", "rate": 2, "duration": 21600000, "deviceId": "tools", "stuff": "feed me", "moar": 0}`)

		jsonDeviceEventData       = []byte(`{ "deviceTime": "2014-06-11T06:00:00.000Z", "time": "2014-06-11T06:00:00.000Z","timezoneOffset": 0, "conversionOffset": 0, "type": "deviceEvent", "subType": "alarm", "deviceId": "platform-tests"}`)
		jsonDeviceEventDataExtras = []byte(`{ "deviceTime": "2014-06-11T06:00:00.000Z", "time": "2014-06-11T06:00:00.000Z","timezoneOffset": 0, "conversionOffset": 0, "type": "deviceEvent", "subType": "alarm", "deviceId": "platform-tests", "stuff": "feed me", "moar": 0}`)
	)

	BeforeEach(func() {
		builder = NewTypeBuilder(injectedFields)
	})

	Context("for unkown json", func() {
		It("should return an error", func() {
			_, errs := builder.BuildFromRaw([]byte(`{"Stuff": "2014-06-11T06:00:00"}`))
			Expect(errs).To(Not(BeNil()))
		})
		It("should tell user what is invalid in error", func() {
			_, errs := builder.BuildFromRaw([]byte(`{"Stuff": "2014-06-11T06:00:00"}`))
			Expect(errs.Error()).To(Equal("processing map[Stuff:2014-06-11T06:00:00] found: there is no match for that type"))
		})
	})

	Context("for data stream", func() {
		var (
			dataSet GenericDataset
		)
		BeforeEach(func() {
			rawTestData, _ := ioutil.ReadFile("./test_data_stream.json")
			json.Unmarshal(rawTestData, &dataSet)
		})
		It("should not return an error as is valid", func() {
			_, errs := builder.BuildFromDataSet(dataSet)
			Expect(errs).To(BeNil())
		})
		It("should return process data when valid", func() {
			data, _ := builder.BuildFromDataSet(dataSet)
			Expect(data).To(Not(BeEmpty()))
		})
	})

	Context("for basal json", func() {
		It("should return a basal when there is a match", func() {
			event, _ := builder.BuildFromRaw(jsonBasalData)
			var basalType *Basal
			Expect(event).To(BeAssignableToTypeOf(basalType))
		})

		It("should return no error when there is a match", func() {
			_, err := builder.BuildFromRaw(jsonBasalData)
			Expect(err).To(BeNil())
		})

		It("should return return a basal even when there are extra feilds", func() {
			event, _ := builder.BuildFromRaw(jsonBasalDataExtras)
			var basalType *Basal
			Expect(event).To(BeAssignableToTypeOf(basalType))
		})

		It("should return no error even when there are extra feilds", func() {
			_, err := builder.BuildFromRaw(jsonBasalDataExtras)
			Expect(err).To(BeNil())
		})
	})

	Context("for deviceEvent json", func() {
		It("should return deviceEvent when there is a match", func() {
			event, _ := builder.BuildFromRaw(jsonDeviceEventData)
			var deviceEventType *DeviceEvent
			Expect(event).To(BeAssignableToTypeOf(deviceEventType))
		})

		It("should return return a basal even when there are extra feilds", func() {
			event, _ := builder.BuildFromRaw(jsonDeviceEventDataExtras)
			var deviceEventType *DeviceEvent
			Expect(event).To(BeAssignableToTypeOf(deviceEventType))
		})
	})
})

var _ = Describe("Base", func() {

	var (
		basalObj = map[string]interface{}{
			"userId":           "b676436f60", //userid would have been injected by now via the builder
			"deviceTime":       "2014-06-11T06:00:00.000Z",
			"time":             "2014-06-11T06:00:00.000Z",
			"timezoneOffset":   0,
			"conversionOffset": 0,
			"type":             "basal",
			"deliveryType":     "scheduled",
			"scheduleName":     "Standard",
			"rate":             2.2,
			"duration":         21600000,
			"deviceId":         "tools",
		}
	)

	Context("can be built from obj", func() {
		It("should return a the base types if the obj is valid", func() {
			base, _ := BuildBase(basalObj)
			var baseType Base
			Expect(base).To(BeAssignableToTypeOf(baseType))
		})
		It("should return and error object that is empty but not nil", func() {
			_, err := BuildBase(basalObj)
			Expect(err).To(Not(BeNil()))
			Expect(err.IsEmpty()).To(BeTrue())
		})
	})
})

var _ = Describe("Basal", func() {

	var (
		basalObj = map[string]interface{}{
			"userId":           "b676436f60", //userid would have been injected by now via the builder
			"time":             "2016-02-25T23:02:00.000Z",
			"timezoneOffset":   -480,
			"clockDriftOffset": 0,
			"conversionOffset": 0,
			"deviceTime":       "2016-02-25T15:02:00.000Z",
			"deviceId":         "IR1285-79-36047-15",
			"type":             "basal",
			"deliveryType":     "scheduled",
			"scheduleName":     "DEFAULT",
			"rate":             1.75,
			"duration":         28800000,
		}
	)

	Context("datum from obj", func() {
		It("should return a basal if the obj is valid", func() {
			basal, _ := BuildBasal(basalObj)
			var basalType *Basal
			Expect(basal).To(BeAssignableToTypeOf(basalType))
		})
		It("should produce no error when valid", func() {
			_, err := BuildBasal(basalObj)
			Expect(err).To(BeNil())
		})
	})

	Context("dataset from builder", func() {
		It("should return a basal if the obj is valid", func() {
			basal, _ := BuildBasal(basalObj)
			var basalType *Basal
			Expect(basal).To(BeAssignableToTypeOf(basalType))
		})
		It("should produce no error when valid", func() {
			_, err := BuildBasal(basalObj)
			Expect(err).To(BeNil())
		})
	})
})

var _ = Describe("DeviceEvent", func() {

	var (
		deviceEventObj = map[string]interface{}{
			"userId":           "b676436f60", //userid would have been injected by now via the builder
			"deviceTime":       "2014-06-11T06:00:00.000Z",
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
		It("should produce no error when valid", func() {
			_, err := BuildDeviceEvent(deviceEventObj)
			Expect(err).To(BeNil())
		})
	})
})
