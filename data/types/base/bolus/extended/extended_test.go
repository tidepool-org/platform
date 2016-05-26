package extended_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types/base/bolus"
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/service"
)

var _ = Describe("Extended", func() {
	var rawObject = testing.RawBaseObject()
	var meta = &bolus.Meta{
		Type:    "bolus",
		SubType: "square",
	}

	BeforeEach(func() {
		rawObject["type"] = "bolus"
		rawObject["subType"] = "square"
		rawObject["duration"] = 0
		rawObject["extended"] = 25.5
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

	Context("extended", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is negative", rawObject, "extended", -0.1,
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(-0.1, 0.0, 100.0), "/extended", meta)},
			),
			Entry("is greater than 100", rawObject, "extended", 100.1,
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(100.1, 0.0, 100.0), "/extended", meta)},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is within bounds", rawObject, "extended", 5.5),
			Entry("is without decimal", rawObject, "extended", 5),
		)
	})
})
