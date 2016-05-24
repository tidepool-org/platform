package reservoirchange_test

import (
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/service"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
)

var _ = Describe("ReservoirChange Event", func() {

	var rawObject = testing.RawBaseObject()

	BeforeEach(func() {

		rawObject["type"] = "deviceEvent"
		rawObject["subType"] = "reservoirChange"
		rawObject["status"] = "some-id"

	})

	Context("status", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("empty", rawObject, "status", "",
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorLengthNotGreaterThan(0, 1), "/status")},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("longer than one character", rawObject, "status", "the-linked-status-id"),
		)

	})

})
