package physical_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/activity/physical"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() interface{} {
	return &types.Meta{
		Type: "physicalActivity",
	}
}

func NewPhysical() *physical.Physical {
	datum := physical.New()
	datum.Base = *testDataTypes.NewBase()
	datum.Type = "physicalActivity"
	datum.ActivityType = pointer.String(test.RandomStringFromArray(physical.ActivityTypes()))
	if datum.ActivityType != nil && *datum.ActivityType == physical.ActivityTypeOther {
		datum.ActivityTypeOther = pointer.String(test.NewText(1, 100))
	}
	datum.Aggregate = pointer.Bool(test.RandomBool())
	datum.Distance = NewDistance()
	datum.Duration = NewDuration()
	datum.ElevationChange = NewElevationChange()
	datum.Energy = NewEnergy()
	datum.Flight = NewFlight()
	datum.Name = pointer.String(test.NewText(1, 100))
	datum.ReportedIntensity = pointer.String(test.RandomStringFromArray(physical.ReportedIntensities()))
	datum.Step = NewStep()
	return datum
}

func ClonePhysical(datum *physical.Physical) *physical.Physical {
	if datum == nil {
		return nil
	}
	clone := physical.New()
	clone.Base = *testDataTypes.CloneBase(&datum.Base)
	clone.ActivityType = test.CloneString(datum.ActivityType)
	clone.ActivityTypeOther = test.CloneString(datum.ActivityTypeOther)
	clone.Aggregate = test.CloneBool(datum.Aggregate)
	clone.Distance = CloneDistance(datum.Distance)
	clone.Duration = CloneDuration(datum.Duration)
	clone.ElevationChange = CloneElevationChange(datum.ElevationChange)
	clone.Energy = CloneEnergy(datum.Energy)
	clone.Flight = CloneFlight(datum.Flight)
	clone.Name = test.CloneString(datum.Name)
	clone.ReportedIntensity = test.CloneString(datum.ReportedIntensity)
	clone.Step = CloneStep(datum.Step)
	return clone
}

