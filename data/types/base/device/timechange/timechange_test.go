package timechange_test

import (
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/service"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
)

var _ = Describe("TimeChange Event", func() {

	var rawObject = testing.RawBaseObject()

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
				Entry("non zulu time", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "manual", "reasons": []string{"travel", "correction"}}),
			)

			DescribeTable("invalid when", testing.ExpectFieldNotValid,
				Entry("zulu time", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06Z", "to": "2016-05-04T07:21:31", "agent": "manual", "reasons": []string{"travel", "correction"}},
					[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorTimeNotValid("2016-05-04T08:18:06Z", "2006-01-02T15:04:05"), "/change/from")},
				),
				Entry("empty time", rawObject, "change",
					map[string]interface{}{"from": "", "to": "2016-05-04T07:21:31", "agent": "manual", "reasons": []string{"travel", "correction"}},
					[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorTimeNotValid("", "2006-01-02T15:04:05"), "/change/from")},
				),
			)

		})

		Context("to", func() {

			DescribeTable("valid when", testing.ExpectFieldIsValid,
				Entry("non zulu time", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "manual", "reasons": []string{"travel", "correction"}}),
			)

			DescribeTable("invalid when", testing.ExpectFieldNotValid,
				Entry("zulu time", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31Z", "agent": "manual", "reasons": []string{"travel", "correction"}},
					[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorTimeNotValid("2016-05-04T07:21:31Z", "2006-01-02T15:04:05"), "/change/to")},
				),
				Entry("empty time", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "", "agent": "manual", "reasons": []string{"travel", "correction"}},
					[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorTimeNotValid("", "2006-01-02T15:04:05"), "/change/to")},
				),
			)

		})

		Context("agent", func() {

			DescribeTable("valid when", testing.ExpectFieldIsValid,
				Entry("manual", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "manual", "reasons": []string{"travel", "correction"}}),
				Entry("automatic", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "automatic", "reasons": []string{"travel", "correction"}}),
			)

			DescribeTable("invalid when", testing.ExpectFieldNotValid,
				Entry("empty", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "", "reasons": []string{"travel", "correction"}},
					[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorStringNotOneOf("", []string{"manual", "automatic"}), "/change/agent")},
				),
				Entry("not predefined type", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "wrong", "reasons": []string{"travel", "correction"}},
					[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorStringNotOneOf("wrong", []string{"manual", "automatic"}), "/change/agent")},
				),
			)

		})

		Context("reasons", func() {

			DescribeTable("valid when", testing.ExpectFieldIsValid,
				Entry("travel", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "manual", "reasons": []string{"travel"}}),
				Entry("correction", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "automatic", "reasons": []string{"correction"}}),
				Entry("from_daylight_savings", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "automatic", "reasons": []string{"from_daylight_savings"}}),
				Entry("to_daylight_savings", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "automatic", "reasons": []string{"to_daylight_savings"}}),
				Entry("other", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "automatic", "reasons": []string{"other"}}),
				Entry("all allowed types", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "automatic", "reasons": []string{"from_daylight_savings", "to_daylight_savings", "travel", "correction", "other"}}),
			)

			DescribeTable("invalid when", testing.ExpectFieldNotValid,
				Entry("empty", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "manual", "reasons": []string{""}},
					[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorStringNotOneOf("", []string{"from_daylight_savings", "to_daylight_savings", "travel", "correction", "other"}), "/change/reasons/0")},
				),
				Entry("not predefined type", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "agent": "manual", "reasons": []string{"wrong"}},
					[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorStringNotOneOf("wrong", []string{"from_daylight_savings", "to_daylight_savings", "travel", "correction", "other"}), "/change/reasons/0")},
				),
			)

		})

		Context("timezone", func() {

			DescribeTable("valid when", testing.ExpectFieldIsValid,
				Entry("set", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "timezone": "US/Central", "agent": "manual", "reasons": []string{"travel"}}),
				Entry("empty", rawObject, "change",
					map[string]interface{}{"from": "2016-05-04T08:18:06", "to": "2016-05-04T07:21:31", "timezone": "", "agent": "manual", "reasons": []string{"travel"}}),
			)

		})
	})

})
