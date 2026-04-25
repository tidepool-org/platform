package appvalidate

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
)

// palmTreeTestTLSPair generates a self-signed ECDSA certificate + key in PEM
// format. The key pair satisfies tls.X509KeyPair but is not trusted by any CA.
func palmTreeTestTLSPair() (certPEM, keyPEM []byte) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		panic(err)
	}
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	privDER, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		panic(err)
	}
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privDER})
	return
}

var _ = Describe("PalmTreeSecrets", func() {
	Describe("palmTreeHTTPTimeout", func() {
		It("is 60 seconds", func() {
			Expect(palmTreeHTTPTimeout).To(Equal(60 * time.Second))
		})
	})

	Describe("GetSecret timeout", func() {
		var (
			server      *httptest.Server
			pt          *PalmTreeSecrets
			partnerData []byte
			ctx         context.Context
		)

		BeforeEach(func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(50 * time.Millisecond)
				w.WriteHeader(http.StatusOK)
			}))
			GinkgoT().Cleanup(server.Close)

			certPEM, keyPEM := palmTreeTestTLSPair()

			cfg := PalmTreeSecretsConfig{
				BaseURL:        "http://example.com",
				CalID:          "test-cal-id",
				ProfileID:      "test-profile-id",
				CertData:       certPEM,
				KeyData:        keyPEM,
				certificateURL: server.URL,
				HTTPTimeout:    5 * time.Millisecond,
			}

			var err error
			pt, err = NewPalmTreeSecrets(logTest.NewLogger(), cfg)
			Expect(err).ToNot(HaveOccurred())

			partnerData, err = json.Marshal(map[string]interface{}{
				"csr": "test-csr",
			})
			Expect(err).ToNot(HaveOccurred())

			ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
		})

		It("times out when the server is slow", func() {
			_, err := pt.GetSecret(ctx, partnerData)
			Expect(err).To(MatchError(ContainSubstring("deadline exceeded")))
		})
	})
})
