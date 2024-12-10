package alerts

import (
	"context"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/platform"
)

var _ = Describe("Client", func() {
	var test404Server, test200Server *httptest.Server

	BeforeEach(func() {
		t := GinkgoT()
		// There's no need to create these before each test, but I can't get
		// Ginkgo to let me start these just once.
		test404Server = testServer(t, func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		})
		test200Server = testServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	Context("Delete", func() {
		It("returns an error on non-200 responses", func() {
			client, ctx := newAlertsClientTest(test404Server)
			err := client.Delete(ctx, &Config{})
			Expect(err).Should(HaveOccurred())
			Expect(err).To(MatchError(ContainSubstring("resource not found")))
		})

		It("returns nil on success", func() {
			client, ctx := newAlertsClientTest(test200Server)
			err := client.Delete(ctx, &Config{})
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Context("Upsert", func() {
		It("returns an error on non-200 responses", func() {
			client, ctx := newAlertsClientTest(test404Server)
			err := client.Upsert(ctx, &Config{})
			Expect(err).Should(HaveOccurred())
			Expect(err).To(MatchError(ContainSubstring("resource not found")))
		})

		It("returns nil on success", func() {
			client, ctx := newAlertsClientTest(test200Server)
			err := client.Upsert(ctx, &Config{})
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})

func buildTestClient(s *httptest.Server) *Client {
	pCfg := &platform.Config{
		Config:        &client.Config{Address: s.URL},
		ServiceSecret: "auth-me",
	}
	pc, err := platform.NewClient(pCfg, platform.AuthorizeAsService)
	Expect(err).ToNot(HaveOccurred())
	client := NewClient(pc, null.NewLogger())
	return client
}

func newAlertsClientTest(server *httptest.Server) (*Client, context.Context) {
	return buildTestClient(server), contextWithNullLogger()
}

func contextWithNullLogger() context.Context {
	return log.NewContextWithLogger(context.Background(), null.NewLogger())
}

func testServer(t GinkgoTInterface, handler http.HandlerFunc) *httptest.Server {
	s := httptest.NewServer(http.HandlerFunc(handler))
	t.Cleanup(s.Close)
	return s
}
