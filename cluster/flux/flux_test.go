package flux_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/cluster"
	"github.com/tidepool-org/platform/cluster/flux"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
)

const (
	kustomizationPath = "/apis/kustomize.toolkit.fluxcd.io/v1/namespaces/fluxv2/kustomizations/fluxv2"

	commitHash       = "868bd7b1c4f0e3a5d2b9c8e7f6a5b4c3d2e1f0a9"
	otherCommitHash  = "2211343ebf21457b9ad063a0e4d9d4ed447a5d11"
	revertCommitHash = "22d0387deab53dc465dd16b95dd540642c247487"

	digestA = "sha256:3202bfa4883d84f357ab66a4c79b639926aa4055306c6c0ba29a62ca97e8a471"
	digestB = "sha256:9435c6211b21d5e801d12b37a62c8ef8e99bde6f095ae42baa7523115b48a0c1"

	readyTransition = "2026-11-06T10:09:03Z"
)

type snapshot struct {
	revision        string
	status          string
	digest          string
	firstReconciled any
	lastReconciled  string
}

func kustomization(lastAppliedRevision string, conditions []map[string]any, snapshots ...snapshot) map[string]any {
	history := []map[string]any{}
	for _, s := range snapshots {
		history = append(history, map[string]any{
			"digest":                 s.digest,
			"firstReconciled":        s.firstReconciled,
			"lastReconciled":         s.lastReconciled,
			"lastReconciledDuration": "12.513656773s",
			"lastReconciledStatus":   s.status,
			"metadata":               map[string]any{"revision": s.revision},
			"totalReconciliations":   18,
		})
	}
	return map[string]any{
		"apiVersion": "kustomize.toolkit.fluxcd.io/v1",
		"kind":       "Kustomization",
		"metadata":   map[string]any{"name": "fluxv2", "namespace": "fluxv2"},
		"status": map[string]any{
			"conditions":            conditions,
			"history":               history,
			"inventory":             map[string]any{"entries": []map[string]any{{"id": "dev1_auth_apps_Deployment", "v": "v1"}}},
			"lastAppliedRevision":   lastAppliedRevision,
			"lastAttemptedRevision": lastAppliedRevision,
		},
	}
}

func certificatePEM(certificate *x509.Certificate) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate.Raw})
}

func otherCertificatePEM() []byte {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	Expect(err).ToNot(HaveOccurred())
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "kubernetes"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	raw, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	Expect(err).ToNot(HaveOccurred())
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: raw})
}

func ready(status string) []map[string]any {
	return []map[string]any{{"type": "Ready", "status": status, "reason": "ReconciliationSucceeded", "lastTransitionTime": readyTransition}}
}

