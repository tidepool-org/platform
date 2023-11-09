package appvalidate

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/kelseyhightower/envconfig"

	"github.com/tidepool-org/platform/structure"
	structValidator "github.com/tidepool-org/platform/structure/validator"
)

var (
	ErrPalmTreeInvalidTLS  = errors.New("invalid PalmTree TLS credentials")
	ErrPalmTreeEmptyConfig = errors.New("empty PalmTree config")
)

const (
	PartnerPalmTree = "PalmTree"
)

type PalmTreeSecretsConfig struct {
	BaseURL   string `envconfig:"PALMTREE_BASE_URL"`
	CalID     string `envconfig:"PALMTREE_CAL_ID"`
	ProfileID string `envconfig:"PALMTREE_PROFILE_ID"`
	// CertData is the raw contents of the tls certificate file
	CertData []byte `envconfig:"PALMTREE_TLS_CERT_DATA"`
	// KeyData is the raw contents of the tls private key file
	KeyData []byte `envconfig:"PALMTREE_TLS_KEY_DATA"`
}

type PalmTreeSecrets struct {
	Config PalmTreeSecretsConfig
	client *http.Client
}

func NewPalmTreeSecretsConfig() (*PalmTreeSecretsConfig, error) {
	cfg := &PalmTreeSecretsConfig{}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func NewPalmTreeSecrets(c *PalmTreeSecretsConfig) (*PalmTreeSecrets, error) {
	if c == nil {
		return nil, ErrPalmTreeEmptyConfig
	}
	cert, err := tls.X509KeyPair(c.CertData, c.KeyData)
	if err != nil {
		return &PalmTreeSecrets{
			Config: *c,
			client: http.DefaultClient,
		}, errors.Join(ErrPalmTreeInvalidTLS, err)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}

	return &PalmTreeSecrets{
		Config: *c,
		client: &http.Client{Transport: tr},
	}, nil
}

type PalmTreePayload struct {
	CSR                       string                            `json:"csr"`
	ProfileID                 string                            `json:"profileId"`
	RequiredFormat            palmTreeRequiredFormat            `json:"requiredFormat"`
	CertificateRequestDetails palmTreeCertificateRequestDetails `json:"optionalCertificateRequestDetails"`
}

type palmTreeCertificateRequestDetails struct {
	SubjectDN string `json:"subjectDn"`
}

type palmTreeRequiredFormat struct {
	Format string `json:"format"`
}

type PalmTreeResponse struct {
	Type string `json:"type"`

	Message struct {
		Message string `json:"message"`
		Details []any  `json:"details"`
	} `json:"message"`

	Enrollment struct {
		ID             string `json:"id"`
		SerialNumber   string `json:"serialNumber"`
		SubjectName    string `json:"subjectName"`
		IssuerName     string `json:"issuerName"`
		ValidityPeroid string `json:"validityPeriod"`
		Status         string `json:"status"`
		Body           string `json:"body"`
	} `json:"enrollment"`
}

func (pt *PalmTreeSecrets) GetSecret(ctx context.Context, partnerDataRaw []byte) (*PalmTreeResponse, error) {
	if len(pt.Config.CertData) == 0 {
		return nil, ErrPalmTreeInvalidTLS
	}
	payload := newPalmtreePayload(pt.Config.ProfileID)

	if err := json.Unmarshal(partnerDataRaw, payload); err != nil {
		return nil, fmt.Errorf("unable to unmarshal PalmTree payload: %w", err)
	}

	if err := structValidator.New().Validate(payload); err != nil {
		return nil, fmt.Errorf("PalmTree: %w: %w", ErrInvalidPartnerPayload, err)
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return nil, err
	}

	u, err := url.Parse(pt.Config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse PalmTree API baseURL: %w", err)
	}
	u.Path = path.Join(u.Path, fmt.Sprintf("v1/certificate-authorities/%s/enrollments", url.PathEscape(pt.Config.CalID)))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Add("content-type", "application/json")
	req.Header.Add("accept", "application/json")

	res, err := pt.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to issue PalmTree API request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("unsuccessful PalmTree API response: %v: %v", res.StatusCode, res.Status)
	}

	var response PalmTreeResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("unable to read PalmTree API response: %w", err)
	}
	return &response, nil
}

func (p *PalmTreePayload) Validate(v structure.Validator) {
	v.String("csr", &p.CSR).NotEmpty()
	v.String("profileId", &p.ProfileID).NotEmpty()
	v.String("requiredFormat.format", &p.RequiredFormat.Format).EqualTo("PEM")
	v.String("optionalCertificateRequestDetails.subjectDn", &p.CertificateRequestDetails.SubjectDN).EqualTo("C=US")
}

func newPalmtreePayload(profileID string) *PalmTreePayload {
	return &PalmTreePayload{
		ProfileID: profileID,
		RequiredFormat: palmTreeRequiredFormat{
			Format: "PEM",
		},
		CertificateRequestDetails: palmTreeCertificateRequestDetails{
			SubjectDN: "C=US",
		},
	}
}
