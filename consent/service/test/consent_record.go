package test

import (
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"

	"time"

	"github.com/tidepool-org/platform/consent"
	consentTest "github.com/tidepool-org/platform/consent/test"
)

func RandomRecordCreateForConsent(cons *consent.Consent) *consent.RecordCreate {
	create := consentTest.RandomRecordCreate()
	create.Type = cons.Type
	create.Version = cons.Version
	if cons.Type != consent.TypeBigDataDonationProject {
		create.Metadata.SupportedOrganizations = nil
	}
	return create
}

func MatchConsentRecord(record consent.Record) types.GomegaMatcher {
	fields := Fields{
		"ID":                 Equal(record.ID),
		"UserID":             Equal(record.UserID),
		"Status":             Equal(record.Status),
		"AgeGroup":           Equal(record.AgeGroup),
		"OwnerName":          Equal(record.OwnerName),
		"ParentGuardianName": Equal(record.ParentGuardianName),
		"GrantorType":        Equal(record.GrantorType),
		"Type":               Equal(record.Type),
		"Version":            Equal(record.Version),
		"Metadata":           BeNil(),
		"GrantTime":          BeTemporally("~", record.GrantTime, time.Millisecond),
		"RevocationTime":     BeNil(),
		"CreatedTime":        BeTemporally("~", record.CreatedTime, time.Millisecond),
		"ModifiedTime":       BeTemporally("~", record.ModifiedTime, time.Millisecond),
	}
	if record.Metadata != nil {
		fields["Metadata"] = Equal(record.Metadata)
	}
	if record.RevocationTime != nil {
		fields["RevocationTime"] = BeTemporally("~", *record.RevocationTime, time.Millisecond)
	}
	return MatchFields(IgnoreExtras, fields)
}
