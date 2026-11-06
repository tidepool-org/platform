package flux

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/kelseyhightower/envconfig"

	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/cluster"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/request"
)

const (
	KustomizationNamespace = "fluxv2"
	KustomizationName      = "fluxv2"

	CacheDuration      = time.Minute
	ErrorCacheDuration = 5 * time.Second
	RequestTimeout     = 10 * time.Second

	ServiceAccountDirectory = "/var/run/secrets/kubernetes.io/serviceaccount"

	reconciliationSucceeded = "ReconciliationSucceeded"
)

var commitHashExpression = regexp.MustCompile(`^[0-9a-f]{40}([0-9a-f]{24})?$`)

// Config is the Kubernetes API server address that Kubernetes sets in every container
type Config struct {
	Host string `envconfig:"KUBERNETES_SERVICE_HOST"`
	Port string `envconfig:"KUBERNETES_SERVICE_PORT"`
}

// ReleaseReporter reads the release from the Flux Kustomization that applies the cluster
// configuration repository, at most once per cache duration
type ReleaseReporter struct {
	client             *client.Client
	url                string
	certificatePath    string
	tokenPath          string
	cacheDuration      time.Duration
	errorCacheDuration time.Duration
	unavailableErr     error

	mutex                sync.Mutex
	expiresAt            time.Time
	release              *cluster.Release
	err                  error
	certificateAuthority []byte
	httpClient           *http.Client
}

