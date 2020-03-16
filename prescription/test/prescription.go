package test

import (
	"fmt"
	"strconv"

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
		FirstName:               pointer.FromString(faker.Name().FirstName()),
		LastName:                pointer.FromString(faker.Name().LastName()),
		Birthday:                pointer.FromString(faker.Date().Birthday(7, 80).Format("")),
		MRN:                     pointer.FromString(faker.Code().Rut()),
		Email:                   pointer.FromString(faker.Internet().Email()),
		Sex:                     RandomSex(),
		Weight:                  RandomWeight(),
		YearOfDiagnosis:         pointer.FromString(strconv.Itoa(faker.RandomInt(1940, 2020))),
		PhoneNumber:             RandomPhoneNumber(),
		Address:                 RandomAddress(),
		InitialSettings:         RandomInitialSettings(),
		Training:                RandomTraining(),
		TherapySettings:         RandomTherapySettings(),
		LoopMode:                RandomLoopMode(),
		PrescriberTermsAccepted: pointer.FromBool(true),
		State:                   RandomState(),
	}
}

func RandomSex() *string {
	return pointer.FromString(faker.RandomChoice([]string{"male", "female", "undisclosed"}))
}

func RandomWeight() *prescription.Weight {
	kgs := float64(faker.RandomInt(10, 100))
	grams := float64(faker.RandomInt(0, 1000))
	weight := kgs + grams/1000.0

	return &prescription.Weight{
		Value: weight,
		Units: "kg",
	}
}

func RandomPhoneNumber() *string {
	return pointer.FromString(fmt.Sprintf("%s-%s-%s", faker.PhoneNumber().AreaCode(), faker.PhoneNumber().ExchangeCode(), faker.PhoneNumber().SubscriberNumber(4)))
}

func RandomAddress() *prescription.Address {
	return &prescription.Address{
		Line1:      pointer.FromString(faker.Address().StreetAddress()),
		Line2:      pointer.FromString(faker.Address().SecondaryAddress()),
		City:       pointer.FromString(faker.Address().City()),
		State:      pointer.FromString(faker.Address().State()),
		PostalCode: pointer.FromString(faker.Address().Postcode()),
		Country:    pointer.FromString("USA"),
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

func RandomTraining() *string {
	return pointer.FromString(faker.RandomChoice([]string{
		prescription.PrescriptionTrainingInPerson,
		prescription.PrescriptionTrainingInModule,
	}))
}

func RandomTherapySettings() *string {
	return pointer.FromString(faker.RandomChoice([]string{
		prescription.PrescriptionTherapySettingInitial,
		prescription.PrescriptionTherapySettingTransferPumpSettings,
		prescription.PrescriptionTherapySettingCertifiedPumpTrainer,
	}))
}

func RandomLoopMode() *string {
	return pointer.FromString(faker.RandomChoice([]string{
		prescription.PrescriptionLoopModeSuspendOnly,
		prescription.PrescriptionLoopModeClosedLoop,
	}))
}

func RandomState() string {
	return faker.RandomChoice([]string{
		prescription.PrescriptionStateDraft,
		prescription.PrescriptionStatePending,
		prescription.PrescriptionStateSubmitted,
		prescription.PrescriptionStateActive,
		prescription.PrescriptionStateInactive,
		prescription.PrescriptionStateExpired,
	})
}
