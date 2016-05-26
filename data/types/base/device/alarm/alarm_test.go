package alarm_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types/base/device"
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/service"
)

var _ = Describe("Alarm", func() {
	var rawObject = testing.RawBaseObject()
	var meta = &device.Meta{
		Type:    "deviceEvent",
		SubType: "alarm",
	}

	BeforeEach(func() {
		rawObject["type"] = "deviceEvent"
		rawObject["subType"] = "alarm"
		rawObject["alarmType"] = "other"
		rawObject["status"] = "OK"
	})

	Context("alarmType", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is empty", rawObject, "alarmType", "",
				[]*service.Error{testing.ComposeError(validator.ErrorStringNotOneOf("", []string{"low_insulin", "no_insulin", "low_power", "no_power", "occlusion", "no_delivery", "auto_off", "over_limit", "other"}), "/alarmType", meta)},
			),
			Entry("is not one of the predefined types", rawObject, "alarmType", "bad-robot",
				[]*service.Error{testing.ComposeError(validator.ErrorStringNotOneOf("bad-robot", []string{"low_insulin", "no_insulin", "low_power", "no_power", "occlusion", "no_delivery", "auto_off", "over_limit", "other"}), "/alarmType", meta)},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is low_insulin type", rawObject, "alarmType", "low_insulin"),
			Entry("is low_power type", rawObject, "alarmType", "low_power"),
			Entry("is no_power type", rawObject, "alarmType", "no_power"),
			Entry("is occlusion type", rawObject, "alarmType", "occlusion"),
			Entry("is no_delivery type", rawObject, "alarmType", "no_delivery"),
			Entry("is auto_off type", rawObject, "alarmType", "auto_off"),
			Entry("is over_limit type", rawObject, "alarmType", "over_limit"),
			Entry("is other type", rawObject, "alarmType", "other"),
		)
	})

	Context("status", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is one character", rawObject, "status", "x",
				[]*service.Error{testing.ComposeError(validator.ErrorLengthNotGreaterThan(1, 1), "/status", meta)},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is more then one character", rawObject, "status", "xx"),
		)
	})
})
