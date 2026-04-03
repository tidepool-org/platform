package service

import (
	"context"
	"slices"

	"github.com/tidepool-org/platform/log"

	"github.com/tidepool-org/platform/consent"
	"github.com/tidepool-org/platform/consent/store/mongo"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	structuredMongo "github.com/tidepool-org/platform/store/structured/mongo"

	mongoDriver "go.mongodb.org/mongo-driver/mongo"
)

type ConsentService struct {
	bddpSharer              BigDataDonationProjectSharer
	consentMailer           *ConsentMailer
	consentRepository       *mongo.ConsentRepository
	consentRecordRepository *mongo.ConsentRecordRepository
	dbClient                *mongoDriver.Client
	logger                  log.Logger
}

func NewConsentService(consentMailer *ConsentMailer, bddpSharer BigDataDonationProjectSharer, consentRepository *mongo.ConsentRepository, consentRecordRepository *mongo.ConsentRecordRepository, dbClient *mongoDriver.Client, logger log.Logger) consent.Service {
	return &ConsentService{
		bddpSharer:              bddpSharer,
		consentMailer:           consentMailer,
		consentRepository:       consentRepository,
		consentRecordRepository: consentRecordRepository,
		dbClient:                dbClient,
		logger:                  logger,
	}
}

func (c *ConsentService) ListConsents(ctx context.Context, filter *consent.Filter, pagination *page.Pagination) (*structuredMongo.ListResult[consent.Consent], error) {
	return c.consentRepository.List(ctx, filter, pagination)
}

func (c *ConsentService) EnsureConsent(ctx context.Context, consent *consent.Consent) error {
	return c.consentRepository.EnsureConsent(ctx, consent)
}

func (c *ConsentService) GetConsentRecord(ctx context.Context, userID string, id string) (*consent.Record, error) {
	return c.consentRecordRepository.GetConsentRecord(ctx, userID, id)
}

func (c *ConsentService) CreateConsentRecord(ctx context.Context, userID string, create *consent.RecordCreate) (*consent.Record, error) {
	records, err := c.CreateConsentRecords(ctx, userID, []*consent.RecordCreate{create})
	if err != nil {
		return nil, err
	}
	return records[0], nil
}

func (c *ConsentService) CreateConsentRecords(ctx context.Context, userID string, creates []*consent.RecordCreate) ([]*consent.Record, error) {
	if len(creates) == 0 {
		return nil, errors.New("creates is empty")
	}

	pagination := page.NewPaginationMinimum()

	type createConsentRecord struct {
		create  consent.RecordCreate
		consent consent.Consent
		record  *consent.Record
	}

	toCreate := make([]createConsentRecord, 0, len(creates))
	for _, create := range creates {
		consents, err := c.ListConsents(ctx, &consent.Filter{
			Type:    pointer.FromAny(create.Type),
			Version: pointer.FromAny(create.Version),
		}, pagination)
		if err != nil {
			return nil, err
		}
		if len(consents.Data) == 0 {
			return nil, errors.Newf("invalid consent type and version combination: %s v%d", create.Type, create.Version)
		}
		createWithConsent := createConsentRecord{
			create:  *create,
			consent: consents.Data[0],
		}

		records, err := c.consentRecordRepository.ListConsentRecords(ctx, userID, &consent.RecordFilter{
			Latest: pointer.FromAny(true),
			Status: pointer.FromAny(consent.RecordStatusActive),
			Type:   pointer.FromAny(create.Type),
		}, pagination)
		if err != nil {
			return nil, err
		}
		if len(records.Data) > 0 {
			createWithConsent.record = pointer.FromAny(records.Data[0])
		}
		toCreate = append(toCreate, createWithConsent)
	}

	res, err := structuredMongo.WithTransaction(ctx, c.dbClient, func(sessCtx mongoDriver.SessionContext) (any, error) {
		records := make([]*consent.Record, 0, len(toCreate))
		shareBDDP := false

		for i, createWithConsent := range toCreate {
			create := createWithConsent.create

			if createWithConsent.record != nil {
				existing := createWithConsent.record

				if existing.Version == createWithConsent.create.Version {
					return nil, errors.Newf("consent record for the same type and version already exists: %s v%d", create.Type, create.Version)
				} else if existing.Version > create.Version {
					return nil, errors.Newf("consent record for a greater version already exists: %s", create.Type)
				}

				revoke := consent.NewRecordRevoke()
				revoke.ID = existing.ID
				revoke.RevocationTime = existing.CreatedTime

				err := c.consentRecordRepository.RevokeConsentRecord(sessCtx, userID, revoke)
				if err != nil {
					return nil, errors.Wrapf(err, "unable to revoke existing consent record for type %s", create.Type)
				}
			}

			record, err := c.consentRecordRepository.CreateConsentRecord(sessCtx, userID, &create)
			if err != nil {
				return nil, errors.Wrapf(err, "unable to create consent record for type %s", create.Type)
			}

			records = append(records, record)
			toCreate[i].record = record

			if create.Type == consent.TypeBigDataDonationProject {
				shareBDDP = true
			}
		}

		if shareBDDP {
			if err := c.bddpSharer.Share(sessCtx, userID); err != nil {
				return nil, errors.New("could not share data with bddp account after granting consent")
			}
		}

		return records, nil
	})
	if err != nil {
		return nil, err
	}

	// Sending emails is executed outside the transaction to prevent a failure from reverting the consent record
	// with an active sharing relationship
	for _, createWithConsent := range toCreate {
		// Just log the error, there's no need to fail the request
		if err := c.consentMailer.SendConsentGrantedEmailNotification(ctx, createWithConsent.consent, *createWithConsent.record); err != nil {
			c.logger.WithError(err).WithField("user", userID).Warn("unable to send email notification")
		}
	}

	records := res.([]*consent.Record)
	return records, nil
}

