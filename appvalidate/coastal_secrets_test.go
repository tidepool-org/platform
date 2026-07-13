package appvalidate

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
)

var _ = Describe("CoastalSecrets", func() {
	Describe("coastalHTTPTimeout", func() {
		It("is 60 seconds", func() {
			Expect(coastalHTTPTimeout).To(Equal(60 * time.Second))
		})
	})

	Describe("GetSecret timeout", func() {
		var (
			server      *httptest.Server
			cs          *CoastalSecrets
			partnerData []byte
			ctx         context.Context
		)

		BeforeEach(func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(50 * time.Millisecond)
				w.WriteHeader(http.StatusOK)
			}))
			GinkgoT().Cleanup(server.Close)

			// Generate an ED25519 private key in PKCS8 PEM format.
			_, priv, err := ed25519.GenerateKey(rand.Reader)
			Expect(err).ToNot(HaveOccurred())
			pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(priv)
			Expect(err).ToNot(HaveOccurred())
			keyData := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8Bytes})

			cfg := CoastalSecretsConfig{
				APIKey:         "test-api-key",
				BaseURL:        "http://example.com",
				ClientID:       "test-client-id",
				ClientSecret:   "test-client-secret",
				RCTypeID:       "test-rc-type",
				KeyData:        keyData,
				certificateURL: server.URL,
				HTTPTimeout:    5 * time.Millisecond,
			}

			cs, err = NewCoastalSecrets(logTest.NewLogger(), cfg)
			Expect(err).ToNot(HaveOccurred())

			partnerData, err = json.Marshal(&CoastalPayload{
				RCTypeID:        "test-rc-type",
				RCInstanceID:    "test-rc-instance",
				HardwareVersion: "v1",
				SoftwareVersion: "v1",
				PHDTypeID:       "test-phd-type",
				PHDInstanceID:   "test-phd-instance",
				CSR:             "test-csr",
			})
			Expect(err).ToNot(HaveOccurred())

			ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
		})

		It("times out when the server is slow", func() {
			_, err := cs.GetSecret(ctx, partnerData)
			Expect(err).To(MatchError(ContainSubstring("deadline exceeded")))
		})
	})
})
