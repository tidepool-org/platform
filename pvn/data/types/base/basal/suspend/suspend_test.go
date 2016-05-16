package suspend_test

import (
	"github.com/tidepool-org/platform/pvn/data/types/base/testing"
	"github.com/tidepool-org/platform/pvn/data/validator"
	"github.com/tidepool-org/platform/service"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
)

var _ = Describe("Susepnd Basal", func() {

	var rawObject = testing.RawBaseObject()

	BeforeEach(func() {

		rawObject["type"] = "basal"
		rawObject["deliveryType"] = "suspend"
		rawObject["duration"] = 0

	})

	Context("duration", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("negative", rawObject, "duration", -1, []*service.Error{validator.ErrorValueNotTrue()}),
			Entry("greater than 86400000", rawObject, "duration", 86400001, []*service.Error{validator.ErrorValueNotTrue()}),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("within bounds", rawObject, "duration", 2400),
		)

	})
})
