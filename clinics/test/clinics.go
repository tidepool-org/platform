package test

import (
	api "github.com/tidepool-org/clinic/client"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"syreclabs.com/go/faker"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewRandomClinic() api.Clinic {
	return api.Clinic{
		Address:          pointer.FromAny(faker.Address().StreetAddress()),
		CanMigrate:       pointer.FromAny(test.RandomBool()),
		City:             pointer.FromAny(faker.Address().City()),
		ClinicType:       pointer.FromAny(test.RandomChoice([]api.ClinicClinicType{api.HealthcareSystem, api.VeterinaryClinic, api.Other})),
		Country:          pointer.FromAny(faker.Address().Country()),
		CreatedTime:      pointer.FromAny(test.RandomTimeFromRange(test.RandomTimeMinimum(), test.RandomTimeMaximum())),
		Id:               pointer.FromAny(primitive.NewObjectIDFromTimestamp(test.RandomTimeFromRange(test.RandomTimeMinimum(), test.RandomTimeMaximum())).Hex()),
		Name:             faker.Company().Name(),
		PhoneNumbers:     pointer.FromAny([]api.PhoneNumber{{Number: faker.PhoneNumber().PhoneNumber()}}),
		PostalCode:       pointer.FromAny(faker.Address().ZipCode()),
		PreferredBgUnits: test.RandomChoice([]api.ClinicPreferredBgUnits{api.MgdL, api.MmolL}),
		ShareCode:        pointer.FromAny(faker.RandomString(15)),
		State:            pointer.FromAny(faker.Address().State()),
		Tier:             pointer.FromAny(test.RandomChoice([]string{"tier1000", "tier2000"})),
		TierDescription:  pointer.FromAny(faker.Lorem().Sentence(5)),
		UpdatedTime:      pointer.FromAny(test.RandomTimeFromRange(test.RandomTimeMinimum(), test.RandomTimeMaximum())),
		Website:          pointer.FromAny(faker.Internet().Url()),
	}
}
