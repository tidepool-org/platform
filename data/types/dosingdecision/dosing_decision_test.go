package dosingdecision_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataTypes "github.com/tidepool-org/platform/data/types"
	dataTypesDosingDecision "github.com/tidepool-org/platform/data/types/dosingdecision"
	dataTypesDosingDecisionTest "github.com/tidepool-org/platform/data/types/dosingdecision/test"
	dataTypesSettingsPump "github.com/tidepool-org/platform/data/types/settings/pump"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
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

	It("TimeFormat is expected", func() {
		Expect(dataTypesDosingDecision.TimeFormat).To(Equal(time.RFC3339Nano))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := dataTypesDosingDecision.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("dosingDecision"))
			Expect(datum.Alerts).To(BeNil())
			Expect(datum.InsulinOnBoard).To(BeNil())
			Expect(datum.CarbohydratesOnBoard).To(BeNil())
			Expect(datum.BloodGlucoseTargetRangeSchedule).To(BeNil())
			Expect(datum.BloodGlucoseForecast).To(BeNil())
			Expect(datum.RecommendedBasal).To(BeNil())
			Expect(datum.RecommendedBolus).To(BeNil())
			Expect(datum.Units).To(BeNil())
		})
	})

	Context("DosingDecision", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",
				func(units *string, mutator func(datum *dataTypesDosingDecision.DosingDecision, units *string), expectedErrors ...error) {
					datum := dataTypesDosingDecisionTest.RandomDosingDecision(units)
					mutator(datum, units)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.DosingDecision, units *string) {},
				),
				Entry("type missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.DosingDecision, units *string) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &dataTypes.Meta{}),
				),
				Entry("type invalid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.DosingDecision, units *string) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "dosingDecision"), "/type", &dataTypes.Meta{Type: "invalidType"}),
				),
				Entry("type dosingDecision",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.DosingDecision, units *string) { datum.Type = "dosingDecision" },
				),
				Entry("insulin on board invalid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.DosingDecision, units *string) { datum.InsulinOnBoard.Amount = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/insulinOnBoard/amount", NewMeta()),
				),
				Entry("carbohydrates on board invalid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.DosingDecision, units *string) {
						datum.CarbohydratesOnBoard.Amount = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/carbohydratesOnBoard/amount", NewMeta()),
				),
				Entry("blood glucose target range schedule invalid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.DosingDecision, units *string) {
						datum.BloodGlucoseTargetRangeSchedule = &dataTypesSettingsPump.BloodGlucoseTargetStartArray{nil}
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bloodGlucoseTargetRangeSchedule/0", NewMeta()),
				),
				Entry("blood glucose forecast invalid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.DosingDecision, units *string) {
						datum.BloodGlucoseForecast = dataTypesDosingDecision.NewForecastArray()
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/bloodGlucoseForecast", NewMeta()),
				),
				Entry("recommended basal invalid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.DosingDecision, units *string) { datum.RecommendedBasal.Rate = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/recommendedBasal/rate", NewMeta()),
				),
				Entry("recommended bolus invalid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.DosingDecision, units *string) {
						datum.RecommendedBolus.Amount = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/recommendedBolus/amount", NewMeta()),
				),
				Entry("units missing",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.DosingDecision, units *string) { datum.Units = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units invalid",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.DosingDecision, units *string) { datum.Units.BloodGlucose = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units/bloodGlucose", NewMeta()),
				),
				Entry("multiple errors",
					pointer.FromString("mmol/L"),
					func(datum *dataTypesDosingDecision.DosingDecision, units *string) {
						datum.InsulinOnBoard.Amount = nil
						datum.CarbohydratesOnBoard.Amount = nil
						datum.BloodGlucoseTargetRangeSchedule = &dataTypesSettingsPump.BloodGlucoseTargetStartArray{nil}
						datum.BloodGlucoseForecast = dataTypesDosingDecision.NewForecastArray()
						datum.RecommendedBasal.Rate = nil
						datum.RecommendedBolus.Amount = nil
						datum.Units = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/insulinOnBoard/amount", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/carbohydratesOnBoard/amount", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bloodGlucoseTargetRangeSchedule/0", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/bloodGlucoseForecast", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/recommendedBasal/rate", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/recommendedBolus/amount", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
			)
		})
	})
})
