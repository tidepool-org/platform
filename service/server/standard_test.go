package server_test

// import (
// 	"net/http"

// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"

// 	"github.com/tidepool-org/platform/log"
// 	nullLog "github.com/tidepool-org/platform/log/null"
// 	"github.com/tidepool-org/platform/service/server"
// 	testService "github.com/tidepool-org/platform/service/test"
// )

// type ServeHTTPInput struct {
// 	response http.ResponseWriter
// 	request  *http.Request
// }

// type TestHandler struct {
// 	ServeHTTPInputs []ServeHTTPInput
// }

// func (t *TestHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
// 	t.ServeHTTPInputs = append(t.ServeHTTPInputs, ServeHTTPInput{response, request})
// }

// var _ = Describe("Standard", func() {
// 	var logger log.Logger
// 	var handler *TestHandler
// 	var api *testService.API
// 	var config *server.Config

// 	BeforeEach(func() {
// 		logger = nullLog.NewLogger()
// 		handler = &TestHandler{}
// 		api = testService.NewAPI()
// 		api.HandlerOutputs = []http.Handler{handler}
// 		config = server.NewConfig()
// 		config.Address = ":9001"
// 		config.TLS = false
// 	})

// 	Context("NewStandard", func() {
// 		It("returns success", func() {
// 			Expect(server.NewStandard(logger, api, config)).ToNot(BeNil())
// 		})

// 		It("returns an error if logger is missing", func() {
// 			standard, err := server.NewStandard(nil, api, config)
// 			Expect(err).To(MatchError("logger is missing"))
// 			Expect(standard).To(BeNil())
// 		})

// 		It("returns an error if api is missing", func() {
// 			standard, err := server.NewStandard(logger, nil, config)
// 			Expect(err).To(MatchError("api is missing"))
// 			Expect(standard).To(BeNil())
// 		})

// 		It("returns an error if config is missing", func() {
// 			standard, err := server.NewStandard(logger, api, nil)
// 			Expect(err).To(MatchError("config is missing"))
// 			Expect(standard).To(BeNil())
// 		})

// 		It("returns an error if config is not valid", func() {
// 			config.Address = ""
// 			standard, err := server.NewStandard(logger, api, config)
// 			Expect(err).To(MatchError("config is invalid; address is missing"))
// 			Expect(standard).To(BeNil())
// 		})
// 	})

// 	// NOTE: Unable to test Serve() function as it actually starts a server (and asks for permission to do on the Mac)
// })
