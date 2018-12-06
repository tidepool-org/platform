package server_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/log"
	nullLog "github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/service/server"
	testService "github.com/tidepool-org/platform/service/test"
)

type ServeHTTPInput struct {
	response http.ResponseWriter
	request  *http.Request
}

type TestHandler struct {
	ServeHTTPInputs []ServeHTTPInput
}

func (t *TestHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	t.ServeHTTPInputs = append(t.ServeHTTPInputs, ServeHTTPInput{response, request})
}

var _ = Describe("Standard", func() {
	var lgr log.Logger
	var hndlr *TestHandler
	var api *testService.API
	var cfg *server.Config

	BeforeEach(func() {
		lgr = nullLog.NewLogger()
		hndlr = &TestHandler{}
		api = testService.NewAPI()
		api.HandlerOutputs = []http.Handler{hndlr}
		cfg = server.NewConfig()
		cfg.Address = ":9001"
		cfg.TLS = false
	})

	Context("NewStandard", func() {
		It("returns success", func() {
			Expect(server.NewStandard(cfg, lgr, api)).ToNot(BeNil())
		})

		It("returns an error if logger is missing", func() {
			standard, err := server.NewStandard(cfg, nil, api)
			Expect(err).To(MatchError("logger is missing"))
			Expect(standard).To(BeNil())
		})

		It("returns an error if api is missing", func() {
			standard, err := server.NewStandard(cfg, lgr, nil)
			Expect(err).To(MatchError("api is missing"))
			Expect(standard).To(BeNil())
		})

		It("returns an error if config is missing", func() {
			standard, err := server.NewStandard(nil, lgr, api)
			Expect(err).To(MatchError("config is missing"))
			Expect(standard).To(BeNil())
		})

		It("returns an error if config is not valid", func() {
			cfg.Address = ""
			standard, err := server.NewStandard(cfg, lgr, api)
			Expect(err).To(MatchError("config is invalid; address is missing"))
			Expect(standard).To(BeNil())
		})
	})

	// NOTE: Unable to test Serve() function as it actually starts a server (and asks for permission to do on the Mac)
})
