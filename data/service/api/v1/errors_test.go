package v1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/service/api/v1"
	"github.com/tidepool-org/platform/service"
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

	Context("ErrorDataSetIDMissing", func() {
		It("matches the expected error", func() {
			Expect(v1.ErrorDataSetIDMissing()).To(Equal(
				&service.Error{
					Code:   "data-set-id-missing",
					Status: 400,
					Title:  "data set id is missing",
					Detail: "Data set id is missing",
				}))
		})
	})

	Context("ErrorDataSetIDNotFound", func() {
		It("matches the expected error", func() {
			Expect(v1.ErrorDataSetIDNotFound("1234567890abcdef")).To(Equal(
				&service.Error{
					Code:   "data-set-id-not-found",
					Status: 404,
					Title:  "data set with specified id not found",
					Detail: "Data set with id 1234567890abcdef not found",
				}))
		})
	})

	Context("ErrorDataSetClosed", func() {
		It("matches the expected error", func() {
			Expect(v1.ErrorDataSetClosed("1234567890abcdef")).To(Equal(
				&service.Error{
					Code:   "data-set-closed",
					Status: 409,
					Title:  "data set with specified id is closed for new data",
					Detail: "Data set with id 1234567890abcdef is closed for new data",
				}))
		})
	})
})
