package alarm_test

import (
	"github.com/tidepool-org/platform/pvn/data/types/base/testing"
	"github.com/tidepool-org/platform/pvn/data/validator"
	"github.com/tidepool-org/platform/service"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
)

var _ = Describe("Alarm Event", func() {

	var rawObject = testing.RawBaseObject()

	BeforeEach(func() {

		rawObject["type"] = "deviceEvent"
		rawObject["subType"] = "alarm"
		rawObject["alarmType"] = "other"
		rawObject["status"] = "OK"

	})

	Context("alarmType", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("empty", rawObject, "alarmType", "", []*service.Error{validator.ErrorValueNotTrue()}),
			Entry("not one of the predefined types", rawObject, "alarmType", "bad-robot", []*service.Error{validator.ErrorValueNotTrue()}),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("low_insulin type", rawObject, "alarmType", "low_insulin"),
			Entry("low_power type", rawObject, "alarmType", "low_power"),
			Entry("no_power type", rawObject, "alarmType", "no_power"),
			Entry("occlusion type", rawObject, "alarmType", "occlusion"),
			Entry("no_delivery type", rawObject, "alarmType", "no_delivery"),
			Entry("auto_off type", rawObject, "alarmType", "auto_off"),
			Entry("over_limit type", rawObject, "alarmType", "over_limit"),
			Entry("other type", rawObject, "alarmType", "other"),
		)

	})

	Context("status", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("one character", rawObject, "status", "x", []*service.Error{validator.ErrorValueNotTrue()}),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("more then one character", rawObject, "status", "xx"),
		)

	})

})
