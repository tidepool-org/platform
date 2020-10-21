package context_test

import (
	"github.com/ant0ine/go-json-rest/rest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/notification/service/context"
	notificationServiceTest "github.com/tidepool-org/platform/notification/service/test"
	"github.com/tidepool-org/platform/notification/store"
	notificationStoreTest "github.com/tidepool-org/platform/notification/store/test"
	testRest "github.com/tidepool-org/platform/test/rest"
)

var _ = Describe("Context", func() {
	var response *testRest.ResponseWriter
	var request *rest.Request
	var svc *notificationServiceTest.Service

	BeforeEach(func() {
		response = testRest.NewResponseWriter()
		request = testRest.NewRequest()
		svc = notificationServiceTest.NewService()
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

		Context("with store repository", func() {
			var repository *notificationStoreTest.NotificationsRepository

			BeforeEach(func() {
				repository = notificationStoreTest.NewNotificationsRepository()
				svc.NotificationStoreImpl.NewNotificationsRepositoryOutputs = []store.NotificationsRepository{repository}
			})

			AfterEach(func() {
				repository.AssertOutputsEmpty()
			})

			Context("Close", func() {
				It("returns successfully", func() {
					Expect(ctx.NotificationsRepository()).To(Equal(repository))
					ctx.Close()
				})
			})

			Context("NotificationsRepository", func() {
				It("returns successfully", func() {
					Expect(ctx.NotificationsRepository()).To(Equal(repository))
				})
			})
		})
	})
})
