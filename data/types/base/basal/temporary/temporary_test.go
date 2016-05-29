package temporary_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types/base/basal"
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/service"
)

func NewRawObject() map[string]interface{} {
	rawObject := testing.RawBaseObject()
	rawObject["type"] = "basal"
	rawObject["deliveryType"] = "temp"
	rawObject["duration"] = 0
	rawObject["rate"] = 5.5
	rawObject["percent"] = 1.1
	return rawObject
}

func NewMeta() interface{} {
	return &basal.Meta{
		Type:         "basal",
		DeliveryType: "temp",
	}
}

var _ = Describe("Temporary", func() {
	Context("duration", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is negative", NewRawObject(), "duration", -1,
				[]*service.Error{testing.ComposeError(validator.ErrorIntegerNotInRange(-1, 0, 86400000), "/duration", NewMeta())},
			),
			Entry("is greater than 86400000", NewRawObject(), "duration", 86400001,
				[]*service.Error{testing.ComposeError(validator.ErrorIntegerNotInRange(86400001, 0, 86400000), "/duration", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is within bounds", NewRawObject(), "duration", 2400),
		)
	})

	Context("rate", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is negative", NewRawObject(), "rate", -0.1,
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(-0.1, 0.0, 20.0), "/rate", NewMeta())},
			),
			Entry("is greater than 20", NewRawObject(), "rate", 20.1,
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(20.1, 0.0, 20.0), "/rate", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is within bounds", NewRawObject(), "rate", 5.5),
			Entry("is without decimal", NewRawObject(), "rate", 5),
		)
	})

	Context("percent", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is negative", NewRawObject(), "percent", -0.1,
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(-0.1, 0.0, 10.0), "/percent", NewMeta())},
			),
			Entry("is greater than 10", NewRawObject(), "percent", 10.1,
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(10.1, 0.0, 10.0), "/percent", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is within bounds", NewRawObject(), "percent", 9.9),
			Entry("is without decimal", NewRawObject(), "percent", 5),
		)
	})
})
