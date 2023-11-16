package test

import (
	"fmt"
	"time"

	dataBloodGlucoseTest "github.com/tidepool-org/platform/data/blood/glucose/test"
	"github.com/tidepool-org/platform/data/types/settings/pump"

	"github.com/google/uuid"

	"github.com/tidepool-org/platform/data/blood/glucose"

	userTest "github.com/tidepool-org/platform/user/test"

	"syreclabs.com/go/faker/locales"

	"syreclabs.com/go/faker"

	pumpTest "github.com/tidepool-org/platform/data/types/settings/pump/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/prescription"
	"github.com/tidepool-org/platform/test"
)

const (
	minBgTarget = 60.0
	maxBgTarget = 180.0
)

func RandomPrescriptions(count int) prescription.Prescriptions {
	prescriptions := make(prescription.Prescriptions, count)
	for i := 0; i < count; i++ {
		prescr := RandomPrescription()

		createdTime := prescr.CreatedTime.Add(time.Hour * time.Duration(i))
		modifiedTime := createdTime.Add(time.Hour * time.Duration(i))

		prescr.CreatedTime = createdTime
		prescr.ModifiedTime = modifiedTime
		prescr.LatestRevision.Attributes.CreatedTime = modifiedTime
		prescr.RevisionHistory[0].Attributes.CreatedTime = modifiedTime

		prescriptions[i] = prescr
	}

	return prescriptions
}

func RandomPrescription() *prescription.Prescription {
	create := RandomRevisionCreate()
	return prescription.NewPrescription(create)
}

func RandomClaimedPrescription() *prescription.Prescription {
	create := RandomRevisionCreate()
	prescr := prescription.NewPrescription(create)
	prescr.AccessCode = ""
	prescr.PatientUserID = userTest.RandomID()
	prescr.State = prescription.StateClaimed

	return prescr
}

func RandomRevisionCreate() *prescription.RevisionCreate {
	userID := userTest.RandomID()
	dataAttributes := RandomAttribtues().DataAttributes
	create := &prescription.RevisionCreate{
		ClinicID:       faker.Number().Hexadecimal(24),
		ClinicianID:    userID,
		CreatedUserID:  userID,
		DataAttributes: dataAttributes,
	}
	ResetRevisionCreateHash(create)
	return create
}

func ResetRevisionCreateHash(create *prescription.RevisionCreate) {
	attrs := prescription.NewIntegrityAttributesFromRevisionCreate(*create)
	hash := prescription.MustGenerateIntegrityHash(attrs)
	create.RevisionHash = hash.Hash
}

func RandomRevision() *prescription.Revision {
	revision := &prescription.Revision{
		RevisionID: faker.RandomInt(0, 10),
		Attributes: RandomAttribtues(),
	}
	ResetRevisionHash(revision)
	return revision
}

func ResetRevisionHash(revision *prescription.Revision) {
	attrs := prescription.NewIntegrityAttributesFromRevision(*revision)
	hash := prescription.MustGenerateIntegrityHash(attrs)
	revision.IntegrityHash = &hash
}

func RandomAttribtues() *prescription.Attributes {
	accountType := faker.RandomChoice(prescription.AccountTypes())
	caregiverFirstName := ""
	caregiverLastName := ""
	if accountType == prescription.AccountTypeCaregiver {
		caregiverFirstName = faker.Name().FirstName()
		caregiverLastName = faker.Name().LastName()
	}
	return &prescription.Attributes{
		DataAttributes: prescription.DataAttributes{
			AccountType:             pointer.FromString(accountType),
			CaregiverFirstName:      pointer.FromString(caregiverFirstName),
			CaregiverLastName:       pointer.FromString(caregiverLastName),
			FirstName:               pointer.FromString(faker.Name().FirstName()),
			LastName:                pointer.FromString(faker.Name().LastName()),
			Birthday:                pointer.FromString(faker.Date().Birthday(7, 80).Format("2006-01-02")),
			MRN:                     pointer.FromString(faker.Code().Rut()),
			Email:                   pointer.FromString(faker.Internet().Email()),
			Sex:                     pointer.FromString(RandomSex()),
			Weight:                  RandomWeight(),
			YearOfDiagnosis:         pointer.FromInt(faker.RandomInt(1940, 2020)),
			PhoneNumber:             RandomPhoneNumber(),
			InitialSettings:         RandomInitialSettings(),
			Calculator:              RandomCalculator(),
			Training:                pointer.FromString(RandomTraining()),
			TherapySettings:         pointer.FromString(RandomTherapySettings()),
			PrescriberTermsAccepted: pointer.FromBool(true),
			State:                   prescription.StateSubmitted,
		},
		CreationAttributes: prescription.CreationAttributes{
			CreatedTime:   time.Now(),
			CreatedUserID: userTest.RandomID(),
		},
	}
}

func RandomSex() string {
	return faker.RandomChoice([]string{"male", "female", "undisclosed"})
}

func RandomWeight() *prescription.Weight {
	kgs := float64(faker.RandomInt(10, 100))
	grams := float64(faker.RandomInt(0, 1000))
	weight := kgs + grams/1000.0

	return &prescription.Weight{
		Value: pointer.FromFloat64(weight),
		Units: "kg",
	}
}

