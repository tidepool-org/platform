package combination_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types/base/bolus"
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/service"
)

func NewRawObject() map[string]interface{} {
	rawObject := testing.RawBaseObject()
	rawObject["type"] = "bolus"
	rawObject["subType"] = "dual/square"
	rawObject["duration"] = 0
	rawObject["extended"] = 25.5
	rawObject["normal"] = 55.5
	return rawObject
}

func NewMeta() interface{} {
	return &bolus.Meta{
		Type:    "bolus",
		SubType: "dual/square",
	}
}

var _ = Describe("Combination", func() {
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

	Context("extended", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is negative", NewRawObject(), "extended", -0.1,
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(-0.1, 0.0, 100.0), "/extended", NewMeta())},
			),
			Entry("is greater than 100", NewRawObject(), "extended", 100.1,
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(100.1, 0.0, 100.0), "/extended", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is within bounds", NewRawObject(), "extended", 5.5),
			Entry("is without decimal", NewRawObject(), "extended", 5),
			Entry("is 0", NewRawObject(), "extended", 0.0),
			Entry("is 100", NewRawObject(), "extended", 100.0),
		)
	})

	Context("normal", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is negative", NewRawObject(), "normal", -0.1,
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(-0.1, 0.0, 100.0), "/normal", NewMeta())},
			),
			Entry("is greater than 100", NewRawObject(), "normal", 100.1,
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(100.1, 0.0, 100.0), "/normal", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is 0", NewRawObject(), "normal", 0),
			Entry("is 100", NewRawObject(), "normal", 100.0),
			Entry("is within bounds", NewRawObject(), "normal", 25.5),
			Entry("is without decimal", NewRawObject(), "normal", 50),
		)
	})
})
