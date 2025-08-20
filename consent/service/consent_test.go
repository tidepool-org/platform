package service_test

import (
	"context"

	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"

	"time"

	logTest "github.com/tidepool-org/platform/log/test"

	userTest "github.com/tidepool-org/platform/user/test"

	"github.com/tidepool-org/platform/consent"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"go.uber.org/mock/gomock"

	authStoreMongo "github.com/tidepool-org/platform/auth/store/mongo"
	"github.com/tidepool-org/platform/consent/service"
	"github.com/tidepool-org/platform/consent/service/test"
	"github.com/tidepool-org/platform/log"
)

var _ = Describe("ConsentService", func() {
	var logger log.Logger

	var bddpCtrl *gomock.Controller
	var bddp *test.MockBigDataDonationProjectSharer
	var mailerCtrl *gomock.Controller
	var mailer *test.MockConsentMailer

	var store *authStoreMongo.Store
	var consentService consent.Service

	var ctx = func() context.Context {
		return log.NewContextWithLogger(context.Background(), logger)
	}

	BeforeEach(func() {
		t := GinkgoT()
		logger = logTest.NewLogger()

		mailerCtrl = gomock.NewController(t)
		mailer = test.NewMockConsentMailer(mailerCtrl)

		bddpCtrl = gomock.NewController(t)
		bddp = test.NewMockBigDataDonationProjectSharer(bddpCtrl)

		store = GetSuiteStore()

		consentRecordRepository := store.NewConsentRecordRepository()
		consentRepository := store.NewConsentRepository()
		consentService = service.NewConsentService(mailer, bddp, consentRepository, consentRecordRepository, store.GetClient(), logger)

		Expect(consentService.EnsureConsent(ctx(), test.ConsentV1)).To(Succeed())
		Expect(consentService.EnsureConsent(ctx(), test.ConsentV2)).To(Succeed())
		Expect(consentService.EnsureConsent(ctx(), test.AnotherConsentV1)).To(Succeed())
		Expect(consentService.EnsureConsent(ctx(), test.MockBDDPConsentV1)).To(Succeed())
	})

	AfterEach(func() {
		bddpCtrl.Finish()
		mailerCtrl.Finish()
	})

	Describe("ListConsents", func() {
		It("should return only the latest version of the consent when latest filter is set to true", func() {
			result, err := consentService.ListConsents(ctx(), &consent.Filter{
				Latest: pointer.FromAny(true),
				Type:   pointer.FromAny("test_consent"),
			}, page.NewPagination())
			Expect(err).To(Not(HaveOccurred()))
			Expect(result.Count).To(Equal(1))
			Expect(result.Data).To(ConsistOf(test.MatchConsent(*test.ConsentV2)))

		})

		It("should return all versions of the consents when latest filter is not set", func() {
			result, err := consentService.ListConsents(ctx(), &consent.Filter{
				//Latest: pointer.FromAny(false),
				Type: pointer.FromAny("test_consent"),
			}, page.NewPagination())
			Expect(err).To(Not(HaveOccurred()))
			Expect(result.Count).To(Equal(2))
			Expect(result.Data).To(HaveLen(2))
			Expect(result.Data).To(ConsistOf(test.MatchConsent(*test.ConsentV2), test.MatchConsent(*test.ConsentV1)))
		})

		It("should return all versions of the consents when latest filter is set to false", func() {
			result, err := consentService.ListConsents(ctx(), &consent.Filter{
				Latest: pointer.FromAny(false),
				Type:   pointer.FromAny("test_consent"),
			}, page.NewPagination())
			Expect(err).To(Not(HaveOccurred()))
			Expect(result.Count).To(Equal(2))
			Expect(result.Data).To(HaveLen(2))
			Expect(result.Data).To(ConsistOf(test.MatchConsent(*test.ConsentV2), test.MatchConsent(*test.ConsentV1)))
		})

		It("should return all consents with an empty filter", func() {
			result, err := consentService.ListConsents(ctx(), &consent.Filter{}, page.NewPagination())
			Expect(err).To(Not(HaveOccurred()))
			Expect(result.Count).To(Equal(4))
			Expect(result.Data).To(HaveLen(4))
			Expect(result.Data).To(ConsistOf(
				test.MatchConsent(*test.ConsentV2),
				test.MatchConsent(*test.ConsentV1),
				test.MatchConsent(*test.AnotherConsentV1),
				test.MatchConsent(*test.MockBDDPConsentV1),
			))
		})

		It("should return correct results with pagination", func() {
			pagination := page.NewPagination()
			pagination.Page = 1
			pagination.Size = 1

			result, err := consentService.ListConsents(ctx(), &consent.Filter{}, pagination)
			Expect(err).To(Not(HaveOccurred()))
			Expect(result.Count).To(Equal(4))
			Expect(result.Data).To(HaveLen(1))
			Expect(result.Data).To(ConsistOf(test.MatchConsent(*test.ConsentV1)))
		})

		It("should return the correct version", func() {
			result, err := consentService.ListConsents(ctx(), &consent.Filter{
				Type:    pointer.FromAny(test.ConsentV1.Type),
				Version: pointer.FromAny(test.ConsentV1.Version),
			}, page.NewPagination())
			Expect(err).To(Not(HaveOccurred()))
			Expect(result.Count).To(Equal(1))
			Expect(result.Data).To(HaveLen(1))
			Expect(result.Data).To(ConsistOf(test.MatchConsent(*test.ConsentV1)))
		})
	})

	Describe("CreateConsentRecord", func() {
		var userID string

		BeforeEach(func() {
			userID = userTest.RandomUserID()
		})

		It("should persist the consent record correctly", func() {
			create := test.RandomRecordCreateForConsent(test.ConsentV2)
			mailer.EXPECT().
				SendConsentGrantedEmailNotification(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			created, err := consentService.CreateConsentRecord(ctx(), userID, create)
			Expect(err).ToNot(HaveOccurred())

			record, err := consentService.GetConsentRecord(ctx(), userID, created.ID)
			Expect(err).ToNot(HaveOccurred())
			Expect(record.ID).To(Equal(created.ID))
			Expect(record.UserID).To(Equal(userID))

			Expect(record.Status).To(Equal(consent.RecordStatusActive))
			Expect(record.AgeGroup).To(Equal(create.AgeGroup))
			Expect(record.OwnerName).To(Equal(create.OwnerName))
			Expect(record.ParentGuardianName).To(Equal(create.ParentGuardianName))
			Expect(record.GrantorType).To(Equal(create.GrantorType))
			Expect(record.Type).To(Equal(create.Type))
			Expect(record.Version).To(Equal(create.Version))
			Expect(record.GrantTime).To(BeTemporally("~", time.Now(), time.Minute))
			Expect(record.RevocationTime).To(BeNil())
			Expect(record.CreatedTime).To(BeTemporally("~", time.Now(), time.Minute))
			Expect(record.ModifiedTime).To(BeTemporally("~", time.Now(), time.Minute))
		})

		It("should send an email with the correct consent and record", func() {
			create := test.RandomRecordCreateForConsent(test.ConsentV2)
			mailer.EXPECT().
				SendConsentGrantedEmailNotification(gomock.Any(), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, cons consent.Consent, record consent.Record) {
					Expect(cons.Type).To(Equal(test.ConsentV2.Type))
					Expect(cons.Version).To(Equal(test.ConsentV2.Version))
					Expect(cons.Content).To(Equal(test.ConsentV2.Content))
					Expect(record.UserID).To(Equal(userID))
					Expect(record.Type).To(Equal(create.Type))
					Expect(record.Version).To(Equal(create.Version))
				}).Return(nil)

			_, err := consentService.CreateConsentRecord(ctx(), userID, create)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should share the user's account with BDDP recipient", func() {
			create := test.RandomRecordCreateForConsent(test.MockBDDPConsentV1)

			mailer.EXPECT().SendConsentGrantedEmailNotification(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			bddp.EXPECT().Share(gomock.Any(), userID).Return(nil)

			_, err := consentService.CreateConsentRecord(ctx(), userID, create)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return an error if the consent type is invalid", func() {
			create := test.RandomRecordCreateForConsent(test.ConsentV2)
			create.Type = "invalid"

			_, err := consentService.CreateConsentRecord(ctx(), userID, create)
			Expect(err).To(HaveOccurred())
		})

		It("should return an error if the consent version is invalid", func() {
			create := test.RandomRecordCreateForConsent(test.ConsentV2)
			create.Version = 3

			_, err := consentService.CreateConsentRecord(ctx(), userID, create)
			Expect(err).To(HaveOccurred())
		})

		It("should return an error if consent with the same version already exists", func() {
			create := test.RandomRecordCreateForConsent(test.ConsentV2)

			mailer.EXPECT().
				SendConsentGrantedEmailNotification(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			_, err := consentService.CreateConsentRecord(ctx(), userID, create)
			Expect(err).ToNot(HaveOccurred())

			_, err = consentService.CreateConsentRecord(ctx(), userID, create)
			Expect(err).To(MatchError("consent record for the same type and version already exists"))
		})

		It("should return an error if consent with a greater version already exists", func() {
			create := test.RandomRecordCreateForConsent(test.ConsentV2)

			mailer.EXPECT().
				SendConsentGrantedEmailNotification(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			_, err := consentService.CreateConsentRecord(ctx(), userID, create)
			Expect(err).ToNot(HaveOccurred())

			create.Version = test.ConsentV1.Version
			_, err = consentService.CreateConsentRecord(ctx(), userID, create)
			Expect(err).To(MatchError("consent record for a greater version already exists"))
		})

		It("should revoke consents with a lower version of the same type", func() {
			create := test.RandomRecordCreateForConsent(test.ConsentV1)

			mailer.EXPECT().
				SendConsentGrantedEmailNotification(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			v1, err := consentService.CreateConsentRecord(ctx(), userID, create)
			Expect(err).ToNot(HaveOccurred())
			Expect(v1).ToNot(BeNil())

			create.Version = test.ConsentV2.Version

			mailer.EXPECT().
				SendConsentGrantedEmailNotification(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			v2, err := consentService.CreateConsentRecord(ctx(), userID, create)
			Expect(err).To(Not(HaveOccurred()))
			Expect(v2).ToNot(BeNil())

			record, err := consentService.GetConsentRecord(ctx(), userID, v1.ID)
			Expect(err).ToNot(HaveOccurred())
			Expect(record.Status).To(Equal(consent.RecordStatusRevoked))
			Expect(record.RevocationTime).To(PointTo(BeTemporally("~", time.Now(), time.Minute)))
			Expect(record.ModifiedTime).ToNot(Equal(v1.ModifiedTime))
		})
	})

	Describe("ListConsentRecords", func() {
		var userID string

		BeforeEach(func() {
			userID = userTest.RandomUserID()

			creates := []*consent.RecordCreate{
				test.RandomRecordCreateForConsent(test.ConsentV1),
				test.RandomRecordCreateForConsent(test.ConsentV2),
				test.RandomRecordCreateForConsent(test.AnotherConsentV1),
			}
			for i, create := range creates {
				mailer.EXPECT().
					SendConsentGrantedEmailNotification(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

				create.CreatedTime = create.CreatedTime.Add(-time.Duration(len(creates)-i) * time.Second)
				created, err := consentService.CreateConsentRecord(ctx(), userID, create)
				Expect(err).ToNot(HaveOccurred())
				Expect(created).ToNot(BeNil())
			}
		})

		It("should return all consent records", func() {
			result, err := consentService.ListConsentRecords(ctx(), userID, &consent.RecordFilter{
				Latest: pointer.FromAny(false),
			}, page.NewPagination())
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Count).To(Equal(3))
			Expect(result.Data).To(HaveExactElements(
				MatchFields(IgnoreExtras, Fields{
					"Type":    Equal(test.AnotherConsentV1.Type),
					"Version": Equal(test.AnotherConsentV1.Version),
					"Status":  Equal(consent.RecordStatusActive),
				}),
				MatchFields(IgnoreExtras, Fields{
					"Type":    Equal(test.ConsentV2.Type),
					"Version": Equal(test.ConsentV2.Version),
					"Status":  Equal(consent.RecordStatusActive),
				}),
				MatchFields(IgnoreExtras, Fields{
					"Type":    Equal(test.ConsentV1.Type),
					"Version": Equal(test.ConsentV1.Version),
					"Status":  Equal(consent.RecordStatusRevoked),
				}),
			))
		})

		It("should filter by type", func() {
			result, err := consentService.ListConsentRecords(ctx(), userID, &consent.RecordFilter{
				Latest: pointer.FromAny(false),
				Type:   pointer.FromAny(test.ConsentType),
			}, page.NewPagination())
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Count).To(Equal(2))
			Expect(result.Data).To(HaveExactElements(
				MatchFields(IgnoreExtras, Fields{
					"Type":    Equal(test.ConsentV2.Type),
					"Version": Equal(test.ConsentV2.Version),
					"Status":  Equal(consent.RecordStatusActive),
				}),
				MatchFields(IgnoreExtras, Fields{
					"Type":    Equal(test.ConsentV1.Type),
					"Version": Equal(test.ConsentV1.Version),
					"Status":  Equal(consent.RecordStatusRevoked),
				}),
			))
		})

		It("should filter by type and version", func() {
			result, err := consentService.ListConsentRecords(ctx(), userID, &consent.RecordFilter{
				Latest:  pointer.FromAny(false),
				Type:    pointer.FromAny(test.ConsentType),
				Version: pointer.FromAny(test.ConsentV2.Version),
			}, page.NewPagination())
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Count).To(Equal(1))
			Expect(result.Data).To(HaveExactElements(
				MatchFields(IgnoreExtras, Fields{
					"Type":    Equal(test.ConsentV2.Type),
					"Version": Equal(test.ConsentV2.Version),
					"Status":  Equal(consent.RecordStatusActive),
				}),
			))
		})

		It("should filter by status", func() {
			result, err := consentService.ListConsentRecords(ctx(), userID, &consent.RecordFilter{
				Latest: pointer.FromAny(false),
				Status: pointer.FromAny(consent.RecordStatusRevoked),
			}, page.NewPagination())
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Count).To(Equal(1))
			Expect(result.Data).To(HaveExactElements(
				MatchFields(IgnoreExtras, Fields{
					"Type":    Equal(test.ConsentV1.Type),
					"Version": Equal(test.ConsentV1.Version),
					"Status":  Equal(consent.RecordStatusRevoked),
				}),
			))
		})

		It("should return the correct results with pagination", func() {
			pagination := page.NewPagination()
			pagination.Page = 1
			pagination.Size = 1

			result, err := consentService.ListConsentRecords(ctx(), userID, &consent.RecordFilter{
				Latest: pointer.FromAny(false),
			}, pagination)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Count).To(Equal(3))
			Expect(result.Data).To(HaveExactElements(
				MatchFields(IgnoreExtras, Fields{
					"Type":    Equal(test.ConsentV2.Type),
					"Version": Equal(test.ConsentV2.Version),
					"Status":  Equal(consent.RecordStatusActive),
				}),
			))
		})

		It("should return latest consent record for each type", func() {
			result, err := consentService.ListConsentRecords(ctx(), userID, &consent.RecordFilter{
				Latest: pointer.FromAny(true),
			}, page.NewPagination())
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Count).To(Equal(2))
			Expect(result.Data).To(ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					"Type":    Equal(test.AnotherConsentV1.Type),
					"Version": Equal(test.AnotherConsentV1.Version),
				}),
				MatchFields(IgnoreExtras, Fields{
					"Type":    Equal(test.ConsentV2.Type),
					"Version": Equal(test.ConsentV2.Version),
				}),
			))
		})

		It("should return the correct results with pagination when latest is true", func() {
			pagination := page.NewPagination()
			pagination.Page = 1
			pagination.Size = 1

			result, err := consentService.ListConsentRecords(ctx(), userID, &consent.RecordFilter{
				Latest: pointer.FromAny(true),
			}, pagination)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Count).To(Equal(2))
			Expect(result.Data).To(ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					"Type":    Equal(test.ConsentV2.Type),
					"Version": Equal(test.ConsentV2.Version),
				}),
			))
		})

	})

	Describe("UpdateConsentRecord", func() {
		var userID string

		BeforeEach(func() {
			userID = userTest.RandomUserID()

			creates := []*consent.RecordCreate{
				test.RandomRecordCreateForConsent(test.ConsentV1),
				test.RandomRecordCreateForConsent(test.ConsentV2),
				test.RandomRecordCreateForConsent(test.AnotherConsentV1),
			}
			for i, create := range creates {
				mailer.EXPECT().
					SendConsentGrantedEmailNotification(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

				create.CreatedTime = create.CreatedTime.Add(-time.Duration(len(creates)-i) * time.Second)
				created, err := consentService.CreateConsentRecord(ctx(), userID, create)
				Expect(err).ToNot(HaveOccurred())
				Expect(created).ToNot(BeNil())
			}
		})

		It("should update the metadata", func() {
			create := test.RandomRecordCreateForConsent(test.MockBDDPConsentV1)

			mailer.EXPECT().
				SendConsentGrantedEmailNotification(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			bddp.EXPECT().Share(gomock.Any(), userID).Return(nil)

			record, err := consentService.CreateConsentRecord(ctx(), userID, create)
			Expect(err).ToNot(HaveOccurred())

			record.Metadata = &consent.RecordMetadata{
				SupportedOrganizations: consent.BigDataDonationProjectOrganizations(),
			}

			_, err = consentService.UpdateConsentRecord(ctx(), record)
			Expect(err).ToNot(HaveOccurred())

			updated, err := consentService.GetConsentRecord(ctx(), userID, record.ID)
			Expect(err).ToNot(HaveOccurred())
			Expect(updated.Metadata).ToNot(BeNil())
			Expect(updated.Metadata.SupportedOrganizations).To(ConsistOf(consent.BigDataDonationProjectOrganizations()))
		})
	})

	Describe("RevokeConsentRecord", func() {
		var userID string

		BeforeEach(func() {
			userID = userTest.RandomUserID()
		})

		It("should revoke the record", func() {
			create := test.RandomRecordCreateForConsent(test.ConsentV2)

			mailer.EXPECT().
				SendConsentGrantedEmailNotification(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			created, err := consentService.CreateConsentRecord(ctx(), userID, create)
			Expect(err).ToNot(HaveOccurred())

			revoke := consent.NewRecordRevoke()
			revoke.ID = created.ID

			mailer.EXPECT().
				SendConsentRevokedEmailNotification(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			Expect(consentService.RevokeConsentRecord(ctx(), userID, revoke)).To(Succeed())

			revoked, err := consentService.GetConsentRecord(ctx(), userID, created.ID)
			Expect(err).ToNot(HaveOccurred())
			Expect(revoked).ToNot(BeNil())
			Expect(revoked.Status).To(Equal(consent.RecordStatusRevoked))
			Expect(revoked.RevocationTime).To(PointTo(BeTemporally("~", time.Now(), time.Minute)))
			Expect(revoked.ModifiedTime).To(BeTemporally("~", time.Now(), time.Minute))
		})

		It("should unshare the user's account with the BDDP recipient the record", func() {
			create := test.RandomRecordCreateForConsent(test.MockBDDPConsentV1)

			mailer.EXPECT().
				SendConsentGrantedEmailNotification(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			bddp.EXPECT().Share(gomock.Any(), userID).Return(nil)

			created, err := consentService.CreateConsentRecord(ctx(), userID, create)
			Expect(err).ToNot(HaveOccurred())

			revoke := consent.NewRecordRevoke()
			revoke.ID = created.ID

			mailer.EXPECT().
				SendConsentRevokedEmailNotification(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			bddp.EXPECT().Unshare(gomock.Any(), userID).Return(nil)

			Expect(consentService.RevokeConsentRecord(ctx(), userID, revoke)).To(Succeed())

			revoked, err := consentService.GetConsentRecord(ctx(), userID, created.ID)
			Expect(err).ToNot(HaveOccurred())
			Expect(revoked).ToNot(BeNil())
			Expect(revoked.Status).To(Equal(consent.RecordStatusRevoked))
			Expect(revoked.RevocationTime).To(PointTo(BeTemporally("~", time.Now(), time.Minute)))
			Expect(revoked.ModifiedTime).To(BeTemporally("~", time.Now(), time.Minute))
		})
	})
})
