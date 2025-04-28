package provider

import "github.com/golang-jwt/jwt/v4"

type Claims struct {
	jwt.RegisteredClaims
	TidepoolLinkID string `json:"custom:tidepool_link_id"`
}
