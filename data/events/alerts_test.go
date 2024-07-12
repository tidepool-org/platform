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

	// Describe("evaluateUrgentLow", func() {
	// 	It("can't function without datum units", func() {
	// 		ctx, _ := addLogger(context.Background())
	// 		alert := newTestUrgentLowAlert()
	// 		datum := newTestStaticDatumMmolL(11)
	// 		datum.Blood.Units = nil
	// 		c := &Consumer{
	// 			Pusher:       newMockPusher(),
	// 			DeviceTokens: newMockDeviceTokensClient(),
	// 		}

	// 		_, err := c.evaluateUrgentLow(ctx, datum, testUserID, alert)

	// 		Expect(err).To(MatchError("Unable to evaluate datum: Units, Value, or Time is nil"))
	// 	})

	// 	It("can't function without datum value", func() {
	// 		ctx, _ := addLogger(context.Background())
	// 		alert := newTestUrgentLowAlert()
	// 		datum := newTestStaticDatumMmolL(11)
	// 		datum.Blood.Value = nil
	// 		c := &Consumer{
	// 			Pusher:       newMockPusher(),
	// 			DeviceTokens: newMockDeviceTokensClient(),
	// 		}

	// 		_, err := c.evaluateUrgentLow(ctx, datum, testUserID, alert)

	// 		Expect(err).To(MatchError("Unable to evaluate datum: Units, Value, or Time is nil"))
	// 	})

	// 	It("can't function without datum time", func() {
	// 		ctx, _ := addLogger(context.Background())
	// 		alert := newTestUrgentLowAlert()
	// 		datum := newTestStaticDatumMmolL(11)
	// 		datum.Blood.Time = nil
	// 		c := &Consumer{
	// 			Pusher:       newMockPusher(),
	// 			DeviceTokens: newMockDeviceTokensClient(),
	// 		}

	// 		_, err := c.evaluateUrgentLow(ctx, datum, testUserID, alert)
	// 		Expect(err).To(MatchError("Unable to evaluate datum: Units, Value, or Time is nil"))
	// 	})

	// 	It("is marked resolved", func() {
	// 		ctx, _ := addLogger(context.Background())
	// 		datum := newTestStaticDatumMmolL(11)
	// 		alert := newTestUrgentLowAlert()
	// 		alert.Threshold.Value = *datum.Blood.Value - 1
	// 		userID := "test-user-id"
	// 		c := &Consumer{
	// 			Pusher:       newMockPusher(),
	// 			DeviceTokens: newMockDeviceTokensClient(),
	// 		}

	// 		updated, err := c.evaluateUrgentLow(ctx, datum, userID, alert)
	// 		Expect(err).To(Succeed())
	// 		Expect(updated).To(BeTrue())
	// 		Expect(alert.Resolved).To(BeTemporally("~", time.Now(), time.Second))
	// 	})

	// 	It("is marked both notified and triggered", func() {
	// 		ctx, _ := addLogger(context.Background())
	// 		datum := newTestStaticDatumMmolL(11)
	// 		alert := newTestUrgentLowAlert()
	// 		alert.Threshold.Value = *datum.Blood.Value + 1
	// 		userID := "test-user-id"
	// 		c := &Consumer{
	// 			Pusher:       newMockPusher(),
	// 			DeviceTokens: newMockDeviceTokensClient(),
	// 		}

	// 		updated, err := c.evaluateUrgentLow(ctx, datum, userID, alert)
	// 		Expect(err).To(Succeed())
	// 		Expect(updated).To(BeTrue())
	// 		Expect(alert.Sent).To(BeTemporally("~", time.Now(), time.Second))
	// 		Expect(alert.Triggered).To(BeTemporally("~", time.Now(), time.Second))
	// 	})

	// 	It("sends notifications regardless of previous notification time", func() {
	// 		ctx, _ := addLogger(context.Background())
	// 		datum := newTestStaticDatumMmolL(11)
	// 		alert := newTestUrgentLowAlert()
	// 		lastTime := time.Now().Add(-10 * time.Second)
	// 		alert.Activity.Sent = lastTime
	// 		alert.Threshold.Value = *datum.Blood.Value + 1
	// 		userID := "test-user-id"
	// 		c := &Consumer{
	// 			Pusher:       newMockPusher(),
	// 			DeviceTokens: newMockDeviceTokensClient(),
	// 		}

	// 		updated, err := c.evaluateUrgentLow(ctx, datum, userID, alert)
	// 		Expect(err).To(Succeed())
	// 		Expect(updated).To(BeTrue())
	// 		Expect(alert.Sent).To(BeTemporally("~", time.Now(), time.Second))
	// 	})
	// })
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
	Tokens      alerts.TokenProvider
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
	tokens := &mockAlertsTokenProvider{Token: "test-token"}
	permissions := newMockPermissionsClient()
	evaluator := newMockStaticEvaluator()

	return &Consumer{
			Alerts:      alertsClient,
			Evaluator:   evaluator,
			Tokens:      tokens,
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
			Tokens:      tokens,
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
	tokens := newMockTokensProvider()
	return &evaluator{
			Alerts:      alertsClient,
			Data:        dataRepo,
			Permissions: permissions,
			Tokens:      tokens,
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

func newTestUrgentLowAlert() *alerts.UrgentLowAlert {
	return &alerts.UrgentLowAlert{
		Base: alerts.Base{
			Enabled:  true,
			Activity: alerts.Activity{},
		},
		Threshold: alerts.Threshold{
			Units: nontypesglucose.MmolL,
		},
	}
}

type mockDeviceTokensClient struct {
	Error  error
	Tokens []*devicetokens.DeviceToken
}

func newMockDeviceTokensClient() *mockDeviceTokensClient {
	return &mockDeviceTokensClient{
		Tokens: []*devicetokens.DeviceToken{},
	}
}

// // testingT is a subset of testing.TB
// type testingT interface {
// 	Errorf(format string, args ...any)
// 	Fatalf(format string, args ...any)
// }

func (m *mockDeviceTokensClient) GetDeviceTokens(ctx context.Context,
	userID string) ([]*devicetokens.DeviceToken, error) {

	if m.Error != nil {
		return nil, m.Error
	}
	return m.Tokens, nil
}

type mockPusher struct {
	Pushes []string
}

func newMockPusher() *mockPusher {
	return &mockPusher{
		Pushes: []string{},
	}
}

func (p *mockPusher) Push(ctx context.Context,
	deviceToken *devicetokens.DeviceToken, notification *push.Notification) error {
	p.Pushes = append(p.Pushes, notification.Message)
	return nil
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

type mockAlertsTokenProvider struct {
	Token string
	Error error
}

func (p *mockAlertsTokenProvider) ServerSessionToken() (string, error) {
	if p.Error != nil {
		return "", p.Error
	}
	return p.Token, nil
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

type mockTokensProvider struct{}

func newMockTokensProvider() *mockTokensProvider {
	return &mockTokensProvider{}
}

func (p *mockTokensProvider) ServerSessionToken() (string, error) {
	return "test-server-session-token", nil
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
