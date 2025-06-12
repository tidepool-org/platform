package dosingdecision_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	dataBloodGlucoseTest "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTypes "github.com/tidepool-org/platform/data/types"
	dataTypesDosingDecision "github.com/tidepool-org/platform/data/types/dosingdecision"
	dataTypesDosingDecisionTest "github.com/tidepool-org/platform/data/types/dosingdecision/test"
	dataTypesSettingsPumpTest "github.com/tidepool-org/platform/data/types/settings/pump/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() interface{} {
	return &dataTypes.Meta{
		Type: "dosingDecision",
	}
}

var _ = Describe("DosingDecision", func() {
	It("Type is expected", func() {
		Expect(dataTypesDosingDecision.Type).To(Equal("dosingDecision"))
	})

	It("ReasonLengthMaximum is expected", func() {
		Expect(dataTypesDosingDecision.ReasonLengthMaximum).To(Equal(100))
	})

	It("ScheduleTimeZoneOffsetMaximum is expected", func() {
		Expect(dataTypesDosingDecision.ScheduleTimeZoneOffsetMaximum).To(Equal(10080))
	})

	It("ScheduleTimeZoneOffsetMinimum is expected", func() {
		Expect(dataTypesDosingDecision.ScheduleTimeZoneOffsetMinimum).To(Equal(-10080))
	})

	Context("DosingDecision", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesDosingDecision.DosingDecision)) {
				unitsBloodGlucose := pointer.FromString(test.RandomStringFromArray(dataBloodGlucose.Units()))
				datum := dataTypesDosingDecisionTest.RandomDosingDecision(unitsBloodGlucose)
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesDosingDecisionTest.NewObjectFromDosingDecision(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesDosingDecisionTest.NewObjectFromDosingDecision(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesDosingDecision.DosingDecision) {},
			),
			Entry("empty",
				func(datum *dataTypesDosingDecision.DosingDecision) {
					*datum = *dataTypesDosingDecision.New()
				},
			),
			Entry("all",
				func(datum *dataTypesDosingDecision.DosingDecision) {
					unitsBloodGlucose := pointer.FromString(test.RandomStringFromArray(dataBloodGlucose.Units()))
					datum.Reason = pointer.FromString(test.RandomStringFromRange(1, dataTypesDosingDecision.ReasonLengthMaximum))
					datum.OriginalFood = dataTypesDosingDecisionTest.RandomFood()
					datum.Food = dataTypesDosingDecisionTest.RandomFood()
					datum.SelfMonitoredBloodGlucose = dataTypesDosingDecisionTest.RandomBloodGlucose(unitsBloodGlucose)
					datum.CarbohydratesOnBoard = dataTypesDosingDecisionTest.RandomCarbohydratesOnBoard()
					datum.InsulinOnBoard = dataTypesDosingDecisionTest.RandomInsulinOnBoard()
					datum.BloodGlucoseTargetSchedule = dataTypesSettingsPumpTest.RandomBloodGlucoseTargetStartArray(unitsBloodGlucose)
					datum.HistoricalBloodGlucose = dataTypesDosingDecisionTest.RandomBloodGlucoseArray(unitsBloodGlucose)
					datum.ForecastBloodGlucose = dataTypesDosingDecisionTest.RandomBloodGlucoseArray(unitsBloodGlucose)
					datum.RecommendedBasal = dataTypesDosingDecisionTest.RandomRecommendedBasal()
					datum.RecommendedBolus = dataTypesDosingDecisionTest.RandomBolus()
					datum.RequestedBolus = dataTypesDosingDecisionTest.RandomBolus()
					datum.Warnings = dataTypesDosingDecisionTest.RandomIssueArray()
					datum.Errors = dataTypesDosingDecisionTest.RandomIssueArray()
					datum.ScheduleTimeZoneOffset = pointer.FromInt(test.RandomIntFromRange(dataTypesDosingDecision.ScheduleTimeZoneOffsetMinimum, dataTypesDosingDecision.ScheduleTimeZoneOffsetMaximum))
					datum.Units = dataTypesDosingDecisionTest.RandomUnits(unitsBloodGlucose)
				},
			),
		)

		Context("New", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesDosingDecision.New()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Type).To(Equal("dosingDecision"))
				Expect(datum.Reason).To(BeNil())
				Expect(datum.OriginalFood).To(BeNil())
				Expect(datum.Food).To(BeNil())
				Expect(datum.SelfMonitoredBloodGlucose).To(BeNil())
				Expect(datum.CarbohydratesOnBoard).To(BeNil())
				Expect(datum.InsulinOnBoard).To(BeNil())
				Expect(datum.BloodGlucoseTargetSchedule).To(BeNil())
				Expect(datum.HistoricalBloodGlucose).To(BeNil())
				Expect(datum.ForecastBloodGlucose).To(BeNil())
				Expect(datum.RecommendedBasal).To(BeNil())
				Expect(datum.RecommendedBolus).To(BeNil())
				Expect(datum.RequestedBolus).To(BeNil())
				Expect(datum.Warnings).To(BeNil())
				Expect(datum.Errors).To(BeNil())
				Expect(datum.ScheduleTimeZoneOffset).To(BeNil())
				Expect(datum.Units).To(BeNil())

			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesDosingDecision.DosingDecision), expectedErrors ...error) {
					unitsBloodGlucose := pointer.FromString(test.RandomStringFromArray(dataBloodGlucose.Units()))
					expectedDatum := dataTypesDosingDecisionTest.RandomDosingDecisionForParser(unitsBloodGlucose)
					object := dataTypesDosingDecisionTest.NewObjectFromDosingDecision(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesDosingDecision.New()
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesDosingDecision.DosingDecision) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesDosingDecision.DosingDecision) {
						object["reason"] = true
						object["originalFood"] = true
						object["food"] = true
						object["smbg"] = true
						object["carbsOnBoard"] = true
						object["insulinOnBoard"] = true
						object["bgTargetSchedule"] = true
						object["bgHistorical"] = true
						object["bgForecast"] = true
						object["recommendedBasal"] = true
						object["recommendedBolus"] = true
						object["requestedBolus"] = true
						object["warnings"] = true
						object["errors"] = true
						object["scheduleTimeZoneOffset"] = true
						object["units"] = true
						expectedDatum.Reason = nil
						expectedDatum.OriginalFood = nil
						expectedDatum.Food = nil
						expectedDatum.SelfMonitoredBloodGlucose = nil
						expectedDatum.CarbohydratesOnBoard = nil
						expectedDatum.InsulinOnBoard = nil
						expectedDatum.BloodGlucoseTargetSchedule = nil
						expectedDatum.HistoricalBloodGlucose = nil
						expectedDatum.ForecastBloodGlucose = nil
						expectedDatum.RecommendedBasal = nil
						expectedDatum.RecommendedBolus = nil
						expectedDatum.RequestedBolus = nil
						expectedDatum.Warnings = nil
						expectedDatum.Errors = nil
						expectedDatum.ScheduleTimeZoneOffset = nil
						expectedDatum.Units = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotString(true), "/reason", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotObject(true), "/originalFood", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotObject(true), "/food", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotObject(true), "/smbg", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotObject(true), "/carbsOnBoard", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotObject(true), "/insulinOnBoard", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotArray(true), "/bgTargetSchedule", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotArray(true), "/bgHistorical", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotArray(true), "/bgForecast", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotObject(true), "/recommendedBasal", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotObject(true), "/recommendedBolus", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotObject(true), "/requestedBolus", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotArray(true), "/warnings", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotArray(true), "/errors", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotInt(true), "/scheduleTimeZoneOffset", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotObject(true), "/units", NewMeta()),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",
				func(mutator func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string), expectedErrors ...error) {
					unitsBloodGlucose := pointer.FromString(test.RandomStringFromArray(dataBloodGlucose.Units()))
					datum := dataTypesDosingDecisionTest.RandomDosingDecision(unitsBloodGlucose)
					mutator(datum, unitsBloodGlucose)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {},
				),
				Entry("type missing",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &dataTypes.Meta{}),
				),
				Entry("type invalid",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.Type = "invalidType"
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "dosingDecision"), "/type", &dataTypes.Meta{Type: "invalidType"}),
				),
				Entry("type dosingDecision",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.Type = "dosingDecision"
					},
				),
				Entry("reason missing",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.Reason = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/reason", NewMeta()),
				),
				Entry("reason empty",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.Reason = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/reason", NewMeta()),
				),
				Entry("reason length in range (upper)",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.Reason = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
				),
				Entry("reason length out of range (upper)",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.Reason = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/reason", NewMeta()),
				),
				Entry("original food missing",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.OriginalFood = nil
					},
				),
				Entry("original food invalid",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.OriginalFood.Nutrition = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/originalFood/nutrition", NewMeta()),
				),
				Entry("original food valid",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.OriginalFood = dataTypesDosingDecisionTest.RandomFood()
					},
				),
				Entry("food missing",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.Food = nil
					},
				),
				Entry("food invalid",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.Food.Nutrition = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/food/nutrition", NewMeta()),
				),
				Entry("food valid",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.Food = dataTypesDosingDecisionTest.RandomFood()
					},
				),
				Entry("self monitored blood glucose missing",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.SelfMonitoredBloodGlucose = nil
					},
				),
				Entry("self monitored blood glucose invalid",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.SelfMonitoredBloodGlucose.Value = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/smbg/value", NewMeta()),
				),
				Entry("self monitored blood glucose valid",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.SelfMonitoredBloodGlucose = dataTypesDosingDecisionTest.RandomBloodGlucose(unitsBloodGlucose)
					},
				),
				Entry("carbohydrates on board missing",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.CarbohydratesOnBoard = nil
					},
				),
				Entry("carbohydrates on board invalid",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.CarbohydratesOnBoard.Amount = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/carbsOnBoard/amount", NewMeta()),
				),
				Entry("carbohydrates on board valid",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.CarbohydratesOnBoard = dataTypesDosingDecisionTest.RandomCarbohydratesOnBoard()
					},
				),
				Entry("insulin on board missing",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.InsulinOnBoard = nil
					},
				),
				Entry("insulin on board invalid",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.InsulinOnBoard.Amount = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/insulinOnBoard/amount", NewMeta()),
				),
				Entry("insulin on board valid",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.InsulinOnBoard = dataTypesDosingDecisionTest.RandomInsulinOnBoard()
					},
				),
				Entry("blood glucose target schedule missing",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.BloodGlucoseTargetSchedule = nil
					},
				),
				Entry("blood glucose target schedule invalid",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						(*datum.BloodGlucoseTargetSchedule)[0].Start = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bgTargetSchedule/0/start", NewMeta()),
				),
				Entry("blood glucose target schedule valid",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.BloodGlucoseTargetSchedule = dataTypesSettingsPumpTest.RandomBloodGlucoseTargetStartArray(unitsBloodGlucose)
					},
				),
				Entry("historical blood glucose missing",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.HistoricalBloodGlucose = nil
					},
				),
				Entry("historical blood glucose invalid",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						(*datum.HistoricalBloodGlucose)[0].Value = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bgHistorical/0/value", NewMeta()),
				),
				Entry("historical blood glucose valid",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.HistoricalBloodGlucose = dataTypesDosingDecisionTest.RandomBloodGlucoseArray(unitsBloodGlucose)
					},
				),
				Entry("forecast blood glucose missing",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.ForecastBloodGlucose = nil
					},
				),
				Entry("forecast blood glucose invalid",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						(*datum.ForecastBloodGlucose)[0].Value = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bgForecast/0/value", NewMeta()),
				),
				Entry("forecast blood glucose valid",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.ForecastBloodGlucose = dataTypesDosingDecisionTest.RandomBloodGlucoseArray(unitsBloodGlucose)
					},
				),
				Entry("recommended basal missing",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.RecommendedBasal = nil
					},
				),
				Entry("recommended basal invalid",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.RecommendedBasal.Rate = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/recommendedBasal/rate", NewMeta()),
				),
				Entry("recommended basal valid",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.RecommendedBasal = dataTypesDosingDecisionTest.RandomRecommendedBasal()
					},
				),
				Entry("recommended bolus missing",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.RecommendedBolus = nil
					},
				),
				Entry("recommended bolus invalid",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.RecommendedBolus = &dataTypesDosingDecision.Bolus{}
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValuesNotExistForAny("amount", "extended", "normal"), "/recommendedBolus", NewMeta()),
				),
				Entry("recommended bolus valid",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.RecommendedBolus = dataTypesDosingDecisionTest.RandomBolus()
					},
				),
				Entry("requested bolus missing",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.RequestedBolus = nil
					},
				),
				Entry("requested bolus invalid",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.RequestedBolus = &dataTypesDosingDecision.Bolus{}
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValuesNotExistForAny("amount", "extended", "normal"), "/requestedBolus", NewMeta()),
				),
				Entry("requested bolus valid",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.RequestedBolus = dataTypesDosingDecisionTest.RandomBolus()
					},
				),
				Entry("warnings missing",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.Warnings = nil
					},
				),
				Entry("warnings invalid",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						(*datum.Warnings)[0].ID = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/warnings/0/id", NewMeta()),
				),
				Entry("warnings valid",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.Warnings = dataTypesDosingDecisionTest.RandomIssueArray()
					},
				),
				Entry("errors missing",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.Errors = nil
					},
				),
				Entry("errors invalid",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						(*datum.Errors)[0].ID = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/errors/0/id", NewMeta()),
				),
				Entry("errors valid",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.Errors = dataTypesDosingDecisionTest.RandomIssueArray()
					},
				),
				Entry("schedule time zone offset missing",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.ScheduleTimeZoneOffset = nil
					},
				),
				Entry("schedule time zone offset out of range (lower)",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.ScheduleTimeZoneOffset = pointer.FromInt(-10081)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-10081, -10080, 10080), "/scheduleTimeZoneOffset", NewMeta()),
				),
				Entry("schedule time zone offset in range (lower)",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.ScheduleTimeZoneOffset = pointer.FromInt(-10080)
					},
				),
				Entry("schedule time zone offset in range (upper)",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.ScheduleTimeZoneOffset = pointer.FromInt(10080)
					},
				),
				Entry("schedule time zone offset out of range (upper)",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.ScheduleTimeZoneOffset = pointer.FromInt(10081)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(10081, -10080, 10080), "/scheduleTimeZoneOffset", NewMeta()),
				),
				Entry("units missing",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.Units = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units invalid",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.Units.Carbohydrate = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units/carb", NewMeta()),
				),
				Entry("units valid",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.Units = dataTypesDosingDecisionTest.RandomUnits(unitsBloodGlucose)
					},
				),
				Entry("multiple errors",
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						datum.Type = "invalidType"
						datum.Reason = nil
						datum.OriginalFood.Nutrition = nil
						datum.Food.Nutrition = nil
						datum.SelfMonitoredBloodGlucose.Value = nil
						datum.CarbohydratesOnBoard.Amount = nil
						datum.InsulinOnBoard.Amount = nil
						(*datum.BloodGlucoseTargetSchedule)[0].Start = nil
						(*datum.HistoricalBloodGlucose)[0].Value = nil
						(*datum.ForecastBloodGlucose)[0].Value = nil
						datum.RecommendedBasal.Rate = nil
						datum.RecommendedBolus = &dataTypesDosingDecision.Bolus{}
						datum.RequestedBolus = &dataTypesDosingDecision.Bolus{}
						(*datum.Warnings)[0].ID = nil
						(*datum.Errors)[0].ID = nil
						datum.ScheduleTimeZoneOffset = pointer.FromInt(-10081)
						datum.Units = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "dosingDecision"), "/type", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/reason", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/originalFood/nutrition", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/food/nutrition", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/smbg/value", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/carbsOnBoard/amount", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/insulinOnBoard/amount", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bgTargetSchedule/0/start", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bgHistorical/0/value", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bgForecast/0/value", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/recommendedBasal/rate", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValuesNotExistForAny("amount", "extended", "normal"), "/recommendedBolus", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValuesNotExistForAny("amount", "extended", "normal"), "/requestedBolus", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/warnings/0/id", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/errors/0/id", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-10081, -10080, 10080), "/scheduleTimeZoneOffset", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", &dataTypes.Meta{Type: "invalidType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum with origin external",
				func(unitsBloodGlucose *string, mutator func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string), expectator func(datum *dataTypesDosingDecision.DosingDecision, expectedDatum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string)) {
					datum := dataTypesDosingDecisionTest.RandomDosingDecision(unitsBloodGlucose)
					mutator(datum, unitsBloodGlucose)
					expectedDatum := dataTypesDosingDecisionTest.CloneDosingDecision(datum)
					normalizer := dataNormalizer.New(logTest.NewLogger())
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(structure.OriginExternal))
					Expect(normalizer.Error()).To(BeNil())
					Expect(normalizer.Data()).To(BeEmpty())
					if expectator != nil {
						expectator(datum, expectedDatum, unitsBloodGlucose)
					}
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("does not modify the datum; units missing",
					nil,
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {},
					nil,
				),
				Entry("modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {},
					func(datum *dataTypesDosingDecision.DosingDecision, expectedDatum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units.BloodGlucose, expectedDatum.Units.BloodGlucose)
					},
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {},
					func(datum *dataTypesDosingDecision.DosingDecision, expectedDatum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.SelfMonitoredBloodGlucose.Value, expectedDatum.SelfMonitoredBloodGlucose.Value, unitsBloodGlucose)
						for index := range *datum.BloodGlucoseTargetSchedule {
							dataBloodGlucoseTest.ExpectNormalizedTarget(&(*datum.BloodGlucoseTargetSchedule)[index].Target, &(*expectedDatum.BloodGlucoseTargetSchedule)[index].Target, unitsBloodGlucose)
						}
						for index := range *datum.HistoricalBloodGlucose {
							dataBloodGlucoseTest.ExpectNormalizedValue((*datum.HistoricalBloodGlucose)[index].Value, (*expectedDatum.HistoricalBloodGlucose)[index].Value, unitsBloodGlucose)
						}
						for index := range *datum.ForecastBloodGlucose {
							dataBloodGlucoseTest.ExpectNormalizedValue((*datum.ForecastBloodGlucose)[index].Value, (*expectedDatum.ForecastBloodGlucose)[index].Value, unitsBloodGlucose)
						}
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units.BloodGlucose, expectedDatum.Units.BloodGlucose)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {},
					func(datum *dataTypesDosingDecision.DosingDecision, expectedDatum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.SelfMonitoredBloodGlucose.Value, expectedDatum.SelfMonitoredBloodGlucose.Value, unitsBloodGlucose)
						for index := range *datum.BloodGlucoseTargetSchedule {
							dataBloodGlucoseTest.ExpectNormalizedTarget(&(*datum.BloodGlucoseTargetSchedule)[index].Target, &(*expectedDatum.BloodGlucoseTargetSchedule)[index].Target, unitsBloodGlucose)
						}
						for index := range *datum.HistoricalBloodGlucose {
							dataBloodGlucoseTest.ExpectNormalizedValue((*datum.HistoricalBloodGlucose)[index].Value, (*expectedDatum.HistoricalBloodGlucose)[index].Value, unitsBloodGlucose)
						}
						for index := range *datum.ForecastBloodGlucose {
							dataBloodGlucoseTest.ExpectNormalizedValue((*datum.ForecastBloodGlucose)[index].Value, (*expectedDatum.ForecastBloodGlucose)[index].Value, unitsBloodGlucose)
						}
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units.BloodGlucose, expectedDatum.Units.BloodGlucose)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(unitsBloodGlucose *string, mutator func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string), expectator func(datum *dataTypesDosingDecision.DosingDecision, expectedDatum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := dataTypesDosingDecisionTest.RandomDosingDecision(unitsBloodGlucose)
						mutator(datum, unitsBloodGlucose)
						expectedDatum := dataTypesDosingDecisionTest.CloneDosingDecision(datum)
						normalizer := dataNormalizer.New(logTest.NewLogger())
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum, unitsBloodGlucose)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum; units missing",
					nil,
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {},
					nil,
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *dataTypesDosingDecision.DosingDecision, unitsBloodGlucose *string) {},
					nil,
				),
			)
		})
	})
})
