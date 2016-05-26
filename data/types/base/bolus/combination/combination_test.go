package combination_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types/base/bolus"
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/service"
)

var _ = Describe("Combination", func() {
	var rawObject = testing.RawBaseObject()
	var meta = &bolus.Meta{
		Type:    "bolus",
		SubType: "dual/square",
	}

	BeforeEach(func() {
		rawObject["type"] = "bolus"
		rawObject["subType"] = "dual/square"
		rawObject["duration"] = 0
		rawObject["extended"] = 25.5
		rawObject["normal"] = 55.5
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
			Entry("is 0", rawObject, "extended", 0.0),
			Entry("is 100", rawObject, "extended", 100.0),
		)
	})

	Context("normal", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is negative", rawObject, "normal", -0.1,
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(-0.1, 0.0, 100.0), "/normal", meta)},
			),
			Entry("is greater than 100", rawObject, "normal", 100.1,
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(100.1, 0.0, 100.0), "/normal", meta)},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is 0", rawObject, "normal", 0),
			Entry("is 100", rawObject, "normal", 100.0),
			Entry("is within bounds", rawObject, "normal", 25.5),
			Entry("is without decimal", rawObject, "normal", 50),
		)
	})
})
