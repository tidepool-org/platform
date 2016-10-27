package combination_test

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
	rawObject["subType"] = "dual/square"
	rawObject["normal"] = 7.6
	rawObject["extended"] = 9.7
	rawObject["duration"] = 12600000
	return rawObject
}

func NewExpectedNormalRawObject() map[string]interface{} {
	rawObject := NewRawObject()
	rawObject["expectedNormal"] = 8.9
	rawObject["extended"] = 0
	rawObject["expectedExtended"] = 10.1
	rawObject["duration"] = 0
	rawObject["expectedDuration"] = 14400000
	return rawObject
}

func NewExpectedExtendedRawObject() map[string]interface{} {
	rawObject := NewRawObject()
	rawObject["expectedExtended"] = 10.1
	rawObject["expectedDuration"] = 14400000
	return rawObject
}

func NewMeta() interface{} {
	return &bolus.Meta{
		Type:    "bolus",
		SubType: "dual/square",
	}
}

var _ = Describe("Combination", func() {
	Context("normal", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("does not exist", NewRawObject(), "normal", nil,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotExists(), "/normal", NewMeta())},
			),
			Entry("is not a number", NewRawObject(), "normal", "not-a-number",
				[]*service.Error{
					testing.ComposeError(service.ErrorTypeNotFloat("not-a-number"), "/normal", NewMeta()),
					testing.ComposeError(service.ErrorValueNotExists(), "/normal", NewMeta()),
				},
			),
			Entry("is less than lower limit", NewRawObject(), "normal", -0.1,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/normal", NewMeta())},
			),
			Entry("is greater than upper limit", NewRawObject(), "normal", 100.1,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(100.1, 0.0, 100.0), "/normal", NewMeta())},
			),
			Entry("is zero without expectedNormal", NewRawObject(), "normal", 0.0,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotExists(), "/expectedNormal", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is approaching lower limit", NewRawObject(), "normal", 0.01),
			Entry("is within lower and upper limit", NewRawObject(), "normal", 14.5),
			Entry("is at upper limit", NewRawObject(), "normal", 100.0),
			Entry("is without decimal", NewRawObject(), "normal", 14),
		)
	})

	Context("expectedNormal", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is not a number", NewExpectedNormalRawObject(), "expectedNormal", "not-a-number",
				[]*service.Error{testing.ComposeError(service.ErrorTypeNotFloat("not-a-number"), "/expectedNormal", NewMeta())},
			),
			Entry("is less than normal", NewExpectedNormalRawObject(), "expectedNormal", 7.5,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(7.5, 7.6, 100.0), "/expectedNormal", NewMeta())},
			),
			Entry("is greater than upper limit", NewExpectedNormalRawObject(), "expectedNormal", 100.1,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(100.1, 7.6, 100.0), "/expectedNormal", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is approaching normal", NewExpectedNormalRawObject(), "expectedNormal", 9.71),
			Entry("is at upper limit", NewExpectedNormalRawObject(), "expectedNormal", 100.0),
			Entry("is without decimal", NewExpectedNormalRawObject(), "expectedNormal", 14),
		)
	})

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
			Entry("does not exist with expected normal", NewExpectedNormalRawObject(), "extended", nil,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotExists(), "/extended", NewMeta())},
			),
			Entry("is not a number with expected normal", NewExpectedNormalRawObject(), "extended", "not-a-number",
				[]*service.Error{
					testing.ComposeError(service.ErrorTypeNotFloat("not-a-number"), "/extended", NewMeta()),
					testing.ComposeError(service.ErrorValueNotExists(), "/extended", NewMeta()),
				},
			),
			Entry("is non-zero with expected normal", NewExpectedNormalRawObject(), "extended", 1.2,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotEqualTo(1.2, 0.0), "/extended", NewMeta())},
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
			Entry("is not a number", NewExpectedExtendedRawObject(), "expectedExtended", "not-a-number",
				[]*service.Error{
					testing.ComposeError(service.ErrorTypeNotFloat("not-a-number"), "/expectedExtended", NewMeta()),
					testing.ComposeError(service.ErrorValueExists(), "/expectedDuration", NewMeta()),
				},
			),
			Entry("is less than extended", NewExpectedExtendedRawObject(), "expectedExtended", 9.6,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(9.6, 9.7, 100.0), "/expectedExtended", NewMeta())},
			),
			Entry("is greater than upper limit", NewExpectedExtendedRawObject(), "expectedExtended", 100.1,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(100.1, 9.7, 100.0), "/expectedExtended", NewMeta())},
			),
			Entry("does not exist with expected normal", NewExpectedNormalRawObject(), "expectedExtended", nil,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotExists(), "/expectedExtended", NewMeta())},
			),
			Entry("is not a number with expected normal", NewExpectedNormalRawObject(), "expectedExtended", "not-a-number",
				[]*service.Error{
					testing.ComposeError(service.ErrorTypeNotFloat("not-a-number"), "/expectedExtended", NewMeta()),
					testing.ComposeError(service.ErrorValueNotExists(), "/expectedExtended", NewMeta()),
				},
			),
			Entry("is less than lower limit with expected normal", NewExpectedNormalRawObject(), "expectedExtended", -0.1,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/expectedExtended", NewMeta())},
			),
			Entry("is greater than upper limit with expected normal", NewExpectedNormalRawObject(), "expectedExtended", 100.1,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(100.1, 0.0, 100.0), "/expectedExtended", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is approaching extended", NewExpectedExtendedRawObject(), "expectedExtended", 9.71),
			Entry("is at upper limit", NewExpectedExtendedRawObject(), "expectedExtended", 100.0),
			Entry("is without decimal", NewExpectedExtendedRawObject(), "expectedExtended", 14),
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
			Entry("does not exist", NewExpectedNormalRawObject(), "duration", nil,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotExists(), "/duration", NewMeta())},
			),
			Entry("is not a number", NewExpectedNormalRawObject(), "duration", "not-a-number",
				[]*service.Error{
					testing.ComposeError(service.ErrorTypeNotInteger("not-a-number"), "/duration", NewMeta()),
					testing.ComposeError(service.ErrorValueNotExists(), "/duration", NewMeta()),
				},
			),
			Entry("is non-zero with expected normal", NewExpectedNormalRawObject(), "duration", 1,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotEqualTo(1, 0), "/duration", NewMeta())},
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
			Entry("exists when expectedExtended does not exist", NewExpectedExtendedRawObject(), "expectedExtended", nil,
				[]*service.Error{testing.ComposeError(service.ErrorValueExists(), "/expectedDuration", NewMeta())},
			),
			Entry("is not a number", NewExpectedExtendedRawObject(), "expectedDuration", "not-a-number",
				[]*service.Error{
					testing.ComposeError(service.ErrorTypeNotInteger("not-a-number"), "/expectedDuration", NewMeta()),
					testing.ComposeError(service.ErrorValueNotExists(), "/expectedDuration", NewMeta()),
				},
			),
			Entry("is less than duration", NewExpectedExtendedRawObject(), "expectedDuration", 12599999,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(12599999, 12600000, 86400000), "/expectedDuration", NewMeta())},
			),
			Entry("is greater than upper limit", NewExpectedExtendedRawObject(), "expectedDuration", 86400001,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(86400001, 12600000, 86400000), "/expectedDuration", NewMeta())},
			),
			Entry("does not exist with expected normal", NewExpectedNormalRawObject(), "expectedDuration", nil,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotExists(), "/expectedDuration", NewMeta())},
			),
			Entry("is not a number with expected normal", NewExpectedNormalRawObject(), "expectedDuration", "not-a-number",
				[]*service.Error{
					testing.ComposeError(service.ErrorTypeNotInteger("not-a-number"), "/expectedDuration", NewMeta()),
					testing.ComposeError(service.ErrorValueNotExists(), "/expectedDuration", NewMeta()),
				},
			),
			Entry("is less than lower limit with expected normal", NewExpectedNormalRawObject(), "expectedDuration", -1,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(-1, 0, 86400000), "/expectedDuration", NewMeta())},
			),
			Entry("is greater than upper limit with expected normal", NewExpectedNormalRawObject(), "expectedDuration", 86400001,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(86400001, 0, 86400000), "/expectedDuration", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is approaching duration", NewExpectedExtendedRawObject(), "expectedDuration", 12600001),
			Entry("is at upper limit", NewExpectedExtendedRawObject(), "expectedDuration", 86400000),
		)
	})
})
