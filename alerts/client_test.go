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

const testToken = "auth-me"
const testUserID = "test-user-id"
const testFollowedUserID = "test-followed-user-id"
const testDataSetID = "upid_000000000000"

var _ = Describe("Client", func() {
	var test404Server *httptest.Server
	var test200Server func(string) *httptest.Server

	BeforeEach(func() {
		t := GinkgoT()
		// There's no need to create these before each test, but I can't get
		// Ginkgo to let me start these just once.
		test404Server = testServer(t, func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		})
		test200Server = func(resp string) *httptest.Server {
			return testServer(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(resp))
			})
		}
	})

	ItReturnsAnErrorOnNon200Responses := func(f func(context.Context, *Client) error) {
		GinkgoHelper()
		It("returns an error on non-200 respnoses", func() {
			client, ctx := newAlertsClientTest(test404Server)
			err := f(ctx, client)
			Expect(err).Should(HaveOccurred())
			Expect(err).To(MatchError(ContainSubstring("resource not found")))
		})
	}

	ItReturnsANilErrorOnSuccess := func(resp string, f func(context.Context, *Client) error) {
		GinkgoHelper()
		It("returns a nil error on success", func() {
			client, ctx := newAlertsClientTest(test200Server(resp))
			err := f(ctx, client)
			Expect(err).To(Succeed())
		})
	}

	Context("Delete", func() {
		ItReturnsAnErrorOnNon200Responses(func(ctx context.Context, client *Client) error {
			return client.Delete(ctx, &Config{})
		})

		ItReturnsANilErrorOnSuccess("", func(ctx context.Context, client *Client) error {
			return client.Delete(ctx, &Config{})
		})
	})

	Context("Upsert", func() {
		ItReturnsAnErrorOnNon200Responses(func(ctx context.Context, client *Client) error {
			return client.Upsert(ctx, &Config{})
		})

		ItReturnsANilErrorOnSuccess("", func(ctx context.Context, client *Client) error {
			return client.Upsert(ctx, &Config{})
		})
	})

	Context("Get", func() {
		ItReturnsAnErrorOnNon200Responses(func(ctx context.Context, client *Client) error {
			_, err := client.Get(ctx, testFollowedUserID, testUserID)
			return err
		})

		ret := `{
                  "userId": "14ee703f-ca9b-4a6b-9ce3-41d886514e7f",
                  "followedUserId": "ce5863bc-cc0b-4177-97d7-e8de0c558820",
                  "uploadId": "upid_00000000000000000000000000000000"
                }`
		ItReturnsANilErrorOnSuccess(ret, func(ctx context.Context, client *Client) error {
			_, err := client.Get(ctx, testFollowedUserID, testUserID)
			return err
		})
	})

	Context("List", func() {
		ItReturnsAnErrorOnNon200Responses(func(ctx context.Context, client *Client) error {
			_, err := client.List(ctx, "")
			return err
		})

		ItReturnsANilErrorOnSuccess("[]", func(ctx context.Context, client *Client) error {
			_, err := client.List(ctx, "")
			return err
		})
	})

	Context("UsersWithoutCommunication", func() {
		ItReturnsAnErrorOnNon200Responses(func(ctx context.Context, client *Client) error {
			_, err := client.UsersWithoutCommunication(ctx)
			return err
		})

		ItReturnsANilErrorOnSuccess("[]", func(ctx context.Context, client *Client) error {
			_, err := client.UsersWithoutCommunication(ctx)
			return err
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

func contextWithNullLoggerDeluxe() (context.Context, log.Logger) {
	lgr := null.NewLogger()
	return log.NewContextWithLogger(context.Background(), lgr), lgr
}

func contextWithNullLogger() context.Context {
	ctx, _ := contextWithNullLoggerDeluxe()
	return ctx
}

func testServer(t GinkgoTInterface, handler http.HandlerFunc) *httptest.Server {
	s := httptest.NewServer(http.HandlerFunc(handler))
	t.Cleanup(s.Close)
	return s
}
