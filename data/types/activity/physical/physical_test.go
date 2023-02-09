package physical_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/activity/physical"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
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
	datum.Base = *dataTypesTest.RandomBase()
	datum.Type = "physicalActivity"
	datum.ActivityType = pointer.FromString(test.RandomStringFromArray(physical.ActivityTypes()))
	if datum.ActivityType != nil && *datum.ActivityType == physical.ActivityTypeOther {
		datum.ActivityTypeOther = pointer.FromString(test.RandomStringFromRange(1, 100))
	}
	datum.Aggregate = pointer.FromBool(test.RandomBool())
	datum.Distance = NewDistance()
	datum.Duration = NewDuration()
	datum.ElevationChange = NewElevationChange()
	datum.Energy = NewEnergy()
	datum.Flight = NewFlight()
	datum.Lap = NewLap()
	datum.Name = pointer.FromString(test.RandomStringFromRange(1, 100))
	datum.ReportedIntensity = pointer.FromString(test.RandomStringFromArray(physical.ReportedIntensities()))
	datum.Step = NewStep()
	return datum
}

func ClonePhysical(datum *physical.Physical) *physical.Physical {
	if datum == nil {
		return nil
	}
	clone := physical.New()
	clone.Base = *dataTypesTest.CloneBase(&datum.Base)
	clone.ActivityType = pointer.CloneString(datum.ActivityType)
	clone.ActivityTypeOther = pointer.CloneString(datum.ActivityTypeOther)
	clone.Aggregate = pointer.CloneBool(datum.Aggregate)
	clone.Distance = CloneDistance(datum.Distance)
	clone.Duration = CloneDuration(datum.Duration)
	clone.ElevationChange = CloneElevationChange(datum.ElevationChange)
	clone.Energy = CloneEnergy(datum.Energy)
	clone.Flight = CloneFlight(datum.Flight)
	clone.Lap = CloneLap(datum.Lap)
	clone.Name = pointer.CloneString(datum.Name)
	clone.ReportedIntensity = pointer.CloneString(datum.ReportedIntensity)
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
			Expect(datum.Lap).To(BeNil())
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
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *physical.Physical) {},
				),
				Entry("type missing",
					func(datum *physical.Physical) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &types.Meta{}),
				),
				Entry("type invalid",
					func(datum *physical.Physical) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "physicalActivity"), "/type", &types.Meta{Type: "invalidType"}),
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
						datum.ActivityTypeOther = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type invalid; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("invalid")
						datum.ActivityTypeOther = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", physical.ActivityTypes()), "/activityType", NewMeta()),
				),
				Entry("activity type invalid; activity type other exists",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("invalid")
						datum.ActivityTypeOther = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", physical.ActivityTypes()), "/activityType", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type americanFootball; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("americanFootball")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type americanFootball; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("americanFootball")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type archery; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("archery")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type archery; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("archery")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type australianFootball; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("australianFootball")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type australianFootball; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("australianFootball")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type badminton; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("badminton")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type badminton; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("badminton")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type barre; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("barre")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type barre; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("barre")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type baseball; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("baseball")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type baseball; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("baseball")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type basketball; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("basketball")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type basketball; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("basketball")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type bowling; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("bowling")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type bowling; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("bowling")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type boxing; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("boxing")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type boxing; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("boxing")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type climbing; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("climbing")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type climbing; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("climbing")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type coreTraining; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("coreTraining")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type coreTraining; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("coreTraining")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type cricket; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("cricket")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type cricket; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("cricket")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type crossCountrySkiing; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("crossCountrySkiing")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type crossCountrySkiing; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("crossCountrySkiing")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type crossTraining; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("crossTraining")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type crossTraining; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("crossTraining")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type curling; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("curling")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type curling; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("curling")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type cycling; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("cycling")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type cycling; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("cycling")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type dance; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("dance")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type dance; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("dance")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type danceInspiredTraining; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("danceInspiredTraining")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type danceInspiredTraining; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("danceInspiredTraining")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type downhillSkiing; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("downhillSkiing")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type downhillSkiing; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("downhillSkiing")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type elliptical; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("elliptical")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type elliptical; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("elliptical")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type equestrianSports; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("equestrianSports")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type equestrianSports; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("equestrianSports")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type fencing; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("fencing")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type fencing; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("fencing")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type fishing; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("fishing")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type fishing; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("fishing")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type flexibility; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("flexibility")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type flexibility; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("flexibility")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type functionalStrengthTraining; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("functionalStrengthTraining")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type functionalStrengthTraining; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("functionalStrengthTraining")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type golf; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("golf")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type golf; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("golf")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type gymnastics; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("gymnastics")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type gymnastics; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("gymnastics")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type handball; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("handball")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type handball; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("handball")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type handCycling; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("handCycling")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type handCycling; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("handCycling")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type highIntensityIntervalTraining; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("highIntensityIntervalTraining")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type highIntensityIntervalTraining; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("highIntensityIntervalTraining")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type hiking; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("hiking")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type hiking; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("hiking")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type hockey; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("hockey")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type hockey; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("hockey")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type hunting; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("hunting")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type hunting; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("hunting")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type jumpRope; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("jumpRope")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type jumpRope; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("jumpRope")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type kickboxing; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("kickboxing")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type kickboxing; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("kickboxing")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type lacrosse; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("lacrosse")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type lacrosse; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("lacrosse")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type martialArts; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("martialArts")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type martialArts; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("martialArts")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type mindAndBody; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("mindAndBody")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type mindAndBody; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("mindAndBody")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type mixedCardio; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("mixedCardio")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type mixedCardio; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("mixedCardio")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type mixedMetabolicCardioTraining; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("mixedMetabolicCardioTraining")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type mixedMetabolicCardioTraining; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("mixedMetabolicCardioTraining")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type other; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("other")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type other; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("other")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type other; activity type other length; in range (upper)",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("other")
						datum.ActivityTypeOther = pointer.FromString(test.RandomStringFromRange(100, 100))
					},
				),
				Entry("activity type other; activity type other length; out of range (upper)",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("other")
						datum.ActivityTypeOther = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type paddleSports; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("paddleSports")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type paddleSports; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("paddleSports")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type pilates; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("pilates")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type pilates; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("pilates")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type play; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("play")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type play; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("play")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type preparationAndRecovery; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("preparationAndRecovery")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type preparationAndRecovery; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("preparationAndRecovery")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type racquetball; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("racquetball")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type racquetball; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("racquetball")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type rowing; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("rowing")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type rowing; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("rowing")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type rugby; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("rugby")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type rugby; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("rugby")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type running; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("running")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type running; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("running")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type sailing; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("sailing")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type sailing; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("sailing")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type skatingSports; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("skatingSports")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type skatingSports; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("skatingSports")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type snowboarding; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("snowboarding")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type snowboarding; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("snowboarding")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type snowSports; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("snowSports")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type snowSports; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("snowSports")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type soccer; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("soccer")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type soccer; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("soccer")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type softball; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("softball")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type softball; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("softball")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type squash; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("squash")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type squash; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("squash")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type stairClimbing; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("stairClimbing")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type stairClimbing; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("stairClimbing")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type stairs; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("stairs")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type stairs; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("stairs")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type stepTraining; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("stepTraining")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type stepTraining; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("stepTraining")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type surfingSports; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("surfingSports")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type surfingSports; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("surfingSports")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type swimming; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("swimming")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type swimming; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("swimming")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type tableTennis; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("tableTennis")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type tableTennis; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("tableTennis")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type taiChi; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("taiChi")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type taiChi; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("taiChi")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type tennis; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("tennis")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type tennis; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("tennis")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type trackAndField; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("trackAndField")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type trackAndField; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("trackAndField")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type traditionalStrengthTraining; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("traditionalStrengthTraining")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type traditionalStrengthTraining; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("traditionalStrengthTraining")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type volleyball; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("volleyball")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type volleyball; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("volleyball")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type walking; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("walking")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type walking; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("walking")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type waterFitness; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("waterFitness")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type waterFitness; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("waterFitness")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type waterPolo; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("waterPolo")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type waterPolo; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("waterPolo")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type waterSports; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("waterSports")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type waterSports; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("waterSports")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type wheelchairRunPace; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("wheelchairRunPace")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type wheelchairRunPace; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("wheelchairRunPace")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type wheelchairWalkPace; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("wheelchairWalkPace")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type wheelchairWalkPace; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("wheelchairWalkPace")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type wrestling; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("wrestling")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type wrestling; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("wrestling")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("activity type yoga; activity type other missing",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("yoga")
						datum.ActivityTypeOther = nil
					},
				),
				Entry("activity type yoga; activity type other empty",
					func(datum *physical.Physical) {
						datum.ActivityType = pointer.FromString("yoga")
						datum.ActivityTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", NewMeta()),
				),
				Entry("distance missing",
					func(datum *physical.Physical) { datum.Distance = nil },
				),
				Entry("distance invalid",
					func(datum *physical.Physical) {
						datum.Distance.Units = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/distance/units", NewMeta()),
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
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration/units", NewMeta()),
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
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/elevationChange/units", NewMeta()),
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
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/energy/units", NewMeta()),
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
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/flight/count", NewMeta()),
				),
				Entry("flight valid",
					func(datum *physical.Physical) { datum.Flight = NewFlight() },
				),
				Entry("lap missing",
					func(datum *physical.Physical) { datum.Lap = nil },
				),
				Entry("lap invalid",
					func(datum *physical.Physical) {
						datum.Lap.Count = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/lap/count", NewMeta()),
				),
				Entry("lap valid",
					func(datum *physical.Physical) { datum.Lap = NewLap() },
				),
				Entry("name missing",
					func(datum *physical.Physical) { datum.Name = nil },
				),
				Entry("name empty",
					func(datum *physical.Physical) { datum.Name = pointer.FromString("") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/name", NewMeta()),
				),
				Entry("name length; in range (upper)",
					func(datum *physical.Physical) { datum.Name = pointer.FromString(test.RandomStringFromRange(100, 100)) },
				),
				Entry("name length; out of range (upper)",
					func(datum *physical.Physical) { datum.Name = pointer.FromString(test.RandomStringFromRange(101, 101)) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/name", NewMeta()),
				),
				Entry("reported intensity missing",
					func(datum *physical.Physical) { datum.ReportedIntensity = nil },
				),
				Entry("reported intensity invalid",
					func(datum *physical.Physical) { datum.ReportedIntensity = pointer.FromString("invalid") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"high", "low", "medium"}), "/reportedIntensity", NewMeta()),
				),
				Entry("reported intensity high",
					func(datum *physical.Physical) { datum.ReportedIntensity = pointer.FromString("high") },
				),
				Entry("reported intensity low",
					func(datum *physical.Physical) { datum.ReportedIntensity = pointer.FromString("low") },
				),
				Entry("reported intensity medium",
					func(datum *physical.Physical) { datum.ReportedIntensity = pointer.FromString("medium") },
				),
				Entry("step missing",
					func(datum *physical.Physical) { datum.Step = nil },
				),
				Entry("step invalid",
					func(datum *physical.Physical) {
						datum.Step.Count = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/step/count", NewMeta()),
				),
				Entry("step valid",
					func(datum *physical.Physical) { datum.Step = NewStep() },
				),
				Entry("multiple errors",
					func(datum *physical.Physical) {
						datum.Type = "invalidType"
						datum.ActivityType = pointer.FromString("invalid")
						datum.ActivityTypeOther = pointer.FromString(test.RandomStringFromRange(1, 100))
						datum.Distance.Units = nil
						datum.Duration.Units = nil
						datum.ElevationChange.Units = nil
						datum.Flight.Count = nil
						datum.Lap.Count = nil
						datum.Name = pointer.FromString("")
						datum.ReportedIntensity = pointer.FromString("invalid")
						datum.Step.Count = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "physicalActivity"), "/type", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", physical.ActivityTypes()), "/activityType", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/activityTypeOther", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/distance/units", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration/units", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/elevationChange/units", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/flight/count", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/lap/count", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/name", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"high", "low", "medium"}), "/reportedIntensity", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/step/count", &types.Meta{Type: "invalidType"}),
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
				Entry("does not modify the datum; lap missing",
					func(datum *physical.Physical) { datum.Lap = nil },
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
