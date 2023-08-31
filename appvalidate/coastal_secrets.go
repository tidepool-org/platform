package appvalidate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/kelseyhightower/envconfig"

	"github.com/tidepool-org/platform/structure"
	structValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	PartnerCoastal = "Coastal"
)

type CoastalSecretsConfig struct {
	APICertificatePath string `envconfig:"COASTAL_API_CERTIFICATE_PATH"`
	APIKey             string `envconfig:"COASTAL_API_KEY"`
	BaseURL            string `envconfig:"COASTAL_BASE_URL"`
	ClientID           string `envconfig:"COASTAL_CLIENT_ID"`
	ClientSecret       string `envconfig:"COSTAL_CLIENT_SECRET"`
	RCTypeID           string `envconfig:"COASTAL_RC_TYPE_ID"`
}

type CoastalSecrets struct {
	Config CoastalSecretsConfig
}

func NewCoastalSecretsConfig() (*CoastalSecretsConfig, error) {
	cfg := &CoastalSecretsConfig{}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

type CoastalPayload struct {
	RCTypeID         string   `json:"rcTypeId"`
	RCInstanceID     string   `json:"rcInstanceId"`
	HardwareVersions []string `json:"rcHWVersions"`
	SoftwareVersions []string `json:"rcSWVersions"`
	PHDTypeID        string   `json:"phdTypeId"`
	PHDInstanceID    string   `json:"phdInstanceId"`
	CSR              string   `json:"csr"`
	RCBMac           string   `json:"rcbMac"` // Note this field will be renamed to rcbSignature in V3 with actual validation, it is currently required to be any non empty string for V2 otherwise a 400 error is returned.
}

type CoastalResponse struct {
	Certificates []struct {
		Content   string `json:"content"`
		TTLInDays int    `json:"ttlInDays"`
		Type      string `json:"type"`
	} `json:"certificates"`
}

func (c *CoastalSecrets) GetSecret(ctx context.Context, partnerDataRaw []byte) (*CoastalResponse, error) {
	payload := newCoastalPayload(c.Config.RCTypeID)
	if err := json.Unmarshal(partnerDataRaw, payload); err != nil {
		return nil, fmt.Errorf("unable to unmarshal Coastal payload: %w", err)
	}
	// Todo: calculate rcbMac / rcbSignature when partner API is updated.

	if err := structValidator.New().Validate(payload); err != nil {
		return nil, fmt.Errorf("unable to validate Coastal payload: %w", err)
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return nil, err
	}

	u, err := url.Parse(c.Config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to prase Coastal API baseURL: %w", err)
	}
	u.Path = path.Join(u.Path, c.Config.APICertificatePath)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Add("apiKey", c.Config.APIKey)
	req.Header.Add("client_id", c.Config.ClientID)
	req.Header.Add("client_secret", c.Config.ClientSecret)
	req.Header.Add("content-type", "application/json")
	req.Header.Add("accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to issue Coastal API request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("unsuccessful Coastal API response: %v: %v", res.StatusCode, res.Status)
	}

	var response CoastalResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("unable to read Coastal API response: %w", err)
	}
	return &response, nil
}

func (p *CoastalPayload) Validate(v structure.Validator) {
	v.String("rcTypeId", &p.RCTypeID).NotEmpty()
	v.String("rcInstanceId", &p.RCInstanceID).NotEmpty()
	v.StringArray("rcHWVersions", &p.HardwareVersions).NotEmpty()
	v.StringArray("rcSWVersions", &p.SoftwareVersions).NotEmpty()
	v.String("phdTypeId", &p.PHDTypeID).NotEmpty()
	v.String("phdInstanceId", &p.PHDInstanceID).NotEmpty()
	v.String("csr", &p.CSR).NotEmpty()
	v.String("rcbMac", &p.RCBMac).NotEmpty()
}

func newCoastalPayload(rcTypeID string) *CoastalPayload {
	return &CoastalPayload{
		RCTypeID: rcTypeID,
	}
}