func (c *ConsentService) ListConsentRecords(ctx context.Context, userID string, filter *consent.RecordFilter, pagination *page.Pagination) (*structuredMongo.ListResult[consent.Record], error) {
	return c.consentRecordRepository.ListConsentRecords(ctx, userID, filter, pagination)
}

func (c *ConsentService) RevokeConsentRecord(ctx context.Context, userID string, revoke *consent.RecordRevoke) error {
	if revoke == nil {
		return errors.New("revoke is missing")
	}
	record, err := c.consentRecordRepository.GetConsentRecord(ctx, userID, revoke.ID)
	if err != nil {
		return errors.Wrapf(err, "consent record doesn't exist")
	}

	type recordRevokeConsent struct {
		record  consent.Record
		revoke  *consent.RecordRevoke
		consent *consent.Consent
	}

	toRevoke := []recordRevokeConsent{{
		record: *record,
		revoke: revoke,
	}}

	// Get all dependent consent records that are active
	dependentTypes := consent.DependentConsentTypes(record.Type, record.Version)
	for _, dependentType := range dependentTypes {
		records, err := c.consentRecordRepository.ListConsentRecords(ctx, userID, &consent.RecordFilter{
			Latest: pointer.FromAny(true),
			Status: pointer.FromAny(consent.RecordStatusActive),
			Type:   pointer.FromAny(dependentType),
		}, page.NewPaginationMinimum())
		if err != nil {
			return errors.Wrapf(err, "unable to list dependent consent records for type %s", dependentType)
		}
		if len(records.Data) > 0 {
			dependentRevoke := consent.NewRecordRevoke()
			dependentRevoke.ID = records.Data[0].ID
			dependentRevoke.RevocationTime = revoke.RevocationTime // Use the same revocation time as the original revoke

			toRevoke = append(toRevoke, recordRevokeConsent{
				record: records.Data[0],
				revoke: dependentRevoke,
			})
		}
	}

	unshareBDDP := false
	for i := range toRevoke {
		if toRevoke[i].record.Type == consent.TypeBigDataDonationProject {
			unshareBDDP = true
		}

		// Get the corresponding consent so we can send the email notification
		consents, err := c.ListConsents(ctx, &consent.Filter{
			Type:    pointer.FromAny(toRevoke[i].record.Type),
			Version: pointer.FromAny(toRevoke[i].record.Version),
		}, page.NewPaginationMinimum())
		if err != nil {
			return errors.Wrapf(err, "unable to list consents for type %s", toRevoke[i].record.Type)
		}
		if len(consents.Data) == 0 {
			continue
		}
		toRevoke[i].consent = pointer.FromAny(consents.Data[0])
	}

	_, err = structuredMongo.WithTransaction(ctx, c.dbClient, func(sessCtx mongoDriver.SessionContext) (any, error) {
		for _, r := range toRevoke {
			if err := c.consentRecordRepository.RevokeConsentRecord(sessCtx, userID, r.revoke); err != nil {
				return nil, err
			}

			if r.consent == nil {
				c.logger.
					WithFields(log.Fields{"user": userID, "type": r.record.Type, "version": r.record.Version}).
					Warn("revoked consent record for an unknown consent type and version")
			}
		}

		// Unshare the data if BDDP was revoked. If this fails, the transaction will be reverted.
		if unshareBDDP {
			err = c.bddpSharer.Unshare(sessCtx, userID)
			if err != nil {
				return nil, errors.New("could not unshare data with bddp account after revoking consent")
			}
		}

		return nil, nil
	})

	// Filter out any revoked consent records that don't have a corresponding consent content
	toRevoke = slices.DeleteFunc(toRevoke, func(r recordRevokeConsent) bool {
		return r.consent == nil
	})

	// Sending an email is executed outside the transaction to prevent a failure from reverting the consent record
	// revocation
	for _, r := range toRevoke {
		if err := c.consentMailer.SendConsentRevokedEmailNotification(ctx, *r.consent, r.record); err != nil {
			// Just log the error, there's no need to fail the entire request
			c.logger.WithError(err).WithField("user", userID).Warn("unable to send email notification")
		}
	}

	return nil
}

func (c *ConsentService) UpdateConsentRecord(ctx context.Context, record *consent.Record) (*consent.Record, error) {
	return c.consentRecordRepository.UpdateConsentRecord(ctx, record)
}

var _ consent.Service = &ConsentService{}
