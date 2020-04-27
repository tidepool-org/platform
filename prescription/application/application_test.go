package application_test

import (
	provider "github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/service"
	"net/http"
	"os"

	"go.uber.org/fx"

	"github.com/tidepool-org/platform/prescription/application"

	authTest "github.com/tidepool-org/platform/auth/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"github.com/tidepool-org/platform/test"

	applicationTest "github.com/tidepool-org/platform/application/test"
	configTest "github.com/tidepool-org/platform/config/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Application", func() {
	Context("with started server, config reporter, and new service", func() {
		var prvdr *applicationTest.Provider
		var prescriptionStoreConfig map[string]interface{}
		var prescriptionServiceConfig map[string]interface{}
		var authClientConfig map[string]interface{}
		var serverSecret string
		var sessionToken string
		var server *Server

		type Result struct {
			fx.In
			Routers  []service.Router `group:"routers"`
		}

		BeforeEach(func() {
			prvdr = applicationTest.NewProviderWithDefaults()
			serverSecret = authTest.NewServiceSecret()
			sessionToken = authTest.NewSessionToken()
			server = NewServer()
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/auth/serverlogin"),
					VerifyHeaderKV("X-Tidepool-Server-Name", *prvdr.NameOutput),
					VerifyHeaderKV("X-Tidepool-Server-Secret", serverSecret),
					VerifyBody(nil),
					RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{sessionToken}})),
			)

			authClientConfig = map[string]interface{}{
				"address":             server.URL(),
				"server_token_secret": authTest.NewServiceSecret(),
				"external": map[string]interface{}{
					"address":                     server.URL(),
					"server_session_token_secret": serverSecret,
				},
			}

			prescriptionStoreConfig = map[string]interface{}{
				"addresses": os.Getenv("TIDEPOOL_STORE_ADDRESSES"),
				"database":  test.RandomStringFromRangeAndCharset(4, 8, test.CharsetLowercase),
				"tls":       "false",
			}

			prescriptionServiceConfig = map[string]interface{}{
				"auth": map[string]interface{}{
					"client": authClientConfig,
				},
				"domain": "test.com",
				"secret": authTest.NewServiceSecret(),
				"prescription": map[string]interface{}{
					"store": prescriptionStoreConfig,
				},
				"server": map[string]interface{}{
					"address": testHttp.NewAddress(),
					"tls":     "false",
				},
				"user": map[string]interface{}{
					"client": map[string]interface{}{
						"address": server.URL(),
					},
				},
			}

			(*prvdr.ConfigReporterOutput).(*configTest.Reporter).Config = prescriptionServiceConfig

		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		Describe("Application Dependency Graph", func() {
			var routers []service.Router = nil
			var dependencies []fx.Option = nil

			BeforeEach(func() {
				dependencies = []fx.Option{
					fx.Provide(func() provider.Provider { return prvdr }),
					provider.ProviderComponentsModule,
					application.Prescription,
					fx.Invoke(func(res Result) {
						routers = res.Routers
					}),
				}
			})

			AfterEach(func() {
				routers = nil
				dependencies = nil
			})

			It("can be built successfully", func() {
				app := fx.New(dependencies...)
				Expect(app.Err()).ToNot(HaveOccurred())
			})

			It("exposes application routers", func() {
				fx.New(dependencies...)
				Expect(routers).ToNot(BeEmpty())
			})
		})
	})
})
