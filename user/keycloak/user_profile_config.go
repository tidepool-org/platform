package keycloak

import (
	"context"
	"strings"

	"github.com/Nerzal/gocloak/v13"
)

type LengthValidator struct {
	Max int `json:"max"`
}

type DoubleValidator struct {
	Min *string `json:"min,omitempty"`
	Max *string `json:"max,omitempty"`
}

type EmailValidator struct {
	MaxLocalLength string `json:"max-local-length"`
}

type IntegerValidator struct {
	Min *string `json:"min,omitempty"`
	Max *string `json:"max,omitempty"`
}

// PatternValidator is a regex pattern validator
type PatternValidator struct {
	Pattern      string `json:"pattern"`
	ErrorMessage string `json:"error-message"`
}

type OptionsValidator struct {
	Options []string `json:"options"`
}

type UPAttributeValidations struct {
	Double  *DoubleValidator  `json:"double,omitempty"`
	Length  *LengthValidator  `json:"length,omitempty"`
	Integer *IntegerValidator `json:"integer,omitempty"`
	Options *OptionsValidator `json:"options,omitempty"`
	Pattern *PatternValidator `json:"pattern,omitempty"`
	Email   *EmailValidator   `json:"email,omitempty"`
}

type UPAttributePermissions struct {
	Edit *[]string `json:"edit,omitempty"`
	View *[]string `json:"view,omitempty"`
}

// UPAttribute is a single attribute definition for a User Profile.
type UPAttribute struct {
	Name        *string                 `json:"name,omitempty"`
	DisplayName *string                 `json:"displayName,omitempty"`
	Required    *bool                   `json:"required,omitempty"`
	MultiValued *bool                   `json:"multivalued,omitempty"`
	Validations *UPAttributeValidations `json:"validations,omitempty"`
	Permissions *UPAttributePermissions `json:"permissions,omitempty"`
	Group       *string                 `json:"group,omitempty"`
	Annotations *map[string]string      `json:"annotations,omitempty"` // might not be omitted? looks empty object if not set
}

// UPConfig represents the Keycloak "schema" for the User Profile that is
// shared across an entire Keycloak realm - see
// https://www.keycloak.org/docs-api/24.0.1/rest-api/index.html#UPConfig
type UPConfig struct {
	Attributes []UPAttribute `json:"attributes"`
}

func (c *keycloakClient) getAdminRealmURL(realm string, path ...string) string {
	path = append([]string{c.cfg.BaseUrl, "admin", "realms", realm}, path...)
	return strings.Join(path, "/")
}

func (c *keycloakClient) SetUserProfileConfig(ctx context.Context, config *UPConfig) error {
	token, err := c.getAdminToken(ctx)
	if err != nil {
		return err
	}

	var res map[string]any
	var errorResponse gocloak.HTTPErrorResponse
	response, err := c.keycloak.RestyClient().R().
		SetContext(ctx).
		SetError(&errorResponse).
		SetAuthToken(token.AccessToken).
		SetBody(config).
		SetResult(&res).
		Put(c.getAdminRealmURL(c.cfg.Realm, "users", "profile"))

	return checkForError(response, err, "unable to set User Profile Config")
}
