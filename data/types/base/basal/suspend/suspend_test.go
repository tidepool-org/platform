package suspend_test

import (
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/service"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
)

var _ = Describe("Suspend Basal", func() {

	var rawObject = testing.RawBaseObject()

	BeforeEach(func() {

		rawObject["type"] = "basal"
		rawObject["deliveryType"] = "suspend"
		rawObject["duration"] = 0

	})

	Context("duration", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("negative", rawObject, "duration", -1,
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorIntegerNotInRange(-1, 0, 604800000), "/duration")},
			),
			Entry("greater than 604800000", rawObject, "duration", 604800001,
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorIntegerNotInRange(604800001, 0, 604800000), "/duration")},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("within bounds", rawObject, "duration", 86400000),
		)

	})
})
