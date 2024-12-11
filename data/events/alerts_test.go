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
	"github.com/tidepool-org/platform/devicetokens"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	logtest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/push"
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

		It("consumes alerts config events", func() {
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

		It("consumes device data events", func() {
			blood := newTestStaticDatumMmolL(7.2)
			kafkaMsg := newAlertsMockConsumerMessage(".data.deviceData.alerts", blood)
			docs := []interface{}{bson.M{}}
			c, deps := newConsumerTestDeps(docs)

			Expect(c.Consume(deps.Context, deps.Session, kafkaMsg)).To(Succeed())
			Expect(deps.Session.MarkCalls).To(Equal(1))
		})

		It("errors out when the datum's UserID is nil", func() {
			blood := newTestStaticDatumMmolL(7.2)
			blood.UserID = nil
			kafkaMsg := newAlertsMockConsumerMessage(".data.deviceData.alerts", blood)
			docs := []interface{}{bson.M{}}
			c, deps := newConsumerTestDeps(docs)

			Expect(c.Consume(deps.Context, deps.Session, kafkaMsg)).
				To(MatchError(ContainSubstring("userID is nil")))
			Expect(deps.Session.MarkCalls).To(Equal(0))
		})

		It("errors out when the datum's UploadID is nil", func() {
			blood := newTestStaticDatumMmolL(7.2)
			blood.UploadID = nil
			kafkaMsg := newAlertsMockConsumerMessage(".data.deviceData.alerts", blood)
			docs := []interface{}{bson.M{}}
			c, deps := newConsumerTestDeps(docs)

			Expect(c.Consume(deps.Context, deps.Session, kafkaMsg)).
				To(MatchError(ContainSubstring("uploadID is nil")))
			Expect(deps.Session.MarkCalls).To(Equal(0))
		})

		It("pushes notifications", func() {
			blood := newTestStaticDatumMmolL(1.0)
			kafkaMsg := newAlertsMockConsumerMessage(".data.deviceData.alerts", blood)
			docs := []interface{}{bson.M{}}
			c, deps := newConsumerTestDeps(docs)
			eval := newMockEvaluator()
			eval.Evaluations[testFollowedUserID+testUploadID] = []mockEvaluatorResponse{
				{
					Notifications: []*alerts.Notification{
						{
							Message:         "something",
							RecipientUserID: testUserID,
							FollowedUserID:  testFollowedUserID,
						},
					},
				},
			}
			c.Evaluator = eval

			Expect(c.Consume(deps.Context, deps.Session, kafkaMsg)).To(Succeed())

			deps.Logger.AssertInfo("logging push notification")
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

				notes, err := eval.Evaluate(ctx, testFollowedUserID, testUploadID)

				Expect(err).To(Succeed())
				Expect(notes).To(ConsistOf(HaveField("RecipientUserID", testUserID)))
			})

			It("checks that alerts configs match the data set id", func() {
				testLogger := logtest.NewLogger()
				ctx := log.NewContextWithLogger(context.Background(), testLogger)
				eval, deps := newEvaluatorTestDeps([]*store.AlertableResponse{testMongoUrgentLowResponse})
				deps.Permissions.Allow(testUserID+"2", permission.Follow, testFollowedUserID)
				deps.Alerts.Configs = append(deps.Alerts.Configs, testAlertsConfigUrgentLow(testUserID+"2"))
				deps.Permissions.Allow(testUserID, permission.Follow, testFollowedUserID)
				wrongDataSetID := testAlertsConfigUrgentLow(testUserID)
				wrongDataSetID.UploadID = "wrong"
				deps.Alerts.Configs = append(deps.Alerts.Configs, wrongDataSetID)

				notes, err := eval.Evaluate(ctx, testFollowedUserID, testUploadID)

				Expect(err).To(Succeed())
				Expect(notes).To(ConsistOf(HaveField("RecipientUserID", testUserID+"2")))
			})

			It("uses the longest delay", func() {
				testLogger := logtest.NewLogger()
				ctx := log.NewContextWithLogger(context.Background(), testLogger)
				eval, deps := newEvaluatorTestDeps([]*store.AlertableResponse{testMongoUrgentLowResponse})
				cfgWithShorterDelay := testAlertsConfigLow(testUserID)
				deps.Alerts.Configs = append(deps.Alerts.Configs, cfgWithShorterDelay)
				deps.Permissions.Allow(testUserID, permission.Follow, testFollowedUserID)
				cfgWithLongerDelay := testAlertsConfigLow(testUserID + "2")
				cfgWithLongerDelay.Alerts.Low.Delay = alerts.DurationMinutes(10 * time.Minute)
				deps.Alerts.Configs = append(deps.Alerts.Configs, cfgWithLongerDelay)
				deps.Permissions.Allow(testUserID+"2", permission.Follow, testFollowedUserID)

				_, err := eval.Evaluate(ctx, testFollowedUserID, testUploadID)

				Expect(err).To(Succeed())
				if Expect(len(deps.Data.GetAlertableDataInputs)).To(Equal(1)) {
					Expect(deps.Data.GetAlertableDataInputs[0].Params.Start).
						To(BeTemporally("~", time.Now().Add(-10*time.Minute), time.Second))
				}
			})
		})

	})
})

