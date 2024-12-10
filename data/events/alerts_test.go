package events

import (
	"context"
	"time"

	"github.com/IBM/sarama"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/alerts"
	nontypesglucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/store"
	storetest "github.com/tidepool-org/platform/data/store/test"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/blood"
	"github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	logtest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/pointer"
)

const (
	testUserID         = "test-user-id"
	testFollowedUserID = "test-followed-user-id"
	testUserNoPermsID  = "test-user-no-perms"
	testUploadID       = "test-upload-id"
)

var (
	testMongoUrgentLowResponse = &store.AlertableResponse{
		Glucose: []*glucose.Glucose{
			newTestStaticDatumMmolL(1.0)},
	}
)

var _ = Describe("Consumer", func() {
	Describe("Consume", func() {
		It("ignores nil messages", func() {
			ctx, _ := addLogger(context.Background())
			c := &Consumer{}

			Expect(c.Consume(ctx, nil, nil)).To(Succeed())
		})

		It("processes alerts config events", func() {
			cfg := &alerts.Config{
				UserID:         testUserID,
				FollowedUserID: testFollowedUserID,
				Alerts: alerts.Alerts{
					Low: &alerts.LowAlert{
						Base: alerts.Base{
							Enabled: true},
						Threshold: alerts.Threshold{
							Value: 101.1,
							Units: "mg/dL",
						},
					},
				},
			}
			kafkaMsg := newAlertsMockConsumerMessage(".data.alerts", cfg)
			docs := []interface{}{bson.M{}}
			c, deps := newConsumerTestDeps(docs)

			Expect(c.Consume(deps.Context, deps.Session, kafkaMsg)).To(Succeed())
			Expect(deps.Session.MarkCalls).To(Equal(1))
		})

		It("processes device data events", func() {
			blood := &glucose.Glucose{
				Blood: blood.Blood{
					Units: pointer.FromAny("mmol/L"),
					Value: pointer.FromAny(7.2),
					Base: types.Base{
						UserID: pointer.FromAny(testFollowedUserID),
					},
				},
			}
			kafkaMsg := newAlertsMockConsumerMessage(".data.deviceData.alerts", blood)
			docs := []interface{}{bson.M{}}
			c, deps := newConsumerTestDeps(docs)

			Expect(c.Consume(deps.Context, deps.Session, kafkaMsg)).To(Succeed())
			Expect(deps.Session.MarkCalls).To(Equal(1))
		})

	})

	Describe("Evaluator", func() {
		Describe("Evaluate", func() {
			It("checks that alerts config owners have permission", func() {
				testLogger := logtest.NewLogger()
				ctx := log.NewContextWithLogger(context.Background(), testLogger)

				eval, deps := newEvaluatorTestDeps([]*store.AlertableResponse{testMongoUrgentLowResponse})
				deps.Permissions.Allow(testUserID, permission.Follow, testFollowedUserID)
				deps.Permissions.DenyAll(testUserNoPermsID, testFollowedUserID)
				deps.Alerts.Configs = append(deps.Alerts.Configs, testAlertsConfigUrgentLow(testUserNoPermsID))
				deps.Alerts.Configs = append(deps.Alerts.Configs, testAlertsConfigUrgentLow(testUserID))

				notes, err := eval.Evaluate(ctx, testFollowedUserID)

				Expect(err).To(Succeed())
				Expect(notes).To(ConsistOf(HaveField("RecipientUserID", testUserID)))
			})

			It("uses the longest delay", func() {

			})
		})

	})
})

type consumerTestDeps struct {
	Alerts      *mockAlertsConfigClient
	Context     context.Context
	Cursor      *mongo.Cursor
	Evaluator   *mockStaticEvaluator
	Logger      log.Logger
	Permissions *mockPermissionsClient
	Repo        *storetest.DataRepository
	Session     *mockConsumerGroupSession
}

func newConsumerTestDeps(docs []interface{}) (*Consumer, *consumerTestDeps) {
	GinkgoHelper()
	ctx, logger := addLogger(context.Background())
	alertsClient := newMockAlertsConfigClient([]*alerts.Config{
		{
			UserID:         testUserID,
			FollowedUserID: testFollowedUserID,
			Alerts:         alerts.Alerts{},
		},
	}, nil)
	dataRepo := storetest.NewDataRepository()
	dataRepo.GetLastUpdatedForUserOutputs = []storetest.GetLastUpdatedForUserOutput{}
	augmentedDocs := augmentMockMongoDocs(docs)
	cur := newMockMongoCursor(augmentedDocs)
	dataRepo.GetDataRangeOutputs = []storetest.GetDataRangeOutput{
		{Error: nil, Cursor: cur},
	}
	permissions := newMockPermissionsClient()
	evaluator := newMockStaticEvaluator()

	return &Consumer{
			Alerts:      alertsClient,
			Evaluator:   evaluator,
			Data:        dataRepo,
			Permissions: permissions,
		}, &consumerTestDeps{
			Alerts:      alertsClient,
			Context:     ctx,
			Cursor:      cur,
			Evaluator:   evaluator,
			Repo:        dataRepo,
			Session:     &mockConsumerGroupSession{},
			Logger:      logger,
			Permissions: permissions,
		}
}

