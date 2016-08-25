package service_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/service"
)

var _ = Describe("Errors", func() {
	Context("ErrorInternalServerFailure", func() {
		It("matches the expected error", func() {
			Expect(service.ErrorInternalServerFailure()).To(Equal(
				&service.Error{
					Code:   "internal-server-failure",
					Status: 500,
					Title:  "internal server failure",
					Detail: "Internal server failure",
				}))
		})
	})

	Context("ErrorJSONMalformed", func() {
		It("matches the expected error", func() {
			Expect(service.ErrorJSONMalformed()).To(Equal(
				&service.Error{
					Code:   "json-malformed",
					Status: 400,
					Title:  "json is malformed",
					Detail: "JSON is malformed",
				}))
		})
	})

	Context("ErrorAuthenticationTokenMissing", func() {
		It("matches the expected error", func() {
			Expect(service.ErrorAuthenticationTokenMissing()).To(Equal(
				&service.Error{
					Code:   "authentication-token-missing",
					Status: 401,
					Title:  "authentication token missing",
					Detail: "Authentication token missing",
				}))
		})
	})

	Context("ErrorUnauthenticated", func() {
		It("matches the expected error", func() {
			Expect(service.ErrorUnauthenticated()).To(Equal(
				&service.Error{
					Code:   "unauthenticated",
					Status: 401,
					Title:  "authentication token is invalid",
					Detail: "Authentication token is invalid",
				}))
		})
	})

	Context("ErrorUnauthorized", func() {
		It("matches the expected error", func() {
			Expect(service.ErrorUnauthorized()).To(Equal(
				&service.Error{
					Code:   "unauthorized",
					Status: 403,
					Title:  "authentication token is not authorized for requested action",
					Detail: "Authentication token is not authorized for requested action",
				}))
		})
	})
})