func RandomPhoneNumber() *prescription.PhoneNumber {
	faker.Locale = locales.En_US
	return &prescription.PhoneNumber{
		CountryCode: 1,
		Number:      fmt.Sprintf("(%s) %s-%s", faker.PhoneNumber().AreaCode(), faker.PhoneNumber().ExchangeCode(), faker.PhoneNumber().SubscriberNumber(4)),
	}
}

func RandomInitialSettings() *prescription.InitialSettings {
	units := glucose.MgdL
	randomPump := pumpTest.NewPump(&units)
	scheduleName := *randomPump.ActiveScheduleName
	bloodGlucoseSchedule := RandomBloodGlucoseTargetSchedule()

	return &prescription.InitialSettings{
		BloodGlucoseUnits:                  units,
		BasalRateSchedule:                  randomPump.BasalRateSchedules.Get(scheduleName),
		GlucoseSafetyLimit:                 randomPump.BloodGlucoseSafetyLimit,
		BloodGlucoseTargetSchedule:         bloodGlucoseSchedule,
		BloodGlucoseTargetPreprandial:      PreprandialBloodGlucoseTarget(bloodGlucoseSchedule),
		BloodGlucoseTargetPhysicalActivity: PhysicalActivityBloodGlucoseTarget(bloodGlucoseSchedule),
		CarbohydrateRatioSchedule:          randomPump.CarbohydrateRatioSchedules.Get(scheduleName),
		InsulinModel:                       RandomInsulinModel(),
		InsulinSensitivitySchedule:         randomPump.InsulinSensitivitySchedules.Get(scheduleName),
		BasalRateMaximum:                   randomPump.Basal.RateMaximum,
		BolusAmountMaximum:                 randomPump.Bolus.AmountMaximum,
		PumpID:                             pointer.FromString(RandomDeviceID()),
		CgmID:                              pointer.FromString(RandomDeviceID()),
	}
}

func RandomCalculator() *prescription.Calculator {
	return &prescription.Calculator{
		Method:                        pointer.FromString(test.RandomStringFromArray(prescription.AllowedCalculatorMethods())),
		RecommendedBasalRate:          pointer.FromFloat64(test.RandomFloat64FromRange(0, 100)),
		RecommendedCarbohydrateRatio:  pointer.FromFloat64(test.RandomFloat64FromRange(0, 100)),
		RecommendedInsulinSensitivity: pointer.FromFloat64(test.RandomFloat64FromRange(0, 100)),
		TotalDailyDose:                pointer.FromFloat64(test.RandomFloat64FromRange(0, 100)),
		TotalDailyDoseScaleFactor:     pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
		Weight:                        pointer.FromFloat64(test.RandomFloat64FromRange(0, 100)),
		WeightUnits:                   pointer.FromString(test.RandomStringFromArray(prescription.AllowedCalculatorWeightUnits())),
	}
}

func PreprandialBloodGlucoseTarget(schedule *pump.BloodGlucoseTargetStartArray) *glucose.Target {
	bounds := schedule.GetBounds()
	low := test.RandomFloat64FromRange(minBgTarget, bounds.Upper)
	high := test.RandomFloat64FromRange(low, bounds.Upper)
	return &glucose.Target{
		Low:  &low,
		High: &high,
	}
}

func PhysicalActivityBloodGlucoseTarget(schedule *pump.BloodGlucoseTargetStartArray) *glucose.Target {
	bounds := schedule.GetBounds()
	low := test.RandomFloat64FromRange(bounds.Upper, maxBgTarget)
	high := test.RandomFloat64FromRange(low, maxBgTarget)
	return &glucose.Target{
		Low:  &low,
		High: &high,
	}
}

func RandomBloodGlucoseTargetSchedule() *pump.BloodGlucoseTargetStartArray {
	startMinimum := pump.BloodGlucoseTargetStartStartMinimum
	schedule := pump.NewBloodGlucoseTargetStartArray()
	for count := test.RandomIntFromRange(1, 3); count > 0; count-- {
		datum := RandomBloodGlucoseTargetStart(startMinimum)
		*schedule = append(*schedule, datum)
		startMinimum = *datum.Start + 1
	}
	return schedule
}

func RandomBloodGlucoseTargetStart(startMinimum int) *pump.BloodGlucoseTargetStart {
	datum := pump.NewBloodGlucoseTargetStart()
	datum.Target = *dataBloodGlucoseTest.RandomLowHighTarget(minBgTarget, maxBgTarget)
	if startMinimum == pump.BloodGlucoseTargetStartStartMinimum {
		datum.Start = pointer.FromInt(pump.BloodGlucoseTargetStartStartMinimum)
	} else {
		datum.Start = pointer.FromInt(test.RandomIntFromRange(startMinimum, pump.BloodGlucoseTargetStartStartMaximum))
	}
	return datum
}

func RandomInsulinModel() *string {
	validInsulinTypes := []string{
		pump.InsulinModelModelTypeRapidAdult,
		pump.InsulinModelModelTypeRapidChild,
	}
	return pointer.FromString(test.RandomStringFromArray(validInsulinTypes))
}

func RandomDeviceID() string {
	return uuid.New().String()
}

func RandomTraining() string {
	return faker.RandomChoice([]string{
		prescription.TrainingInPerson,
		prescription.TrainingInModule,
	})
}

func RandomTherapySettings() string {
	return faker.RandomChoice([]string{
		prescription.TherapySettingInitial,
		prescription.TherapySettingTransferPumpSettings,
	})
}
