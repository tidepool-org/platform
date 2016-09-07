package scheduled_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types/base/basal"
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/service"
)

func NewRawObject() map[string]interface{} {
	rawObject := testing.RawBaseObject()
	rawObject["type"] = "basal"
	rawObject["deliveryType"] = "scheduled"
	rawObject["scheduleName"] = "test"
	rawObject["rate"] = 1.0
	rawObject["duration"] = 0
	return rawObject
}

func NewMeta() interface{} {
	return &basal.Meta{
		Type:         "basal",
		DeliveryType: "scheduled",
	}
}

var _ = Describe("Scheduled", func() {
	Context("duration", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is negative", NewRawObject(), "duration", -1,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(-1, 0, 432000000), "/duration", NewMeta())},
			),
			Entry("is greater than 432000000", NewRawObject(), "duration", 432000001,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(432000001, 0, 432000000), "/duration", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is within bounds", NewRawObject(), "duration", 2400),
		)
	})

	Context("rate", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is negative", NewRawObject(), "rate", -0.1,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(-0.1, 0.0, 20.0), "/rate", NewMeta())},
			),
			Entry("is greater than 20", NewRawObject(), "rate", 20.1,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(20.1, 0.0, 20.0), "/rate", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is within bounds", NewRawObject(), "rate", 5.5),
			Entry("is without decimal", NewRawObject(), "rate", 5),
		)
	})

	Context("scheduleName", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is one character", NewRawObject(), "scheduleName", "a",
				[]*service.Error{testing.ComposeError(service.ErrorLengthNotGreaterThan(1, 1), "/scheduleName", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is more than one character", NewRawObject(), "scheduleName", "ab"),
		)
	})
})
