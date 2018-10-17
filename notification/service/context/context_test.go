package context_test

import (
	"github.com/ant0ine/go-json-rest/rest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/notification/service/context"
	testService "github.com/tidepool-org/platform/notification/service/test"
	"github.com/tidepool-org/platform/notification/store"
	testStore "github.com/tidepool-org/platform/notification/store/test"
	testRest "github.com/tidepool-org/platform/test/rest"
)

var _ = Describe("Context", func() {
	var response *testRest.ResponseWriter
	var request *rest.Request
	var svc *testService.Service

	BeforeEach(func() {
		response = testRest.NewResponseWriter()
		request = testRest.NewRequest()
		svc = testService.NewService()
	})

	AfterEach(func() {
		Expect(svc.UnusedOutputsCount()).To(Equal(0))
		response.AssertOutputsEmpty()
	})

	Context("MustNew", func() {
		It("panics if notification service is missing", func() {
			Expect(func() { context.MustNew(nil, response, request) }).To(Panic())
		})

		It("panics if response is missing", func() {
			Expect(func() { context.MustNew(svc, nil, request) }).To(Panic())
		})

		It("panics if request is missing", func() {
			Expect(func() { context.MustNew(svc, response, nil) }).To(Panic())
		})

		It("returns successfully", func() {
			ctx := context.MustNew(svc, response, request)
			Expect(ctx).ToNot(BeNil())
		})
	})

	Context("New", func() {
		It("returns an error if service is missing", func() {
			ctx, err := context.New(nil, response, request)
			Expect(err).To(MatchError("service is missing"))
			Expect(ctx).To(BeNil())
		})

		It("returns an error if response is missing", func() {
			ctx, err := context.New(svc, nil, request)
			Expect(err).To(MatchError("response is missing"))
			Expect(ctx).To(BeNil())
		})

		It("returns an error if request is missing", func() {
			ctx, err := context.New(svc, response, nil)
			Expect(err).To(MatchError("request is missing"))
			Expect(ctx).To(BeNil())
		})

		It("returns successfully", func() {
			ctx, err := context.New(svc, response, request)
			Expect(err).ToNot(HaveOccurred())
			Expect(ctx).ToNot(BeNil())
		})
	})

	Context("with new context", func() {
		var ctx *context.Context

		BeforeEach(func() {
			var err error
			ctx, err = context.New(svc, response, request)
			Expect(err).ToNot(HaveOccurred())
			Expect(ctx).ToNot(BeNil())
		})

		Context("Close", func() {
			It("returns successfully", func() {
				ctx.Close()
			})
		})

		Context("with store session", func() {
			var ssn *testStore.NotificationsSession

			BeforeEach(func() {
				ssn = testStore.NewNotificationsSession()
				svc.NotificationStoreImpl.NewNotificationsSessionOutputs = []store.NotificationsSession{ssn}
			})

			AfterEach(func() {
				ssn.AssertOutputsEmpty()
			})

			Context("Close", func() {
				It("returns successfully", func() {
					ssn.CloseOutputs = []error{nil}
					Expect(ctx.NotificationsSession()).To(Equal(ssn))
					ctx.Close()
					Expect(ssn.CloseInvocations).To(Equal(1))
				})
			})

			Context("NotificationsSession", func() {
				It("returns successfully", func() {
					Expect(ctx.NotificationsSession()).To(Equal(ssn))
				})
			})
		})
	})
})
