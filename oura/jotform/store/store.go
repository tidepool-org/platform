package store

import (
	"context"
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/errors"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

const CollectionName = "jotformSubmissions"

type ProcessedSubmission struct {
	SubmissionID string          `bson:"submissionId"`
	FormID       string          `bson:"formId"`
	RawPayload   json.RawMessage `bson:"rawPayload"`
	CreatedTime  time.Time       `bson:"createdTime"`
}

type Store interface {
	GetProcessedSubmission(ctx context.Context, formID, submissionID string) (*ProcessedSubmission, error)
	SaveProcessedSubmission(ctx context.Context, submission *ProcessedSubmission) error
}

type store struct {
	*storeStructuredMongo.Repository
}

func NewStore(mongoStore *storeStructuredMongo.Store) (Store, error) {
	if mongoStore == nil {
		return nil, errors.New("mongo store is missing")
	}

	s := &store{
		Repository: mongoStore.GetRepository(CollectionName),
	}

	if err := s.EnsureIndexes(context.Background()); err != nil {
		return nil, errors.Wrap(err, "unable to ensure indexes")
	}

	return s, nil
}

func (s *store) EnsureIndexes(ctx context.Context) error {
	return s.CreateAllIndexes(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "formId", Value: 1}, {Key: "submissionId", Value: 1}},
			Options: options.Index().
				SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "formId", Value: 1}, {Key: "createdTime", Value: -1}},
			Options: options.Index(),
		},
	})
}

func (s *store) GetProcessedSubmission(ctx context.Context, formID, submissionID string) (*ProcessedSubmission, error) {
	if submissionID == "" {
		return nil, errors.New("submission ID is missing")
	}

	var submission ProcessedSubmission
	err := s.FindOne(ctx, bson.M{
		"formId":       formID,
		"submissionId": submissionID,
	}).Decode(&submission)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to get processed submission")
	}

	return &submission, nil
}

func (s *store) SaveProcessedSubmission(ctx context.Context, submission *ProcessedSubmission) error {
	if submission == nil {
		return errors.New("submission is missing")
	}
	if submission.FormID == "" {
		return errors.New("form id is missing")
	}
	if submission.SubmissionID == "" {
		return errors.New("submission id is missing")
	}

	_, err := s.InsertOne(ctx, submission)
	if err != nil {
		if storeStructuredMongo.IsDup(err) {
			return errors.New("submission already exists")
		}
		return errors.Wrap(err, "unable to save processed submission")
	}

	return nil
}
