package test

import (
	"fmt"

	"syreclabs.com/go/faker"

	"github.com/tidepool-org/platform/data/types/settings/cgm"
	cgmTest "github.com/tidepool-org/platform/data/types/settings/cgm/test"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/data/types/settings/pump/test"
	"github.com/tidepool-org/platform/device"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/prescription"
)

func RandomRevisionCreate() *prescription.RevisionCreate {
	return &prescription.RevisionCreate{
		FirstName:               faker.Name().FirstName(),
		LastName:                faker.Name().LastName(),
		Birthday:                faker.Date().Birthday(7, 80).Format("2020-03-19"),
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
		PrescriberTermsAccepted: pointer.FromBool(true),
		State:                   RandomRevisionState(),
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
	return fmt.Sprintf("(%s) %s-%s", faker.PhoneNumber().AreaCode(), faker.PhoneNumber().ExchangeCode(), faker.PhoneNumber().SubscriberNumber(4))
}

func RandomAddress() *prescription.Address {
	return &prescription.Address{
		Line1:      faker.Address().StreetAddress(),
		Line2:      faker.Address().SecondaryAddress(),
		City:       faker.Address().City(),
		State:      faker.Address().State(),
		PostalCode: faker.Address().Postcode(),
		Country:    "us",
	}
}

func RandomInitialSettings() *prescription.InitialSettings {
	units := pointer.FromString("mg/dL")
	randomPump := test.NewPump(units)
	randomCGM := cgmTest.RandomCGM(units)

	return &prescription.InitialSettings{
		BasalRateSchedule:          randomPump.BasalRateSchedule,
		BloodGlucoseTargetSchedule: randomPump.BloodGlucoseTargetSchedule,
		CarbohydrateRatioSchedule:  randomPump.CarbohydrateRatioSchedule,
		InsulinSensitivitySchedule: randomPump.InsulinSensitivitySchedule,
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

func RandomRevisionState() string {
	return faker.RandomChoice([]string{
		prescription.StateDraft,
		prescription.StatePending,
		prescription.StateSubmitted,
	})
}
