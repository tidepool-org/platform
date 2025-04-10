package service_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	applicationTest "github.com/tidepool-org/platform/application/test"
	authTest "github.com/tidepool-org/platform/auth/test"
	configTest "github.com/tidepool-org/platform/config/test"
	serviceService "github.com/tidepool-org/platform/service/service"
)

var _ = Describe("Authenticated", func() {

	var provider *applicationTest.Provider
	var svc *serviceService.Authenticated
	var serverSecret string
	var sessionToken string
	var authClientConfig map[string]interface{}
	var serviceConfig map[string]interface{}
	var serverConfig map[string]interface{}
	var testServer *Server

	BeforeEach(func() {
		provider = applicationTest.NewProviderWithDefaults()

		serverSecret = authTest.NewServiceSecret()
		sessionToken = authTest.NewSessionToken()

		testServer = NewServer()
		testServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest("POST", "/auth/serverlogin"),
				VerifyHeaderKV("X-Tidepool-Server-Name", *provider.NameOutput),
				VerifyHeaderKV("X-Tidepool-Server-Secret", serverSecret),
				VerifyBody(nil),
				RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{sessionToken}})),
		)

		authClientConfig = map[string]interface{}{
			"address":             testServer.URL(),
			"server_token_secret": authTest.NewServiceSecret(),
			"external": map[string]interface{}{
				"address":                     testServer.URL(),
				"server_session_token_secret": serverSecret,
			},
		}
		serverConfig = map[string]interface{}{
			"address": testServer.URL(),
			"tls":     "false",
		}
		serviceConfig = map[string]interface{}{
			"secret": authTest.NewServiceSecret(),
			"server": serverConfig,
			"auth": map[string]interface{}{
				"client": authClientConfig,
			},
		}
		(*provider.ConfigReporterOutput).(*configTest.Reporter).Config = serviceConfig

		svc = serviceService.NewAuthenticated()
		Expect(svc).ToNot(BeNil())

		Expect(svc.Initialize(provider)).To(Succeed())
	})

	AfterEach(func() {
		if svc != nil {
			svc.Terminate()
		}
		provider.AssertOutputsEmpty()
	})

	It("returns the secret", func() {
		Expect(svc.Secret()).To(Equal(serviceConfig["secret"]))
	})

	Context("AuthClient", func() {
		It("returns successfully with server token", func() {
			authClient := svc.AuthClient()
			Expect(authClient).ToNot(BeNil())
			Eventually(authClient.ServerSessionToken).Should(Equal(sessionToken))
		})
	})
})
