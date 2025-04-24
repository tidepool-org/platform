package auth_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"

	"github.com/tidepool-org/platform/auth"
	authTest "github.com/tidepool-org/platform/auth/test"
)

var _ = Describe("Client", func() {
	Context("TidepoolSessionTokenHeaderKey", func() {
		It("is the correct header name", func() {
			Expect(auth.TidepoolSessionTokenHeaderKey).To(Equal("X-Tidepool-Session-Token"))
		})
	})

	Context("NewContextWithServerSessionTokenProvider", func() {
		var serverSessionTokenProviderController *gomock.Controller
		var serverSessionTokenProvider *authTest.MockServerSessionTokenProvider
		var ctx context.Context

		BeforeEach(func() {
			serverSessionTokenProviderController = gomock.NewController(GinkgoT())
			serverSessionTokenProvider = authTest.NewMockServerSessionTokenProvider(serverSessionTokenProviderController)
			ctx = context.Background()
		})

		AfterEach(func() {
			serverSessionTokenProviderController.Finish()
		})

		It("persists the server session token provider", func() {
			ctx = auth.NewContextWithServerSessionTokenProvider(ctx, serverSessionTokenProvider)
			Expect(auth.ServerSessionTokenProviderFromContext(ctx)).To(Equal(serverSessionTokenProvider))
		})
	})

	Context("ServerSessionTokenProviderFromContext", func() {
		var serverSessionTokenProviderController *gomock.Controller
		var serverSessionTokenProvider *authTest.MockServerSessionTokenProvider
		var ctx context.Context

		BeforeEach(func() {
			serverSessionTokenProviderController = gomock.NewController(GinkgoT())
			serverSessionTokenProvider = authTest.NewMockServerSessionTokenProvider(serverSessionTokenProviderController)
			ctx = context.Background()
		})

		AfterEach(func() {
			serverSessionTokenProviderController.Finish()
		})

		It("returs nil if the context is nil", func() {
			Expect(auth.ServerSessionTokenProviderFromContext(nil)).To(BeNil())
		})

		It("returns nil if there is no server session token provider", func() {
			Expect(auth.ServerSessionTokenProviderFromContext(ctx)).To(BeNil())
		})

		It("obtains the server session token provider", func() {
			ctx = auth.NewContextWithServerSessionTokenProvider(ctx, serverSessionTokenProvider)
			Expect(auth.ServerSessionTokenProviderFromContext(ctx)).To(Equal(serverSessionTokenProvider))
		})
	})
})
