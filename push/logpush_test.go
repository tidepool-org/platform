package push

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/devicetokens"
	"github.com/tidepool-org/platform/log"
	logtest "github.com/tidepool-org/platform/log/test"
)

var _ = Describe("NewLogPusher", func() {
	It("succeeds", func() {
		testLog := logtest.NewLogger()

		Expect(NewLogPusher(testLog)).ToNot(Equal(nil))
	})

	It("implements Push by logging a message", func() {
		testLog := logtest.NewLogger()
		ctx := context.Background()
		testToken := &devicetokens.DeviceToken{}
		testNotification := &Notification{}

		pusher := NewLogPusher(testLog)
		Expect(pusher).ToNot(Equal(nil))

		Expect(pusher.Push(ctx, testToken, testNotification)).To(Succeed())
		testFields := log.Fields{
			"deviceToken":  testToken,
			"notification": testNotification,
		}
		testLog.AssertInfo("logging push notification", testFields)
	})

	It("handles being passed a nil logger", func() {
		ctx := context.Background()
		testToken := &devicetokens.DeviceToken{}
		testNotification := &Notification{}

		pusher := NewLogPusher(nil)
		Expect(pusher).ToNot(Equal(nil))

		Expect(func() {
			Expect(pusher.Push(ctx, testToken, testNotification)).To(Succeed())
		}).ToNot(Panic())
	})
})
