package client_test

import (
	"context"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	authClient "github.com/tidepool-org/platform/auth/client"
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("External", func() {
	var config *authClient.ExternalConfig
	var name string
	var logger *logTest.Logger

	BeforeEach(func() {
		config = authClient.NewExternalConfig()
		config.AuthenticationConfig.UserAgent = testHttp.NewUserAgent()
		config.ServerSessionTokenSecret = authTest.NewServiceSecret()
		name = test.RandomString()
		logger = logTest.NewLogger()
	})

	Context("NewExternal", func() {
		BeforeEach(func() {
			config.AuthenticationConfig.Address = testHttp.NewAddress()
		})

		It("returns an error when the config is missing", func() {
			config = nil
			client, err := authClient.NewExternal(config, name, logger)
			errorsTest.ExpectEqual(err, errors.New("config is missing"))
			Expect(client).To(BeNil())
		})

		It("returns an error when the name is missing", func() {
			name = ""
			client, err := authClient.NewExternal(config, name, logger)
			errorsTest.ExpectEqual(err, errors.New("name is missing"))
			Expect(client).To(BeNil())
		})

		It("returns an error when the logger is missing", func() {
			logger = nil
			client, err := authClient.NewExternal(config, name, nil)
			errorsTest.ExpectEqual(err, errors.New("logger is missing"))
			Expect(client).To(BeNil())
		})

		It("returns success", func() {
			Expect(authClient.NewExternal(config, name, logger)).ToNot(BeNil())
		})
	})

	Context("with server and new client", func() {
		var server *Server
		var requestHandlers []http.HandlerFunc
		var client *authClient.External
		var sessionToken string
		var details request.Details
		var ctx context.Context

		BeforeEach(func() {
			server = NewServer()
			requestHandlers = nil
			sessionToken = authTest.NewSessionToken()
			details = request.NewDetails(request.MethodSessionToken, "", sessionToken, "patient")
			ctx = context.Background()
			ctx = log.NewContextWithLogger(ctx, logger)
		})

		JustBeforeEach(func() {
			server.AppendHandlers(CombineHandlers(requestHandlers...))
			var err error
			config.AuthenticationConfig.Address = server.URL()
			client, err = authClient.NewExternal(config, name, logger)
			Expect(err).ToNot(HaveOccurred())
			Expect(client).ToNot(BeNil())
			ctx = request.NewContextWithDetails(ctx, details)
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})
	})
})
