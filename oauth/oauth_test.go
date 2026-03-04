package oauth_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/oauth"
	"github.com/tidepool-org/platform/request"
)

var _ = Describe("OAuth", func() {
	It("ProviderType is expected", func() {
		Expect(oauth.ProviderType).To(Equal("oauth"))
	})

	It("ActionAuthorize is expected", func() {
		Expect(oauth.ActionAuthorize).To(Equal("authorize"))
	})

	It("ActionRevoke is expected", func() {
		Expect(oauth.ActionRevoke).To(Equal("revoke"))
	})

	Context("IsAccessTokenError", func() {
		It("returns false if the error is nil", func() {
			Expect(oauth.IsAccessTokenError(nil)).To(BeFalse())
		})

		It("returns false if the cause of the error is not unauthenticated", func() {
			Expect(oauth.IsAccessTokenError(request.ErrorUnauthorized())).To(BeFalse())
		})

		It("returns true if the error is unauthenticated", func() {
			Expect(oauth.IsAccessTokenError(request.ErrorUnauthenticated())).To(BeTrue())
		})

		It("returns true if the cause of the error is unauthenticated", func() {
			Expect(oauth.IsAccessTokenError(errors.Wrap(request.ErrorUnauthenticated(), "a wrapper error"))).To(BeTrue())
		})
	})

	Context("IsRefreshTokenError", func() {
		It("returns false if the error is nil", func() {
			Expect(oauth.IsRefreshTokenError(nil)).To(BeFalse())
		})

		It("returns false if the error cause is nil", func() {
			testErr := errorsTest.RandomError()
			Expect(oauth.IsRefreshTokenError(testErr)).To(BeFalse())
		})

		It("returns false if the error cause is not one of the expected refresh token errors", func() {
			testCause := errors.New("one oauth2: failure three")
			testErr := errors.Wrap(testCause, "test error")
			Expect(oauth.IsRefreshTokenError(testErr)).To(BeFalse())
		})

		It("returns false if the error cause contains 'oauth2: cannot fetch token:'", func() {
			testCause := errors.New("one oauth2: cannot fetch token: three")
			testErr := errors.Wrap(testCause, "test error")
			Expect(oauth.IsRefreshTokenError(testErr)).To(BeFalse())
		})

		It(`returns true if the error cause contains 'oauth2: "invalid_grant"'`, func() {
			testCause := errors.New(`one 'oauth2: "invalid_grant" three`)
			testErr := errors.Wrap(testCause, "test error")
			Expect(oauth.IsRefreshTokenError(testErr)).To(BeTrue())
		})
	})
})
