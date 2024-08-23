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

	"github.com/kelseyhightower/envconfig"

	"github.com/tidepool-org/platform/structure"
	structValidator "github.com/tidepool-org/platform/structure/validator"
)

var (
	ErrPalmTreeEmptyConfig = errors.New("empty PalmTree config")

	partners = []string{
		PartnerCoastal,
		PartnerPalmTree,
	}
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
	KeyData        []byte `envconfig:"PALMTREE_TLS_KEY_DATA"`
	certificateURL string
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
	if err := structValidator.New().Validate(cfg); err != nil {
		return nil, errors.Join(ErrInvalidPartnerCredentials, err)
	}
	fullPath, err := url.JoinPath(cfg.BaseURL, fmt.Sprintf("v1/certificate-authorities/%s/enrollments", url.PathEscape(cfg.CalID)))
	if err != nil {
		return nil, fmt.Errorf("unable to parse PalmTree API certificate path: %w", err)
	}
	uri, err := url.ParseRequestURI(fullPath)
	if err != nil {
		return nil, fmt.Errorf("unable to parse PalmTree API certificate URI: %w", err)
	}
	cfg.certificateURL = uri.String()

	return cfg, nil
}

func NewPalmTreeSecrets(cfg PalmTreeSecretsConfig) (*PalmTreeSecrets, error) {
	if err := structValidator.New().Validate(&cfg); err != nil {
		return nil, errors.Join(ErrInvalidPartnerCredentials, err)
	}
	cert, err := tls.X509KeyPair(cfg.CertData, cfg.KeyData)
	if err != nil {
		return nil, errors.Join(ErrInvalidPartnerCredentials, err)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}

	return &PalmTreeSecrets{
		Config: cfg,
		client: &http.Client{Transport: tr},
	}, nil
}

type PalmTreePayload struct {
	CSR            string                 `json:"csr"`
	ProfileID      string                 `json:"profileId"`
	RequiredFormat palmTreeRequiredFormat `json:"requiredFormat"`
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
		return nil, ErrInvalidPartnerCredentials
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, pt.Config.certificateURL, &buf)
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
		var body map[string]any
		_ = json.NewDecoder(res.Body).Decode(&body)
		return nil, fmt.Errorf("unsuccessful PalmTree API response: %v: %v, body: %v", res.StatusCode, res.Status, body)
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
}

func newPalmtreePayload(profileID string) *PalmTreePayload {
	return &PalmTreePayload{
		ProfileID: profileID,
		RequiredFormat: palmTreeRequiredFormat{
			Format: "PEM",
		},
	}
}

func (cfg *PalmTreeSecretsConfig) Validate(v structure.Validator) {
	v.String("BaseURL", &cfg.BaseURL).NotEmpty()
	v.String("CalID", &cfg.CalID).NotEmpty()
	v.String("ProfileID", &cfg.ProfileID).NotEmpty()
	v.Bytes("CertData", cfg.CertData).NotEmpty()
	v.Bytes("KeyData", cfg.KeyData).NotEmpty()
}
