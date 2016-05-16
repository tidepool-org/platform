package scheduled_test

import (
	"github.com/tidepool-org/platform/pvn/data/types/base/testing"
	"github.com/tidepool-org/platform/pvn/data/validator"
	"github.com/tidepool-org/platform/service"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
)

var _ = Describe("Scheduled Basal", func() {

	var rawObject = testing.RawBaseObject()

	BeforeEach(func() {

		rawObject["type"] = "basal"
		rawObject["deliveryType"] = "scheduled"
		rawObject["scheduleName"] = "test"
		rawObject["rate"] = 1.0
		rawObject["duration"] = 0

	})

	Context("duration", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("negative", rawObject, "duration", -1, []*service.Error{validator.ErrorValueNotTrue()}),
			Entry("greater than 432000000", rawObject, "duration", 432000001, []*service.Error{validator.ErrorValueNotTrue()}),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("within bounds", rawObject, "duration", 2400),
		)

	})

	Context("rate", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("negative", rawObject, "rate", -0.1, []*service.Error{validator.ErrorValueNotTrue()}),
			Entry("greater than 20", rawObject, "rate", 20.1, []*service.Error{validator.ErrorValueNotTrue()}),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("within bounds", rawObject, "rate", 5.5),
			Entry("also without decimal", rawObject, "rate", 5),
		)

	})

	Context("scheduleName", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("one character", rawObject, "scheduleName", "a", []*service.Error{validator.ErrorValueNotTrue()}),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("more than one character", rawObject, "scheduleName", "ab"),
		)

	})
})