var _ = Describe("Physical", func() {
	It("Type is expected", func() {
		Expect(physical.Type).To(Equal("physicalActivity"))
	})

	It("ActivityTypeOther is expected", func() {
		Expect(physical.ActivityTypeOther).To(Equal("other"))
	})

	It("ReportedIntensityHigh is expected", func() {
		Expect(physical.ReportedIntensityHigh).To(Equal("high"))
	})

	It("ReportedIntensityLow is expected", func() {
		Expect(physical.ReportedIntensityLow).To(Equal("low"))
	})

	It("ReportedIntensityMedium is expected", func() {
		Expect(physical.ReportedIntensityMedium).To(Equal("medium"))
	})

	It("ActivityTypes returns expected", func() {
		Expect(physical.ActivityTypes()).To(Equal([]string{
			"americanFootball", "archery", "australianFootball", "badminton", "barre", "baseball", "basketball", "bowling", "boxing", "climbing", "coreTraining",
			"cricket", "crossCountrySkiing", "crossTraining", "curling", "cycling", "dance", "danceInspiredTraining", "downhillSkiing", "elliptical",
			"equestrianSports", "fencing", "fishing", "flexibility", "functionalStrengthTraining", "golf", "gymnastics", "handball", "handCycling",
			"highIntensityIntervalTraining", "hiking", "hockey", "hunting", "jumpRope", "kickboxing", "lacrosse", "martialArts", "mindAndBody", "mixedCardio",
			"mixedMetabolicCardioTraining", "other", "paddleSports", "pilates", "play", "preparationAndRecovery", "racquetball", "rowing", "rugby", "running",
			"sailing", "skatingSports", "snowboarding", "snowSports", "soccer", "softball", "squash", "stairClimbing", "stairs", "stepTraining",
			"surfingSports", "swimming", "tableTennis", "taiChi", "tennis", "trackAndField", "traditionalStrengthTraining", "volleyball", "walking", "waterFitness",
			"waterPolo", "waterSports", "wheelchairRunPace", "wheelchairWalkPace", "wrestling", "yoga",
		}))
	})

	It("ReportedIntensities returns expected", func() {
		Expect(physical.ReportedIntensities()).To(Equal([]string{"high", "low", "medium"}))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := physical.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("physicalActivity"))
			Expect(datum.ActivityType).To(BeNil())
			Expect(datum.ActivityTypeOther).To(BeNil())
			Expect(datum.Aggregate).To(BeNil())
			Expect(datum.Distance).To(BeNil())
			Expect(datum.Duration).To(BeNil())
			Expect(datum.ElevationChange).To(BeNil())
			Expect(datum.Energy).To(BeNil())
			Expect(datum.Flight).To(BeNil())
			Expect(datum.Name).To(BeNil())
			Expect(datum.ReportedIntensity).To(BeNil())
			Expect(datum.Step).To(BeNil())
		})
	})

	Context("Physical", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *physical.Physical), expectedErrors ...error) {
					datum := NewPhysical()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *physical.Physical) {},
				),
				Entry("type missing",
					func(datum *physical.Physical) { datum.Type = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &types.Meta{}),
				),
				Entry("type invalid",
					func(datum *physical.Physical) { datum.Type = "invalidType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "physicalActivity"), "/type", &types.Meta{Type: "invalidType"}),
				),
				Entry("type physicalActivity",
					func(datum *physical.Physical) { datum.Type = "physicalActivity" },
				),
				Entry("activity type missing; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = nil
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type missing; activity type other exists",
					func(datum *physical.Physical) {
						datum.ActivityType = nil
						datum.ActivityTypeOther = pointer.String(test.NewText(1, 100))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type invalid; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("invalid")
						datum.ActivityTypeOther = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", physical.ActivityTypes()), "/activityType", NewMeta()),
				),
				Entry("activity type invalid; activity type other exists",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("invalid")
						datum.ActivityTypeOther = pointer.String(test.NewText(1, 100))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", physical.ActivityTypes()), "/activityType", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type americanFootball; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("americanFootball")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type americanFootball; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("americanFootball")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type archery; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("archery")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type archery; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("archery")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type australianFootball; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("australianFootball")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type australianFootball; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("australianFootball")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type badminton; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("badminton")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type badminton; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("badminton")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type barre; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("barre")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type barre; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("barre")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type baseball; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("baseball")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type baseball; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("baseball")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type basketball; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("basketball")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type basketball; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("basketball")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type bowling; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("bowling")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type bowling; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("bowling")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type boxing; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("boxing")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type boxing; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("boxing")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type climbing; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("climbing")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type climbing; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("climbing")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type coreTraining; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("coreTraining")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type coreTraining; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("coreTraining")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type cricket; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("cricket")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type cricket; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("cricket")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type crossCountrySkiing; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("crossCountrySkiing")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type crossCountrySkiing; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("crossCountrySkiing")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type crossTraining; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("crossTraining")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type crossTraining; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("crossTraining")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type curling; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("curling")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type curling; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("curling")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type cycling; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("cycling")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type cycling; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("cycling")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type dance; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("dance")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type dance; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("dance")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type danceInspiredTraining; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("danceInspiredTraining")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type danceInspiredTraining; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("danceInspiredTraining")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type downhillSkiing; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("downhillSkiing")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type downhillSkiing; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("downhillSkiing")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type elliptical; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("elliptical")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type elliptical; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("elliptical")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type equestrianSports; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("equestrianSports")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type equestrianSports; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("equestrianSports")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type fencing; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("fencing")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type fencing; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("fencing")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type fishing; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("fishing")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type fishing; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("fishing")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type flexibility; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("flexibility")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type flexibility; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("flexibility")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type functionalStrengthTraining; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("functionalStrengthTraining")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type functionalStrengthTraining; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("functionalStrengthTraining")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type golf; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("golf")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type golf; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("golf")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type gymnastics; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("gymnastics")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type gymnastics; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("gymnastics")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type handball; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("handball")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type handball; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("handball")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type handCycling; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("handCycling")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type handCycling; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("handCycling")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type highIntensityIntervalTraining; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("highIntensityIntervalTraining")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type highIntensityIntervalTraining; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("highIntensityIntervalTraining")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type hiking; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("hiking")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type hiking; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("hiking")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type hockey; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("hockey")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type hockey; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("hockey")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type hunting; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("hunting")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type hunting; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("hunting")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type jumpRope; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("jumpRope")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type jumpRope; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("jumpRope")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type kickboxing; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("kickboxing")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type kickboxing; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("kickboxing")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type lacrosse; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("lacrosse")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type lacrosse; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("lacrosse")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type martialArts; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("martialArts")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type martialArts; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("martialArts")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type mindAndBody; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("mindAndBody")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type mindAndBody; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("mindAndBody")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type mixedCardio; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("mixedCardio")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type mixedCardio; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("mixedCardio")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type mixedMetabolicCardioTraining; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("mixedMetabolicCardioTraining")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type mixedMetabolicCardioTraining; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("mixedMetabolicCardioTraining")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type other; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("other")
						datum.ActivityTypeOther = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type other; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("other")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type other; activity type other length; in range (upper)",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("other")
						datum.ActivityTypeOther = pointer.String(test.NewText(100, 100))
					},
				),
				Entry("activity type other; activity type other length; out of range (upper)",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("other")
						datum.ActivityTypeOther = pointer.String(test.NewText(101, 101))
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type paddleSports; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("paddleSports")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type paddleSports; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("paddleSports")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type pilates; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("pilates")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type pilates; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("pilates")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type play; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("play")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type play; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("play")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type preparationAndRecovery; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("preparationAndRecovery")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type preparationAndRecovery; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("preparationAndRecovery")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type racquetball; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("racquetball")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type racquetball; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("racquetball")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type rowing; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("rowing")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type rowing; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("rowing")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type rugby; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("rugby")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type rugby; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("rugby")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type running; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("running")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type running; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("running")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type sailing; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("sailing")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type sailing; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("sailing")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type skatingSports; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("skatingSports")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type skatingSports; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("skatingSports")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type snowboarding; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("snowboarding")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type snowboarding; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("snowboarding")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type snowSports; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("snowSports")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type snowSports; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("snowSports")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type soccer; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("soccer")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type soccer; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("soccer")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type softball; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("softball")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type softball; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("softball")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type squash; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("squash")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type squash; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("squash")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type stairClimbing; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("stairClimbing")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type stairClimbing; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("stairClimbing")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type stairs; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("stairs")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type stairs; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("stairs")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type stepTraining; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("stepTraining")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type stepTraining; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("stepTraining")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type surfingSports; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("surfingSports")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type surfingSports; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("surfingSports")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type swimming; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("swimming")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type swimming; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("swimming")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type tableTennis; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("tableTennis")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type tableTennis; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("tableTennis")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type taiChi; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("taiChi")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type taiChi; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("taiChi")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type tennis; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("tennis")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type tennis; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("tennis")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type trackAndField; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("trackAndField")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type trackAndField; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("trackAndField")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type traditionalStrengthTraining; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("traditionalStrengthTraining")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type traditionalStrengthTraining; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("traditionalStrengthTraining")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type volleyball; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("volleyball")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type volleyball; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("volleyball")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type walking; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("walking")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type walking; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("walking")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type waterFitness; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("waterFitness")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type waterFitness; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("waterFitness")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type waterPolo; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("waterPolo")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type waterPolo; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("waterPolo")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type waterSports; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("waterSports")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type waterSports; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("waterSports")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type wheelchairRunPace; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("wheelchairRunPace")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type wheelchairRunPace; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("wheelchairRunPace")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type wheelchairWalkPace; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("wheelchairWalkPace")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type wheelchairWalkPace; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("wheelchairWalkPace")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type wrestling; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("wrestling")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type wrestling; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("wrestling")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type yoga; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("yoga")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type yoga; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.String("yoga")
						datum.ActivityTypeOther = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("distance missing",
					func(datum *physical.Physical) { datum.Distance = nil },
				),
				Entry("distance invalid",
					func(datum *physical.Physical) {
						datum.Distance.Units = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/distance/units", NewMeta()),
				),
				Entry("distance valid",
					func(datum *physical.Physical) { datum.Distance = NewDistance() },
				),
				Entry("duration missing",
					func(datum *physical.Physical) { datum.Duration = nil },
				),
				Entry("duration invalid",
					func(datum *physical.Physical) {
						datum.Duration.Units = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration/units", NewMeta()),
				),
				Entry("duration valid",
					func(datum *physical.Physical) { datum.Duration = NewDuration() },
				),
				Entry("elevation change missing",
					func(datum *physical.Physical) { datum.ElevationChange = nil },
				),
				Entry("elevation change invalid",
					func(datum *physical.Physical) {
						datum.ElevationChange.Units = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/elevationChange/units", NewMeta()),
				),
				Entry("elevation change valid",
					func(datum *physical.Physical) { datum.ElevationChange = NewElevationChange() },
				),
				Entry("energy change missing",
					func(datum *physical.Physical) { datum.Energy = nil },
				),
				Entry("energy change invalid",
					func(datum *physical.Physical) {
						datum.Energy.Units = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/energy/units", NewMeta()),
				),
				Entry("energy change valid",
					func(datum *physical.Physical) { datum.Energy = NewEnergy() },
				),
				Entry("flight missing",
					func(datum *physical.Physical) { datum.Flight = nil },
				),
				Entry("flight invalid",
					func(datum *physical.Physical) {
						datum.Flight.Count = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/flight/count", NewMeta()),
				),
				Entry("flight valid",
					func(datum *physical.Physical) { datum.Flight = NewFlight() },
				),
				Entry("name missing",
					func(datum *physical.Physical) { datum.Name = nil },
				),
				Entry("name empty",
					func(datum *physical.Physical) { datum.Name = pointer.String("") },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/name", NewMeta()),
				),
				Entry("name length; in range (upper)",
					func(datum *physical.Physical) { datum.Name = pointer.String(test.NewText(100, 100)) },
				),
				Entry("name length; out of range (upper)",
					func(datum *physical.Physical) { datum.Name = pointer.String(test.NewText(101, 101)) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/name", NewMeta()),
				),
				Entry("reported intensity missing",
					func(datum *physical.Physical) { datum.ReportedIntensity = nil },
				),
				Entry("reported intensity invalid",
					func(datum *physical.Physical) { datum.ReportedIntensity = pointer.String("invalid") },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"high", "low", "medium"}), "/reportedIntensity", NewMeta()),
				),
				Entry("reported intensity high",
					func(datum *physical.Physical) { datum.ReportedIntensity = pointer.String("high") },
				),
				Entry("reported intensity low",
					func(datum *physical.Physical) { datum.ReportedIntensity = pointer.String("low") },
				),
				Entry("reported intensity medium",
					func(datum *physical.Physical) { datum.ReportedIntensity = pointer.String("medium") },
				),
				Entry("step missing",
					func(datum *physical.Physical) { datum.Flight = nil },
				),
				Entry("step invalid",
					func(datum *physical.Physical) {
						datum.Step.Count = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/step/count", NewMeta()),
				),
				Entry("step valid",
					func(datum *physical.Physical) { datum.Flight = NewFlight() },
				),
				Entry("multiple errors",
					func(datum *physical.Physical) {
						datum.Type = "invalidType"
						datum.ActivityType = pointer.String("invalid")
						datum.ActivityTypeOther = pointer.String(test.NewText(1, 100))
						datum.Distance.Units = nil
						datum.Duration.Units = nil
						datum.ElevationChange.Units = nil
						datum.Name = pointer.String("")
						datum.ReportedIntensity = pointer.String("invalid")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "physicalActivity"), "/type", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", physical.ActivityTypes()), "/activityType", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/distance/units", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration/units", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/elevationChange/units", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/name", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"high", "low", "medium"}), "/reportedIntensity", &types.Meta{Type: "invalidType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *physical.Physical)) {
					for _, origin := range structure.Origins() {
						datum := NewPhysical()
						mutator(datum)
						expectedDatum := ClonePhysical(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *physical.Physical) {},
				),
				Entry("does not modify the datum; activity type missing",
					func(datum *physical.Physical) { datum.ActivityType = nil },
				),
				Entry("does not modify the datum; activity type other missing",
					func(datum *physical.Physical) { datum.ActivityTypeOther = nil },
				),
				Entry("does not modify the datum; aggregate missing",
					func(datum *physical.Physical) { datum.Aggregate = nil },
				),
				Entry("does not modify the datum; distance missing",
					func(datum *physical.Physical) { datum.Distance = nil },
				),
				Entry("does not modify the datum; duration missing",
					func(datum *physical.Physical) { datum.Duration = nil },
				),
				Entry("does not modify the datum; elevation change missing",
					func(datum *physical.Physical) { datum.ElevationChange = nil },
				),
				Entry("does not modify the datum; energy missing",
					func(datum *physical.Physical) { datum.Energy = nil },
				),
				Entry("does not modify the datum; flight missing",
					func(datum *physical.Physical) { datum.Flight = nil },
				),
				Entry("does not modify the datum; name missing",
					func(datum *physical.Physical) { datum.Name = nil },
				),
				Entry("does not modify the datum; reported intensity missing",
					func(datum *physical.Physical) { datum.ReportedIntensity = nil },
				),
				Entry("does not modify the datum; step missing",
					func(datum *physical.Physical) { datum.Step = nil },
				),
			)
		})
	})
})
