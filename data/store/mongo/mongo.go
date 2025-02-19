package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/alerts"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/dosingdecision"
	"github.com/tidepool-org/platform/errors"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

func NewStore(config *storeStructuredMongo.Config) (*Store, error) {
	baseStore, err := storeStructuredMongo.NewStore(config)
	if err != nil {
		return nil, err
	}

	return NewStoreFromBase(baseStore), nil
}

func NewStoreFromBase(base *storeStructuredMongo.Store) *Store {
	return &Store{
		Store: base,
	}
}

type Store struct {
	*storeStructuredMongo.Store
}

func (s *Store) EnsureIndexes() error {
	dataRepository := s.NewDataRepository()
	summaryRepository := s.NewSummaryRepository()
	alertsRepository := s.NewAlertsRepository()
	lastCommunicationsRepository := s.NewLastCommunicationsRepository()

	if err := dataRepository.EnsureIndexes(); err != nil {
		return err
	}

	if err := summaryRepository.EnsureIndexes(); err != nil {
		return err
	}

	if err := alertsRepository.EnsureIndexes(); err != nil {
		return err
	}

	if err := lastCommunicationsRepository.EnsureIndexes(); err != nil {
		return err
	}

	return nil
}

func (s *Store) NewDataRepository() store.DataRepository {
	return &DataRepository{
		DatumRepository: &DatumRepository{
			s.Store.GetRepository("deviceData"),
		},
		DataSetRepository: &DataSetRepository{
			s.Store.GetRepository("deviceDataSets"),
		},
	}
}

func (s *Store) NewSummaryRepository() store.SummaryRepository {
	return &SummaryRepository{
		s.Store.GetRepository("summary"),
	}
}

func (s *Store) NewAlertsRepository() alerts.Repository {
	r := alertsRepo(*s.Store.GetRepository("alerts"))
	return &r
}

func (s *Store) NewLastCommunicationsRepository() alerts.LastCommunicationsRepository {
	r := lastCommunicationsRepo(*s.Store.GetRepository("lastCommunications"))
	return &r
}

func (s *Store) NewAlertsDataRepository() alerts.DataRepository {
	r := deviceDataForAlertsRepo(*s.Store.GetRepository("deviceData"))
	return &r
}

type deviceDataForAlertsRepo storeStructuredMongo.Repository

func (r *deviceDataForAlertsRepo) GetAlertableData(ctx context.Context,
	params alerts.GetAlertableDataParams) (*alerts.GetAlertableDataResponse, error) {

	if params.End.IsZero() {
		params.End = time.Now()
	}

	cursor, err := r.getAlertableData(ctx, params, dosingdecision.Type)
	if err != nil {
		return nil, err
	}
	dosingDecisions := []*dosingdecision.DosingDecision{}
	if err := cursor.All(ctx, &dosingDecisions); err != nil {
		return nil, errors.Wrap(err, "Unable to load alertable dosing documents")
	}
	cursor, err = r.getAlertableData(ctx, params, continuous.Type)
	if err != nil {
		return nil, err
	}
	glucoseData := []*glucose.Glucose{}
	if err := cursor.All(ctx, &glucoseData); err != nil {
		return nil, errors.Wrap(err, "Unable to load alertable glucose documents")
	}
	response := &alerts.GetAlertableDataResponse{
		DosingDecisions: dosingDecisions,
		Glucose:         glucoseData,
	}

	return response, nil
}

func (r *deviceDataForAlertsRepo) getAlertableData(ctx context.Context,
	params alerts.GetAlertableDataParams, typ string) (*mongo.Cursor, error) {

	selector := bson.M{
		"_active":  true,
		"uploadId": params.UploadID,
		"type":     typ,
		"_userId":  params.UserID,
		"time":     bson.M{"$gte": params.Start, "$lte": params.End},
	}
	findOptions := options.Find().SetSort(bson.D{{Key: "time", Value: -1}})
	cursor, err := r.Find(ctx, selector, findOptions)
	if err != nil {
		format := "Unable to find alertable %s data in dataset %s"
		return nil, errors.Wrapf(err, format, typ, params.UploadID)
	}
	return cursor, nil
}
