package reservoirchange_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/device"
)

func NewRawStatusObject() map[string]interface{} {
	rawStatusObject := testData.RawBaseObject()
	rawStatusObject["type"] = "deviceEvent"
	rawStatusObject["subType"] = "status"
	rawStatusObject["status"] = "suspended"
	rawStatusObject["reason"] = map[string]interface{}{}
	return rawStatusObject
}

func NewRawObject() map[string]interface{} {
	rawObject := testData.RawBaseObject()
	rawObject["type"] = "deviceEvent"
	rawObject["subType"] = "reservoirChange"
	rawObject["status"] = NewRawStatusObject()
	return rawObject
}

func NewMeta() interface{} {
	return &device.Meta{
		Type:    "deviceEvent",
		SubType: "reservoirChange",
	}

}

var _ = Describe("Reservoirchange", func() {
	Context("status", func() {
		DescribeTable("valid when", testData.ExpectFieldIsValid,
			Entry("is longer than one character", NewRawObject(), "status", NewRawStatusObject()),
		)
	})
})
