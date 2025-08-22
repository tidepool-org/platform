package test

import (
	"syreclabs.com/go/faker"

	"github.com/tidepool-org/platform/consent"
	"github.com/tidepool-org/platform/pointer"
)

func RandomRecordCreate() *consent.RecordCreate {
	ageGroups := consent.AgeGroups()
	ageGroup := ageGroups[faker.RandomInt(0, len(ageGroups)-1)]

	create := consent.NewRecordCreate()
	create.AgeGroup = ageGroup
	create.OwnerName = faker.Name().Name()
	create.Type = "big_data_donation_project"
	create.Version = faker.RandomInt(1, 10)

	if ageGroup != consent.AgeGroupEighteenOrOver {
		create.ParentGuardianName = pointer.FromString(faker.Name().Name())
		create.GrantorType = consent.GrantorTypeParentGuardian
	} else {
		create.GrantorType = consent.GrantorTypeOwner
	}

	orgs := consent.BigDataDonationProjectOrganizations()
	selectedOrgs := make([]consent.BigDataDonationProjectOrganization, 0)
	numOrgs := faker.RandomInt(1, 3)
	for i := 0; i < numOrgs; i++ {
		org := orgs[faker.RandomInt(0, len(orgs)-1)]
		selectedOrgs = append(selectedOrgs, org)
	}
	create.Metadata = &consent.RecordMetadata{
		SupportedOrganizations: selectedOrgs,
	}

	return create
}
