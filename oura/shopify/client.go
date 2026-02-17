package shopify

import (
	"context"

	"github.com/tidepool-org/platform/oura/shopify/generated"
)

type ClientConfig struct {
	StoreID      string `envconfig:"TIDEPOOL_OURA_SHOPIFY_STORE_ID"`
	ClientID     string `envconfig:"TIDEPOOL_OURA_SHOPIFY_CLIENT_ID"`
	ClientSecret string `envconfig:"TIDEPOOL_OURA_SHOPIFY_CLIENT_SECRET"`
}

//go:generate mockgen -source=client.go -destination=./test/client.go -package=test Client
type Client interface {
	CreateDiscountCode(ctx context.Context, discountCodeInput DiscountCodeInput) error
	GetOrder(ctx context.Context, orderID string) (*generated.GetOrderOrderByIdentifierOrder, error)
	GetProductsFromOrder(order *generated.GetOrderOrderByIdentifierOrder) *Products
}

type DiscountCodeInput struct {
	Title     string
	Code      string
	ProductID string
}

type Products struct {
	IDs          []string `json:"products"`
	DiscountCode string   `json:"discount_code"`
}
