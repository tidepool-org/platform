package types_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

var _ = Describe("Time Validation", func() {

	type TimeStruct struct {
		//e.g. `deviceTime`
		NonZuluTimeString string `json:"nonZuluTimeString" bson:"nonZuluTimeString" valid:"nonZuluTimeString"`
		//e.g. `createdTime`
		ZuluTimeString   string `json:"zuluTimeString" bson:"zuluTimeString" valid:"zuluTimeString"`
		OffsetTimeString string `json:"offsetTimeString" bson:"offsetTimeString" valid:"offsetTimeString"`
		//e.g. `time`
		OffsetOrZuluTimeString string `json:"offsetOrZuluTimeString" bson:"offsetOrZuluTimeString" valid:"offsetOrZuluTimeString"`
	}

	var (
		helper *types.TestingHelper

		timeStruct TimeStruct

		failureReasons = validate.FailureReasons{
			"NonZuluTimeString":      validate.ValidationInfo{FieldName: "nonZuluTimeString", Message: types.NonZuluTimeStringField.Message},
			"ZuluTimeString":         validate.ValidationInfo{FieldName: "zuluTimeString", Message: types.ZuluTimeStringField.Message},
			"OffsetTimeString":       validate.ValidationInfo{FieldName: "offsetTimeString", Message: types.OffsetTimeStringField.Message},
			"OffsetOrZuluTimeString": validate.ValidationInfo{FieldName: "offsetOrZuluTimeString", Message: types.OffsetOrZuluTimeStringField.Message},
		}
	)

	BeforeEach(func() {
		helper = types.NewTestingHelper()
		timeStruct = TimeStruct{
			NonZuluTimeString:      "2013-05-04T03:58:44.584",
			ZuluTimeString:         "2013-05-04T03:58:44.584Z",
			OffsetTimeString:       "2013-05-04T03:58:44-08:00",
			OffsetOrZuluTimeString: "2013-05-04T03:58:44-08:00",
		}
	})

	Context("ZuluTimeStringField", func() {

		It("valid when matches example", func() {
			timeStruct.ZuluTimeString = "2013-05-04T03:58:44.584Z"
			types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(timeStruct, helper.ErrorProcessing)
			Expect(helper.ValidDataType(timeStruct)).To(BeNil())
		})

		It("invalid if before 2007-01-01T00:00:00Z", func() {
			timeStruct.ZuluTimeString = "2006-01-01T00:00:00Z"
			types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(timeStruct, helper.ErrorProcessing)
			Expect(
				helper.ErrorIsExpected(
					timeStruct,
					types.ExpectedErrorDetails{
						Path:   "0/zuluTimeString",
						Detail: "An ISO 8601-formatted UTC timestamp with a final Z for 'Zulu' time e.g 2013-05-04T03:58:44.584Z given '2006-01-01T00:00:00Z'",
					}),
			).To(BeNil())
		})

		It("invalid if in the future", func() {

			future := time.Now().AddDate(0, 0, 10)

			timeStruct.ZuluTimeString = future.Format(types.ZuluTimeStringField.Format)
			types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(timeStruct, helper.ErrorProcessing)
			Expect(
				helper.ErrorIsExpected(
					timeStruct,
					types.ExpectedErrorDetails{
						Path:   "0/zuluTimeString",
						Detail: fmt.Sprintf("An ISO 8601-formatted UTC timestamp with a final Z for 'Zulu' time e.g 2013-05-04T03:58:44.584Z given '%s'", timeStruct.ZuluTimeString),
					}),
			).To(BeNil())
		})

		It("invalid if wrong format", func() {
			timeStruct.ZuluTimeString = "02 Jan 2013 15:04 -0700"
			types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(timeStruct, helper.ErrorProcessing)
			Expect(
				helper.ErrorIsExpected(
					timeStruct,
					types.ExpectedErrorDetails{
						Path:   "0/zuluTimeString",
						Detail: "An ISO 8601-formatted UTC timestamp with a final Z for 'Zulu' time e.g 2013-05-04T03:58:44.584Z given '02 Jan 2013 15:04 -0700'",
					}),
			).To(BeNil())

		})

		It("invalid non zulu", func() {
			timeStruct.ZuluTimeString = "2013-05-04T03:58:44.584"
			types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(timeStruct, helper.ErrorProcessing)
			Expect(
				helper.ErrorIsExpected(
					timeStruct,
					types.ExpectedErrorDetails{
						Path:   "0/zuluTimeString",
						Detail: "An ISO 8601-formatted UTC timestamp with a final Z for 'Zulu' time e.g 2013-05-04T03:58:44.584Z given '2013-05-04T03:58:44.584'",
					}),
			).To(BeNil())

		})

	})

	Context("NonZuluTimeStringField", func() {

		It("valid when matches example", func() {
			timeStruct.NonZuluTimeString = "2013-05-04T03:58:44.584"
			types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(timeStruct, helper.ErrorProcessing)
			Expect(helper.ValidDataType(timeStruct)).To(BeNil())
		})

		It("invalid if before 2007-01-01T00:00:00", func() {
			timeStruct.NonZuluTimeString = "2006-01-01T00:00:00"
			types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(timeStruct, helper.ErrorProcessing)
			Expect(
				helper.ErrorIsExpected(
					timeStruct,
					types.ExpectedErrorDetails{
						Path:   "0/nonZuluTimeString",
						Detail: "An ISO 8601 formatted timestamp without any timezone offset information e.g 2013-05-04T03:58:44.584 given '2006-01-01T00:00:00'",
					}),
			).To(BeNil())

		})

		It("invalid if in the future", func() {

			future := time.Now().AddDate(0, 0, 10)

			timeStruct.NonZuluTimeString = future.Format(types.NonZuluTimeStringField.Format)
			types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(timeStruct, helper.ErrorProcessing)
			Expect(
				helper.ErrorIsExpected(
					timeStruct,
					types.ExpectedErrorDetails{
						Path:   "0/nonZuluTimeString",
						Detail: fmt.Sprintf("An ISO 8601 formatted timestamp without any timezone offset information e.g 2013-05-04T03:58:44.584 given '%s'", timeStruct.NonZuluTimeString),
					}),
			).To(BeNil())
		})

		It("invalid if wrong format", func() {
			timeStruct.NonZuluTimeString = "2013-05-04T03:58:44.584Z"
			types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(timeStruct, helper.ErrorProcessing)
			Expect(
				helper.ErrorIsExpected(
					timeStruct,
					types.ExpectedErrorDetails{
						Path:   "0/nonZuluTimeString",
						Detail: "An ISO 8601 formatted timestamp without any timezone offset information e.g 2013-05-04T03:58:44.584 given '2013-05-04T03:58:44.584Z'",
					}),
			).To(BeNil())

		})

	})

	Context("OffsetTimeString", func() {

		It("valid when matches example", func() {
			timeStruct.OffsetTimeString = "2013-05-04T03:58:44-08:00"
			types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(timeStruct, helper.ErrorProcessing)
			Expect(helper.ValidDataType(timeStruct)).To(BeNil())
		})

		It("invalid if before 2007-01-01T00:00:00", func() {
			timeStruct.OffsetTimeString = "2006-01-01T00:00:00-08:00"
			types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(timeStruct, helper.ErrorProcessing)
			Expect(
				helper.ErrorIsExpected(
					timeStruct,
					types.ExpectedErrorDetails{
						Path:   "0/offsetTimeString",
						Detail: "An ISO 8601-formatted timestamp including a timezone offset from UTC e.g 2013-05-04T03:58:44-08:00 given '2006-01-01T00:00:00-08:00'",
					}),
			).To(BeNil())

		})

		It("invalid if in the future", func() {

			future := time.Now().AddDate(0, 0, 10)

			timeStruct.OffsetTimeString = future.Format(types.OffsetTimeStringField.Format)
			types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(timeStruct, helper.ErrorProcessing)
			Expect(
				helper.ErrorIsExpected(
					timeStruct,
					types.ExpectedErrorDetails{
						Path:   "0/offsetTimeString",
						Detail: fmt.Sprintf("An ISO 8601-formatted timestamp including a timezone offset from UTC e.g 2013-05-04T03:58:44-08:00 given '%s'", timeStruct.OffsetTimeString),
					}),
			).To(BeNil())
		})

		It("invalid if wrong format", func() {
			timeStruct.OffsetTimeString = "2013-05-04T03:58:44.584Z"
			types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(timeStruct, helper.ErrorProcessing)
			Expect(
				helper.ErrorIsExpected(
					timeStruct,
					types.ExpectedErrorDetails{
						Path:   "0/offsetTimeString",
						Detail: "An ISO 8601-formatted timestamp including a timezone offset from UTC e.g 2013-05-04T03:58:44-08:00 given '2013-05-04T03:58:44.584Z'",
					}),
			).To(BeNil())
		})

	})

	Context("OffsetOrZuluTimeStringField", func() {

		It("valid when matches zulu example", func() {
			timeStruct.OffsetOrZuluTimeString = "2013-05-04T03:58:44.584Z"
			types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(timeStruct, helper.ErrorProcessing)
			Expect(helper.ValidDataType(timeStruct)).To(BeNil())
		})

		It("valid when matches offset example", func() {
			timeStruct.OffsetOrZuluTimeString = "2013-05-04T03:58:44-08:00"
			types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(timeStruct, helper.ErrorProcessing)
			Expect(helper.ValidDataType(timeStruct)).To(BeNil())
		})

	})

})
