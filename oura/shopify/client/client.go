package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Khan/genqlient/graphql"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/tidepool-org/platform/oura/shopify"
)

type defaultClient struct {
	gql graphql.Client
}

func NewClient(ctx context.Context, cfg shopify.ClientConfig) (shopify.Client, error) {
	oauthConfig := clientcredentials.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		TokenURL:     fmt.Sprintf("https://%s.myshopify.com/admin/oauth/access_token", cfg.StoreID),
	}

	httpClient := http.Client{
		Transport: newAuthedTransport(oauthConfig.TokenSource(ctx), http.DefaultTransport),
	}

	endpoint := fmt.Sprintf("https://%s.myshopify.com/admin/api/2025-10/graphql.json", cfg.StoreID)

	return &defaultClient{
		gql: graphql.NewClient(endpoint, &httpClient),
	}, nil
}