var _ = Describe("Flux", func() {
	var ctx context.Context
	var server *Server
	var tokenPath string

	BeforeEach(func() {
		ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
		server = NewServer()
		tokenPath = filepath.Join(GinkgoT().TempDir(), "token")
		Expect(os.WriteFile(tokenPath, []byte("test-token\n"), 0600)).To(Succeed())
	})

	AfterEach(func() {
		server.Close()
	})

	newReleaseReporterWithErrorCache := func(cacheDuration time.Duration, errorCacheDuration time.Duration) *flux.ReleaseReporter {
		releaseReporter, err := flux.NewReleaseReporter(&client.Config{Address: server.URL()}, "", tokenPath, cacheDuration, errorCacheDuration)
		Expect(err).ToNot(HaveOccurred())
		return releaseReporter
	}

	newReleaseReporter := func(cacheDuration time.Duration) *flux.ReleaseReporter {
		return newReleaseReporterWithErrorCache(cacheDuration, cacheDuration)
	}

	respondWith := func(statusCode int, body any) http.HandlerFunc {
		return CombineHandlers(
			VerifyRequest(http.MethodGet, kustomizationPath),
			VerifyHeaderKV("Authorization", "Bearer test-token"),
			VerifyHeaderKV("Accept", "application/json"),
			RespondWithJSONEncoded(statusCode, body),
		)
	}

	releaseOf := func(body map[string]any) (*cluster.Release, error) {
		server.AppendHandlers(respondWith(http.StatusOK, body))
		return newReleaseReporter(0).Release(ctx)
	}

	at := func(value string) time.Time {
		parsed, err := time.Parse(time.RFC3339, value)
		Expect(err).ToNot(HaveOccurred())
		return parsed
	}

	applied := func() map[string]any {
		return kustomization("master@sha1:"+commitHash, ready("True"),
			snapshot{"master@sha1:" + commitHash, "ReconciliationSucceeded", digestA, "2026-11-03T18:42:07Z", "2026-11-06T10:09:03Z"},
		)
	}
	appliedRelease := &cluster.Release{Revision: commitHash, Time: time.Date(2026, 11, 3, 18, 42, 7, 0, time.UTC)}

	Context("Release from the history", func() {
		It("returns the commit hash and the first reconciliation of the applied manifests", func() {
			Expect(releaseOf(applied())).To(Equal(appliedRelease))
		})

		It("returns the first reconciliation of the new manifests after a deploy", func() {
			Expect(releaseOf(kustomization("master@sha1:"+commitHash, ready("True"),
				snapshot{"master@sha1:" + commitHash, "ReconciliationSucceeded", digestA, "2026-11-03T18:42:07Z", "2026-11-06T10:09:03Z"},
				snapshot{"master@sha1:" + otherCommitHash, "ReconciliationSucceeded", digestB, "2026-11-01T09:00:00Z", "2026-11-03T18:40:00Z"},
			))).To(Equal(appliedRelease))
		})

		It("returns when the replaced manifests were last reconciled after a revert to manifests applied before", func() {
			Expect(releaseOf(kustomization("master@sha1:"+revertCommitHash, ready("True"),
				snapshot{"master@sha1:" + revertCommitHash, "ReconciliationSucceeded", digestA, "2026-11-01T09:00:00Z", "2026-11-06T10:09:03Z"},
				snapshot{"master@sha1:" + otherCommitHash, "ReconciliationSucceeded", digestB, "2026-11-03T18:42:07Z", "2026-11-05T09:59:30Z"},
			))).To(Equal(&cluster.Release{Revision: revertCommitHash, Time: at("2026-11-05T09:59:30Z")}))
		})

		It("keeps the date of the applied manifests when a later commit leaves them unchanged", func() {
			Expect(releaseOf(kustomization("master@sha1:"+otherCommitHash, ready("True"),
				snapshot{"master@sha1:" + otherCommitHash, "ReconciliationSucceeded", digestA, "2026-11-03T18:42:07Z", "2026-11-06T10:09:03Z"},
			))).To(Equal(&cluster.Release{Revision: otherCommitHash, Time: appliedRelease.Time}))
		})

		It("ignores failed reconciliations of other manifests", func() {
			Expect(releaseOf(kustomization("master@sha1:"+commitHash, ready("False"),
				snapshot{"master@sha1:" + otherCommitHash, "ReconciliationFailed", digestB, "2026-11-04T09:00:00Z", "2026-11-06T10:14:00Z"},
				snapshot{"master@sha1:" + commitHash, "ReconciliationSucceeded", digestA, "2026-11-03T18:42:07Z", "2026-11-04T08:55:00Z"},
			))).To(Equal(appliedRelease))
		})

		It("ignores a failed reconciliation of the same manifests", func() {
			Expect(releaseOf(kustomization("master@sha1:"+commitHash, ready("True"),
				snapshot{"master@sha1:" + commitHash, "ReconciliationSucceeded", digestA, "2026-11-03T18:42:07Z", "2026-11-06T10:09:03Z"},
				snapshot{"master@sha1:" + commitHash, "ReconciliationFailed", digestA, "2026-11-04T09:00:00Z", "2026-11-04T09:05:00Z"},
			))).To(Equal(appliedRelease))
		})
	})

	Context("Release without a usable history", func() {
		readyRelease := &cluster.Release{Revision: commitHash, Time: time.Date(2026, 11, 6, 10, 9, 3, 0, time.UTC)}

		It("returns the Ready transition without history, as before Flux 2.7", func() {
			Expect(releaseOf(kustomization("master@sha1:"+commitHash, ready("True")))).To(Equal(readyRelease))
		})

		It("returns the Ready transition when failed reconciliations pushed the applied manifests out of the history", func() {
			Expect(releaseOf(kustomization("master@sha1:"+commitHash, ready("False"),
				snapshot{"master@sha1:" + otherCommitHash, "ReconciliationFailed", digestB, "2026-11-04T09:00:00Z", "2026-11-06T10:14:00Z"},
			))).To(Equal(readyRelease))
		})

		It("returns the Ready transition while reconciling", func() {
			Expect(releaseOf(kustomization("master@sha1:"+commitHash, ready("Unknown")))).To(Equal(readyRelease))
		})

		It("returns the Ready transition when the first reconciliation time is missing", func() {
			Expect(releaseOf(kustomization("master@sha1:"+commitHash, ready("True"),
				snapshot{"master@sha1:" + commitHash, "ReconciliationSucceeded", digestA, nil, "2026-11-06T10:09:03Z"},
			))).To(Equal(readyRelease))
		})

		It("returns an error if Flux has no time at all", func() {
			_, err := releaseOf(kustomization("master@sha1:"+commitHash, nil))
			Expect(err).To(MatchError(`kustomization has no time for revision "master@sha1:` + commitHash + `"`))
		})
	})

	Context("Release", func() {
		It("returns the cached release until the cache duration expires", func() {
			server.AppendHandlers(respondWith(http.StatusOK, applied()))
			releaseReporter := newReleaseReporter(time.Hour)
			Expect(releaseReporter.Release(ctx)).To(Equal(appliedRelease))
			Expect(releaseReporter.Release(ctx)).To(Equal(appliedRelease))
			Expect(server.ReceivedRequests()).To(HaveLen(1))
		})

		It("returns the cached error until the cache duration expires", func() {
			server.AppendHandlers(respondWith(http.StatusForbidden, map[string]any{"kind": "Status"}))
			releaseReporter := newReleaseReporter(time.Hour)
			_, err := releaseReporter.Release(ctx)
			Expect(err).To(MatchError("unable to get kustomization; authentication token is not authorized for requested action"))
			_, err = releaseReporter.Release(ctx)
			Expect(err).To(MatchError("unable to get kustomization; authentication token is not authorized for requested action"))
			Expect(server.ReceivedRequests()).To(HaveLen(1))
		})

		It("retries after an error once the shorter error cache expires and then caches the release", func() {
			server.AppendHandlers(
				respondWith(http.StatusInternalServerError, map[string]any{"kind": "Status"}),
				respondWith(http.StatusOK, applied()),
			)
			releaseReporter := newReleaseReporterWithErrorCache(time.Hour, 0)
			_, err := releaseReporter.Release(ctx)
			Expect(err).To(MatchError(ContainSubstring("unexpected response status code 500")))
			Expect(releaseReporter.Release(ctx)).To(Equal(appliedRelease))
			Expect(releaseReporter.Release(ctx)).To(Equal(appliedRelease))
			Expect(server.ReceivedRequests()).To(HaveLen(2))
		})

		It("returns a copy of the cached release", func() {
			server.AppendHandlers(respondWith(http.StatusOK, applied()))
			releaseReporter := newReleaseReporter(time.Hour)
			release, err := releaseReporter.Release(ctx)
			Expect(err).ToNot(HaveOccurred())
			release.Revision = otherCommitHash
			Expect(releaseReporter.Release(ctx)).To(Equal(appliedRelease))
		})

		It("returns the Kubernetes message of a rejected request", func() {
			server.AppendHandlers(respondWith(http.StatusForbidden, map[string]any{
				"kind":    "Status",
				"status":  "Failure",
				"reason":  "Forbidden",
				"code":    403,
				"message": `kustomizations.kustomize.toolkit.fluxcd.io "fluxv2" is forbidden: User "system:serviceaccount:dev1:auth" cannot get resource "kustomizations" in API group "kustomize.toolkit.fluxcd.io" in the namespace "fluxv2"`,
			}))
			_, err := newReleaseReporter(0).Release(ctx)
			Expect(err).To(MatchError(`unable to get kustomization; unexpected response status code 403: kustomizations.kustomize.toolkit.fluxcd.io "fluxv2" is forbidden: User "system:serviceaccount:dev1:auth" cannot get resource "kustomizations" in API group "kustomize.toolkit.fluxcd.io" in the namespace "fluxv2"`))
		})

		It("returns a newly applied revision once the cache expires", func() {
			server.AppendHandlers(
				respondWith(http.StatusOK, applied()),
				respondWith(http.StatusOK, kustomization("master@sha1:"+otherCommitHash, ready("True"),
					snapshot{"master@sha1:" + otherCommitHash, "ReconciliationSucceeded", digestB, "2026-11-06T08:55:09Z", "2026-11-06T10:09:03Z"},
					snapshot{"master@sha1:" + commitHash, "ReconciliationSucceeded", digestA, "2026-11-03T18:42:07Z", "2026-11-06T08:50:00Z"},
				)),
			)
			releaseReporter := newReleaseReporter(0)
			Expect(releaseReporter.Release(ctx)).To(Equal(appliedRelease))
			Expect(releaseReporter.Release(ctx)).To(Equal(&cluster.Release{Revision: otherCommitHash, Time: at("2026-11-06T08:55:09Z")}))
		})

		It("returns an error if the token cannot be read", func() {
			Expect(os.Remove(tokenPath)).To(Succeed())
			_, err := newReleaseReporter(0).Release(ctx)
			Expect(err).To(MatchError(ContainSubstring("unable to read service account token")))
			Expect(server.ReceivedRequests()).To(BeEmpty())
		})

		It("returns an error if the response is not JSON", func() {
			server.AppendHandlers(RespondWith(http.StatusOK, "not json"))
			_, err := newReleaseReporter(0).Release(ctx)
			Expect(err).To(MatchError(ContainSubstring("unable to get kustomization")))
		})

		It("returns an error if no revision was applied", func() {
			_, err := releaseOf(kustomization("", ready("Unknown")))
			Expect(err).To(MatchError("kustomization has no applied revision"))
		})

		DescribeTable("returns an error if the revision is invalid",
			func(revision string) {
				_, err := releaseOf(kustomization(revision, ready("True"), snapshot{revision, "ReconciliationSucceeded", digestA, "2026-11-03T18:42:07Z", "2026-11-06T10:09:03Z"}))
				Expect(err).To(MatchError(`kustomization revision "` + revision + `" is invalid`))
			},
			Entry("no hash", "master"),
			Entry("short hash", "master@sha1:868bd7b"),
			Entry("upper case hash", "master@sha1:868BD7B1C4F0E3A5D2B9C8E7F6A5B4C3D2E1F0A9"),
			Entry("not a hash", "master@sha1:not-a-commit-hash-but-forty-characters-x"),
		)
	})

	Context("certificate authority", func() {
		var tlsServer *Server
		var certificatePath string

		BeforeEach(func() {
			tlsServer = NewTLSServer()
			certificatePath = filepath.Join(GinkgoT().TempDir(), "ca.crt")
		})

		AfterEach(func() {
			tlsServer.Close()
		})

		newTLSReleaseReporter := func() *flux.ReleaseReporter {
			releaseReporter, err := flux.NewReleaseReporter(&client.Config{Address: tlsServer.URL()}, certificatePath, tokenPath, time.Hour, 0)
			Expect(err).ToNot(HaveOccurred())
			return releaseReporter
		}

		It("verifies the API server with the certificate authority", func() {
			Expect(os.WriteFile(certificatePath, certificatePEM(tlsServer.HTTPTestServer.Certificate()), 0600)).To(Succeed())
			tlsServer.AppendHandlers(respondWith(http.StatusOK, applied()))
			Expect(newTLSReleaseReporter().Release(ctx)).To(Equal(appliedRelease))
		})

		It("re-reads the certificate authority, so a rotated authority is used without a restart", func() {
			Expect(os.WriteFile(certificatePath, otherCertificatePEM(), 0600)).To(Succeed())
			releaseReporter := newTLSReleaseReporter()
			_, err := releaseReporter.Release(ctx)
			Expect(err).To(MatchError(ContainSubstring("certificate")))

			bundle := append(otherCertificatePEM(), certificatePEM(tlsServer.HTTPTestServer.Certificate())...)
			Expect(os.WriteFile(certificatePath, bundle, 0600)).To(Succeed())
			tlsServer.AppendHandlers(respondWith(http.StatusOK, applied()))
			Expect(releaseReporter.Release(ctx)).To(Equal(appliedRelease))
		})

		It("returns an error if the certificate authority cannot be read", func() {
			_, err := newTLSReleaseReporter().Release(ctx)
			Expect(err).To(MatchError(ContainSubstring("unable to read service account certificate authority")))
		})

		It("returns an error if the certificate authority is invalid", func() {
			Expect(os.WriteFile(certificatePath, []byte("not a certificate"), 0600)).To(Succeed())
			_, err := newTLSReleaseReporter().Release(ctx)
			Expect(err).To(MatchError("service account certificate authority is invalid"))
		})
	})

	Context("NewInClusterReleaseReporter", func() {
		It("returns an error from Release when not running in kubernetes", func() {
			GinkgoT().Setenv("KUBERNETES_SERVICE_HOST", "")
			_, err := flux.NewInClusterReleaseReporter("test").Release(ctx)
			Expect(err).To(MatchError("not running in kubernetes"))
		})

		It("reads the certificate authority when it fetches, not when it is created", func() {
			GinkgoT().Setenv("KUBERNETES_SERVICE_HOST", "127.0.0.1")
			GinkgoT().Setenv("KUBERNETES_SERVICE_PORT", "6443")
			if _, err := os.Stat(filepath.Join(flux.ServiceAccountDirectory, "ca.crt")); err == nil {
				Skip("running in kubernetes")
			}
			_, err := flux.NewInClusterReleaseReporter("test").Release(ctx)
			Expect(err).To(MatchError(ContainSubstring("unable to read service account certificate authority")))
		})
	})
})
