package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"
	"net/http"
	"net/url"

	"github.com/tidepool-org/platform/userservices/client"
)

var _ = Describe("Client", func() {
	Context("OwnerPermission", func() {
		It("has the expected permissions", func() {
			Expect(client.OwnerPermission).To(Equal("root"))
		})
	})

	Context("CustodianPermission", func() {
		It("has the expected permissions", func() {
			Expect(client.CustodianPermission).To(Equal("custodian"))
		})
	})

	Context("UploadPermission", func() {
		It("has the expected permissions", func() {
			Expect(client.UploadPermission).To(Equal("upload"))
		})
	})

	Context("ViewPermission", func() {
		It("has the expected permissions", func() {
			Expect(client.ViewPermission).To(Equal("view"))
		})
	})

	Context("UnauthorizedError", func() {
		var unauthorizedError *client.UnauthorizedError

		BeforeEach(func() {
			unauthorizedError = client.NewUnauthorizedError()
		})

		It("is valid", func() {
			Expect(unauthorizedError).ToNot(BeNil())
		})

		It("has the expected error", func() {
			Expect(unauthorizedError.Error()).To(Equal("client: unauthorized"))
		})
	})

	Context("IsUnauthorizedError", func() {
		It("returns true for an UnauthorizedError", func() {
			Expect(client.IsUnauthorizedError(client.NewUnauthorizedError())).To(BeTrue())
		})

		It("returns false for any other type of error", func() {
			Expect(client.IsUnauthorizedError(errors.New("other"))).To(BeFalse())
		})
	})

	Context("NewUnexpectedResponseError", func() {
		var unexpectedResponseError *client.UnexpectedResponseError

		BeforeEach(func() {
			url, _ := url.Parse("http://localhost:1234")
			request := &http.Request{Method: "GET", URL: url}
			response := &http.Response{StatusCode: 400}
			unexpectedResponseError = client.NewUnexpectedResponseError(response, request)
		})

		It("is valid", func() {
			Expect(unexpectedResponseError).ToNot(BeNil())
		})

		It("has the expected error", func() {
			Expect(unexpectedResponseError.Error()).To(Equal("client: unexpected response status code 400 from GET http://localhost:1234"))
		})
	})
})
