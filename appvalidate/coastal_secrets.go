package appvalidate

import (
	"bytes"
	"context"
	"encoding/json"
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
	BaseURL            string `env:"COASTAL_BASE_URL"`
	APICertificatePath string `env:"COASTAL_API_CERTIFICATE_PATH"`
	APIKey             string `env:"COASTAL_API_KEY"`
	ClientID           string `env:"COASTAL_CLIENT_ID"`
	ClientSecret       string `env:"COSTAL_CLIENT_SECRET"`
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
	RCBMac           string   `json:"rcbMac"`
}

type CoastalResponse struct {
	Certificates []struct {
		Content   string `json:"content"`
		TTLInDays int    `json:"ttlInDays"`
		Type      string `json:"type"`
	} `json:"certificates"`
}

func (c *CoastalSecrets) GetSecret(ctx context.Context, payloadRaw []byte) (*CoastalResponse, error) {
	var payload CoastalPayload
	if err := json.Unmarshal(payloadRaw, &payload); err != nil {
		return nil, err
	}

	if err := structValidator.New().Validate(&payload); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return nil, err
	}

	u, err := url.Parse(c.Config.BaseURL)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	defer res.Body.Close()
	var response CoastalResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
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
