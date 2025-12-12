package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Khan/genqlient/graphql"
	"golang.org/x/oauth2/clientcredentials"
)

type ClientConfig struct {
	StoreID      string `envconfig:"TIDEPOOL_SHOPIFY_STORE_ID" required:"true"`
	ClientID     string `envconfig:"TIDEPOOL_SHOPIFY_CLIENT_ID" required:"true"`
	ClientSecret string `envconfig:"TIDEPOOL_SHOPIFY_CLIENT_SECRET" required:"true"`
}

type Client struct {
	gql graphql.Client
}

func NewClient(ctx context.Context, cfg ClientConfig) (*Client, error) {
	oauthConfig := clientcredentials.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		TokenURL:     fmt.Sprintf("https://%s.myshopify.com/admin/oauth/access_token", cfg.StoreID),
	}

	httpClient := http.Client{
		Transport: NewAuthedTransport(oauthConfig.TokenSource(ctx), http.DefaultTransport),
	}

	endpoint := fmt.Sprintf("https://%s.myshopify.com/admin/api/2025-10/graphql.json", cfg.StoreID)

	return &Client{
		gql: graphql.NewClient(endpoint, &httpClient),
	}, nil
}
