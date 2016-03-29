package data

import (
	"encoding/json"
	"io/ioutil"

	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
)

var _ = Describe("Builder", func() {

	var (
		builder        Builder
		injectedFields = map[string]interface{}{"userId": "b676436f60", "uploadId": "43099shgs55", "groupId": "upid_b856b0e6e519"}
	)

	BeforeEach(func() {
		builder = NewTypeBuilder(injectedFields)
	})

	Context("for unkown json", func() {
		It("returns an error", func() {
			_, errs := builder.BuildFromBytes([]byte(`{"Stuff": "2014-06-11T06:00:00"}`))
			Expect(errs).To(Not(BeNil()))
		})
		It("error tells the user what is invalid", func() {
			_, errs := builder.BuildFromBytes([]byte(`{"Stuff": "2014-06-11T06:00:00"}`))
			Expect(errs.Error()).To(Equal(`data.Datum{"Stuff":"2014-06-11T06:00:00"} there is no match for that type`))
		})
	})

	Context("for data stream", func() {
		var (
			datumArray DatumArray
		)
		BeforeEach(func() {
			rawTestData, _ := ioutil.ReadFile("./_fixtures/test_data_stream.json")
			json.Unmarshal(rawTestData, &datumArray)
		})
		It("should not return an error as is valid", func() {
			_, errs := builder.BuildFromDatumArray(datumArray)
			Expect(errs).To(BeNil())
		})
		It("should return process data when valid", func() {
			data, _ := builder.BuildFromDatumArray(datumArray)
			Expect(data).To(Not(BeEmpty()))
		})
	})

	Context("for basal json", func() {
		Context("with all fields", func() {
			jsonBasalData := []byte(`{ "deviceTime": "2014-06-11T06:00:00.000Z", "time": "2014-06-11T06:00:00.000Z","timezoneOffset": 0, "conversionOffset": 0, "type": "basal", "deliveryType": "scheduled", "scheduleName": "Standard", "rate": 2, "duration": 21600000, "deviceId": "InsOmn-3333333333"}`)
			jsonBasalDataExtras := []byte(`{ "deviceTime": "2014-06-11T06:00:00.000Z", "time": "2014-06-11T06:00:00.000Z","timezoneOffset": 0, "conversionOffset": 0, "type": "basal", "deliveryType": "scheduled", "scheduleName": "Standard", "rate": 2, "duration": 21600000, "deviceId": "InsOmn-3333333333", "stuff": "feed me", "moar": 0}`)

			It("should return a basal when there is a match", func() {
				event, _ := builder.BuildFromBytes(jsonBasalData)
				var basalType *Basal
				Expect(event).To(BeAssignableToTypeOf(basalType))
			})

			It("should return no error when there is a match", func() {
				_, err := builder.BuildFromBytes(jsonBasalData)
				Expect(err).To(BeNil())
			})

			It("should return return a basal even when there are extra fields", func() {
				event, _ := builder.BuildFromBytes(jsonBasalDataExtras)
				var basalType *Basal
				Expect(event).To(BeAssignableToTypeOf(basalType))
			})

			It("should return no error even when there are extra fields", func() {
				_, err := builder.BuildFromBytes(jsonBasalDataExtras)
				Expect(err).To(BeNil())
			})
		})
		Context("with only core fields returns", func() {
			jsonCoreBasalData := []byte(`{  "time": "2014-06-11T06:00:00.000Z","type": "basal", "deliveryType": "scheduled", "scheduleName": "Standard", "rate": 2, "duration": 21600000, "deviceId": "InsOmn-3333333333"}`)

			It("a basal when there is a match", func() {
				event, _ := builder.BuildFromBytes(jsonCoreBasalData)
				var basalType *Basal
				Expect(event).To(BeAssignableToTypeOf(basalType))
			})

			It("an error when there is a match", func() {
				_, err := builder.BuildFromBytes(jsonCoreBasalData)
				Expect(err).To(BeNil())
			})
		})
	})

	Context("for deviceEvent json", func() {
		Context("with all fields returns", func() {
			jsonDeviceEventData := []byte(`{ "deviceTime": "2014-06-11T06:00:00.000Z", "time": "2014-06-11T06:00:00.000Z","timezoneOffset": 0, "conversionOffset": 0, "type": "deviceEvent", "subType": "alarm", "deviceId": "InsOmn-3333333333"}`)
			jsonDeviceEventDataExtras := []byte(`{ "deviceTime": "2014-06-11T06:00:00.000Z", "time": "2014-06-11T06:00:00.000Z","timezoneOffset": 0, "conversionOffset": 0, "type": "deviceEvent", "subType": "alarm", "deviceId": "InsOmn-3333333333", "stuff": "feed me", "moar": 0}`)

			It("deviceEvent when there is a match", func() {
				event, _ := builder.BuildFromBytes(jsonDeviceEventData)
				var deviceEventType *DeviceEvent
				Expect(event).To(BeAssignableToTypeOf(deviceEventType))
			})

			It("a basal even when there are extra fields", func() {
				event, _ := builder.BuildFromBytes(jsonDeviceEventDataExtras)
				var deviceEventType *DeviceEvent
				Expect(event).To(BeAssignableToTypeOf(deviceEventType))
			})
		})
		Context("with only core fields returns", func() {
			jsonCoreDeviceEventData := []byte(`{ "time": "2014-06-11T06:00:00.000Z","type": "deviceEvent", "subType": "alarm", "deviceId": "InsOmn-3333333333"}`)

			It("deviceEvent when there is a match", func() {
				event, _ := builder.BuildFromBytes(jsonCoreDeviceEventData)
				var deviceEventType *DeviceEvent
				Expect(event).To(BeAssignableToTypeOf(deviceEventType))
			})
		})
	})
})
