package alerts

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/devicetokens"
	"github.com/tidepool-org/platform/log"
	logtest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/push"
	"github.com/tidepool-org/platform/task"
)

var _ = Describe("CarePartnerRunner", func() {
	Describe("Run", func() {
		It("schedules its next run", func() {
			runner, test := newCarePartnerRunnerTest()

			runner.Run(test.Ctx, test.Task)

			Expect(test.Task.AvailableTime).ToNot(BeZero())
			Expect(test.Task.DeadlineTime).To(BeNil())
			Expect(test.Task.State).To(Equal(task.TaskStatePending))
		})

		Context("continues after logging errors", func() {
			It("retrieving users without communication", func() {
				runner, test := newCarePartnerRunnerTest()
				test.Alerts.OverdueCommunicationsError = fmt.Errorf("test error")

				runner.Run(test.Ctx, test.Task)

				test.Logger.AssertWarn("running care partner no communication check")
			})

			It("retrieving an alerts config", func() {
				runner, test := newCarePartnerRunnerTest()
				test.Alerts.ListError = fmt.Errorf("test error")

				runner.Run(test.Ctx, test.Task)

				Expect(func() {
					test.Logger.AssertInfo("Unable to evaluate no communication", log.Fields{
						"followedUserID": mockUserID2,
					})
				}).ToNot(Panic(), map[string]any{
					"got": quickJSON(test.Logger.SerializedFields),
				})
			})

			It("upsetting alerts configs", func() {
				runner, test := newCarePartnerRunnerTest()
				test.Alerts.UpsertError = fmt.Errorf("test error")

				runner.Run(test.Ctx, test.Task)

				Expect(func() {
					test.Logger.AssertError("Unable to upsert changed alerts config", log.Fields{
						"userID":         mockUserID1,
						"followedUserID": mockUserID2,
						"dataSetID":      mockDataSetID,
					})
				}).ToNot(Panic(), quickJSON(map[string]any{
					"got": test.Logger.SerializedFields,
				}))
			})

			It("retrieving device tokens", func() {
				runner, test := newCarePartnerRunnerTest()
				test.Tokens.GetError = fmt.Errorf("test error")

				runner.Run(test.Ctx, test.Task)

				Expect(func() {
					test.Logger.AssertInfo("unable to retrieve device tokens", log.Fields{
						"recipientUserID": mockUserID1,
					})
				}, quickJSON(map[string]any{
					"got": test.Logger.SerializedFields,
				}))
			})

			It("pushing notifications", func() {
				runner, test := newCarePartnerRunnerTest()
				test.Pusher.PushErrors = append(test.Pusher.PushErrors, fmt.Errorf("test error"))

				runner.Run(test.Ctx, test.Task)

				Expect(len(test.Pusher.PushCalls)).To(Equal(1))
				Expect(func() {
					test.Logger.AssertInfo("unable to push notification", log.Fields{
						"recipientUserID": testUserID,
					})
				}, quickJSON(map[string]any{
					"got": test.Logger.SerializedFields,
				}))
			})
		})

		It("ignores Configs that don't match the data set id", func() {
			runner, test := newCarePartnerRunnerTest()
			firstResp := test.Alerts.OverdueCommunicationsResponses[0]
			test.Alerts.OverdueCommunicationsResponses[0] = append(firstResp, LastCommunication{
				UserID:                 firstResp[0].UserID,
				DataSetID:              "non-matching",
				LastReceivedDeviceData: firstResp[0].LastReceivedDeviceData,
			})

			runner.Run(test.Ctx, test.Task)

			Expect(len(test.Pusher.PushCalls)).To(Equal(1))
		})

		It("pushes to each token", func() {
			runner, test := newCarePartnerRunnerTest()
			test.Tokens.GetResponses[0] = append(test.Tokens.GetResponses[0],
				test.Tokens.GetResponses[0][0])

			runner.Run(test.Ctx, test.Task)

			Expect(len(test.Pusher.PushCalls)).To(Equal(2))
		})

		It("pushes to each token, continuing if any experience an error", func() {
			runner, test := newCarePartnerRunnerTest()
			test.Tokens.GetResponses[0] = append(test.Tokens.GetResponses[0],
				test.Tokens.GetResponses[0][0])
			test.Pusher.PushErrors = append([]error{fmt.Errorf("test error")}, test.Pusher.PushErrors...)

			runner.Run(test.Ctx, test.Task)

			Expect(len(test.Pusher.PushCalls)).To(Equal(2))
		})

		It("ignores Configs that don't have permission", func() {
			runner, test := newCarePartnerRunnerTest()
			// disable permissions, no configs should be used
			test.Permissions.AlwaysAllow = false

			runner.Run(test.Ctx, test.Task)
			Expect(len(test.Pusher.PushCalls)).To(Equal(0))

			// reset, add a user *with* perms, and check that it works
			runner, test = newCarePartnerRunnerTest()
			test.Permissions.AlwaysAllow = false
			test.Permissions.Allow(mockUserID3, mockUserID2, permission.Follow, permission.Read)
			cfg := *test.Config
			cfg.UserID = mockUserID3
			test.Alerts.ListResponses[0] = append(test.Alerts.ListResponses[0], &cfg)
			runner.Run(test.Ctx, test.Task)
			Expect(len(test.Pusher.PushCalls)).To(Equal(1))
		})

		It("upserts configs that need it", func() {
			runner, test := newCarePartnerRunnerTest()
			runner.Run(test.Ctx, test.Task)

			// One call from needsUpsert, another when the notification is sent.
			Expect(len(test.Alerts.UpsertCalls)).To(Equal(2))
			act0 := test.Alerts.UpsertCalls[0].Activity.NoCommunication
			Expect(act0.Triggered).ToNot(BeZero())
			Expect(act0.Sent).To(BeZero())
			act1 := test.Alerts.UpsertCalls[1].Activity.NoCommunication
			Expect(act1.Sent).ToNot(BeZero())
		})

		It("upserts configs that need it, even without a notification", func() {
			runner, test := newCarePartnerRunnerTest()
			act := test.Alerts.ListResponses[0][0].Activity.NoCommunication
			act.Triggered = time.Now().Add(-time.Hour)
			act.Sent = time.Now().Add(-time.Hour)
			test.Alerts.ListResponses[0][0].Activity.NoCommunication = act
			test.Alerts.OverdueCommunicationsResponses[0][0].LastReceivedDeviceData = time.Now()

			runner.Run(test.Ctx, test.Task)

			// One call from needsUpsert, no call from sent (no notification to send)
			Expect(len(test.Alerts.UpsertCalls)).To(Equal(1))
			act0 := test.Alerts.UpsertCalls[0].Activity.NoCommunication
			Expect(act0.Resolved).To(BeTemporally("~", time.Now()))
		})

		It("doesn't re-mark itself resolved", func() {
			runner, test := newCarePartnerRunnerTest()
			act := test.Alerts.ListResponses[0][0].Activity.NoCommunication
			act.Triggered = time.Now().Add(-time.Hour)
			act.Sent = time.Now().Add(-time.Hour)
			act.Resolved = time.Now().Add(-time.Minute)
			test.Alerts.ListResponses[0][0].Activity.NoCommunication = act
			test.Alerts.OverdueCommunicationsResponses[0][0].LastReceivedDeviceData = time.Now()

			runner.Run(test.Ctx, test.Task)
			Expect(len(test.Alerts.UpsertCalls)).To(Equal(0))
		})

		It("doesn't re-send before delay", func() {
			runner, test := newCarePartnerRunnerTest()
			act := test.Alerts.ListResponses[0][0].Activity.NoCommunication
			orig := time.Now().Add(-time.Minute)
			act.Triggered = orig
			act.Sent = orig
			test.Alerts.ListResponses[0][0].Activity.NoCommunication = act

			runner.Run(test.Ctx, test.Task)
			Expect(len(test.Alerts.UpsertCalls)).To(Equal(0))
		})
	})
})