// NewReleaseReporter reads the Kustomization from the API server at the client address; without a certificate path
// it uses the default HTTP client
func NewReleaseReporter(clientConfig *client.Config, certificatePath string, tokenPath string, cacheDuration time.Duration, errorCacheDuration time.Duration) (*ReleaseReporter, error) {
	clnt, err := client.NewWithErrorParser(clientConfig, &statusErrorParser{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to create kubernetes client")
	}

	return &ReleaseReporter{
		client:             clnt,
		url:                clnt.ConstructURL("apis", "kustomize.toolkit.fluxcd.io", "v1", "namespaces", KustomizationNamespace, "kustomizations", KustomizationName),
		certificatePath:    certificatePath,
		tokenPath:          tokenPath,
		cacheDuration:      cacheDuration,
		errorCacheDuration: errorCacheDuration,
		httpClient:         http.DefaultClient,
	}, nil
}

// NewInClusterReleaseReporter never fails so the service starts outside Kubernetes; Release reports why it is unavailable
func NewInClusterReleaseReporter(userAgent string) *ReleaseReporter {
	cfg := &Config{}
	if err := envconfig.Process("", cfg); err != nil {
		return &ReleaseReporter{unavailableErr: errors.Wrap(err, "unable to load kubernetes config")}
	}
	if cfg.Host == "" || cfg.Port == "" {
		return &ReleaseReporter{unavailableErr: errors.New("not running in kubernetes")}
	}

	clientConfig := &client.Config{Address: "https://" + net.JoinHostPort(cfg.Host, cfg.Port), UserAgent: userAgent}
	releaseReporter, err := NewReleaseReporter(clientConfig, filepath.Join(ServiceAccountDirectory, "ca.crt"), filepath.Join(ServiceAccountDirectory, "token"), CacheDuration, ErrorCacheDuration)
	if err != nil {
		return &ReleaseReporter{unavailableErr: err}
	}
	return releaseReporter
}

func (r *ReleaseReporter) Release(ctx context.Context) (*cluster.Release, error) {
	if r.unavailableErr != nil {
		return nil, r.unavailableErr
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !time.Now().Before(r.expiresAt) {
		r.release, r.err = r.fetch(context.WithoutCancel(ctx))
		if r.err == nil {
			r.expiresAt = time.Now().Add(r.cacheDuration)
		} else {
			r.expiresAt = time.Now().Add(r.errorCacheDuration)
		}
	}

	if r.err != nil {
		return nil, r.err
	}
	release := *r.release
	return &release, nil
}

func (r *ReleaseReporter) fetch(ctx context.Context) (*cluster.Release, error) {
	httpClient, err := r.currentHTTPClient()
	if err != nil {
		return nil, err
	}

	token, err := os.ReadFile(r.tokenPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read service account token")
	}

	mutators := []request.RequestMutator{
		request.NewHeaderMutator("Authorization", "Bearer "+strings.TrimSpace(string(token))),
		request.NewHeaderMutator("Accept", "application/json"),
	}
	var kstmztn kustomization
	if err = r.client.RequestDataWithHTTPClient(ctx, http.MethodGet, r.url, mutators, nil, &kstmztn, nil, httpClient); err != nil {
		return nil, errors.Wrap(err, "unable to get kustomization")
	}

	return kstmztn.release()
}

// The cluster certificate authority bundle holds both authorities while EKS rotates them, so it is read on every fetch
func (r *ReleaseReporter) currentHTTPClient() (*http.Client, error) {
	if r.certificatePath == "" {
		return r.httpClient, nil
	}

	certificateAuthority, err := os.ReadFile(r.certificatePath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read service account certificate authority")
	}
	if r.certificateAuthority != nil && bytes.Equal(certificateAuthority, r.certificateAuthority) {
		return r.httpClient, nil
	}

	rootCAs := x509.NewCertPool()
	if !rootCAs.AppendCertsFromPEM(certificateAuthority) {
		return nil, errors.New("service account certificate authority is invalid")
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{RootCAs: rootCAs}

	r.certificateAuthority = certificateAuthority
	r.httpClient = &http.Client{Transport: transport, Timeout: RequestTimeout}
	return r.httpClient, nil
}

// Kubernetes explains a rejected request, such as missing RBAC, in the message of a Status body
type statusErrorParser struct{}

func (s *statusErrorParser) ParseErrorResponse(ctx context.Context, res *http.Response, req *http.Request) error {
	var status struct {
		Message string `json:"message"`
	}
	if json.NewDecoder(res.Body).Decode(&status) != nil || status.Message == "" {
		return nil
	}
	return errors.Newf("unexpected response status code %d: %s", res.StatusCode, status.Message)
}

type kustomization struct {
	Status struct {
		LastAppliedRevision string `json:"lastAppliedRevision"`
		Conditions          []struct {
			Type               string    `json:"type"`
			LastTransitionTime time.Time `json:"lastTransitionTime"`
		} `json:"conditions"`
		History []snapshot `json:"history"`
	} `json:"status"`
}

type snapshot struct {
	Digest               string            `json:"digest"`
	FirstReconciled      time.Time         `json:"firstReconciled"`
	LastReconciled       time.Time         `json:"lastReconciled"`
	LastReconciledStatus string            `json:"lastReconciledStatus"`
	Metadata             map[string]string `json:"metadata"`
}

func (k *kustomization) release() (*cluster.Release, error) {
	revision := k.Status.LastAppliedRevision
	if revision == "" {
		return nil, errors.New("kustomization has no applied revision")
	}
	commitHash, err := parseRevision(revision)
	if err != nil {
		return nil, err
	}
	releaseTime := k.releaseTime(revision)
	if releaseTime.IsZero() {
		return nil, errors.Newf("kustomization has no time for revision %q", revision)
	}
	return &cluster.Release{Revision: commitHash, Time: releaseTime}, nil
}

// History snapshots are keyed by the digest of the applied manifests, so the release time is when the applied
// manifests were first reconciled or, after a revert to manifests applied before, when the replaced manifests were
// last reconciled. Without a snapshot (Flux before 2.7, or pushed out of the five kept by failed reconciliations)
// the Ready transition is the best Flux has: the start or end of the last reconciliation.
func (k *kustomization) releaseTime(revision string) time.Time {
	var current *snapshot
	for index := range k.Status.History {
		if s := &k.Status.History[index]; s.LastReconciledStatus == reconciliationSucceeded && s.Metadata["revision"] == revision {
			current = s
			break
		}
	}
	if current != nil && !current.FirstReconciled.IsZero() {
		releaseTime := current.FirstReconciled
		for _, s := range k.Status.History {
			if s.LastReconciledStatus == reconciliationSucceeded && s.Digest != current.Digest && s.LastReconciled.After(releaseTime) {
				releaseTime = s.LastReconciled
			}
		}
		return releaseTime
	}
	for _, condition := range k.Status.Conditions {
		if condition.Type == "Ready" {
			return condition.LastTransitionTime
		}
	}
	return time.Time{}
}

// Flux revisions are formatted <ref>@<algorithm>:<hash>, for example master@sha1:da9edeab...
func parseRevision(revision string) (string, error) {
	commitHash := revision[strings.LastIndex(revision, ":")+1:]
	if !commitHashExpression.MatchString(commitHash) {
		return "", errors.Newf("kustomization revision %q is invalid", revision)
	}
	return commitHash, nil
}
