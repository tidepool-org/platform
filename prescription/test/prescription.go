package test

import (
	"fmt"
	"strings"
	"time"

	"github.com/tidepool-org/platform/data/blood/glucose"

	userTest "github.com/tidepool-org/platform/user/test"

	"syreclabs.com/go/faker/locales"

	"syreclabs.com/go/faker"

	"github.com/tidepool-org/platform/data/types/settings/cgm"
	cgmTest "github.com/tidepool-org/platform/data/types/settings/cgm/test"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/data/types/settings/pump/test"
	"github.com/tidepool-org/platform/device"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/prescription"
)

func RandomPrescriptions(count int) prescription.Prescriptions {
	prescriptions := make(prescription.Prescriptions, count)
	for i := 0; i < count; i++ {
		prescriptions[i] = RandomPrescription()
	}

	return prescriptions
}

func RandomPrescription() *prescription.Prescription {
	create := RandomRevisionCreate()
	return prescription.NewPrescription(userTest.RandomID(), create)
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
		Address:                 RandomAddress(),
		InitialSettings:         RandomInitialSettings(),
		Training:                RandomTraining(),
		TherapySettings:         RandomTherapySettings(),
		LoopMode:                RandomLoopMode(),
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
		Address:                 RandomAddress(),
		InitialSettings:         RandomInitialSettings(),
		Training:                RandomTraining(),
		TherapySettings:         RandomTherapySettings(),
		LoopMode:                RandomLoopMode(),
		PrescriberTermsAccepted: true,
		State:                   prescription.StateSubmitted,
		ModifiedTime:            time.Now(),
		ModifiedUserID:          userTest.RandomID(),
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

func RandomPhoneNumber() string {
	faker.Locale = locales.En_US
	return fmt.Sprintf("(%s) %s-%s", faker.PhoneNumber().AreaCode(), faker.PhoneNumber().ExchangeCode(), faker.PhoneNumber().SubscriberNumber(4))
}

func RandomAddress() *prescription.Address {
	return &prescription.Address{
		Line1:      faker.Address().StreetAddress(),
		Line2:      faker.Address().SecondaryAddress(),
		City:       faker.Address().City(),
		State:      strings.ToUpper(faker.Address().StateAbbr()),
		PostalCode: faker.Address().Postcode(),
		Country:    "US",
	}
}

func RandomInitialSettings() *prescription.InitialSettings {
	units := faker.RandomChoice(glucose.Units())
	randomPump := test.NewPump(&units)
	scheduleName := *randomPump.ActiveScheduleName
	randomCGM := cgmTest.RandomCGM(&units)

	return &prescription.InitialSettings{
		BloodGlucoseUnits:          units,
		BasalRateSchedule:          randomPump.BasalRateSchedules.Get(scheduleName),
		BloodGlucoseTargetSchedule: randomPump.BloodGlucoseTargetSchedules.Get(scheduleName),
		CarbohydrateRatioSchedule:  randomPump.CarbohydrateRatioSchedules.Get(scheduleName),
		InsulinSensitivitySchedule: randomPump.InsulinSensitivitySchedules.Get(scheduleName),
		BasalRateMaximum:           randomPump.Basal.RateMaximum,
		BolusAmountMaximum:         randomPump.Bolus.AmountMaximum,
		PumpType:                   getPumpType(randomPump),
		CGMType:                    getCGMType(randomCGM),
	}
}

func getPumpType(pump *pump.Pump) *device.Device {
	manufacturers := *pump.Manufacturers
	manufacturer := manufacturers[0]

	return &device.Device{
		Type:         device.DeviceTypePump,
		Manufacturer: manufacturer,
		Model:        *pump.Model,
	}
}

func getCGMType(cgm *cgm.CGM) *device.Device {
	manufacturers := *cgm.Manufacturers
	manufacturer := manufacturers[0]

	return &device.Device{
		Type:         device.DeviceTypePump,
		Manufacturer: manufacturer,
		Model:        *cgm.Model,
	}
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
		prescription.TherapySettingCertifiedPumpTrainer,
	})
}

func RandomLoopMode() string {
	return faker.RandomChoice([]string{
		prescription.LoopModeSuspendOnly,
		prescription.LoopModeClosedLoop,
	})
}
