package temporary_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types/base/basal"
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/service"
)

var _ = Describe("Temporary", func() {
	var rawObject = testing.RawBaseObject()
	var meta = &basal.Meta{
		Type:         "basal",
		DeliveryType: "temporary",
	}

	BeforeEach(func() {
		rawObject["type"] = "basal"
		rawObject["deliveryType"] = "temporary"
		rawObject["duration"] = 0
		rawObject["rate"] = 5.5
		rawObject["percent"] = 1.1
	})

	Context("duration", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is negative", rawObject, "duration", -1,
				[]*service.Error{testing.ComposeError(validator.ErrorIntegerNotInRange(-1, 0, 86400000), "/duration", meta)},
			),
			Entry("is greater than 86400000", rawObject, "duration", 86400001,
				[]*service.Error{testing.ComposeError(validator.ErrorIntegerNotInRange(86400001, 0, 86400000), "/duration", meta)},
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

	Context("percent", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is negative", rawObject, "percent", -0.1,
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(-0.1, 0.0, 10.0), "/percent", meta)},
			),
			Entry("is greater than 10", rawObject, "percent", 10.1,
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(10.1, 0.0, 10.0), "/percent", meta)},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is within bounds", rawObject, "percent", 9.9),
			Entry("is without decimal", rawObject, "percent", 5),
		)
	})
})
