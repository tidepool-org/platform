package upload

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/data/types"
)

var _ = Describe("Upload", func() {

	var uploadObj = fixtures.TestingDatumBase()
	var helper *types.TestingHelper

	BeforeEach(func() {
		helper = types.NewTestingHelper()

		uploadObj["type"] = "upload"
		uploadObj["computerTime"] = "2014-01-01T14:00:00"
		uploadObj["uploadId"] = "123-my-upload-id"
		uploadObj["byUser"] = "123-my-user-id"
		uploadObj["version"] = "tidepool-uploader 0.1.0"
		uploadObj["deviceManufacturers"] = []string{"Medtronic"}
		uploadObj["deviceModel"] = "Paradigm 522"
		uploadObj["deviceSerialNumber"] = "123456-blah"
		uploadObj["deviceTags"] = []string{"insulin-pump"}
		uploadObj["deviceId"] = "123-my-upload-id"
		uploadObj["timeProcessing"] = "none"
	})

	Context("upload record from obj", func() {

		It("when valid", func() {
			Expect(helper.ValidDataType(Build(uploadObj, helper.ErrorProcessing))).To(BeNil())
		})

		Context("validation", func() {
			Context("computerTime", func() {
				It("fails if not valid time", func() {
					uploadObj["computerTime"] = "Tuesday 14th May, 2015"

					Expect(
						helper.ErrorIsExpected(
							Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/computerTime",
								Detail: "Times need to be ISO 8601 format and not in the future given 'Tuesday 14th May, 2015'",
							}),
					).To(BeNil())

				})

				It("is required", func() {
					delete(uploadObj, "computerTime")

					Expect(
						helper.ErrorIsExpected(
							Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/computerTime",
								Detail: "Times need to be ISO 8601 format and not in the future given '<nil>'",
							}),
					).To(BeNil())
				})

				It("cannot be empty", func() {
					uploadObj["computerTime"] = ""
					Expect(
						helper.ErrorIsExpected(
							Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/computerTime",
								Detail: "Times need to be ISO 8601 format and not in the future given ''",
							}),
					).To(BeNil())
				})

			})

			Context("uploadId", func() {
				It("is required", func() {
					delete(uploadObj, "uploadId")
					Expect(
						helper.ErrorIsExpected(
							Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/uploadId",
								Detail: "This is a required field need needs to be 10+ characters in length given '<nil>'",
							}),
					).To(BeNil())
				})
				It("cannot be empty", func() {
					uploadObj["uploadId"] = ""
					Expect(
						helper.ErrorIsExpected(
							Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/uploadId",
								Detail: "This is a required field need needs to be 10+ characters in length given ''",
							}),
					).To(BeNil())
				})
			})
			Context("byUser", func() {

				It("is required", func() {
					delete(uploadObj, "byUser")
					Expect(
						helper.ErrorIsExpected(
							Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/byUser",
								Detail: "This is a required field need needs to be 10+ characters in length given '<nil>'",
							}),
					).To(BeNil())
				})

				It("cannot be empty", func() {
					uploadObj["byUser"] = ""

					Expect(
						helper.ErrorIsExpected(
							Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/byUser",
								Detail: "This is a required field need needs to be 10+ characters in length given ''",
							}),
					).To(BeNil())
				})
			})
			Context("version", func() {

				It("is required", func() {
					delete(uploadObj, "version")

					Expect(
						helper.ErrorIsExpected(
							Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/version",
								Detail: "This is a required field need needs to be 10+ characters in length given '<nil>'",
							}),
					).To(BeNil())
				})

				It("cannot be empty", func() {
					uploadObj["version"] = ""

					Expect(
						helper.ErrorIsExpected(
							Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/version",
								Detail: "This is a required field need needs to be 10+ characters in length given ''",
							}),
					).To(BeNil())
				})

			})
			Context("deviceModel", func() {

				It("is required", func() {
					delete(uploadObj, "deviceModel")

					Expect(
						helper.ErrorIsExpected(
							Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/deviceModel",
								Detail: "This is a required field need needs to be 10+ characters in length given '<nil>'",
							}),
					).To(BeNil())
				})

				It("cannot be empty", func() {
					uploadObj["deviceModel"] = ""

					Expect(
						helper.ErrorIsExpected(
							Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/deviceModel",
								Detail: "This is a required field need needs to be 10+ characters in length given ''",
							}),
					).To(BeNil())
				})

			})
			Context("deviceSerialNumber", func() {

				It("is required", func() {
					delete(uploadObj, "deviceSerialNumber")

					Expect(
						helper.ErrorIsExpected(
							Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/deviceSerialNumber",
								Detail: "This is a required field need needs to be 10+ characters in length given '<nil>'",
							}),
					).To(BeNil())
				})

				It("cannot be empty", func() {
					uploadObj["deviceSerialNumber"] = ""

					Expect(
						helper.ErrorIsExpected(
							Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/deviceSerialNumber",
								Detail: "This is a required field need needs to be 10+ characters in length given ''",
							}),
					).To(BeNil())
				})

			})
			Context("deviceId", func() {
				It("is required", func() {
					delete(uploadObj, "deviceId")

					Expect(
						helper.ErrorIsExpected(
							Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/deviceId",
								Detail: "This is a required field need needs to be 10+ characters in length given '<nil>'",
							}),
					).To(BeNil())
				})
				It("cannot be empty", func() {
					uploadObj["deviceId"] = ""

					Expect(
						helper.ErrorIsExpected(
							Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/deviceId",
								Detail: "This is a required field need needs to be 10+ characters in length given ''",
							}),
					).To(BeNil())
				})
			})
			Context("deviceManufacturers", func() {

				It("is required", func() {
					delete(uploadObj, "deviceManufacturers")

					Expect(
						helper.ErrorIsExpected(
							Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/deviceManufacturers",
								Detail: "Must contain at least one manufacturer name given '<nil>'",
							}),
					).To(BeNil())

				})

				It("cannot be empty", func() {
					uploadObj["deviceManufacturers"] = []string{}

					Expect(
						helper.ErrorIsExpected(
							Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/deviceManufacturers",
								Detail: "Must contain at least one manufacturer name given '[]'",
							}),
					).To(BeNil())
				})

			})
			Context("deviceTags", func() {

				It("is required", func() {
					delete(uploadObj, "deviceTags")

					Expect(
						helper.ErrorIsExpected(
							Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/deviceTags",
								Detail: "Must be one of insulin-pump, cgm, bgm given '<nil>'",
							}),
					).To(BeNil())
				})

				It("cannot be empty", func() {
					uploadObj["deviceTags"] = []string{""}

					Expect(
						helper.ErrorIsExpected(
							Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/deviceTags",
								Detail: "Must be one of insulin-pump, cgm, bgm given '[]'",
							}),
					).To(BeNil())
				})

				It("has to be in approved list", func() {
					uploadObj["deviceTags"] = []string{"unknown"}

					Expect(
						helper.ErrorIsExpected(
							Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/deviceTags",
								Detail: "Must be one of insulin-pump, cgm, bgm given '[unknown]'",
							}),
					).To(BeNil())
				})

				It("can be any of insulin-pump, cgm, bgm", func() {
					uploadObj["deviceTags"] = []string{"insulin-pump", "cgm", "bgm"}
					Expect(helper.ValidDataType(Build(uploadObj, helper.ErrorProcessing))).To(BeNil())
				})
			})
			Context("timeProcessing", func() {

				It("is required", func() {
					delete(uploadObj, "timeProcessing")

					Expect(
						helper.ErrorIsExpected(
							Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/timeProcessing",
								Detail: "Must be one of across-the-board-timezone, utc-bootstrapping, none given '<nil>'",
							}),
					).To(BeNil())
				})

				It("cannot be empty", func() {
					uploadObj["timeProcessing"] = ""

					Expect(
						helper.ErrorIsExpected(
							Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/timeProcessing",
								Detail: "Must be one of across-the-board-timezone, utc-bootstrapping, none given ''",
							}),
					).To(BeNil())
				})

				It("can be across-the-board-timezone", func() {
					uploadObj["timeProcessing"] = "across-the-board-timezone"
					Expect(helper.ValidDataType(Build(uploadObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("can be utc-bootstrapping", func() {
					uploadObj["timeProcessing"] = "utc-bootstrapping"
					Expect(helper.ValidDataType(Build(uploadObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("can be none", func() {
					uploadObj["timeProcessing"] = "none"
					Expect(helper.ValidDataType(Build(uploadObj, helper.ErrorProcessing))).To(BeNil())
				})
			})
		})

	})
})
