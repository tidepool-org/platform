package push

import (
	"context"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sideshow/apns2"

	"github.com/tidepool-org/platform/devicetokens"
	"github.com/tidepool-org/platform/log"
	testlog "github.com/tidepool-org/platform/log/test"
)

const (
	testBundleID = "test-bundle-id"
)

var (
	testDeviceToken []byte = []byte("dGVzdGluZyAxIDIgMw==")
)

type pushTestDeps struct {
	Client       *mockAPNS2Client
	Token        *devicetokens.DeviceToken
	Notification *Notification
}

func testDeps() (context.Context, *APNSPusher, *pushTestDeps) {
	ctx := context.Background()
	mockClient := &mockAPNS2Client{
		Response: &apns2.Response{
			StatusCode: http.StatusOK,
		},
	}
	pusher := NewAPNSPusher(mockClient, testBundleID)
	deps := &pushTestDeps{
		Client: mockClient,
		Token: &devicetokens.DeviceToken{
			Apple: &devicetokens.AppleDeviceToken{
				Token: testDeviceToken,
			},
		},
		Notification: &Notification{},
	}
	return ctx, pusher, deps
}

var _ = Describe("APNSPusher", func() {
	Describe("Push", func() {
		It("requires an Apple token", func() {
			ctx, pusher, deps := testDeps()
			deps.Token.Apple = nil

			err := pusher.Push(ctx, deps.Token, deps.Notification)

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(ContainSubstring("can only use Apple device tokens")))
		})

		Context("its environment", func() {

			for _, env := range []string{devicetokens.AppleEnvProduction, devicetokens.AppleEnvSandbox} {
				It("is set via its token", func() {
					ctx, pusher, deps := testDeps()
					deps.Token.Apple.Environment = env

					err := pusher.Push(ctx, deps.Token, deps.Notification)

					Expect(err).To(Succeed())
					// This is reaching into the implementation of
					// APNS2Client, but there's no other way to test this.
					Expect(deps.Client.Env).To(Equal(env))
				})
			}
		})

		It("reports upstream errors", func() {
			ctx, pusher, deps := testDeps()
			deps.Client.Error = fmt.Errorf("test error")

			err := pusher.Push(ctx, deps.Token, deps.Notification)

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(ContainSubstring("test error")))
		})

		Context("when a logger is available", func() {
			It("logs", func() {
				ctx, pusher, deps := testDeps()
				testLogger := testlog.NewLogger()
				ctx = log.NewContextWithLogger(ctx, testLogger)
				deps.Client.Response = &apns2.Response{
					StatusCode: http.StatusOK,
					ApnsID:     "test-id",
				}

				err := pusher.Push(ctx, deps.Token, deps.Notification)

				Expect(err).To(Succeed())
				testLogger.AssertInfo("notification pushed", log.Fields{
					"apnsID": "test-id",
				})
			})
		})

		It("reports non-200 responses as errors", func() {
			ctx, pusher, deps := testDeps()
			deps.Client.Response = &apns2.Response{
				StatusCode: http.StatusBadRequest,
			}

			err := pusher.Push(ctx, deps.Token, deps.Notification)

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(ContainSubstring("APNs returned non-200 status")))
		})
	})
})

type mockAPNS2Client struct {
	Response *apns2.Response
	Error    error
	Env      string
}

func (c *mockAPNS2Client) Development() APNS2Client {
	c.Env = devicetokens.AppleEnvSandbox
	return c
}

func (c *mockAPNS2Client) Production() APNS2Client {
	c.Env = devicetokens.AppleEnvProduction
	return c
}

func (c *mockAPNS2Client) PushWithContext(_ apns2.Context, _ *apns2.Notification) (*apns2.Response, error) {
	if c.Error != nil {
		return nil, c.Error
	}
	if c.Response != nil {
		return c.Response, nil
	}
	return nil, nil
}