func newEvaluatorTestDeps(responses []*store.AlertableResponse) (*evaluator, *evaluatorTestDeps) {
	alertsClient := newMockAlertsConfigClient(nil, nil)
	dataRepo := storetest.NewDataRepository()
	dataRepo.GetLastUpdatedForUserOutputs = []storetest.GetLastUpdatedForUserOutput{}
	for _, r := range responses {
		out := storetest.GetAlertableDataOutput{Response: r}
		dataRepo.GetAlertableDataOutputs = append(dataRepo.GetAlertableDataOutputs, out)
	}
	permissions := newMockPermissionsClient()
	return &evaluator{
			Alerts:      alertsClient,
			Data:        dataRepo,
			Permissions: permissions,
		}, &evaluatorTestDeps{
			Alerts:      alertsClient,
			Permissions: permissions,
		}
}

type evaluatorTestDeps struct {
	Alerts      *mockAlertsConfigClient
	Permissions *mockPermissionsClient
}

// mockEvaluator implements Evaluator.
type mockEvaluator struct {
	Evaluations   map[string][]mockEvaluatorResponse
	EvaluateCalls map[string]int
}

type mockEvaluatorResponse struct {
	Notifications []*alerts.Notification
	Error         error
}

func newMockEvaluator() *mockEvaluator {
	return &mockEvaluator{
		Evaluations:   map[string][]mockEvaluatorResponse{},
		EvaluateCalls: map[string]int{},
	}
}

func (e *mockEvaluator) Evaluate(ctx context.Context, followedUserID string) (
	[]*alerts.Notification, error) {

	if _, found := e.Evaluations[followedUserID]; !found {
		return nil, nil
	}
	resp := e.Evaluations[followedUserID][0]
	if len(e.Evaluations[followedUserID]) > 1 {
		e.Evaluations[followedUserID] = e.Evaluations[followedUserID][1:]
	}
	e.EvaluateCalls[followedUserID] += 1
	if resp.Error != nil {
		return nil, resp.Error
	}
	return resp.Notifications, nil
}

func (e *mockEvaluator) EvaluateCallsTotal() int {
	total := 0
	for _, val := range e.EvaluateCalls {
		total += val
	}
	return total
}

// mockStaticEvaluator wraps mock evaluator with a static response.
//
// Useful when testing Consumer behavior, when the behavior of the Evaulator
// isn't relevant to the Consumer test.
type mockStaticEvaluator struct {
	*mockEvaluator
}

func newMockStaticEvaluator() *mockStaticEvaluator {
	return &mockStaticEvaluator{newMockEvaluator()}
}

func (e *mockStaticEvaluator) Evaluate(ctx context.Context, followedUserID string) (
	[]*alerts.Notification, error) {

	e.EvaluateCalls[followedUserID] += 1
	return nil, nil
}

func newAlertsMockConsumerMessage(topic string, v any) *sarama.ConsumerMessage {
	GinkgoHelper()
	doc := &struct {
		FullDocument any `json:"fullDocument" bson:"fullDocument"`
	}{FullDocument: v}
	vBytes, err := bson.MarshalExtJSON(doc, false, false)
	Expect(err).To(Succeed())
	return &sarama.ConsumerMessage{
		Value: vBytes,
		Topic: topic,
	}
}

func addLogger(ctx context.Context) (context.Context, log.Logger) {
	GinkgoHelper()
	if ctx == nil {
		ctx = context.Background()
	}

	lgr := newTestLogger()
	return log.NewContextWithLogger(ctx, lgr), lgr
}

func newTestLogger() log.Logger {
	GinkgoHelper()
	lgr := logtest.NewLogger()
	return lgr
}

func augmentMockMongoDocs(inDocs []interface{}) []interface{} {
	defaultDoc := bson.M{
		"_userId": testFollowedUserID,
		"_active": true,
		"type":    "upload",
		"time":    time.Now(),
	}
	outDocs := []interface{}{}
	for _, inDoc := range inDocs {
		newDoc := defaultDoc
		switch v := (inDoc).(type) {
		case map[string]interface{}:
			for key, val := range v {
				newDoc[key] = val
			}
			outDocs = append(outDocs, newDoc)
		default:
			outDocs = append(outDocs, inDoc)
		}
	}
	return outDocs
}

