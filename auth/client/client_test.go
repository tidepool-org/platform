package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	authClient "github.com/tidepool-org/platform/auth/client"
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Client", func() {
	var serverTokenSecret string
	var name string
	var logger log.Logger

	BeforeEach(func() {
		serverTokenSecret = authTest.NewServiceSecret()
		name = test.RandomStringFromRangeAndCharset(4, 16, test.CharsetAlphaNumeric)
		logger = logTest.NewLogger()
		Expect(logger).ToNot(BeNil())
	})

	Context("NewClient", func() {
		var config *authClient.Config

		BeforeEach(func() {
			config = authClient.NewConfig()
			Expect(config).ToNot(BeNil())
			config.ExternalConfig.AuthenticationConfig.Address = testHttp.NewAddress()
			config.ExternalConfig.AuthenticationConfig.UserAgent = testHttp.NewUserAgent()
			config.ExternalConfig.ServerSessionTokenSecret = serverTokenSecret
		})

		It("returns an error if config is missing", func() {
			client, err := authClient.NewClient(nil, name, logger)
			errorsTest.ExpectEqual(err, errors.New("config is missing"))
			Expect(client).To(BeNil())
		})

		It("returns an error if name is missing", func() {
			client, err := authClient.NewClient(config, "", logger)
			errorsTest.ExpectEqual(err, errors.New("name is missing"))
			Expect(client).To(BeNil())
		})

		It("returns an error if logger is missing", func() {
			client, err := authClient.NewClient(config, name, nil)
			errorsTest.ExpectEqual(err, errors.New("logger is missing"))
			Expect(client).To(BeNil())
		})

		It("returns an error if config server session token secret is missing", func() {
			config.ExternalConfig.ServerSessionTokenSecret = ""
			client, err := authClient.NewClient(config, name, logger)
			errorsTest.ExpectEqual(err, errors.New("config is invalid"))
			Expect(client).To(BeNil())
		})

		It("returns an error if config external authentication address is missing", func() {
			config.ExternalConfig.AuthenticationConfig.Address = ""
			client, err := authClient.NewClient(config, name, logger)
			errorsTest.ExpectEqual(err, errors.New("config is invalid"))
			Expect(client).To(BeNil())
		})

		It("returns success", func() {
			client, err := authClient.NewClient(config, name, logger)
			Expect(err).ToNot(HaveOccurred())
			Expect(client).ToNot(BeNil())
		})
	})

	Context("with started server and new client", func() {
		var server *Server
		var config *authClient.Config
		var client *authClient.Client

		BeforeEach(func() {
			server = NewServer()
			config = authClient.NewConfig()
			Expect(config).ToNot(BeNil())
			config.ExternalConfig.AuthenticationConfig.Address = server.URL()
			config.ExternalConfig.AuthenticationConfig.UserAgent = testHttp.NewUserAgent()
			config.ExternalConfig.ServerSessionTokenSecret = serverTokenSecret
		})

		JustBeforeEach(func() {
			var err error
			client, err = authClient.NewClient(config, name, logger)
			Expect(err).ToNot(HaveOccurred())
			Expect(client).ToNot(BeNil())
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

	})
})
