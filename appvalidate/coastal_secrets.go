package appvalidate

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/kelseyhightower/envconfig"

	"github.com/tidepool-org/platform/structure"
	structValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	PartnerCoastal = "Coastal"
)

type CoastalSecretsConfig struct {
	APIKey       string `envconfig:"COASTAL_API_KEY"`
	BaseURL      string `envconfig:"COASTAL_BASE_URL"`
	ClientID     string `envconfig:"COASTAL_CLIENT_ID"`
	ClientSecret string `envconfig:"COASTAL_CLIENT_SECRET"`
	RCTypeID     string `envconfig:"COASTAL_RC_TYPE_ID"`
	// KeyData is the raw contents of the ED25519 private key file in PEM format.
	KeyData        []byte `envconfig:"COASTAL_PRIVATE_KEY_DATA"`
	certificateURL string
}

type CoastalSecrets struct {
	Config CoastalSecretsConfig
	pk     ed25519.PrivateKey
}

func NewCoastalSecretsConfig() (*CoastalSecretsConfig, error) {
	cfg := &CoastalSecretsConfig{}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, err
	}
	if err := structValidator.New().Validate(cfg); err != nil {
		return nil, errors.Join(ErrInvalidPartnerCredentials, err)
	}
	fullPath, err := url.JoinPath(cfg.BaseURL, "devices/api/v1/certificates")
	if err != nil {
		return nil, fmt.Errorf("unable to parse Coastal API certificate path: %w", err)
	}
	uri, err := url.ParseRequestURI(fullPath)
	if err != nil {
		return nil, fmt.Errorf("unable to parse Coastal API certificate URI: %w", err)
	}
	cfg.certificateURL = uri.String()

	return cfg, nil
}

func NewCoastalSecrets(cfg CoastalSecretsConfig) (*CoastalSecrets, error) {
	if err := structValidator.New().Validate(&cfg); err != nil {
		return nil, errors.Join(ErrInvalidPartnerCredentials, err)
	}
	keyBlock, _ := pem.Decode(cfg.KeyData)
	if keyBlock == nil {
		return nil, fmt.Errorf("Coastal key data is not in PEM format: %w", ErrInvalidPartnerCredentials)
	}
	privKeyAny, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, errors.Join(ErrInvalidPartnerCredentials, fmt.Errorf("unable to parse EC private key: %w", err))
	}
	privateKey, ok := privKeyAny.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("unexpected coastal private key type: %T", privKeyAny)
	}
	return &CoastalSecrets{
		Config: cfg,
		pk:     privateKey,
	}, nil
}

// CoastalPayload is the external front facing payload sent to platform from clients.
type CoastalPayload struct {
	RCTypeID        string `json:"rcTypeId"`
	RCInstanceID    string `json:"rcInstanceId"`
	HardwareVersion string `json:"rcHWVersion"`
	SoftwareVersion string `json:"rcSWVersion"`
	PHDTypeID       string `json:"phdTypeId"`
	PHDInstanceID   string `json:"phdInstanceId"`
	CSR             string `json:"csr"`
}

func (c *CoastalPayload) toInternalPayload(pk ed25519.PrivateKey) (payload *coastalPayload, signature string, err error) {
	payload = &coastalPayload{
		RCTypeID:         c.RCTypeID,
		RCInstanceID:     c.RCInstanceID,
		HardwareVersions: []string{c.HardwareVersion},
		SoftwareVersions: []string{c.SoftwareVersion},
		PHDTypeID:        c.PHDTypeID,
		PHDInstanceID:    c.PHDInstanceID,
		CSR:              c.CSR,
	}

	bytesRaw, err := json.Marshal(payload)
	if err != nil {
		return nil, "", fmt.Errorf("unable to marshal payload when calculating signature: %w", err)
	}
	signature = base64.StdEncoding.EncodeToString(ed25519.Sign(pk, bytesRaw))
	return payload, signature, nil
}

// coastalPayload is the internal facing actual Coastal payload that platform sends to coastal's API.
type coastalPayload struct {
	RCTypeID         string   `json:"rcTypeId"`
	RCInstanceID     string   `json:"rcInstanceId"`
	HardwareVersions []string `json:"rcHWVersions"`
	SoftwareVersions []string `json:"rcSWVersions"`
	PHDTypeID        string   `json:"phdTypeId"`
	PHDInstanceID    string   `json:"phdInstanceId"`
	CSR              string   `json:"csr"`
}

type CoastalResponse struct {
	Certificates []struct {
		Content   string `json:"content"`
		TTLInDays int    `json:"ttlInDays"`
		Type      string `json:"type"`
	} `json:"certificates"`
}

func (c *CoastalSecrets) GetSecret(ctx context.Context, partnerDataRaw []byte) (*CoastalResponse, error) {
	if c.pk == nil {
		return nil, ErrInvalidPartnerCredentials
	}
	payload := newCoastalPayload(c.Config.RCTypeID)
	if err := json.Unmarshal(partnerDataRaw, payload); err != nil {
		return nil, fmt.Errorf("unable to unmarshal Coastal payload: %w", err)
	}

	if err := structValidator.New().Validate(payload); err != nil {
		return nil, fmt.Errorf("Coastal: %w: %w", ErrInvalidPartnerPayload, err)
	}

	internalPayload, signature, err := payload.toInternalPayload(c.pk)
	if err != nil {
		return nil, fmt.Errorf("unable to genarate internal Coastal payload: %w", err)
	}

	b, err := json.Marshal(internalPayload)
	if err != nil {
		return nil, fmt.Errorf("Coastal marshal error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Config.certificateURL, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Add("apiKey", c.Config.APIKey)
	req.Header.Add("client_id", c.Config.ClientID)
	req.Header.Add("client_secret", c.Config.ClientSecret)
	req.Header.Add("content-type", "application/json")
	req.Header.Add("accept", "application/json")
	req.Header.Add("signature", signature)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to issue Coastal API request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		var msg string
		if bits, err := io.ReadAll(res.Body); err == nil {
			msg = string(bits)
		}
		return nil, fmt.Errorf("unsuccessful Coastal API response: %v: %v: %v", res.StatusCode, res.Status, msg)
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
	v.String("rcHWVersion", &p.HardwareVersion).NotEmpty()
	v.String("rcSWVersion", &p.SoftwareVersion).NotEmpty()
	v.String("phdTypeId", &p.PHDTypeID).NotEmpty()
	v.String("phdInstanceId", &p.PHDInstanceID).NotEmpty()
	v.String("csr", &p.CSR).NotEmpty()
}

func newCoastalPayload(rcTypeID string) *CoastalPayload {
	return &CoastalPayload{
		RCTypeID: rcTypeID,
	}
}

func (cfg *CoastalSecretsConfig) Validate(v structure.Validator) {
	v.String("APIKey", &cfg.APIKey).NotEmpty()
	v.String("BaseURL", &cfg.BaseURL).NotEmpty()
	v.String("ClientID", &cfg.ClientID).NotEmpty()
	v.String("ClientSecret", &cfg.ClientSecret).NotEmpty()
	v.String("RCTypeID", &cfg.RCTypeID).NotEmpty()
	v.Bytes("KeyData", cfg.KeyData).NotEmpty()
}
