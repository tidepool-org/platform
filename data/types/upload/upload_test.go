package upload

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
	"github.com/tidepool-org/platform/validate"
)

func TestingDatumBase() map[string]interface{} {
	return map[string]interface{}{
		"userId":           "b676436f60",
		"groupId":          "43099shgs55",
		"uploadId":         "upid_b856b0e6e519",
		"deviceTime":       "2014-06-11T06:00:00.000Z",
		"time":             "2014-06-11T06:00:00.000Z",
		"timezoneOffset":   0,
		"conversionOffset": 0,
		"clockDriftOffset": 0,
		"deviceId":         "InsOmn-111111111",
	}
}

var _ = Describe("Upload", func() {

	var uploadObj = TestingDatumBase()
	var processing validate.ErrorProcessing

	Context("upload record from obj", func() {

		BeforeEach(func() {
			processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
			uploadObj["type"] = "upload"
			uploadObj["computerTime"] = "2014-01-01T14:00:00"
			uploadObj["uploadId"] = "123-my-upload-id"
			uploadObj["byUser"] = "123-my-user-id"
			uploadObj["version"] = "tidepool-uploader 0.1.0"
			uploadObj["deviceManufacturers"] = []string{"Medtronic"}
			uploadObj["deviceModel"] = "Paradigm 522"
			uploadObj["deviceSerialNumber"] = "12345"
			uploadObj["deviceTags"] = []string{"insulin-pump"}
			uploadObj["deviceId"] = "123-my-upload-id"
			uploadObj["timeProcessing"] = "none"
		})

		It("when valid", func() {
			uploadRec := Build(uploadObj, processing)
			var recordType *Record
			Expect(uploadRec).To(BeAssignableToTypeOf(recordType))
			Expect(processing.HasErrors()).To(BeFalse())
		})

		Context("validation", func() {
			Context("computerTime", func() {
				It("fails if not valid time", func() {
					uploadObj["computerTime"] = "Tuesday 14th May, 2015"
					uploadRec := Build(uploadObj, processing)
					Expect(processing.HasErrors()).To(BeTrue())
					Expect(uploadRec).To(Not(BeNil()))
				})
				/*It("is required", func() {
					delete(uploadObj, "computerTime")
					uploadRec := Build(uploadObj, processing)
					Expect(processing.HasErrors()).To(BeTrue(), "No errors found when expected")
					Expect(uploadRec).To(Not(BeNil()))
				})
				It("cannot be empty", func() {
					uploadObj["computerTime"] = ""
					uploadRec := Build(uploadObj, processing)
					Expect(processing.HasErrors()).To(BeTrue(), "No errors found when expected")
					Expect(uploadRec).To(Not(BeNil()))
				})*/
			})
			/*
				Context("uploadId", func() {
					It("is required", func() {
						delete(uploadObj, "uploadId")
						uploadRec := Build(uploadObj, processing)
						Expect(processing.HasErrors()).To(BeTrue(), "No errors found when expected")
						Expect(uploadRec).To(Not(BeNil()))
					})
					It("cannot be empty", func() {
						uploadObj["uploadId"] = ""
						uploadRec := Build(uploadObj, processing)
						Expect(processing.HasErrors()).To(BeTrue(), "No errors found when expected")
						Expect(uploadRec).To(Not(BeNil()))
					})
				})
				Context("byUser", func() {
					It("is required", func() {
						delete(uploadObj, "byUser")
						uploadRec := Build(uploadObj, processing)
						Expect(processing.HasErrors()).To(BeTrue(), "No errors found when expected")
						Expect(uploadRec).To(Not(BeNil()))
					})
					It("cannot be empty", func() {
						uploadObj["byUser"] = ""
						uploadRec := Build(uploadObj, processing)
						Expect(processing.HasErrors()).To(BeTrue(), "No errors found when expected")
						Expect(uploadRec).To(Not(BeNil()))
					})
				})
				Context("version", func() {
					It("is required", func() {
						delete(uploadObj, "version")
						uploadRec := Build(uploadObj, processing)
						Expect(processing.HasErrors()).To(BeTrue(), "No errors found when expected")
						Expect(uploadRec).To(Not(BeNil()))
					})
					It("cannot be empty", func() {
						uploadObj["version"] = ""
						uploadRec := Build(uploadObj, processing)
						Expect(processing.HasErrors()).To(BeTrue(), "No errors found when expected")
						Expect(uploadRec).To(Not(BeNil()))
					})
				})
				Context("deviceModel", func() {
					It("is required", func() {
						delete(uploadObj, "deviceModel")
						uploadRec := Build(uploadObj, processing)
						Expect(processing.HasErrors()).To(BeTrue(), "No errors found when expected")
						Expect(uploadRec).To(Not(BeNil()))
					})
					It("cannot be empty", func() {
						uploadObj["deviceModel"] = ""
						uploadRec := Build(uploadObj, processing)
						Expect(processing.HasErrors()).To(BeTrue(), "No errors found when expected")
						Expect(uploadRec).To(Not(BeNil()))
					})
				})
				Context("deviceSerialNumber", func() {
					It("is required", func() {
						delete(uploadObj, "deviceSerialNumber")
						uploadRec := Build(uploadObj, processing)
						Expect(processing.HasErrors()).To(BeTrue(), "No errors found when expected")
						Expect(uploadRec).To(Not(BeNil()))
					})
					It("cannot be empty", func() {
						uploadObj["deviceSerialNumber"] = ""
						uploadRec := Build(uploadObj, processing)
						Expect(processing.HasErrors()).To(BeTrue(), "No errors found when expected")
						Expect(uploadRec).To(Not(BeNil()))
					})
				})
				Context("deviceId", func() {
					It("is required", func() {
						delete(uploadObj, "deviceId")
						uploadRec := Build(uploadObj, processing)
						Expect(processing.HasErrors()).To(BeTrue(), "No errors found when expected")
						Expect(uploadRec).To(Not(BeNil()))
					})
					It("cannot be empty", func() {
						uploadObj["deviceId"] = ""
						uploadRec := Build(uploadObj, processing)
						Expect(processing.HasErrors()).To(BeTrue(), "No errors found when expected")
						Expect(uploadRec).To(Not(BeNil()))
					})
				})*/
			Context("deviceManufacturers", func() {
				/*It("is required", func() {
					delete(uploadObj, "deviceManufacturers")
					uploadRec := Build(uploadObj, processing)
					Expect(processing.HasErrors()).To(BeTrue(), "No errors found when expected")
					Expect(uploadRec).To(Not(BeNil()))
				})
				It("cannot be empty", func() {
					uploadObj["deviceManufacturers"] = []string{""}
					uploadRec := Build(uploadObj, processing)
					Expect(processing.HasErrors()).To(BeTrue(), "No errors found when expected")
					Expect(uploadRec).To(Not(BeNil()))
				})*/
			})
			Context("deviceTags", func() {
				/*It("is required", func() {
					delete(uploadObj, "deviceTags")
					uploadRec := Build(uploadObj, processing)
					Expect(processing.HasErrors()).To(BeTrue())
					Expect(uploadRec).To(Not(BeNil()))
				})
				It("cannot be empty", func() {
					uploadObj["deviceTags"] = []string{""}
					uploadRec := Build(uploadObj, processing)
					Expect(processing.HasErrors()).To(BeTrue(), "No errors found when expected")
					Expect(uploadRec).To(Not(BeNil()))
				})
				It("has to be in approved list", func() {
					uploadObj["deviceTags"] = []string{"unknown"}
					uploadRec := Build(uploadObj, processing)
					Expect(processing.HasErrors()).To(BeTrue(), "No errors found when expected")
					Expect(uploadRec).To(Not(BeNil()))
				})*/
				It("can be any of insulin-pump, cgm, bgm", func() {
					uploadObj["deviceTags"] = []string{"insulin-pump", "cgm", "bgm"}
					uploadRec := Build(uploadObj, processing)
					Expect(processing.HasErrors()).To(BeFalse())
					Expect(uploadRec).To(Not(BeNil()))
				})
			})
			Context("timeProcessing", func() {
				/*It("is required", func() {
					delete(uploadObj, "timeProcessing")
					uploadRec := Build(uploadObj, processing)
					Expect(processing.HasErrors()).To(BeTrue(), "No errors found when expected")
					Expect(uploadRec).To(Not(BeNil()))
				})
				It("cannot be empty", func() {
					uploadObj["timeProcessing"] = ""
					uploadRec := Build(uploadObj, processing)
					Expect(processing.HasErrors()).To(BeTrue(), "No errors found when expected")
					Expect(uploadRec).To(Not(BeNil()))
				})*/
				It("can be across-the-board-timezone", func() {
					uploadObj["timeProcessing"] = "across-the-board-timezone"
					uploadRec := Build(uploadObj, processing)
					Expect(processing.HasErrors()).To(BeFalse())
					Expect(uploadRec).To(Not(BeNil()))
				})
				It("can be utc-bootstrapping", func() {
					uploadObj["timeProcessing"] = "utc-bootstrapping"
					uploadRec := Build(uploadObj, processing)
					Expect(processing.HasErrors()).To(BeFalse())
					Expect(uploadRec).To(Not(BeNil()))
				})
				It("can be none", func() {
					uploadObj["timeProcessing"] = "none"
					uploadRec := Build(uploadObj, processing)
					Expect(processing.HasErrors()).To(BeFalse())
					Expect(uploadRec).To(Not(BeNil()))
				})
			})
		})

	})
})
