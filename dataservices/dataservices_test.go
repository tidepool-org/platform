package main_test

import (
	"net/http"

	. "github.com/tidepool-org/platform/dataservices"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/version"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("The Dataservices client", func() {
	var client *DataServiceClient

	BeforeEach(func() {
		client = NewDataServiceClient()
	})

	AfterEach(func() {
		//shut down the server between tests
	})

	Describe("version", func() {
		It("should return status 200", func() {
			recorded := service.RunRequest(client.GetVersion, service.MakeSimpleRequest("GET", "http://localhost/version", nil))
			Expect(recorded.CodeIs(http.StatusOK)).To(BeTrue(), "Should have been 200 OK")
		})
		It("should return version as the body", func() {
			recorded := service.RunRequest(client.GetVersion, service.MakeSimpleRequest("GET", "http://localhost/version", nil))
			Expect(recorded.BodyIs(version.String)).To(BeTrue(), "Expected "+recorded.Recorder.Body.String()+" to be "+version.String)
		})
	})

})
