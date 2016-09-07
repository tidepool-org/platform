package suspend_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types/base/basal"
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/service"
)

func NewRawObject() map[string]interface{} {
	rawObject := testing.RawBaseObject()
	rawObject["type"] = "basal"
	rawObject["deliveryType"] = "suspend"
	return rawObject
}

func NewMeta() interface{} {
	return &basal.Meta{
		Type:         "basal",
		DeliveryType: "suspend",
	}
}

var _ = Describe("Suspend", func() {
	Context("duration", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is negative", NewRawObject(), "duration", -1,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta())},
			),
			Entry("is greater than 604800000", NewRawObject(), "duration", 604800001,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is missing", NewRawObject(), "type", "basal"),
			Entry("is zero", NewRawObject(), "duration", 0),
			Entry("is within bounds", NewRawObject(), "duration", 86400000),
		)
	})
})
