package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/platform"
	testHttp "github.com/tidepool-org/platform/test/http"
	userClient "github.com/tidepool-org/platform/user/client"
)

var _ = Describe("Client", func() {
	var config *platform.Config
	var authorizeAs platform.AuthorizeAs

	BeforeEach(func() {
		config = platform.NewConfig()
		config.UserAgent = testHttp.NewUserAgent()
		authorizeAs = platform.AuthorizeAsService
	})

	Context("New", func() {
		BeforeEach(func() {
			config.Address = testHttp.NewAddress()
		})

		It("returns an error when the config is missing", func() {
			config = nil
			client, err := userClient.New(nil, authorizeAs)
			errorsTest.ExpectEqual(err, errors.New("config is missing"))
			Expect(client).To(BeNil())
		})

		It("returns an error when the authorize as is invalid", func() {
			authorizeAs = platform.AuthorizeAs(-1)
			client, err := userClient.New(config, authorizeAs)
			errorsTest.ExpectEqual(err, errors.New("authorize as is invalid"))
			Expect(client).To(BeNil())
		})

		It("returns success", func() {
			Expect(userClient.New(config, authorizeAs)).ToNot(BeNil())
		})
	})
})
