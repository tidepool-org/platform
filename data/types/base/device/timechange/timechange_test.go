package timechange_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types/base/device"
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/service"
)

var _ = Describe("Timechange", func() {
	var rawObject = testing.RawBaseObject()
	var meta = &device.Meta{
		Type:    "deviceEvent",
		SubType: "timeChange",
	}

	BeforeEach(func() {
		rawObject["type"] = "deviceEvent"
		rawObject["subType"] = "timeChange"
		rawObject["change"] = map[string]interface{}{
			"from":    "2016-05-04T08:18:06",
			"to":      "2016-05-04T07:21:31",
			"agent":   "manual",
			"reasons": []string{"travel", "correction"},
		}
	})

	Context("change", func() {
		Context("from", func() {
			DescribeTable("valid when", testing.ExpectFieldIsValid,
				Entry("is non zulu time", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "manual", "reasons": []string{"travel", "correction"}}),
			)

			DescribeTable("invalid when", testing.ExpectFieldNotValid,
				Entry("is zulu time", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06Z", "to": "2016-05-04T07:21:31", "agent": "manual", "reasons": []string{"travel", "correction"}},
					[]*service.Error{testing.ComposeError(validator.ErrorTimeNotValid("2016-05-04T08:18:06Z", "2006-01-02T15:04:05"), "/change/from", meta)},
				),
				Entry("is empty time", rawObject, "change",
					map[string]interface{}{"from": "", "to": "2016-05-04T07:21:31", "agent": "manual", "reasons": []string{"travel", "correction"}},
					[]*service.Error{testing.ComposeError(validator.ErrorTimeNotValid("", "2006-01-02T15:04:05"), "/change/from", meta)},
				),
			)
		})

		Context("to", func() {
			DescribeTable("valid when", testing.ExpectFieldIsValid,
				Entry("is non zulu time", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "manual", "reasons": []string{"travel", "correction"}}),
			)

			DescribeTable("invalid when", testing.ExpectFieldNotValid,
				Entry("is zulu time", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31Z", "agent": "manual", "reasons": []string{"travel", "correction"}},
					[]*service.Error{testing.ComposeError(validator.ErrorTimeNotValid("2016-05-04T07:21:31Z", "2006-01-02T15:04:05"), "/change/to", meta)},
				),
				Entry("is empty time", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "", "agent": "manual", "reasons": []string{"travel", "correction"}},
					[]*service.Error{testing.ComposeError(validator.ErrorTimeNotValid("", "2006-01-02T15:04:05"), "/change/to", meta)},
				),
			)
		})

		Context("agent", func() {
			DescribeTable("valid when", testing.ExpectFieldIsValid,
				Entry("is manual", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "manual", "reasons": []string{"travel", "correction"}}),
				Entry("is automatic", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "automatic", "reasons": []string{"travel", "correction"}}),
			)

			DescribeTable("invalid when", testing.ExpectFieldNotValid,
				Entry("is empty", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "", "reasons": []string{"travel", "correction"}},
					[]*service.Error{testing.ComposeError(validator.ErrorStringNotOneOf("", []string{"manual", "automatic"}), "/change/agent", meta)},
				),
				Entry("is not predefined type", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "wrong", "reasons": []string{"travel", "correction"}},
					[]*service.Error{testing.ComposeError(validator.ErrorStringNotOneOf("wrong", []string{"manual", "automatic"}), "/change/agent", meta)},
				),
			)
		})

		Context("reasons", func() {
			DescribeTable("valid when", testing.ExpectFieldIsValid,
				Entry("is travel", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "manual", "reasons": []string{"travel"}}),
				Entry("is correction", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "automatic", "reasons": []string{"correction"}}),
				Entry("is from_daylight_savings", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "automatic", "reasons": []string{"from_daylight_savings"}}),
				Entry("is to_daylight_savings", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "automatic", "reasons": []string{"to_daylight_savings"}}),
				Entry("is other", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "automatic", "reasons": []string{"other"}}),
				Entry("is all allowed types", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "automatic", "reasons": []string{"from_daylight_savings", "to_daylight_savings", "travel", "correction", "other"}}),
			)

			DescribeTable("invalid when", testing.ExpectFieldNotValid,
				Entry("is empty", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "manual", "reasons": []string{""}},
					[]*service.Error{testing.ComposeError(validator.ErrorStringNotOneOf("", []string{"from_daylight_savings", "to_daylight_savings", "travel", "correction", "other"}), "/change/reasons/0", meta)},
				),
				Entry("is not predefined type", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "manual", "reasons": []string{"wrong"}},
					[]*service.Error{testing.ComposeError(validator.ErrorStringNotOneOf("wrong", []string{"from_daylight_savings", "to_daylight_savings", "travel", "correction", "other"}), "/change/reasons/0", meta)},
				),
			)
		})

		Context("timezone", func() {
			DescribeTable("valid when", testing.ExpectFieldIsValid,
				Entry("is set", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "timezone": "US/Central", "agent": "manual", "reasons": []string{"travel"}}),
				Entry("is empty", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "timezone": "", "agent": "manual", "reasons": []string{"travel"}}),
			)
		})
	})
})
