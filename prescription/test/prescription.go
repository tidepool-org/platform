package test

import (
	"fmt"
	"time"

	dataTypesSettingsPump "github.com/tidepool-org/platform/data/types/settings/pump"

	dataBloodGlucoseTest "github.com/tidepool-org/platform/data/blood/glucose/test"
	"github.com/tidepool-org/platform/data/types/settings/pump"

	"github.com/google/uuid"

	"github.com/tidepool-org/platform/data/blood/glucose"
	testUtils "github.com/tidepool-org/platform/test"

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
	create := RandomRevisionCreate(userTest.RandomID())
	return prescription.NewPrescription(create)
}

func RandomClaimedPrescription() *prescription.Prescription {
	create := RandomRevisionCreate(userTest.RandomID())
	prescr := prescription.NewPrescription(create)
	prescr.AccessCode = ""
	prescr.PatientUserID = userTest.RandomID()
	prescr.State = prescription.StateClaimed

	return prescr
}

func RandomRevisionCreate(userID string) *prescription.RevisionCreate {
	if userID == "" {
		userID = userTest.RandomID()
	}
	accountType := faker.RandomChoice(prescription.AccountTypes())
	dataAttributes := prescription.DataAttributes{
		AccountType:             accountType,
		FirstName:               faker.Name().FirstName(),
		LastName:                faker.Name().LastName(),
		Birthday:                faker.Date().Birthday(7, 80).Format("2006-01-02"),
		MRN:                     faker.Code().Rut(),
		Email:                   faker.Internet().Email(),
		Sex:                     RandomSex(),
		Weight:                  RandomWeight(),
		YearOfDiagnosis:         faker.RandomInt(1940, 2020),
		PhoneNumber:             RandomPhoneNumber(),
		InitialSettings:         RandomInitialSettings(),
		Calculator:              RandomCalculator(),
		Training:                RandomTraining(),
		TherapySettings:         RandomTherapySettings(),
		PrescriberTermsAccepted: true,
		State:                   prescription.StateSubmitted,
	}
	if accountType == prescription.AccountTypeCaregiver {
		dataAttributes.CaregiverFirstName = faker.Name().FirstName()
		dataAttributes.CaregiverLastName = faker.Name().LastName()
	}
	hash := prescription.MustGenerateIntegrityHash(prescription.IntegrityAttributes{
		DataAttributes: dataAttributes,
		CreatedUserId:  userID,
	})
	return &prescription.RevisionCreate{
		ClinicID:       faker.Number().Hexadecimal(24),
		ClinicianID:    userID,
		CreatedUserId:  userID,
		DataAttributes: dataAttributes,
		RevisionHash:   hash.Hash,
	}
}

func RandomRevision() *prescription.Revision {
	revision := prescription.Revision{
		RevisionID: faker.RandomInt(0, 10),
		Attributes: RandomAttribtues(),
	}
	hash := prescription.MustGenerateIntegrityHash(prescription.NewIntegrityAttributesFromRevision(revision))
	revision.IntegrityHash = &hash
	return &revision
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
			AccountType:             accountType,
			CaregiverFirstName:      caregiverFirstName,
			CaregiverLastName:       caregiverLastName,
			FirstName:               faker.Name().FirstName(),
			LastName:                faker.Name().LastName(),
			Birthday:                faker.Date().Birthday(7, 80).Format("2006-01-02"),
			MRN:                     faker.Code().Rut(),
			Email:                   faker.Internet().Email(),
			Sex:                     RandomSex(),
			Weight:                  RandomWeight(),
			YearOfDiagnosis:         faker.RandomInt(1940, 2020),
			PhoneNumber:             RandomPhoneNumber(),
			InitialSettings:         RandomInitialSettings(),
			Calculator:              RandomCalculator(),
			Training:                RandomTraining(),
			TherapySettings:         RandomTherapySettings(),
			PrescriberTermsAccepted: true,
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
		GlucoseSafetyLimit:                 randomPump.BloodGlucoseSuspendThreshold,
		BloodGlucoseTargetSchedule:         bloodGlucoseSchedule,
		BloodGlucoseTargetPreprandial:      PreprandialBloodGlucoseTarget(bloodGlucoseSchedule),
		BloodGlucoseTargetPhysicalActivity: PhysicalActivityBloodGlucoseTarget(bloodGlucoseSchedule),
		CarbohydrateRatioSchedule:          randomPump.CarbohydrateRatioSchedules.Get(scheduleName),
		InsulinModel:                       RandomInsulinModel(),
		InsulinSensitivitySchedule:         randomPump.InsulinSensitivitySchedules.Get(scheduleName),
		BasalRateMaximum:                   randomPump.Basal.RateMaximum,
		BolusAmountMaximum:                 randomPump.Bolus.AmountMaximum,
		PumpID:                             RandomDeviceID(),
		CgmID:                              RandomDeviceID(),
	}
}

func RandomCalculator() *prescription.Calculator {
	return &prescription.Calculator{
		Method:                        testUtils.RandomStringFromArray(prescription.AllowedCalculatorMethods()),
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
	datum.Target = *dataBloodGlucoseTest.NewLowHighTarget(minBgTarget, maxBgTarget)
	if startMinimum == pump.BloodGlucoseTargetStartStartMinimum {
		datum.Start = pointer.FromInt(pump.BloodGlucoseTargetStartStartMinimum)
	} else {
		datum.Start = pointer.FromInt(test.RandomIntFromRange(startMinimum, pump.BloodGlucoseTargetStartStartMaximum))
	}
	return datum
}

func RandomInsulinModel() *string {
	validInsulinTypes := []string{
		dataTypesSettingsPump.InsulinModelModelTypeRapidAdult,
		dataTypesSettingsPump.InsulinModelModelTypeRapidChild,
	}
	return pointer.FromString(testUtils.RandomStringFromArray(validInsulinTypes))
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
