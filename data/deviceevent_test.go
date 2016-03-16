package data_test

import (
	. "github.com/tidepool-org/platform/data"

	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
)

var _ = Describe("DeviceEvent", func() {

	const (
		userid   = "b676436f60"
		groupid  = "43099shgs55"
		uploadid = "upid_b856b0e6e519"
	)

	var (
		deviceEventObj = map[string]interface{}{
			"userId":           userid, //userid would have been injected by now via the builder
			"groupId":          groupid,
			"uploadId":         uploadid,
			"deviceTime":       "2014-06-11T06:00:00.000Z",
			"time":             "2014-06-11T06:00:00.000Z",
			"timezoneOffset":   0,
			"conversionOffset": 0,
			"clockDriftOffset": 0,
			"type":             "deviceEvent",
			"subType":          "alarm",
			"deviceId":         "InsOmn-888888888",
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
