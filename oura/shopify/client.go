package shopify

import (
	"context"
	"strconv"
	"time"
)

const (
	ProductGIDPrefix = "gid://shopify/Product/"
	OrderGIDPrefix   = "gid://shopify/Order/"
)

type ClientConfig struct {
	StoreID      string `envconfig:"TIDEPOOL_OURA_SHOPIFY_STORE_ID"`
	ClientID     string `envconfig:"TIDEPOOL_OURA_SHOPIFY_CLIENT_ID"`
	ClientSecret string `envconfig:"TIDEPOOL_OURA_SHOPIFY_CLIENT_SECRET"`
}

//go:generate mockgen -source=client.go -destination=./test/client.go -package=test Client
type Client interface {
	CreateDiscountCode(ctx context.Context, discountCodeInput DiscountCodeInput) error
	GetOrderSummary(ctx context.Context, orderID string) (*OrderSummary, error)
}

type DiscountCodeInput struct {
	Title     string
	Code      string
	ProductID string
}

type OrderSummary struct {
	GID                 string
	CreatedTime         time.Time
	OrderedProductIDs   []string
	DeliveredProductIDs []string
	DiscountCode        string
}

func GetOrderGID(id int64) string {
	return OrderGIDPrefix + strconv.FormatInt(id, 10)
}

func GetProductGID(id int64) string {
	return ProductGIDPrefix + strconv.FormatInt(id, 10)
}
