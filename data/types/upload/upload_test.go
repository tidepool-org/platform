package upload_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fixtures "github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/upload"
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
		uploadObj["deviceManufacturers"] = []interface{}{"Medtronic"}
		uploadObj["deviceModel"] = "Paradigm 522"
		uploadObj["deviceSerialNumber"] = "123456-blah"
		uploadObj["deviceTags"] = []interface{}{"insulin-pump"}
		uploadObj["timeProcessing"] = "none"
	})

	Context("upload record from obj", func() {

		It("when valid", func() {
			Expect(helper.ValidDataType(upload.Build(uploadObj, helper.ErrorProcessing))).To(BeNil())
		})

		Context("validation", func() {
			Context("computerTime", func() {
				It("fails if not valid time", func() {
					uploadObj["computerTime"] = "Tuesday 14th May, 2015"

					Expect(
						helper.ErrorIsExpected(
							upload.Build(uploadObj, helper.ErrorProcessing),
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
							upload.Build(uploadObj, helper.ErrorProcessing),
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
							upload.Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/computerTime",
								Detail: "Times need to be ISO 8601 format and not in the future given ''",
							}),
					).To(BeNil())
				})

			})
			Context("byUser", func() {

				It("is required", func() {
					delete(uploadObj, "byUser")
					Expect(
						helper.ErrorIsExpected(
							upload.Build(uploadObj, helper.ErrorProcessing),
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
							upload.Build(uploadObj, helper.ErrorProcessing),
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
							upload.Build(uploadObj, helper.ErrorProcessing),
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
							upload.Build(uploadObj, helper.ErrorProcessing),
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
							upload.Build(uploadObj, helper.ErrorProcessing),
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
							upload.Build(uploadObj, helper.ErrorProcessing),
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
							upload.Build(uploadObj, helper.ErrorProcessing),
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
							upload.Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/deviceSerialNumber",
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
							upload.Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/deviceManufacturers",
								Detail: "Must contain at least one manufacturer name given '<nil>'",
							}),
					).To(BeNil())

				})

				It("cannot be empty", func() {
					uploadObj["deviceManufacturers"] = []interface{}{}

					Expect(
						helper.ErrorIsExpected(
							upload.Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/deviceManufacturers/0",
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
							upload.Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/deviceTags",
								Detail: "Must be one of insulin-pump, cgm, bgm given '<nil>'",
							}),
					).To(BeNil())
				})

				It("cannot be empty", func() {
					uploadObj["deviceTags"] = []interface{}{""}

					Expect(
						helper.ErrorIsExpected(
							upload.Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/deviceTags/0",
								Detail: "Must be one of insulin-pump, cgm, bgm given '[]'",
							}),
					).To(BeNil())
				})

				It("cannot have any invalid entries", func() {
					uploadObj["deviceTags"] = []interface{}{"insulin-pump", "nope", "cgm"}

					Expect(
						helper.ErrorIsExpected(
							upload.Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/deviceTags/1",
								Detail: "Must be one of insulin-pump, cgm, bgm given '[insulin-pump nope cgm]'",
							}),
					).To(BeNil())
				})

				It("has to be in approved list", func() {
					uploadObj["deviceTags"] = []interface{}{"unknown"}

					Expect(
						helper.ErrorIsExpected(
							upload.Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/deviceTags/0",
								Detail: "Must be one of insulin-pump, cgm, bgm given '[unknown]'",
							}),
					).To(BeNil())
				})

				It("can be any of insulin-pump, cgm, bgm", func() {
					uploadObj["deviceTags"] = []interface{}{"insulin-pump", "cgm", "bgm"}
					Expect(helper.ValidDataType(upload.Build(uploadObj, helper.ErrorProcessing))).To(BeNil())
				})
			})
			Context("timeProcessing", func() {

				It("is required", func() {
					delete(uploadObj, "timeProcessing")

					Expect(
						helper.ErrorIsExpected(
							upload.Build(uploadObj, helper.ErrorProcessing),
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
							upload.Build(uploadObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/timeProcessing",
								Detail: "Must be one of across-the-board-timezone, utc-bootstrapping, none given ''",
							}),
					).To(BeNil())
				})

				It("can be across-the-board-timezone", func() {
					uploadObj["timeProcessing"] = "across-the-board-timezone"
					Expect(helper.ValidDataType(upload.Build(uploadObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("can be utc-bootstrapping", func() {
					uploadObj["timeProcessing"] = "utc-bootstrapping"
					Expect(helper.ValidDataType(upload.Build(uploadObj, helper.ErrorProcessing))).To(BeNil())
				})

				It("can be none", func() {
					uploadObj["timeProcessing"] = "none"
					Expect(helper.ValidDataType(upload.Build(uploadObj, helper.ErrorProcessing))).To(BeNil())
				})
			})
		})

	})
})
