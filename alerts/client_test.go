package alerts

import (
	"context"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/platform"
)

const testToken = "auth-me"

var _ = Describe("Client", func() {
	var test404Server, test200Server *httptest.Server
	var testAuthServer func(*string) *httptest.Server

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
		testAuthServer = func(token *string) *httptest.Server {
			return testServer(t, func(w http.ResponseWriter, r *http.Request) {
				*token = r.Header.Get(auth.TidepoolSessionTokenHeaderKey)
				w.WriteHeader(http.StatusOK)
			})
		}
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

		It("injects an auth token", func() {
			token := ""
			client, ctx := newAlertsClientTest(testAuthServer(&token))
			_ = client.Delete(ctx, &Config{})
			Expect(token).To(Equal(testToken))
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

		It("injects an auth token", func() {
			token := ""
			client, ctx := newAlertsClientTest(testAuthServer(&token))
			_ = client.Upsert(ctx, &Config{})
			Expect(token).To(Equal(testToken))
		})
	})
})

func buildTestClient(s *httptest.Server) *Client {
	pCfg := &platform.Config{
		Config: &client.Config{
			Address: s.URL,
		},
	}
	token := mockTokenProvider(testToken)
	pc, err := platform.NewClient(pCfg, platform.AuthorizeAsService)
	Expect(err).ToNot(HaveOccurred())
	client := NewClient(pc, token, null.NewLogger())
	return client
}

func newAlertsClientTest(server *httptest.Server) (*Client, context.Context) {
	return buildTestClient(server), contextWithNullLogger()
}

func contextWithNullLogger() context.Context {
	return log.NewContextWithLogger(context.Background(), null.NewLogger())
}

type mockTokenProvider string

func (p mockTokenProvider) ServerSessionToken() (string, error) {
	return string(p), nil
}

func testServer(t GinkgoTInterface, handler http.HandlerFunc) *httptest.Server {
	s := httptest.NewServer(http.HandlerFunc(handler))
	t.Cleanup(s.Close)
	return s
}
