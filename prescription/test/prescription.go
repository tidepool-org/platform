package test

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/tidepool-org/platform/data/blood/glucose"

	userTest "github.com/tidepool-org/platform/user/test"

	"syreclabs.com/go/faker/locales"

	"syreclabs.com/go/faker"

	"github.com/tidepool-org/platform/data/types/settings/pump/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/prescription"
)

func RandomPrescriptions(count int) prescription.Prescriptions {
	prescriptions := make(prescription.Prescriptions, count)
	for i := 0; i < count; i++ {
		prescr := RandomPrescription()

		createdTime := prescr.CreatedTime.Add(time.Second*time.Duration(i) - time.Hour)
		modifiedTime := prescr.LatestRevision.Attributes.CreatedTime.Add(time.Second * time.Duration(i))

		prescr.CreatedTime = createdTime
		prescr.LatestRevision.Attributes.CreatedTime = modifiedTime
		prescr.RevisionHistory[0].Attributes.CreatedTime = modifiedTime

		prescriptions[i] = prescr
	}

	return prescriptions
}

func RandomPrescription() *prescription.Prescription {
	create := RandomRevisionCreate()
	return prescription.NewPrescription(userTest.RandomID(), create)
}

func RandomClaimedPrescription() *prescription.Prescription {
	create := RandomRevisionCreate()
	prescr := prescription.NewPrescription(userTest.RandomID(), create)
	prescr.AccessCode = ""
	prescr.PatientUserID = userTest.RandomID()
	prescr.State = prescription.StateReviewed

	return prescr
}

func RandomRevisionCreate() *prescription.RevisionCreate {
	return &prescription.RevisionCreate{
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
		Training:                RandomTraining(),
		TherapySettings:         RandomTherapySettings(),
		PrescriberTermsAccepted: true,
		State:                   prescription.StateSubmitted,
	}
}

func RandomRevision() *prescription.Revision {
	return &prescription.Revision{
		RevisionID: faker.RandomInt(0, 10),
		Signature:  nil,
		Attributes: RandomAttribtues(),
	}
}

func RandomAttribtues() *prescription.Attributes {
	return &prescription.Attributes{
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
		Training:                RandomTraining(),
		TherapySettings:         RandomTherapySettings(),
		PrescriberTermsAccepted: true,
		State:                   prescription.StateSubmitted,
		CreatedTime:             time.Now(),
		CreatedUserID:           userTest.RandomID(),
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
	units := faker.RandomChoice(glucose.Units())
	randomPump := test.NewPump(&units)
	scheduleName := *randomPump.ActiveScheduleName

	return &prescription.InitialSettings{
		BloodGlucoseUnits:          units,
		BasalRateSchedule:          randomPump.BasalRateSchedules.Get(scheduleName),
		BloodGlucoseTargetSchedule: randomPump.BloodGlucoseTargetSchedules.Get(scheduleName),
		CarbohydrateRatioSchedule:  randomPump.CarbohydrateRatioSchedules.Get(scheduleName),
		InsulinSensitivitySchedule: randomPump.InsulinSensitivitySchedules.Get(scheduleName),
		BasalRateMaximum:           randomPump.Basal.RateMaximum,
		BolusAmountMaximum:         randomPump.Bolus.AmountMaximum,
		PumpID:                     RandomDeviceID(),
		CgmID:                      RandomDeviceID(),
	}
}

func RandomDeviceID() *primitive.ObjectID {
	id := primitive.NewObjectID()
	return &id
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
