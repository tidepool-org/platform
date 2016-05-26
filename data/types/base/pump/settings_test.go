package pump_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/base"
	"github.com/tidepool-org/platform/data/types/base/pump"
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/data/types/common/bloodglucose"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/service"
)

var _ = Describe("Settings", func() {
	var rawObject = testing.RawBaseObject()
	var meta = &base.Meta{
		Type: "pumpSettings",
	}

	BeforeEach(func() {
		rawObject["type"] = "pumpSettings"
		rawObject["activeSchedule"] = "standard"

		rawObject["units"] = map[string]interface{}{
			"carb": "grams",
			"bg":   "mmol/L",
		}

		rawObject["carbRatio"] = []interface{}{
			map[string]interface{}{"amount": 12, "start": 0},
			map[string]interface{}{"amount": 10, "start": 21600000},
		}

		rawObject["bgTarget"] = []interface{}{
			map[string]interface{}{"low": 5.5, "high": 6.7, "start": 0},
			map[string]interface{}{"low": 5.0, "high": 6.1, "start": 18000000},
		}

		rawObject["insulinSensitivity"] = []interface{}{
			map[string]interface{}{"amount": 3.6, "start": 0},
			map[string]interface{}{"amount": 2.5, "start": 18000000},
		}

		rawObject["basalSchedules"] = map[string]interface{}{
			"standard": []interface{}{
				map[string]interface{}{"rate": 0.6, "start": 0},
			},
			"pattern a": []interface{}{
				map[string]interface{}{"rate": 1.25, "start": 0},
			},
			"pattern b": []interface{}{},
		}
	})

	Context("activeSchedule", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is empty", rawObject, "activeSchedule", "",
				[]*service.Error{testing.ComposeError(validator.ErrorLengthNotGreaterThanOrEqualTo(0, 1), "/activeSchedule", meta)},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is more than 1 characters", rawObject, "activeSchedule", "A"),
			Entry("is freetext", rawObject, "activeSchedule", "standard"),
		)
	})

	Context("units", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("has bg empty", rawObject, "units", map[string]interface{}{"carb": "grams", "bg": ""},
				[]*service.Error{testing.ComposeError(validator.ErrorStringNotOneOf("", []string{bloodglucose.Mmoll, bloodglucose.MmolL, bloodglucose.Mgdl, bloodglucose.MgdL}), "/units/bg", meta)},
			),
			Entry("has bg not predefined type", rawObject, "units", map[string]interface{}{"carb": "grams", "bg": "na"},
				[]*service.Error{testing.ComposeError(validator.ErrorStringNotOneOf("na", []string{bloodglucose.Mmoll, bloodglucose.MmolL, bloodglucose.Mgdl, bloodglucose.MgdL}), "/units/bg", meta)},
			),
			Entry("has carb empty", rawObject, "units", map[string]interface{}{"carb": "", "bg": "mmol/L"},
				[]*service.Error{testing.ComposeError(validator.ErrorLengthNotGreaterThanOrEqualTo(0, 1), "/units/carb", meta)},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("has carbs set and bg set as mmol/L", rawObject, "units", map[string]interface{}{"carb": "grams", "bg": "mmol/L"}),
			Entry("has carbs set and bg set as mg/dl", rawObject, "units", map[string]interface{}{"carb": "grams", "bg": "mg/dl"}),
		)
	})

	Context("carbRatio", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("has start negative", rawObject, "carbRatio",
				[]interface{}{map[string]interface{}{"amount": 12, "start": -1}},
				[]*service.Error{testing.ComposeError(validator.ErrorIntegerNotInRange(-1, 0, 86400000), "/carbRatio/0/start", meta)},
			),
			Entry("has start greater than 86400000", rawObject, "carbRatio",
				[]interface{}{map[string]interface{}{"amount": 12, "start": 86400001}},
				[]*service.Error{testing.ComposeError(validator.ErrorIntegerNotInRange(86400001, 0, 86400000), "/carbRatio/0/start", meta)},
			),
			Entry("has amount negative", rawObject, "carbRatio",
				[]interface{}{map[string]interface{}{"amount": -1, "start": 21600000}},
				[]*service.Error{testing.ComposeError(validator.ErrorIntegerNotInRange(-1, 0, 250), "/carbRatio/0/amount", meta)},
			),
			Entry("has amount greater than 250", rawObject, "carbRatio",
				[]interface{}{map[string]interface{}{"amount": 251, "start": 21600000}},
				[]*service.Error{testing.ComposeError(validator.ErrorIntegerNotInRange(251, 0, 250), "/carbRatio/0/amount", meta)},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("has start and amount within bounds", rawObject, "carbRatio",
				[]interface{}{map[string]interface{}{"amount": 12.0, "start": 0}},
			),
		)
	})

	Context("basalSchedules", func() {
		basalSchedules := map[string]interface{}{
			"standard": []interface{}{
				map[string]interface{}{"rate": 0.6, "start": 0},
				map[string]interface{}{"rate": 0.6, "start": 10800000},
				map[string]interface{}{"rate": 0.6, "start": 23400000},
				map[string]interface{}{"rate": 0.6, "start": 43200000},
				map[string]interface{}{"rate": 0.6, "start": 63000000},
				map[string]interface{}{"rate": 0.6, "start": 81000000},
			},
			"pattern a": []interface{}{
				map[string]interface{}{"rate": 1.25, "start": 0},
				map[string]interface{}{"rate": 1.25, "start": 10800000},
				map[string]interface{}{"rate": 1.25, "start": 25200000},
				map[string]interface{}{"rate": 1.25, "start": 43200000},
				map[string]interface{}{"rate": 1.25, "start": 72000000},
			},
			"pattern b": []interface{}{},
		}

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("has start and rate within bounds", rawObject, "basalSchedules", basalSchedules),
		)

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("has start negative", rawObject, "basalSchedules",
				map[string]interface{}{
					"standard": []interface{}{
						map[string]interface{}{"rate": 0.6, "start": -1},
					}},
				[]*service.Error{testing.ComposeError(validator.ErrorIntegerNotInRange(-1, 0, 86400000), "/basalSchedules/0/start", meta)},
			),
			Entry("has start to large", rawObject, "basalSchedules",
				map[string]interface{}{
					"standard": []interface{}{
						map[string]interface{}{"rate": 0.6, "start": 86400001},
					}},
				[]*service.Error{testing.ComposeError(validator.ErrorIntegerNotInRange(86400001, 0, 86400000), "/basalSchedules/0/start", meta)},
			),
			Entry("has nested start to large", rawObject, "basalSchedules",
				map[string]interface{}{
					"standard": []interface{}{
						map[string]interface{}{"rate": 0.6, "start": 5},
						map[string]interface{}{"rate": 0.6, "start": 86400001},
					}},
				[]*service.Error{testing.ComposeError(validator.ErrorIntegerNotInRange(86400001, 0, 86400000), "/basalSchedules/1/start", meta)},
			),
			Entry("has start negative", rawObject, "basalSchedules",
				map[string]interface{}{
					"standard": []interface{}{
						map[string]interface{}{"rate": -0.1, "start": 10800000},
					}},
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(-0.1, 0.0, 20.0), "/basalSchedules/0/rate", meta)},
			),
			Entry("has start to large", rawObject, "basalSchedules",
				map[string]interface{}{
					"standard": []interface{}{
						map[string]interface{}{"rate": 20.1, "start": 10800000},
					}},
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(20.1, 0.0, 20.0), "/basalSchedules/0/rate", meta)},
			),
			Entry("has nested rate to large", rawObject, "basalSchedules",
				map[string]interface{}{
					"standard": []interface{}{
						map[string]interface{}{"rate": 0.6, "start": 0},
						map[string]interface{}{"rate": 25.1, "start": 10800000},
					}},
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(25.1, 0.0, 20.0), "/basalSchedules/1/rate", meta)},
			),
		)

	})

	Context("insulinSensitivity", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("has start negative", rawObject, "insulinSensitivity",
				[]interface{}{map[string]interface{}{"amount": 12, "start": -1}},
				[]*service.Error{testing.ComposeError(validator.ErrorIntegerNotInRange(-1, 0, 86400000), "/insulinSensitivity/0/start", meta)},
			),
			Entry("has start greater than 86400000", rawObject, "insulinSensitivity",
				[]interface{}{map[string]interface{}{"amount": 12, "start": 86400001}},
				[]*service.Error{testing.ComposeError(validator.ErrorIntegerNotInRange(86400001, 0, 86400000), "/insulinSensitivity/0/start", meta)},
			),
			Entry("has amount negative", rawObject, "insulinSensitivity",
				[]interface{}{map[string]interface{}{"amount": -0.1, "start": 21600000}},
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(-0.1, bloodglucose.MgdLFromValue, bloodglucose.MgdLToValue), "/insulinSensitivity/0/amount", meta)},
			),
			Entry("has amount greater than 1000.0", rawObject, "insulinSensitivity",
				[]interface{}{map[string]interface{}{"amount": 1000.1, "start": 21600000}},
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(1000.1, bloodglucose.MgdLFromValue, bloodglucose.MgdLToValue), "/insulinSensitivity/0/amount", meta)},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("has start and amount within bounds", rawObject, "insulinSensitivity",
				[]interface{}{map[string]interface{}{"amount": 12, "start": 0}},
			),
		)
	})

	Context("bgTarget", func() {
		Context("start, target, range", func() {
			DescribeTable("invalid when", testing.ExpectFieldNotValid,
				Entry("has start negative", rawObject, "bgTarget",
					[]interface{}{map[string]interface{}{"start": -1, "target": 99.0, "range": 15}},
					[]*service.Error{testing.ComposeError(validator.ErrorIntegerNotInRange(-1, 0, 86400000), "/bgTarget/0/start", meta)},
				),
				Entry("has start greater than 86400000", rawObject, "bgTarget",
					[]interface{}{map[string]interface{}{"start": 86400001, "target": 99.0, "range": 15}},
					[]*service.Error{testing.ComposeError(validator.ErrorIntegerNotInRange(86400001, 0, 86400000), "/bgTarget/0/start", meta)},
				),
				Entry("has target negative", rawObject, "bgTarget",
					[]interface{}{map[string]interface{}{"start": 21600000, "target": -0.1, "range": 15}},
					[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(-0.1, bloodglucose.MgdLFromValue, bloodglucose.MgdLToValue), "/bgTarget/0/target", meta)},
				),
				Entry("has target greater than 1000.0", rawObject, "bgTarget",
					[]interface{}{map[string]interface{}{"start": 21600000, "target": 1000.1, "range": 15}},
					[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(1000.1, bloodglucose.MgdLFromValue, bloodglucose.MgdLToValue), "/bgTarget/0/target", meta)},
				),
				Entry("has range negative", rawObject, "bgTarget",
					[]interface{}{map[string]interface{}{"start": 21600000, "target": 99.0, "range": -1}},
					[]*service.Error{testing.ComposeError(validator.ErrorIntegerNotInRange(-1, 0, 50), "/bgTarget/0/range", meta)},
				),
				Entry("has range greater than 51", rawObject, "bgTarget",
					[]interface{}{map[string]interface{}{"start": 21600000, "target": 199.0, "range": 51}},
					[]*service.Error{testing.ComposeError(validator.ErrorIntegerNotInRange(51, 0, 50), "/bgTarget/0/range", meta)},
				),
			)

			DescribeTable("valid when", testing.ExpectFieldIsValid,
				Entry("is within bounds", rawObject, "bgTarget",
					[]interface{}{map[string]interface{}{"start": 21600000, "target": 99.9, "range": 10}},
				),
			)
		})

		Context("start, target, high", func() {

			DescribeTable("invalid when", testing.ExpectFieldNotValid,
				Entry("has start negative", rawObject, "bgTarget",
					[]interface{}{map[string]interface{}{"start": -1, "target": 99.0, "high": 180.0}},
					[]*service.Error{testing.ComposeError(validator.ErrorIntegerNotInRange(-1, 0, 86400000), "/bgTarget/0/start", meta)},
				),
				Entry("has start greater than 86400000", rawObject, "bgTarget",
					[]interface{}{map[string]interface{}{"start": 86400001, "target": 99.0, "high": 180.0}},
					[]*service.Error{testing.ComposeError(validator.ErrorIntegerNotInRange(86400001, 0, 86400000), "/bgTarget/0/start", meta)},
				),
				Entry("has target negative", rawObject, "bgTarget",
					[]interface{}{map[string]interface{}{"start": 21600000, "target": -0.1, "high": 180.0}},
					[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(-0.1, bloodglucose.MgdLFromValue, bloodglucose.MgdLToValue), "/bgTarget/0/target", meta)},
				),
				Entry("has target greater than 1000.0", rawObject, "bgTarget",
					[]interface{}{map[string]interface{}{"start": 21600000, "target": 1000.1, "high": 180.0}},
					[]*service.Error{
						testing.ComposeError(validator.ErrorFloatNotInRange(1000.1, bloodglucose.MgdLFromValue, bloodglucose.MgdLToValue), "/bgTarget/0/target", meta),
						testing.ComposeError(validator.ErrorValueNotGreaterThan(180, 1000.1), "/bgTarget/0/high", meta),
					},
				),
				Entry("has high less than target", rawObject, "bgTarget",
					[]interface{}{map[string]interface{}{"start": 21600000, "target": 90.0, "high": 80.0}},
					[]*service.Error{testing.ComposeError(validator.ErrorValueNotGreaterThan(80.0, 90.0), "/bgTarget/0/high", meta)},
				),
				Entry("has high greater than 1000.0", rawObject, "bgTarget",
					[]interface{}{map[string]interface{}{"start": 21600000, "target": 90.0, "high": 1000.1}},
					[]*service.Error{testing.ComposeError(validator.ErrorValueNotLessThanOrEqualTo(1000.1, bloodglucose.MgdLToValue), "/bgTarget/0/high", meta)},
				),
			)

			DescribeTable("valid when", testing.ExpectFieldIsValid,
				Entry("is within bounds", rawObject, "bgTarget",
					[]interface{}{map[string]interface{}{"start": 21600000, "target": 99.9, "high": 180.0}},
				),
			)
		})
	})

	Context("bgTarget normalized", func() {
		DescribeTable("normalization when low mmol/L", func(val, expected float64) {
			pumpSettings, err := pump.New()
			Expect(err).To(BeNil())
			pumpSettings.Units = &pump.Units{BloodGlucose: &bloodglucose.MmolL}

			high := val + 5.0
			target := val + 2.0

			pumpSettings.BloodGlucoseTargets = &[]*pump.BloodGlucoseTarget{
				{High: &high, Low: &val, Target: &target},
				{High: &high, Low: &val, Target: &target},
			}

			testContext := context.NewStandard()
			standardNormalizer, err := normalizer.NewStandard(testContext)
			Expect(err).To(BeNil())
			pumpSettings.Normalize(standardNormalizer)
			Expect(pumpSettings.Units.BloodGlucose).To(Equal(&bloodglucose.MmolL))

			for _, bgTarget := range *pumpSettings.BloodGlucoseTargets {
				Expect(bgTarget.Low).To(Equal(&expected))
			}
		},
			Entry("is very low", 0.1, 0.1),
			Entry("is very high", 50.0, 50.0),
			Entry("is normal", 3.8, 3.8),
		)

		DescribeTable("normalization when high mmol/L", func(val, expected float64) {
			pumpSettings, err := pump.New()
			Expect(err).To(BeNil())
			pumpSettings.Units = &pump.Units{BloodGlucose: &bloodglucose.MmolL}

			low := val - 5.0
			target := val - 2.0

			pumpSettings.BloodGlucoseTargets = &[]*pump.BloodGlucoseTarget{
				{High: &val, Low: &low, Target: &target},
				{High: &val, Low: &low, Target: &target},
			}

			testContext := context.NewStandard()
			standardNormalizer, err := normalizer.NewStandard(testContext)
			Expect(err).To(BeNil())
			pumpSettings.Normalize(standardNormalizer)
			Expect(pumpSettings.Units.BloodGlucose).To(Equal(&bloodglucose.MmolL))

			for _, bgTarget := range *pumpSettings.BloodGlucoseTargets {
				Expect(bgTarget.High).To(Equal(&expected))
			}
		},
			Entry("is very low", 8.1, 8.1),
			Entry("is very high", 55.0, 55.0),
			Entry("is normal", 3.8, 3.8),
		)

		DescribeTable("normalization when target mmol/L", func(val, expected float64) {
			pumpSettings, err := pump.New()
			Expect(err).To(BeNil())
			pumpSettings.Units = &pump.Units{BloodGlucose: &bloodglucose.MmolL}

			low := val - 5.0
			high := val + 5.0

			pumpSettings.BloodGlucoseTargets = &[]*pump.BloodGlucoseTarget{
				{High: &high, Low: &low, Target: &val},
				{High: &high, Low: &low, Target: &val},
			}

			testContext := context.NewStandard()
			standardNormalizer, err := normalizer.NewStandard(testContext)
			Expect(err).To(BeNil())
			pumpSettings.Normalize(standardNormalizer)
			Expect(pumpSettings.Units.BloodGlucose).To(Equal(&bloodglucose.MmolL))

			for _, bgTarget := range *pumpSettings.BloodGlucoseTargets {
				Expect(bgTarget.Target).To(Equal(&expected))
			}
		},
			Entry("is very low", 8.1, 8.1),
			Entry("is very high", 49.0, 49.0),
			Entry("is normal", 10.1, 10.1),
		)

		DescribeTable("normalization when low mg/dL", func(val, expected float64) {
			pumpSettings, err := pump.New()
			Expect(err).To(BeNil())
			pumpSettings.Units = &pump.Units{BloodGlucose: &bloodglucose.MgdL}

			high := val + 5.0
			target := val + 2.0

			pumpSettings.BloodGlucoseTargets = &[]*pump.BloodGlucoseTarget{
				{High: &high, Low: &val, Target: &target},
				{High: &high, Low: &val, Target: &target},
			}

			testContext := context.NewStandard()
			standardNormalizer, err := normalizer.NewStandard(testContext)
			Expect(err).To(BeNil())
			pumpSettings.Normalize(standardNormalizer)
			Expect(pumpSettings.Units.BloodGlucose).To(Equal(&bloodglucose.MmolL))

			for _, bgTarget := range *pumpSettings.BloodGlucoseTargets {
				Expect(bgTarget.Low).To(Equal(&expected))
			}
		},
			Entry("is very low", 60.1, 3.3359995426183655),
			Entry("is very high", 800.0, 44.405983928364265),
			Entry("is normal", 160.0, 8.881196785672854),
		)

		DescribeTable("normalization when high mg/dL", func(val, expected float64) {
			pumpSettings, err := pump.New()
			Expect(err).To(BeNil())
			pumpSettings.Units = &pump.Units{BloodGlucose: &bloodglucose.MgdL}

			low := val - 5.0
			target := val - 2.0

			pumpSettings.BloodGlucoseTargets = &[]*pump.BloodGlucoseTarget{
				{High: &val, Low: &low, Target: &target},
				{High: &val, Low: &low, Target: &target},
			}

			testContext := context.NewStandard()
			standardNormalizer, err := normalizer.NewStandard(testContext)
			Expect(err).To(BeNil())
			pumpSettings.Normalize(standardNormalizer)
			Expect(pumpSettings.Units.BloodGlucose).To(Equal(&bloodglucose.MmolL))

			for _, bgTarget := range *pumpSettings.BloodGlucoseTargets {
				Expect(bgTarget.High).To(Equal(&expected))
			}
		},
			Entry("is very low", 100.0, 5.550747991045533),
			Entry("is very high", 950.0, 52.73210591493257),
			Entry("is normal", 200.0, 11.101495982091066),
		)

		DescribeTable("normalization when target mg/dL", func(val, expected float64) {
			pumpSettings, err := pump.New()
			Expect(err).To(BeNil())
			pumpSettings.Units = &pump.Units{BloodGlucose: &bloodglucose.MgdL}

			low := val - 5.0
			high := val + 5.0

			pumpSettings.BloodGlucoseTargets = &[]*pump.BloodGlucoseTarget{
				{High: &high, Low: &low, Target: &val},
				{High: &high, Low: &low, Target: &val},
			}

			testContext := context.NewStandard()
			standardNormalizer, err := normalizer.NewStandard(testContext)
			Expect(err).To(BeNil())
			pumpSettings.Normalize(standardNormalizer)
			Expect(pumpSettings.Units.BloodGlucose).To(Equal(&bloodglucose.MmolL))

			for _, bgTarget := range *pumpSettings.BloodGlucoseTargets {
				Expect(bgTarget.Target).To(Equal(&expected))
			}
		},
			Entry("is very low", 70.1, 3.8910743417229186),
			Entry("is very high", 500.0, 27.75373995522767),
			Entry("is normal", 180.1, 9.996897131873006),
		)
	})

	Context("insulinSensitivity normalized", func() {
		DescribeTable("when mmol/L", func(val, expected float64) {
			pumpSettings, err := pump.New()
			Expect(err).To(BeNil())
			pumpSettings.Units = &pump.Units{BloodGlucose: &bloodglucose.MmolL}

			start := 21600000

			pumpSettings.InsulinSensitivities = &[]*pump.InsulinSensitivity{
				{Amount: &val, Start: &start},
				{Amount: &val, Start: &start},
			}

			testContext := context.NewStandard()
			standardNormalizer, err := normalizer.NewStandard(testContext)
			Expect(err).To(BeNil())
			pumpSettings.Normalize(standardNormalizer)
			Expect(pumpSettings.Units.BloodGlucose).To(Equal(&bloodglucose.MmolL))

			for _, insulinSensitivity := range *pumpSettings.InsulinSensitivities {
				Expect(insulinSensitivity.Amount).To(Equal(&expected))
			}
		},
			Entry("is very low", 0.1, 0.1),
			Entry("is very high", 55.0, 55.0),
			Entry("is normal", 8.3, 8.3),
		)

		DescribeTable("when mg/dL", func(val, expected float64) {
			pumpSettings, err := pump.New()
			Expect(err).To(BeNil())
			pumpSettings.Units = &pump.Units{BloodGlucose: &bloodglucose.MgdL}

			start := 21600000

			pumpSettings.InsulinSensitivities = &[]*pump.InsulinSensitivity{
				{Amount: &val, Start: &start},
				{Amount: &val, Start: &start},
			}

			testContext := context.NewStandard()
			standardNormalizer, err := normalizer.NewStandard(testContext)
			Expect(err).To(BeNil())
			pumpSettings.Normalize(standardNormalizer)
			Expect(pumpSettings.Units.BloodGlucose).To(Equal(&bloodglucose.MmolL))

			for _, insulinSensitivity := range *pumpSettings.InsulinSensitivities {
				Expect(insulinSensitivity.Amount).To(Equal(&expected))
			}
		},
			Entry("is very low", 60.0, 3.33044879462732),
			Entry("is very high", 990.85745, 55.0),
			Entry("is normal", 160.0, 8.881196785672854),
		)
	})
})
