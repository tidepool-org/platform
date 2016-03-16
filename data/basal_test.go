package data_test

import (
	. "github.com/tidepool-org/platform/data"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Basal", func() {

	const (
		userid   = "b676436f60"
		groupid  = "43099shgs55"
		uploadid = "upid_b856b0e6e519"
	)

	var (
		basalObj = map[string]interface{}{
			"userId":           userid, //userid would have been injected by now via the builder
			"groupId":          groupid,
			"uploadId":         uploadid,
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
