package oauth_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/oauth"
)

var _ = Describe("OAuth", func() {
	Context("IsRefreshTokenError", func() {
		It("returns false is the error is nil", func() {
			Expect(oauth.IsRefreshTokenError(nil)).To(BeFalse())
		})

		It("returns false is the error cause is nil", func() {
			testErr := errorsTest.RandomError()
			Expect(oauth.IsRefreshTokenError(testErr)).To(BeFalse())
		})

		It("returns false is the error cause is not one of the expected refresh token errors", func() {
			testCause := errors.New("one oauth2: failure three")
			testErr := errors.Wrap(testCause, "test error")
			Expect(oauth.IsRefreshTokenError(testErr)).To(BeFalse())
		})

		It("returns false is the error cause contains 'oauth2: cannot fetch token:'", func() {
			testCause := errors.New("one oauth2: cannot fetch token: three")
			testErr := errors.Wrap(testCause, "test error")
			Expect(oauth.IsRefreshTokenError(testErr)).To(BeFalse())
		})

		It(`returns true is the error cause contains 'oauth2: "invalid_grant"'`, func() {
			testCause := errors.New(`one 'oauth2: "invalid_grant" three`)
			testErr := errors.Wrap(testCause, "test error")
			Expect(oauth.IsRefreshTokenError(testErr)).To(BeTrue())
		})
	})
})
