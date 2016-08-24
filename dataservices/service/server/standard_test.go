package server_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dataservices/service/server"
	"github.com/tidepool-org/platform/log"
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

type TestAPI struct {
	HandlerOutputs []http.Handler
}

func (t *TestAPI) Close() {
	panic("Unexpected invocation of Close on TestAPI")
}

func (t *TestAPI) Handler() http.Handler {
	output := t.HandlerOutputs[0]
	t.HandlerOutputs = t.HandlerOutputs[1:]
	return output
}

var _ = Describe("Standard", func() {
	var logger log.Logger
	var handler *TestHandler
	var api *TestAPI
	var config *server.Config

	BeforeEach(func() {
		logger = log.NewNull()
		handler = &TestHandler{}
		api = &TestAPI{
			HandlerOutputs: []http.Handler{handler},
		}
		config = &server.Config{
			Address: ":8077",
		}
	})

	Context("NewStandard", func() {
		It("returns success", func() {
			Expect(server.NewStandard(logger, api, config)).ToNot(BeNil())
		})

		It("returns an error if logger is missing", func() {
			standard, err := server.NewStandard(nil, api, config)
			Expect(err).To(MatchError("server: logger is missing"))
			Expect(standard).To(BeNil())
		})

		It("returns an error if api is missing", func() {
			standard, err := server.NewStandard(logger, nil, config)
			Expect(err).To(MatchError("server: api is missing"))
			Expect(standard).To(BeNil())
		})

		It("returns an error if config is missing", func() {
			standard, err := server.NewStandard(logger, api, nil)
			Expect(err).To(MatchError("server: config is missing"))
			Expect(standard).To(BeNil())
		})

		It("returns an error if config is not valid", func() {
			config.Address = ""
			standard, err := server.NewStandard(logger, api, config)
			Expect(err).To(MatchError("server: config is invalid; server: address is missing"))
			Expect(standard).To(BeNil())
		})
	})

	// NOTE: Unable to test Serve() function as it actually starts a server (and asks for permission to do on the Mac)
})
