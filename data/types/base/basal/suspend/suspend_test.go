package suspend_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types/base/basal"
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/service"
)

var _ = Describe("Suspend", func() {
	var rawObject = testing.RawBaseObject()
	var meta = &basal.Meta{
		Type:         "basal",
		DeliveryType: "suspend",
	}

	BeforeEach(func() {
		rawObject["type"] = "basal"
		rawObject["deliveryType"] = "suspend"
		rawObject["duration"] = 0
	})

	Context("duration", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is negative", rawObject, "duration", -1,
				[]*service.Error{testing.ComposeError(validator.ErrorIntegerNotInRange(-1, 0, 604800000), "/duration", meta)},
			),
			Entry("is greater than 604800000", rawObject, "duration", 604800001,
				[]*service.Error{testing.ComposeError(validator.ErrorIntegerNotInRange(604800001, 0, 604800000), "/duration", meta)},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is within bounds", rawObject, "duration", 86400000),
		)
	})
})
