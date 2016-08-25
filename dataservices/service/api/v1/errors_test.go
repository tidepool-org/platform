package v1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dataservices/service/api/v1"
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

	Context("ErrorDatasetIDMissing", func() {
		It("matches the expected error", func() {
			Expect(v1.ErrorDatasetIDMissing()).To(Equal(
				&service.Error{
					Code:   "dataset-id-missing",
					Status: 400,
					Title:  "dataset id is missing",
					Detail: "Dataset id is missing",
				}))
		})
	})

	Context("ErrorDatasetIDNotFound", func() {
		It("matches the expected error", func() {
			Expect(v1.ErrorDatasetIDNotFound("1234567890abcdef")).To(Equal(
				&service.Error{
					Code:   "dataset-id-not-found",
					Status: 404,
					Title:  "dataset with specified id not found",
					Detail: "Dataset with id 1234567890abcdef not found",
				}))
		})
	})

	Context("ErrorDatasetClosed", func() {
		It("matches the expected error", func() {
			Expect(v1.ErrorDatasetClosed("1234567890abcdef")).To(Equal(
				&service.Error{
					Code:   "dataset-closed",
					Status: 409,
					Title:  "dataset with specified id is closed for new data",
					Detail: "Dataset with id 1234567890abcdef is closed for new data",
				}))
		})
	})
})
