package v1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/user/service/api/v1"
)

var _ = Describe("Errors", func() {
	Context("ErrorUserIDMissing", func() {
		It("matches the expected error", func() {
			Expect(v1.ErrorUserIDMissing()).To(Equal(
				&service.Error{
					Code:   "user-id-missing",
					Status: 400,
					Title:  "user id is missing",
					Detail: "User id is missing",
				}))
		})
	})

	Context("ErrorUserIDNotFound", func() {
		It("matches the expected error", func() {
			Expect(v1.ErrorUserIDNotFound("1234567890abcdef")).To(Equal(
				&service.Error{
					Code:   "user-id-not-found",
					Status: 404,
					Title:  "user with specified id not found",
					Detail: "User with id 1234567890abcdef not found",
				}))
		})
	})
})
