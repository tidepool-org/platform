package service

import (
	"context"

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
	consentRepository       *mongo.ConsentRepository
	consentRecordRepository *mongo.ConsentRecordRepository
	dbClient                *mongoDriver.Client
}

func NewConsentService(bddpSharer BigDataDonationProjectSharer, consentRepository *mongo.ConsentRepository, consentRecordRepository *mongo.ConsentRecordRepository, dbClient *mongoDriver.Client) consent.Service {
	return &ConsentService{
		bddpSharer:              bddpSharer,
		consentRepository:       consentRepository,
		consentRecordRepository: consentRecordRepository,
		dbClient:                dbClient,
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

	res, err := structuredMongo.WithTransaction(ctx, c.dbClient, func(sessCtx mongoDriver.SessionContext) (any, error) {
		consents, err := c.ListConsents(sessCtx, &consent.Filter{
			Type:    pointer.FromAny(create.Type),
			Version: pointer.FromAny(create.Version),
		}, pagination)
		if err != nil {
			return nil, err
		}
		if len(consents.Data) == 0 {
			return nil, errors.New("invalid consent type and version combination")
		}

		records, err := c.consentRecordRepository.ListConsentRecords(sessCtx, userID, &consent.RecordFilter{
			Latest: pointer.FromAny(true),
			Status: pointer.FromAny(consent.RecordStatusActive),
			Type:   pointer.FromAny(create.Type),
		}, pagination)
		if err != nil {
			return nil, err
		}

		if len(records.Data) > 0 {
			if create.Type == consent.TypeBigDataDonationProject {
				err = c.bddpSharer.Unshare(ctx, userID)
				if err != nil {
					return nil, errors.New("could not unshare data with bddp account before revoking consent")
				}
			}

			revoke := consent.NewConsentRecordRevoke()
			revoke.ID = records.Data[0].ID

			// Ensure non-interrupted stream of data if the user re-consents
			revoke.RevocationTime = create.CreatedTime

			err = c.consentRecordRepository.RevokeConsentRecord(sessCtx, userID, revoke)
			if err != nil {
				return nil, errors.Wrapf(err, "unable to revoke existing consent record for type %s", create.Type)
			}
		}

		cons, err := c.consentRecordRepository.CreateConsentRecord(sessCtx, userID, create)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to create consent record for type %s", create.Type)
		}

		if create.Type == consent.TypeBigDataDonationProject {
			err = c.bddpSharer.Share(ctx, userID)
			if err != nil {
				return nil, errors.New("could not share data with bddp account after granting consent")
			}
		}

		return cons, err
	})

	if err != nil {
		return nil, err
	}

	return res.(*consent.Record), nil
}

func (c *ConsentService) ListConsentRecords(ctx context.Context, userID string, filter *consent.RecordFilter, pagination *page.Pagination) (*structuredMongo.ListResult[consent.Record], error) {
	return c.consentRecordRepository.ListConsentRecords(ctx, userID, filter, pagination)
}

func (c *ConsentService) RevokeConsentRecord(ctx context.Context, userID string, revoke *consent.RecordRevoke) error {
	return c.consentRecordRepository.RevokeConsentRecord(ctx, userID, revoke)
}

func (c *ConsentService) UpdateConsentRecord(ctx context.Context, record *consent.Record) (*consent.Record, error) {
	return c.consentRecordRepository.UpdateConsentRecord(ctx, record)
}

var _ consent.Service = &ConsentService{}
