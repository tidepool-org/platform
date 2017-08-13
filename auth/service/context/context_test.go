package context_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth/service/context"
	testService "github.com/tidepool-org/platform/auth/service/test"
	"github.com/tidepool-org/platform/auth/store"
	testStore "github.com/tidepool-org/platform/auth/store/test"
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
		Expect(response.UnusedOutputsCount()).To(Equal(0))
	})

	Context("MustNew", func() {
		It("panics if auth service is missing", func() {
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
			Expect(err).To(MatchError("context: service is missing"))
			Expect(ctx).To(BeNil())
		})

		It("returns an error if response is missing", func() {
			ctx, err := context.New(svc, nil, request)
			Expect(err).To(MatchError("context: response is missing"))
			Expect(ctx).To(BeNil())
		})

		It("returns an error if request is missing", func() {
			ctx, err := context.New(svc, response, nil)
			Expect(err).To(MatchError("context: request is missing"))
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
			var ss *testStore.StoreSession

			BeforeEach(func() {
				ss = testStore.NewStoreSession()
				svc.AuthStoreImpl.NewSessionOutputs = []store.StoreSession{ss}
			})

			AfterEach(func() {
				Expect(ss.UnusedOutputsCount()).To(Equal(0))
			})

			Context("Close", func() {
				It("returns successfully", func() {
					Expect(ctx.AuthStoreSession()).To(Equal(ss))
					ctx.Close()
					Expect(ss.CloseInvocations).To(Equal(1))
				})
			})

			Context("AuthStoreSession", func() {
				It("returns successfully", func() {
					Expect(ctx.AuthStoreSession()).To(Equal(ss))
					Expect(ss.SetAgentInvocations).To(Equal(1))
				})
			})
		})
	})
})