type consumerTestDeps struct {
	Alerts       *mockAlertsConfigClient
	Context      context.Context
	Cursor       *mongo.Cursor
	DeviceTokens *mockDeviceTokens
	Evaluator    *mockStaticEvaluator
	Logger       *logtest.Logger
	Permissions  *mockPermissionsClient
	Pusher       Pusher
	Repo         *storetest.DataRepository
	Session      *mockConsumerGroupSession
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
	pusher := push.NewLogPusher(logger)
	deviceTokens := newMockDeviceTokens()
	deviceTokens.Tokens = append(deviceTokens.Tokens, []*devicetokens.DeviceToken{
		{Apple: &devicetokens.AppleDeviceToken{}},
	})

	return &Consumer{
			Alerts:       alertsClient,
			Evaluator:    evaluator,
			Data:         dataRepo,
			DeviceTokens: deviceTokens,
			Permissions:  permissions,
			Pusher:       pusher,
		}, &consumerTestDeps{
			Alerts:       alertsClient,
			Context:      ctx,
			Cursor:       cur,
			DeviceTokens: deviceTokens,
			Evaluator:    evaluator,
			Pusher:       pusher,
			Repo:         dataRepo,
			Session:      &mockConsumerGroupSession{},
			Logger:       logger,
			//Tokens:      tokens,
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
			Data:        dataRepo,
		}
}

type evaluatorTestDeps struct {
	Alerts      *mockAlertsConfigClient
	Permissions *mockPermissionsClient
	Data        *storetest.DataRepository
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

func (e *mockEvaluator) Evaluate(ctx context.Context, followedUserID, dataSetID string) (
	[]*alerts.Notification, error) {

	key := followedUserID + dataSetID
	if _, found := e.Evaluations[key]; !found {
		return nil, nil
	}
	resp := e.Evaluations[key][0]
	if len(e.Evaluations[key]) > 1 {
		e.Evaluations[key] = e.Evaluations[key][1:]
	}
	e.EvaluateCalls[key] += 1
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

func (e *mockStaticEvaluator) Evaluate(ctx context.Context, followedUserID, dataSetID string) (
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

func addLogger(ctx context.Context) (context.Context, *logtest.Logger) {
	GinkgoHelper()
	if ctx == nil {
		ctx = context.Background()
	}

	lgr := logtest.NewLogger()
	return log.NewContextWithLogger(ctx, lgr), lgr
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
				UserID:   pointer.FromAny(testFollowedUserID),
				Time:     pointer.FromTime(time.Now()),
				UploadID: pointer.FromAny(testUploadID),
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

func testAlertsConfigLow(userID string) *alerts.Config {
	return &alerts.Config{
		UserID:         userID,
		FollowedUserID: testFollowedUserID,
		UploadID:       testUploadID,
		Alerts: alerts.Alerts{
			Low: &alerts.LowAlert{
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

type mockDeviceTokens struct {
	Error  error
	Tokens [][]*devicetokens.DeviceToken
}

func newMockDeviceTokens() *mockDeviceTokens {
	return &mockDeviceTokens{
		Tokens: [][]*devicetokens.DeviceToken{},
	}
}

func (t *mockDeviceTokens) GetDeviceTokens(ctx context.Context, userID string) ([]*devicetokens.DeviceToken, error) {
	if t.Error != nil {
		return nil, t.Error
	}
	if len(t.Tokens) > 0 {
		ret := t.Tokens[0]
		t.Tokens = t.Tokens[1:]
		return ret, nil
	}
	return nil, nil
}
