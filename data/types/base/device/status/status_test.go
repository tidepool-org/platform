package status_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types/base/device"
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/service"
)

var _ = Describe("Status", func() {
	var rawObject = testing.RawBaseObject()
	var meta = &device.Meta{
		Type:    "deviceEvent",
		SubType: "status",
	}

	BeforeEach(func() {
		rawObject["type"] = "deviceEvent"
		rawObject["subType"] = "status"
		rawObject["duration"] = 0
		rawObject["status"] = "suspended"
		rawObject["reason"] = map[string]interface{}{"suspended": "manual"}
	})

	Context("duration", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is less than 0", rawObject, "duration", -1,
				[]*service.Error{testing.ComposeError(validator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/duration", meta)},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is 0", rawObject, "duration", 0),
			Entry("is max of 999999999999999999", rawObject, "duration", 999999999999999999),
		)
	})

	Context("status", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is empty", rawObject, "status", "",
				[]*service.Error{testing.ComposeError(validator.ErrorStringNotOneOf("", []string{"suspended"}), "/status", meta)},
			),
			Entry("is not one of the predefined types", rawObject, "status", "bad",
				[]*service.Error{testing.ComposeError(validator.ErrorStringNotOneOf("bad", []string{"suspended"}), "/status", meta)},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is suspended type", rawObject, "status", "suspended"),
		)
	})

	Context("reason", func() {
		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is manual", rawObject, "reason", map[string]interface{}{"suspended": "manual"}),
			Entry("is automatic", rawObject, "reason", map[string]interface{}{"suspended": "automatic"}),
		)
	})
})
