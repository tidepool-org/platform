package scheduled_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types/base/basal"
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/service"
)

var _ = Describe("Scheduled", func() {
	var rawObject = testing.RawBaseObject()
	var meta = &basal.Meta{
		Type:         "basal",
		DeliveryType: "scheduled",
	}

	BeforeEach(func() {
		rawObject["type"] = "basal"
		rawObject["deliveryType"] = "scheduled"
		rawObject["scheduleName"] = "test"
		rawObject["rate"] = 1.0
		rawObject["duration"] = 0
	})

	Context("duration", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is negative", rawObject, "duration", -1,
				[]*service.Error{testing.ComposeError(validator.ErrorIntegerNotInRange(-1, 0, 432000000), "/duration", meta)},
			),
			Entry("is greater than 432000000", rawObject, "duration", 432000001,
				[]*service.Error{testing.ComposeError(validator.ErrorIntegerNotInRange(432000001, 0, 432000000), "/duration", meta)},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is within bounds", rawObject, "duration", 2400),
		)
	})

	Context("rate", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is negative", rawObject, "rate", -0.1,
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(-0.1, 0.0, 20.0), "/rate", meta)},
			),
			Entry("is greater than 20", rawObject, "rate", 20.1,
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(20.1, 0.0, 20.0), "/rate", meta)},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is within bounds", rawObject, "rate", 5.5),
			Entry("is without decimal", rawObject, "rate", 5),
		)
	})

	Context("scheduleName", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is one character", rawObject, "scheduleName", "a",
				[]*service.Error{testing.ComposeError(validator.ErrorLengthNotGreaterThan(1, 1), "/scheduleName", meta)},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is more than one character", rawObject, "scheduleName", "ab"),
		)
	})
})