type carePartnerRunnerTest struct {
	Alerts      *mockAlertsClient
	Config      *Config
	Ctx         context.Context
	Logger      *logtest.Logger
	Permissions *mockPermissionClient
	Pusher      *mockPusher
	Task        *task.Task
	Tokens      *mockDeviceTokensClient
}

func newCarePartnerRunnerTest() (*CarePartnerRunner, *carePartnerRunnerTest) {
	alerts := newMockAlertsClient()
	ctx, lgr, cfg := newConfigTest()
	cfg.Alerts.NoCommunication.Enabled = true
	pusher := newMockPusher()
	tsk := &task.Task{}
	tokens := newMockDeviceTokensClient()
	perms := newMockPermissionClient()
	perms.AlwaysAllow = true
	authClient := newMockAuthTokenProvider()

	runner, err := NewCarePartnerRunner(lgr, alerts, tokens, pusher, perms, authClient)
	Expect(err).To(Succeed())

	last := time.Now().Add(-(DefaultNoCommunicationDelay + time.Second))
	alerts.OverdueCommunicationsResponses = [][]LastCommunication{{
		{
			UserID:                 mockUserID2,
			DataSetID:              mockDataSetID,
			LastReceivedDeviceData: last,
		},
	}}
	alerts.ListResponses = [][]*Config{{cfg}}
	tokens.GetResponses = [][]*devicetokens.DeviceToken{
		{
			{Apple: &devicetokens.AppleDeviceToken{}},
		},
	}

	return runner, &carePartnerRunnerTest{
		Alerts:      alerts,
		Config:      cfg,
		Ctx:         ctx,
		Logger:      lgr,
		Permissions: perms,
		Pusher:      pusher,
		Task:        tsk,
		Tokens:      tokens,
	}
}

type mockDeviceTokensClient struct {
	GetError     error
	GetResponses [][]*devicetokens.DeviceToken
}

func newMockDeviceTokensClient() *mockDeviceTokensClient {
	return &mockDeviceTokensClient{
		GetResponses: [][]*devicetokens.DeviceToken{},
	}
}

func (c *mockDeviceTokensClient) GetDeviceTokens(ctx context.Context, userID string) ([]*devicetokens.DeviceToken, error) {
	if c.GetError != nil {
		return nil, c.GetError
	}
	if len(c.GetResponses) > 0 {
		ret := c.GetResponses[0]
		c.GetResponses = c.GetResponses[1:]
		return ret, nil
	}
	return nil, nil
}

type mockPusher struct {
	PushCalls  []pushCall
	PushErrors []error
}

type pushCall struct {
	Token        *devicetokens.DeviceToken
	Notification *push.Notification
}

func newMockPusher() *mockPusher {
	return &mockPusher{}
}

func (p *mockPusher) Push(_ context.Context,
	token *devicetokens.DeviceToken, notification *push.Notification) error {

	p.PushCalls = append(p.PushCalls, pushCall{token, notification})
	if len(p.PushErrors) > 0 {
		err := p.PushErrors[0]
		p.PushErrors = p.PushErrors[1:]
		return err
	}
	return nil
}

type mockAuthTokenProvider struct{}

func newMockAuthTokenProvider() *mockAuthTokenProvider {
	return &mockAuthTokenProvider{}
}

func (p *mockAuthTokenProvider) ServerSessionToken() (string, error) {
	return "", nil
}
