package client_test

import (
	"context"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/task"
	taskClient "github.com/tidepool-org/platform/task/client"
	testHttp "github.com/tidepool-org/platform/test/http"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Client", func() {
	var config *platform.Config
	var authorizeAs platform.AuthorizeAs

	BeforeEach(func() {
		config = platform.NewConfig()
		config.UserAgent = testHttp.NewUserAgent()
	})

	Context("New", func() {
		BeforeEach(func() {
			config.Address = testHttp.NewAddress()
			authorizeAs = platform.AuthorizeAsService
		})

		It("returns an error when the config is missing", func() {
			config = nil
			client, err := taskClient.New(nil, authorizeAs)
			errorsTest.ExpectEqual(err, errors.New("config is missing"))
			Expect(client).To(BeNil())
		})

		It("returns an error when the authorize as is invalid", func() {
			authorizeAs = platform.AuthorizeAs(-1)
			client, err := taskClient.New(config, authorizeAs)
			errorsTest.ExpectEqual(err, errors.New("authorize as is invalid"))
			Expect(client).To(BeNil())
		})

		It("returns success", func() {
			Expect(taskClient.New(config, authorizeAs)).ToNot(BeNil())
		})
	})

	Context("with server and new client", func() {
		var server *Server
		var requestHandlers []http.HandlerFunc
		var logger *logTest.Logger
		var ctx context.Context
		var client task.Client

		BeforeEach(func() {
			server = NewServer()
			requestHandlers = nil
			logger = logTest.NewLogger()
			ctx = log.NewContextWithLogger(context.Background(), logger)
		})

		JustBeforeEach(func() {
			server.AppendHandlers(CombineHandlers(requestHandlers...))
			config.Address = server.URL()
			var err error
			client, err = taskClient.New(config, authorizeAs)
			Expect(err).ToNot(HaveOccurred())
			Expect(client).ToNot(BeNil())
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		clientAssertions := func() {
		}

		When("client must authorize as service", func() {
			BeforeEach(func() {
				config.ServiceSecret = authTest.NewServiceSecret()
				authorizeAs = platform.AuthorizeAsService
				requestHandlers = append(requestHandlers, VerifyHeaderKV("X-Tidepool-Service-Secret", config.ServiceSecret))
			})

			clientAssertions()
		})

		When("client must authorize as user", func() {
			BeforeEach(func() {
				sessionToken := authTest.NewSessionToken()
				authorizeAs = platform.AuthorizeAsUser
				requestHandlers = append(requestHandlers, VerifyHeaderKV("X-Tidepool-Session-Token", sessionToken))
				ctx = request.NewContextWithDetails(ctx, request.NewDetails(request.MethodAccessToken, userTest.RandomID(), sessionToken))
			})

			clientAssertions()
		})
	})
})
