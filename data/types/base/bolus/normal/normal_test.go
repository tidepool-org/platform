package normal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types/base/bolus"
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/service"
)

var _ = Describe("Normal", func() {
	var rawObject = testing.RawBaseObject()
	var meta = &bolus.Meta{
		Type:    "bolus",
		SubType: "normal",
	}

	BeforeEach(func() {
		rawObject["type"] = "bolus"
		rawObject["subType"] = "normal"
		rawObject["normal"] = 52.1
	})

	Context("normal", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is negative", rawObject, "normal", -0.1,
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(-0.1, 0.0, 100.0), "/normal", meta)},
			),
			Entry("is greater than 20", rawObject, "normal", 100.1,
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(100.1, 0.0, 100.0), "/normal", meta)},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is within bounds", rawObject, "normal", 25.5),
			Entry("is without decimal", rawObject, "normal", 50),
		)
	})
})
