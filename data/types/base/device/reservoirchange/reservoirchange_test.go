package reservoirchange_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types/base/device"
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/service"
)

var _ = Describe("Reservoirchange", func() {
	var rawObject = testing.RawBaseObject()
	var meta = &device.Meta{
		Type:    "deviceEvent",
		SubType: "reservoirChange",
	}

	BeforeEach(func() {
		rawObject["type"] = "deviceEvent"
		rawObject["subType"] = "reservoirChange"
		rawObject["status"] = "some-id"
	})

	Context("status", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is empty", rawObject, "status", "",
				[]*service.Error{testing.ComposeError(validator.ErrorLengthNotGreaterThan(0, 1), "/status", meta)},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is longer than one character", rawObject, "status", "the-linked-status-id"),
		)
	})
})
