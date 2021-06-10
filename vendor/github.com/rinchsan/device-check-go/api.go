package devicecheck

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	developmentBaseURL = "https://api.development.devicecheck.apple.com/v1"
	productionBaseURL  = "https://api.devicecheck.apple.com/v1"
)

func newBaseURL(env Environment) string {
	switch env {
	case Development:
		return developmentBaseURL
	case Production:
		return productionBaseURL
	default:
		return developmentBaseURL
	}
}

type api struct {
	client  *http.Client
	baseURL string
}

func newAPI(env Environment) api {
	return api{
		client:  http.DefaultClient,
		baseURL: newBaseURL(env),
	}
}

func newAPIWithHTTPClient(client *http.Client, env Environment) api {
	return api{
		client:  client,
		baseURL: newBaseURL(env),
	}
}

func (api api) do(jwt, path string, requestBody interface{}) (*http.Response, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(requestBody); err != nil {
		return nil, fmt.Errorf("json: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, api.baseURL+path, buf)
	if err != nil {
		return nil, fmt.Errorf("http: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))
	req.Header.Set("User-Agent", "device-check-go (+https://github.com/rinchsan/device-check-go)")

	return api.client.Do(req)
}