func newMockMongoCursor(docs []interface{}) *mongo.Cursor {
	GinkgoHelper()
	cur, err := mongo.NewCursorFromDocuments(docs, nil, nil)
	Expect(err).To(Succeed())
	return cur
}

func newTestStaticDatumMmolL(value float64) *glucose.Glucose {
	return &glucose.Glucose{
		Blood: blood.Blood{
			Base: types.Base{
				Time: pointer.FromTime(time.Now()),
			},
			Units: pointer.FromString(nontypesglucose.MmolL),
			Value: pointer.FromFloat64(value),
		},
	}
}

type mockAlertsConfigClient struct {
	Error   error
	Configs []*alerts.Config
}

func newMockAlertsConfigClient(c []*alerts.Config, err error) *mockAlertsConfigClient {
	if c == nil {
		c = []*alerts.Config{}
	}
	return &mockAlertsConfigClient{
		Configs: c,
		Error:   err,
	}
}

func (c *mockAlertsConfigClient) Delete(_ context.Context, _ *alerts.Config) error {
	return c.Error
}

func (c *mockAlertsConfigClient) Get(_ context.Context, _ *alerts.Config) (*alerts.Config, error) {
	if c.Error != nil {
		return nil, c.Error
	} else if len(c.Configs) > 0 {
		return c.Configs[0], nil
	}
	return nil, nil
}

func (c *mockAlertsConfigClient) List(_ context.Context, userID string) ([]*alerts.Config, error) {
	if c.Error != nil {
		return nil, c.Error
	} else if len(c.Configs) > 0 {
		return c.Configs, nil
	}
	return nil, nil
}

func (c *mockAlertsConfigClient) Upsert(_ context.Context, _ *alerts.Config) error {
	return c.Error
}

type mockConsumerGroupSession struct {
	MarkCalls int
}

func (s *mockConsumerGroupSession) Claims() map[string][]int32 {
	panic("not implemented") // TODO: Implement
}

func (s *mockConsumerGroupSession) MemberID() string {
	panic("not implemented") // TODO: Implement
}

func (s *mockConsumerGroupSession) GenerationID() int32 {
	panic("not implemented") // TODO: Implement
}

func (s *mockConsumerGroupSession) MarkOffset(topic string, partition int32, offset int64, metadata string) {
	panic("not implemented") // TODO: Implement
}

func (s *mockConsumerGroupSession) Commit() {
	panic("not implemented") // TODO: Implement
}

func (s *mockConsumerGroupSession) ResetOffset(topic string, partition int32, offset int64, metadata string) {
	panic("not implemented") // TODO: Implement
}

func (s *mockConsumerGroupSession) MarkMessage(msg *sarama.ConsumerMessage, metadata string) {
	s.MarkCalls++
}

func (s *mockConsumerGroupSession) Context() context.Context {
	panic("not implemented") // TODO: Implement
}

type mockPermissionsClient struct {
	Error error
	Perms map[string]permission.Permissions
}

func newMockPermissionsClient() *mockPermissionsClient {
	return &mockPermissionsClient{
		Perms: map[string]permission.Permissions{},
	}
}

func (c *mockPermissionsClient) Key(requesterUserID, targetUserID string) string {
	return requesterUserID + targetUserID
}

func (c *mockPermissionsClient) Allow(requestUserID, perm, targetUserID string) {
	key := c.Key(requestUserID, targetUserID)
	if _, found := c.Perms[key]; !found {
		c.Perms[key] = permission.Permissions{}
	}
	c.Perms[key][perm] = permission.Permission{}
}

func (c *mockPermissionsClient) DenyAll(requestUserID, targetUserID string) {
	key := c.Key(requestUserID, targetUserID)
	delete(c.Perms, key)
}

func (c *mockPermissionsClient) GetUserPermissions(ctx context.Context, requestUserID string, targetUserID string) (permission.Permissions, error) {
	if c.Error != nil {
		return nil, c.Error
	}
	if p, ok := c.Perms[c.Key(requestUserID, targetUserID)]; ok {
		return p, nil
	} else {
		return nil, errors.New("test error NOT FOUND")
	}
}

func testAlertsConfigUrgentLow(userID string) *alerts.Config {
	return &alerts.Config{
		UserID:         userID,
		FollowedUserID: testFollowedUserID,
		UploadID:       testUploadID,
		Alerts: alerts.Alerts{
			UrgentLow: &alerts.UrgentLowAlert{
				Base: alerts.Base{
					Enabled:  true,
					Activity: alerts.Activity{},
				},
				Threshold: alerts.Threshold{
					Value: 10.0,
					Units: nontypesglucose.MgdL,
				},
			},
		},
	}
}
