package timechange_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types/base/device"
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/service"
)

func NewRawObject() map[string]interface{} {
	rawObject := testing.RawBaseObject()
	rawObject["type"] = "deviceEvent"
	rawObject["subType"] = "timeChange"
	rawObject["change"] = map[string]interface{}{
		"from":    "2016-05-04T08:18:06",
		"to":      "2016-05-04T07:21:31",
		"agent":   "manual",
		"reasons": []string{"travel", "correction"},
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
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "manual", "reasons": []string{"travel", "correction"}}),
			)

			DescribeTable("invalid when", testing.ExpectFieldNotValid,
				Entry("is zulu time", NewRawObject(), "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06Z", "to": "2016-05-04T07:21:31", "agent": "manual", "reasons": []string{"travel", "correction"}},
					[]*service.Error{testing.ComposeError(validator.ErrorTimeNotValid("2016-05-04T08:18:06Z", "2006-01-02T15:04:05"), "/change/from", NewMeta())},
				),
				Entry("is empty time", NewRawObject(), "change",
					map[string]interface{}{"from": "", "to": "2016-05-04T07:21:31", "agent": "manual", "reasons": []string{"travel", "correction"}},
					[]*service.Error{testing.ComposeError(validator.ErrorTimeNotValid("", "2006-01-02T15:04:05"), "/change/from", NewMeta())},
				),
			)
		})

		Context("to", func() {
			DescribeTable("valid when", testing.ExpectFieldIsValid,
				Entry("is non zulu time", NewRawObject(), "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "manual", "reasons": []string{"travel", "correction"}}),
			)

			DescribeTable("invalid when", testing.ExpectFieldNotValid,
				Entry("is zulu time", NewRawObject(), "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31Z", "agent": "manual", "reasons": []string{"travel", "correction"}},
					[]*service.Error{testing.ComposeError(validator.ErrorTimeNotValid("2016-05-04T07:21:31Z", "2006-01-02T15:04:05"), "/change/to", NewMeta())},
				),
				Entry("is empty time", NewRawObject(), "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "", "agent": "manual", "reasons": []string{"travel", "correction"}},
					[]*service.Error{testing.ComposeError(validator.ErrorTimeNotValid("", "2006-01-02T15:04:05"), "/change/to", NewMeta())},
				),
			)
		})

		Context("agent", func() {
			DescribeTable("valid when", testing.ExpectFieldIsValid,
				Entry("is manual", NewRawObject(), "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "manual", "reasons": []string{"travel", "correction"}}),
				Entry("is automatic", NewRawObject(), "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "automatic", "reasons": []string{"travel", "correction"}}),
			)

			DescribeTable("invalid when", testing.ExpectFieldNotValid,
				Entry("is empty", NewRawObject(), "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "", "reasons": []string{"travel", "correction"}},
					[]*service.Error{testing.ComposeError(validator.ErrorStringNotOneOf("", []string{"manual", "automatic"}), "/change/agent", NewMeta())},
				),
				Entry("is not predefined type", NewRawObject(), "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "wrong", "reasons": []string{"travel", "correction"}},
					[]*service.Error{testing.ComposeError(validator.ErrorStringNotOneOf("wrong", []string{"manual", "automatic"}), "/change/agent", NewMeta())},
				),
			)
		})

		Context("reasons", func() {
			DescribeTable("valid when", testing.ExpectFieldIsValid,
				Entry("is travel", NewRawObject(), "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "manual", "reasons": []string{"travel"}}),
				Entry("is correction", NewRawObject(), "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "automatic", "reasons": []string{"correction"}}),
				Entry("is from_daylight_savings", NewRawObject(), "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "automatic", "reasons": []string{"from_daylight_savings"}}),
				Entry("is to_daylight_savings", NewRawObject(), "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "automatic", "reasons": []string{"to_daylight_savings"}}),
				Entry("is other", NewRawObject(), "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "automatic", "reasons": []string{"other"}}),
				Entry("is all allowed types", NewRawObject(), "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "automatic", "reasons": []string{"from_daylight_savings", "to_daylight_savings", "travel", "correction", "other"}}),
			)

			DescribeTable("invalid when", testing.ExpectFieldNotValid,
				Entry("is empty", NewRawObject(), "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "manual", "reasons": []string{""}},
					[]*service.Error{testing.ComposeError(validator.ErrorStringNotOneOf("", []string{"from_daylight_savings", "to_daylight_savings", "travel", "correction", "other"}), "/change/reasons/0", NewMeta())},
				),
				Entry("is not predefined type", NewRawObject(), "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "manual", "reasons": []string{"wrong"}},
					[]*service.Error{testing.ComposeError(validator.ErrorStringNotOneOf("wrong", []string{"from_daylight_savings", "to_daylight_savings", "travel", "correction", "other"}), "/change/reasons/0", NewMeta())},
				),
			)
		})

		Context("timezone", func() {
			DescribeTable("valid when", testing.ExpectFieldIsValid,
				Entry("is set", NewRawObject(), "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "timezone": "US/Central", "agent": "manual", "reasons": []string{"travel"}}),
				Entry("is empty", NewRawObject(), "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "timezone": "", "agent": "manual", "reasons": []string{"travel"}}),
			)
		})
	})
})
