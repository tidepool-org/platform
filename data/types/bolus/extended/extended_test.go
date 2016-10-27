package extended_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types/bolus"
	"github.com/tidepool-org/platform/data/types/testing"
	"github.com/tidepool-org/platform/service"
)

func NewRawObject() map[string]interface{} {
	rawObject := testing.RawBaseObject()
	rawObject["type"] = "bolus"
	rawObject["subType"] = "square"
	rawObject["extended"] = 7.6
	rawObject["duration"] = 12600000
	return rawObject
}

func NewExpectedRawObject() map[string]interface{} {
	rawObject := NewRawObject()
	rawObject["expectedExtended"] = 8.9
	rawObject["expectedDuration"] = 14400000
	return rawObject
}

func NewMeta() interface{} {
	return &bolus.Meta{
		Type:    "bolus",
		SubType: "square",
	}
}

var _ = Describe("Extended", func() {
	Context("extended", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("does not exist", NewRawObject(), "extended", nil,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotExists(), "/extended", NewMeta())},
			),
			Entry("is not a number", NewRawObject(), "extended", "not-a-number",
				[]*service.Error{
					testing.ComposeError(service.ErrorTypeNotFloat("not-a-number"), "/extended", NewMeta()),
					testing.ComposeError(service.ErrorValueNotExists(), "/extended", NewMeta()),
				},
			),
			Entry("is less than lower limit", NewRawObject(), "extended", -0.1,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/extended", NewMeta())},
			),
			Entry("is greater than upper limit", NewRawObject(), "extended", 100.1,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(100.1, 0.0, 100.0), "/extended", NewMeta())},
			),
			Entry("is zero without expectedExtended", NewRawObject(), "extended", 0.0,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotExists(), "/expectedExtended", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is approaching lower limit", NewRawObject(), "extended", 0.01),
			Entry("is within lower and upper limit", NewRawObject(), "extended", 14.5),
			Entry("is at upper limit", NewRawObject(), "extended", 100.0),
			Entry("is without decimal", NewRawObject(), "extended", 14),
		)
	})

	Context("expectedExtended", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is not a number", NewExpectedRawObject(), "expectedExtended", "not-a-number",
				[]*service.Error{
					testing.ComposeError(service.ErrorTypeNotFloat("not-a-number"), "/expectedExtended", NewMeta()),
					testing.ComposeError(service.ErrorValueExists(), "/expectedDuration", NewMeta()),
				},
			),
			Entry("is less than extended", NewExpectedRawObject(), "expectedExtended", 7.5,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(7.5, 7.6, 100.0), "/expectedExtended", NewMeta())},
			),
			Entry("is greater than upper limit", NewExpectedRawObject(), "expectedExtended", 100.1,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(100.1, 7.6, 100.0), "/expectedExtended", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is approaching extended", NewExpectedRawObject(), "expectedExtended", 7.61),
			Entry("is at upper limit", NewExpectedRawObject(), "expectedExtended", 100.0),
			Entry("is without decimal", NewExpectedRawObject(), "expectedExtended", 14),
		)
	})

	Context("duration", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("does not exist", NewRawObject(), "duration", nil,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotExists(), "/duration", NewMeta())},
			),
			Entry("is not a number", NewRawObject(), "duration", "not-a-number",
				[]*service.Error{
					testing.ComposeError(service.ErrorTypeNotInteger("not-a-number"), "/duration", NewMeta()),
					testing.ComposeError(service.ErrorValueNotExists(), "/duration", NewMeta()),
				},
			),
			Entry("is less than lower limit", NewRawObject(), "duration", -1,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(-1, 0, 86400000), "/duration", NewMeta())},
			),
			Entry("is greater than upper limit", NewRawObject(), "duration", 86400001,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(86400001, 0, 86400000), "/duration", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is at lower limit", NewRawObject(), "duration", 0),
			Entry("is within lower and upper limit", NewRawObject(), "duration", 14400000),
			Entry("is at upper limit", NewRawObject(), "duration", 86400000),
		)
	})

	Context("expectedDuration", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("exists when expectedExtended does not exist", NewExpectedRawObject(), "expectedExtended", nil,
				[]*service.Error{testing.ComposeError(service.ErrorValueExists(), "/expectedDuration", NewMeta())},
			),
			Entry("is not a number", NewExpectedRawObject(), "expectedDuration", "not-a-number",
				[]*service.Error{
					testing.ComposeError(service.ErrorTypeNotInteger("not-a-number"), "/expectedDuration", NewMeta()),
					testing.ComposeError(service.ErrorValueNotExists(), "/expectedDuration", NewMeta()),
				},
			),
			Entry("is less than duration", NewExpectedRawObject(), "expectedDuration", 12599999,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(12599999, 12600000, 86400000), "/expectedDuration", NewMeta())},
			),
			Entry("is greater than upper limit", NewExpectedRawObject(), "expectedDuration", 86400001,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(86400001, 12600000, 86400000), "/expectedDuration", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is approaching duration", NewExpectedRawObject(), "expectedDuration", 12600001),
			Entry("is at upper limit", NewExpectedRawObject(), "expectedDuration", 86400000),
		)
	})
})
