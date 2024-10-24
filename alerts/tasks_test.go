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
				test.Alerts.UsersWithoutCommsError = fmt.Errorf("test error")

				runner.Run(test.Ctx, test.Task)

				test.Logger.AssertWarn("running care partner no communication check")
			})

			It("retrieving an alerts config", func() {
				runner, test := newCarePartnerRunnerTest()
				test.Alerts.ListError = fmt.Errorf("test error")

				runner.Run(test.Ctx, test.Task)

				test.Logger.AssertInfo("unable to evaluate no communication", log.Fields{
					"followedUserID": testFollowedUserID,
				})
			})

			It("upserting alerts configs", func() {
				runner, test := newCarePartnerRunnerTest()
				test.Alerts.UpsertError = fmt.Errorf("test error")

				runner.Run(test.Ctx, test.Task)

				test.Logger.AssertInfo("Unable to upsert alerts config", log.Fields{
					"UserID":         testUserID,
					"FollowedUserID": testFollowedUserID,
				})
			})

			It("retrieving device tokens", func() {
				runner, test := newCarePartnerRunnerTest()
				test.Tokens.GetError = fmt.Errorf("test error")

				runner.Run(test.Ctx, test.Task)

				test.Logger.AssertInfo("unable to retrieve device tokens", log.Fields{
					"recipientUserID": testUserID,
				})
			})

			It("pushes notifications", func() {
				runner, test := newCarePartnerRunnerTest()
				test.Pusher.PushErrors = append(test.Pusher.PushErrors, fmt.Errorf("test error"))

				runner.Run(test.Ctx, test.Task)

				Expect(len(test.Pusher.PushCalls)).To(Equal(1))
				test.Logger.AssertInfo("unable to push notification", log.Fields{
					"recipientUserID": testUserID,
				})
			})
		})

		It("ignores Configs that don't match the data set id", func() {
			runner, test := newCarePartnerRunnerTest()
			firstResp := test.Alerts.UsersWithoutCommsResponses[0]
			test.Alerts.UsersWithoutCommsResponses[0] = append(firstResp, LastCommunication{
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

		It("pushes to each token, even if the first experiences an error", func() {
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
			userIDWithPerm := testUserID + "2"
			test.Permissions.AlwaysAllow = false
			test.Permissions.Allow(userIDWithPerm, permission.Follow, testFollowedUserID)
			test.Alerts.ListResponses[0] = append(test.Alerts.ListResponses[0],
				&Config{
					UserID:         userIDWithPerm,
					FollowedUserID: testFollowedUserID,
					UploadID:       testDataSetID,
					Alerts: Alerts{
						NoCommunicationAlert: &NoCommunicationAlert{},
					},
				},
			)
			runner.Run(test.Ctx, test.Task)
			Expect(len(test.Pusher.PushCalls)).To(Equal(1))
		})
	})
})

type carePartnerRunnerTest struct {
	Alerts      *mockAlertsClient
	Ctx         context.Context
	Logger      *logtest.Logger
	Permissions *mockPermissionClient
	Pusher      *mockPusher
	Task        *task.Task
	Tokens      *mockDeviceTokensClient
}

func newCarePartnerRunnerTest() (*CarePartnerRunner, *carePartnerRunnerTest) {
	alerts := newMockAlertsClient()
	lgr := logtest.NewLogger()
	ctx := log.NewContextWithLogger(context.Background(), lgr)
	pusher := newMockPusher()
	tsk := &task.Task{}
	tokens := newMockDeviceTokensClient()
	perms := newMockPermissionClient()
	authClient := newMockAuthTokenProvider()
	perms.AlwaysAllow = true

	runner, err := NewCarePartnerRunner(lgr, alerts, tokens, pusher, perms, authClient)
	Expect(err).To(Succeed())

	alerts.UsersWithoutCommsResponses = [][]LastCommunication{
		{
			{
				UserID:                 testFollowedUserID,
				DataSetID:              testDataSetID,
				LastReceivedDeviceData: time.Now().Add(-12 * time.Hour),
			},
		},
	}
	alerts.ListResponses = [][]*Config{
		{
			{
				UserID:         testUserID,
				FollowedUserID: testFollowedUserID,
				UploadID:       testDataSetID,
				Alerts: Alerts{
					NoCommunicationAlert: &NoCommunicationAlert{},
				},
			},
		},
	}
	tokens.GetResponses = [][]*devicetokens.DeviceToken{
		{
			{
				Apple: &devicetokens.AppleDeviceToken{},
			},
		},
	}

	return runner, &carePartnerRunnerTest{
		Alerts:      alerts,
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
