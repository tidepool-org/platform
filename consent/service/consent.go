package service

import (
	"context"

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
	consentMailer           ConsentMailer
	consentRepository       *mongo.ConsentRepository
	consentRecordRepository *mongo.ConsentRecordRepository
	dbClient                *mongoDriver.Client
	logger                  log.Logger
}

func NewConsentService(consentMailer ConsentMailer, bddpSharer BigDataDonationProjectSharer, consentRepository *mongo.ConsentRepository, consentRecordRepository *mongo.ConsentRecordRepository, dbClient *mongoDriver.Client, logger log.Logger) consent.Service {
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
	pagination := page.NewPagination()
	pagination.Size = 1

	consents, err := c.ListConsents(ctx, &consent.Filter{
		Type:    pointer.FromAny(create.Type),
		Version: pointer.FromAny(create.Version),
	}, pagination)
	if err != nil {
		return nil, err
	}
	if len(consents.Data) == 0 {
		return nil, errors.New("invalid consent type and version combination")
	}
	cons := consents.Data[0]

	records, err := c.consentRecordRepository.ListConsentRecords(ctx, userID, &consent.RecordFilter{
		Latest: pointer.FromAny(true),
		Status: pointer.FromAny(consent.RecordStatusActive),
		Type:   pointer.FromAny(create.Type),
	}, pagination)
	if err != nil {
		return nil, err
	}

	res, err := structuredMongo.WithTransaction(ctx, c.dbClient, func(sessCtx mongoDriver.SessionContext) (any, error) {
		if len(records.Data) > 0 {
			existing := &records.Data[0]
			if existing.Version == create.Version {
				return nil, errors.New("consent record for the same type and version already exists")
			} else if existing.Version > create.Version {
				return nil, errors.New("consent record for a greater version already exists")
			}

			revoke := consent.NewRecordRevoke()
			revoke.ID = existing.ID
			revoke.RevocationTime = existing.CreatedTime // Ensure non-interrupted stream of data on re-consent

			err = c.consentRecordRepository.RevokeConsentRecord(sessCtx, userID, revoke)
			if err != nil {
				return nil, errors.Wrapf(err, "unable to revoke existing consent record for type %s", create.Type)
			}
		}

		record, err := c.consentRecordRepository.CreateConsentRecord(sessCtx, userID, create)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to create consent record for type %s", create.Type)
		}

		if create.Type == consent.TypeBigDataDonationProject {
			err = c.bddpSharer.Share(ctx, userID)
			if err != nil {
				return nil, errors.New("could not share data with bddp account after granting consent")
			}
		}

		return record, err
	})
	if err != nil || res == nil {
		return nil, err
	}

	record := res.(*consent.Record)

	// Sending an email is executed outside the transaction to prevent a failure from reverting the consent record
	// with an active sharing relationship
	if err := c.consentMailer.SendConsentGrantedEmailNotification(ctx, cons, *record); err != nil {
		// Just log the error, there's no need to fail the request
		c.logger.WithError(err).WithField("user", userID).Warn("unable to send email notification")
	}

	return record, nil
}

func (c *ConsentService) ListConsentRecords(ctx context.Context, userID string, filter *consent.RecordFilter, pagination *page.Pagination) (*structuredMongo.ListResult[consent.Record], error) {
	return c.consentRecordRepository.ListConsentRecords(ctx, userID, filter, pagination)
}

func (c *ConsentService) RevokeConsentRecord(ctx context.Context, userID string, revoke *consent.RecordRevoke) error {
	record, err := c.consentRecordRepository.GetConsentRecord(ctx, userID, revoke.ID)
	if err != nil {
		return errors.Wrapf(err, "consent record doesn't exist")
	}

	consents, err := c.ListConsents(ctx, &consent.Filter{
		Type:    pointer.FromAny(record.Type),
		Version: pointer.FromAny(record.Version),
	}, page.NewPagination())
	if err != nil {
		return err
	}

	_, err = structuredMongo.WithTransaction(ctx, c.dbClient, func(sessCtx mongoDriver.SessionContext) (any, error) {
		if err := c.consentRecordRepository.RevokeConsentRecord(ctx, userID, revoke); err != nil {
			return nil, err
		}
		if record.Type == consent.TypeBigDataDonationProject {
			err = c.bddpSharer.Unshare(ctx, userID)
			if err != nil {
				return nil, errors.New("could not unshare data with bddp account after revoking consent")
			}
		}
		return nil, nil
	})

	if len(consents.Data) == 0 {
		c.logger.WithField("user", userID).Warn("revoking record for missing consent type and version")
		return nil
	}

	// Sending an email is executed outside the transaction to prevent a failure from reverting the consent record
	// revocation
	if err := c.consentMailer.SendConsentRevokedEmailNotification(ctx, consents.Data[0], *record); err != nil {
		// Just log the error, there's no need to fail the request
		c.logger.WithError(err).WithField("user", userID).Warn("unable to send email notification")
	}

	return nil
}

func (c *ConsentService) UpdateConsentRecord(ctx context.Context, record *consent.Record) (*consent.Record, error) {
	return c.consentRecordRepository.UpdateConsentRecord(ctx, record)
}

var _ consent.Service = &ConsentService{}
