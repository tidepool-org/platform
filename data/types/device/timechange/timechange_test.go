package timechange_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/data/types/testing"
	"github.com/tidepool-org/platform/service"
)

func NewRawObject() map[string]interface{} {
	rawObject := testing.RawBaseObject()
	rawObject["type"] = "deviceEvent"
	rawObject["subType"] = "timeChange"
	rawObject["change"] = map[string]interface{}{
		"from":  "2016-05-04T08:18:06",
		"to":    "2016-05-04T07:21:31",
		"agent": "manual",
	}
	return rawObject
}

func NewMeta() interface{} {
	return &device.Meta{
		Type:    "deviceEvent",
		SubType: "timeChange",
	}
}

var _ = Describe("Timechange", func() {
	Context("change", func() {
		Context("from", func() {
			DescribeTable("valid when", testing.ExpectFieldIsValid,
				Entry("is non zulu time", NewRawObject(), "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "manual"}),
			)

			DescribeTable("invalid when", testing.ExpectFieldNotValid,
				Entry("is zulu time", NewRawObject(), "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06Z", "to": "2016-05-04T07:21:31", "agent": "manual"},
					[]*service.Error{testing.ComposeError(service.ErrorValueTimeNotValid("2016-05-04T08:18:06Z", "2006-01-02T15:04:05"), "/change/from", NewMeta())},
				),
				Entry("is empty time", NewRawObject(), "change",
					map[string]interface{}{"from": "", "to": "2016-05-04T07:21:31", "agent": "manual"},
					[]*service.Error{testing.ComposeError(service.ErrorValueTimeNotValid("", "2006-01-02T15:04:05"), "/change/from", NewMeta())},
				),
			)
		})

		Context("to", func() {
			DescribeTable("valid when", testing.ExpectFieldIsValid,
				Entry("is non zulu time", NewRawObject(), "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "manual"}),
			)

			DescribeTable("invalid when", testing.ExpectFieldNotValid,
				Entry("is zulu time", NewRawObject(), "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31Z", "agent": "manual"},
					[]*service.Error{testing.ComposeError(service.ErrorValueTimeNotValid("2016-05-04T07:21:31Z", "2006-01-02T15:04:05"), "/change/to", NewMeta())},
				),
				Entry("is empty time", NewRawObject(), "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "", "agent": "manual"},
					[]*service.Error{testing.ComposeError(service.ErrorValueTimeNotValid("", "2006-01-02T15:04:05"), "/change/to", NewMeta())},
				),
			)
		})

		Context("agent", func() {
			DescribeTable("valid when", testing.ExpectFieldIsValid,
				Entry("is manual", NewRawObject(), "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "manual"}),
				Entry("is automatic", NewRawObject(), "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "automatic"}),
			)

			DescribeTable("invalid when", testing.ExpectFieldNotValid,
				Entry("is empty", NewRawObject(), "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": ""},
					[]*service.Error{testing.ComposeError(service.ErrorValueStringNotOneOf("", []string{"manual", "automatic"}), "/change/agent", NewMeta())},
				),
				Entry("is not predefined type", NewRawObject(), "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "wrong"},
					[]*service.Error{testing.ComposeError(service.ErrorValueStringNotOneOf("wrong", []string{"manual", "automatic"}), "/change/agent", NewMeta())},
				),
			)
		})

	})
})
