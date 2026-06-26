package test

import (
	bsonPrimitive "go.mongodb.org/mongo-driver/bson/primitive"
	"syreclabs.com/go/faker"

	clinicClient "github.com/tidepool-org/clinic/client"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func NewRandomClinic() clinicClient.ClinicV1 {
	return clinicClient.ClinicV1{
		Address:          pointer.FromAny(faker.Address().StreetAddress()),
		CanMigrate:       pointer.FromAny(test.RandomBool()),
		City:             pointer.FromAny(faker.Address().City()),
		ClinicType:       pointer.FromAny(test.RandomChoice([]clinicClient.ClinicV1ClinicType{clinicClient.ClinicV1ClinicTypeHealthcareSystem, clinicClient.ClinicV1ClinicTypeVeterinaryClinic, clinicClient.ClinicV1ClinicTypeOther})),
		Country:          pointer.FromAny(faker.Address().Country()),
		CreatedTime:      pointer.FromAny(test.RandomTime()),
		Id:               pointer.FromAny(bsonPrimitive.NewObjectIDFromTimestamp(test.RandomTime()).Hex()),
		Name:             faker.Company().Name(),
		PhoneNumbers:     pointer.FromAny([]clinicClient.PhoneNumberV1{{Number: faker.PhoneNumber().PhoneNumber()}}),
		PostalCode:       pointer.FromAny(faker.Address().ZipCode()),
		PreferredBgUnits: test.RandomChoice([]clinicClient.ClinicV1PreferredBgUnits{clinicClient.ClinicV1PreferredBgUnitsMgdL, clinicClient.ClinicV1PreferredBgUnitsMmolL}),
		ShareCode:        pointer.FromAny(faker.RandomString(15)),
		State:            pointer.FromAny(faker.Address().State()),
		Tier:             pointer.FromAny(test.RandomChoice([]string{"tier1000", "tier2000"})),
		TierDescription:  pointer.FromAny(faker.Lorem().Sentence(5)),
		UpdatedTime:      pointer.FromAny(test.RandomTime()),
		Website:          pointer.FromAny(faker.Internet().Url()),
	}
}

func NewRandomEHRSettings() *clinicClient.EhrSettingsV1 {
	return &clinicClient.EhrSettingsV1{
		DestinationIds: &clinicClient.EhrDestinationsV1{
			Flowsheet: faker.RandomString(16),
			Notes:     faker.RandomString(16),
			Results:   faker.RandomString(16),
		},
		Enabled:   true,
		MrnIdType: "MRN",
		ProcedureCodes: clinicClient.EhrProceduresV1{
			CreateAccount:                 pointer.FromAny(faker.RandomString(5)),
			CreateAccountAndEnableReports: pointer.FromAny(faker.RandomString(5)),
			DisableSummaryReports:         pointer.FromAny(faker.RandomString(5)),
			EnableSummaryReports:          pointer.FromAny(faker.RandomString(5)),
		},
		Provider: "redox",
		ScheduledReports: clinicClient.ScheduledReportsV1{
			Cadence:               clinicClient.N14d,
			OnUploadEnabled:       true,
			OnUploadNoteEventType: pointer.FromAny(clinicClient.ScheduledReportsV1OnUploadNoteEventTypeNew),
		},
		SourceId: faker.RandomString(16),
	}
}
