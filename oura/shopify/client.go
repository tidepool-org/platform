package shopify

import "context"

type ClientConfig struct {
	StoreID      string `envconfig:"TIDEPOOL_SHOPIFY_STORE_ID"`
	ClientID     string `envconfig:"TIDEPOOL_SHOPIFY_CLIENT_ID"`
	ClientSecret string `envconfig:"TIDEPOOL_SHOPIFY_CLIENT_SECRET"`
}

//go:generate mockgen -source=client.go -destination=./test/client.go -package=test Client
type Client interface {
	CreateDiscountCode(ctx context.Context, discountCodeInput DiscountCodeInput) error
	GetDeliveredProducts(ctx context.Context, orderID string) (*DeliveredProducts, error)
}

type DiscountCodeInput struct {
	Title     string
	Code      string
	ProductID string
}

type DeliveredProducts struct {
	OrderID      string   `json:"order_id"`
	IDs          []string `json:"products"`
	DiscountCode string   `json:"discount_code"`
}
