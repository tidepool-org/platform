package data_test

import (
	"encoding/json"
	"io/ioutil"

	. "github.com/tidepool-org/platform/data"

	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
)

var _ = Describe("Builder", func() {

	const (
		userid   = "b676436f60"
		groupid  = "43099shgs55"
		uploadid = "upid_b856b0e6e519"
	)

	var (
		builder Builder

		injectedFields = map[string]interface{}{"userId": userid, "uploadId": uploadid, "groupId": groupid}
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
			dataSet Dataset
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
		Context("with all fields", func() {
			jsonBasalData := []byte(`{ "deviceTime": "2014-06-11T06:00:00.000Z", "time": "2014-06-11T06:00:00.000Z","timezoneOffset": 0, "conversionOffset": 0, "type": "basal", "deliveryType": "scheduled", "scheduleName": "Standard", "rate": 2, "duration": 21600000, "deviceId": "InsOmn-3333333333"}`)
			jsonBasalDataExtras := []byte(`{ "deviceTime": "2014-06-11T06:00:00.000Z", "time": "2014-06-11T06:00:00.000Z","timezoneOffset": 0, "conversionOffset": 0, "type": "basal", "deliveryType": "scheduled", "scheduleName": "Standard", "rate": 2, "duration": 21600000, "deviceId": "InsOmn-3333333333", "stuff": "feed me", "moar": 0}`)

			It("should return a basal when there is a match", func() {
				event, _ := builder.BuildFromRaw(jsonBasalData)
				var basalType *Basal
				Expect(event).To(BeAssignableToTypeOf(basalType))
			})

			It("should return no error when there is a match", func() {
				_, err := builder.BuildFromRaw(jsonBasalData)
				Expect(err).To(BeNil())
			})

			It("should return return a basal even when there are extra fields", func() {
				event, _ := builder.BuildFromRaw(jsonBasalDataExtras)
				var basalType *Basal
				Expect(event).To(BeAssignableToTypeOf(basalType))
			})

			It("should return no error even when there are extra fields", func() {
				_, err := builder.BuildFromRaw(jsonBasalDataExtras)
				Expect(err).To(BeNil())
			})
		})
		Context("with only core fields", func() {
			jsonCoreBasalData := []byte(`{  "time": "2014-06-11T06:00:00.000Z","type": "basal", "deliveryType": "scheduled", "scheduleName": "Standard", "rate": 2, "duration": 21600000, "deviceId": "InsOmn-3333333333"}`)

			It("should return a basal when there is a match", func() {
				event, _ := builder.BuildFromRaw(jsonCoreBasalData)
				var basalType *Basal
				Expect(event).To(BeAssignableToTypeOf(basalType))
			})

			It("should return no error when there is a match", func() {
				_, err := builder.BuildFromRaw(jsonCoreBasalData)
				Expect(err).To(BeNil())
			})
		})
	})

	Context("for deviceEvent json", func() {
		Context("with all fields", func() {
			jsonDeviceEventData := []byte(`{ "deviceTime": "2014-06-11T06:00:00.000Z", "time": "2014-06-11T06:00:00.000Z","timezoneOffset": 0, "conversionOffset": 0, "type": "deviceEvent", "subType": "alarm", "deviceId": "InsOmn-3333333333"}`)
			jsonDeviceEventDataExtras := []byte(`{ "deviceTime": "2014-06-11T06:00:00.000Z", "time": "2014-06-11T06:00:00.000Z","timezoneOffset": 0, "conversionOffset": 0, "type": "deviceEvent", "subType": "alarm", "deviceId": "InsOmn-3333333333", "stuff": "feed me", "moar": 0}`)

			It("should return deviceEvent when there is a match", func() {
				event, _ := builder.BuildFromRaw(jsonDeviceEventData)
				var deviceEventType *DeviceEvent
				Expect(event).To(BeAssignableToTypeOf(deviceEventType))
			})

			It("should return return a basal even when there are extra fields", func() {
				event, _ := builder.BuildFromRaw(jsonDeviceEventDataExtras)
				var deviceEventType *DeviceEvent
				Expect(event).To(BeAssignableToTypeOf(deviceEventType))
			})
		})
		Context("with only core fields", func() {
			jsonCoreDeviceEventData := []byte(`{ "time": "2014-06-11T06:00:00.000Z","type": "deviceEvent", "subType": "alarm", "deviceId": "InsOmn-3333333333"}`)

			It("should return deviceEvent when there is a match", func() {
				event, _ := builder.BuildFromRaw(jsonCoreDeviceEventData)
				var deviceEventType *DeviceEvent
				Expect(event).To(BeAssignableToTypeOf(deviceEventType))
			})
		})
	})
})
