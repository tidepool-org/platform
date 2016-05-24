package status_test

import (
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/service"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
)

var _ = Describe("Status Event", func() {

	var rawObject = testing.RawBaseObject()

	BeforeEach(func() {

		rawObject["type"] = "deviceEvent"
		rawObject["subType"] = "status"
		rawObject["duration"] = 0
		rawObject["status"] = "suspended"
		rawObject["reason"] = map[string]interface{}{"suspended": "manual"}

	})

	Context("duration", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("less than 0", rawObject, "duration", -1,
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/duration")},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("0", rawObject, "duration", 0),
			Entry("max of 999999999999999999", rawObject, "duration", 999999999999999999),
		)

	})

	Context("status", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("empty", rawObject, "status", "",
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorStringNotOneOf("", []string{"suspended"}), "/status")},
			),
			Entry("not one of the predefined types", rawObject, "status", "bad",
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorStringNotOneOf("bad", []string{"suspended"}), "/status")},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("suspended type", rawObject, "status", "suspended"),
		)

	})

	Context("reason", func() {

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("manual", rawObject, "reason", map[string]interface{}{"suspended": "manual"}),
			Entry("automatic", rawObject, "reason", map[string]interface{}{"suspended": "automatic"}),
		)

	})

})
