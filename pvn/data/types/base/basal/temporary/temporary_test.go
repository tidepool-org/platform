package temporary_test

import (
	"github.com/tidepool-org/platform/pvn/data/types/base/testing"
	"github.com/tidepool-org/platform/pvn/data/validator"
	"github.com/tidepool-org/platform/service"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
)

var _ = Describe("Temporary Basal", func() {

	var rawObject = testing.RawBaseObject()

	BeforeEach(func() {

		rawObject["type"] = "basal"
		rawObject["deliveryType"] = "temporary"
		rawObject["duration"] = 0
		rawObject["rate"] = 5.5
		rawObject["percent"] = 1.1

	})

	Context("duration", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("negative", rawObject, "duration", -1,
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorIntegerNotInRange(-1, 0, 86400000), "/duration")},
			),
			Entry("greater than 86400000", rawObject, "duration", 86400001,
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorIntegerNotInRange(86400001, 0, 86400000), "/duration")},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("within bounds", rawObject, "duration", 2400),
		)

	})

	Context("rate", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("negative", rawObject, "rate", -0.1,
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorFloatNotInRange(-0.1, 0.0, 20.0), "/rate")},
			),
			Entry("greater than 20", rawObject, "rate", 20.1,
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorFloatNotInRange(20.1, 0.0, 20.0), "/rate")},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("within bounds", rawObject, "rate", 5.5),
			Entry("also without decimal", rawObject, "rate", 5),
		)

	})

	Context("percent", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("negative", rawObject, "percent", -0.1,
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorFloatNotInRange(-0.1, 0.0, 10.0), "/percent")},
			),
			Entry("greater than 10", rawObject, "percent", 10.1,
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorFloatNotInRange(10.1, 0.0, 10.0), "/percent")},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("within bounds", rawObject, "percent", 9.9),
			Entry("also without decimal", rawObject, "percent", 5),
		)

	})
})
