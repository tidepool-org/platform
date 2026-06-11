package keycloak

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/Nerzal/gocloak/v13/pkg/jwx"
	"github.com/go-resty/resty/v2"
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/oauth2"

	"github.com/tidepool-org/platform/pointer"
	userlib "github.com/tidepool-org/platform/user"
)

const (
	masterRealm            = "master"
	termsAcceptedAttribute = "terms_and_conditions"
)

type KeycloakConfig struct {
	ClientID              string `envconfig:"TIDEPOOL_KEYCLOAK_CLIENT_ID" required:"true"`
	ClientSecret          string `envconfig:"TIDEPOOL_KEYCLOAK_CLIENT_SECRET" required:"true"`
	LongLivedClientID     string `envconfig:"TIDEPOOL_KEYCLOAK_LONG_LIVED_CLIENT_ID" required:"true"`
	LongLivedClientSecret string `envconfig:"TIDEPOOL_KEYCLOAK_LONG_LIVED_CLIENT_SECRET" required:"true"`
	BackendClientID       string `envconfig:"TIDEPOOL_KEYCLOAK_BACKEND_CLIENT_ID" required:"true"`
	BackendClientSecret   string `envconfig:"TIDEPOOL_KEYCLOAK_BACKEND_CLIENT_SECRET" required:"true"`
	BaseUrl               string `envconfig:"TIDEPOOL_KEYCLOAK_BASE_URL" required:"true"`
	Realm                 string `envconfig:"TIDEPOOL_KEYCLOAK_REALM" required:"true"`
	AdminUsername         string `envconfig:"TIDEPOOL_KEYCLOAK_ADMIN_USERNAME" required:"true"`
	AdminPassword         string `envconfig:"TIDEPOOL_KEYCLOAK_ADMIN_PASSWORD" required:"true"`
}

func (c *KeycloakConfig) FromEnv() error {
	return envconfig.Process("", c)
}

type keycloakClient struct {
	cfg                      *KeycloakConfig
	adminToken               *oauth2.Token
	adminTokenRefreshExpires time.Time
	keycloak                 *gocloak.GoCloak
	adminTokenLock           *sync.RWMutex
}

func newKeycloakClient(config *KeycloakConfig) *keycloakClient {
	return &keycloakClient{
		cfg:            config,
		keycloak:       gocloak.NewClient(config.BaseUrl),
		adminTokenLock: &sync.RWMutex{},
	}
}

func (c *keycloakClient) Login(ctx context.Context, username, password string) (*oauth2.Token, error) {
	return c.doLogin(ctx, c.cfg.ClientID, c.cfg.ClientSecret, username, password)
}

func (c *keycloakClient) LoginLongLived(ctx context.Context, username, password string) (*oauth2.Token, error) {
	return c.doLogin(ctx, c.cfg.LongLivedClientID, c.cfg.LongLivedClientSecret, username, password)
}

func (c *keycloakClient) doLogin(ctx context.Context, clientId, clientSecret, username, password string) (*oauth2.Token, error) {
	jwt, err := c.keycloak.Login(
		ctx,
		clientId,
		clientSecret,
		c.cfg.Realm,
		username,
		password,
	)
	if err != nil {
		return nil, err
	}
	return c.jwtToAccessToken(jwt), nil
}

func (c *keycloakClient) GetBackendServiceToken(ctx context.Context) (*oauth2.Token, error) {
	jwt, err := c.keycloak.LoginClient(ctx, c.cfg.BackendClientID, c.cfg.BackendClientSecret, c.cfg.Realm)
	if err != nil {
		return nil, err
	}
	return c.jwtToAccessToken(jwt), nil
}

func (c *keycloakClient) jwtToAccessToken(jwt *gocloak.JWT) *oauth2.Token {
	if jwt == nil {
		return nil
	}
	return (&oauth2.Token{
		AccessToken:  jwt.AccessToken,
		TokenType:    jwt.TokenType,
		RefreshToken: jwt.RefreshToken,
		Expiry:       time.Now().Add(time.Duration(jwt.ExpiresIn) * time.Second),
	}).WithExtra(map[string]interface{}{
		"refresh_expires_in": jwt.RefreshExpiresIn,
	})
}

func (c *keycloakClient) RevokeToken(ctx context.Context, token oauth2.Token) error {
	clientId, clientSecret := c.getClientAndSecretFromToken(ctx, token)
	return c.keycloak.Logout(
		ctx,
		clientId,
		clientSecret,
		c.cfg.Realm,
		token.RefreshToken,
	)
}

func (c *keycloakClient) RefreshToken(ctx context.Context, token oauth2.Token) (*oauth2.Token, error) {
	clientId, clientSecret := c.getClientAndSecretFromToken(ctx, token)

	jwt, err := c.keycloak.RefreshToken(
		ctx,
		token.RefreshToken,
		clientId,
		clientSecret,
		c.cfg.Realm,
	)
	if err != nil {
		return nil, err
	}
	return c.jwtToAccessToken(jwt), nil
}

func (c *keycloakClient) GetUserById(ctx context.Context, id string) (*userlib.User, error) {
	if id == "" {
		return nil, nil
	}

	users, err := c.FindUsersWithIds(ctx, []string{id})
	if err != nil || len(users) == 0 {
		return nil, err
	}

	return users[0], nil
}

func (c *keycloakClient) GetUserByEmail(ctx context.Context, email string) (*userlib.User, error) {
	if email == "" {
		return nil, nil
	}
	token, err := c.getAdminToken(ctx)
	if err != nil {
		return nil, err
	}

	users, err := c.keycloak.GetUsers(ctx, token.AccessToken, c.cfg.Realm, gocloak.GetUsersParams{
		Email: &email,
		Exact: gocloak.BoolP(true),
	})
	if err != nil || len(users) == 0 {
		return nil, err
	}

	return c.GetUserById(ctx, *users[0].ID)
}

func (c *keycloakClient) UpdateUser(ctx context.Context, user *userlib.User) error {
	token, err := c.getAdminToken(ctx)
	if err != nil {
		return err
	}

	gocloakUser := gocloak.User{
		ID:            user.UserID,
		Username:      user.Username,
		Enabled:       &user.Enabled,
		EmailVerified: user.EmailVerified,
		Email:         user.Username,
	}

	attrs := map[string][]string{}
	if terms := pointer.ToString(user.TermsAccepted); terms != "" {
		attrs[termsAcceptedAttribute] = []string{terms}
	}

	if user.Profile != nil {
		maps.Copy(attrs, user.Profile.ToAttributes())
	}

	gocloakUser.Attributes = &attrs
	if err := c.keycloak.UpdateUser(ctx, token.AccessToken, c.cfg.Realm, gocloakUser); err != nil {
		return err
	}
	if err := c.updateRolesForUser(ctx, user); err != nil {
		return err
	}
	return nil
}

func (c *keycloakClient) UpdateUserProfile(ctx context.Context, id string, p *userlib.Profile) error {
	user, err := c.GetUserById(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return userlib.ErrUserNotFound
	}
	user.Profile = p
	return c.UpdateUser(ctx, user)
}

func (c *keycloakClient) FindUsersWithIds(ctx context.Context, ids []string) (users []*userlib.User, err error) {
	const errMessage = "could not retrieve users by ids"

	token, err := c.getAdminToken(ctx)
	if err != nil {
		return nil, err
	}

	var res []*gocloak.User
	var errorResponse gocloak.HTTPErrorResponse
	response, err := c.keycloak.RestyClient().R().
		SetContext(ctx).
		SetError(&errorResponse).
		SetAuthToken(token.AccessToken).
		SetResult(&res).
		SetQueryParam("ids", strings.Join(ids, ",")).
		Get(c.getRealmURL(c.cfg.Realm, "tidepool-admin", "users"))

	err = checkForError(response, err, errMessage)
	if err != nil {
		return nil, err
	}

	users = make([]*userlib.User, len(res))
	for i, u := range res {
		users[i] = newUserFromGocloakUser(u)
	}

	return users, nil
}

func (c *keycloakClient) IntrospectToken(ctx context.Context, token oauth2.Token) (*userlib.TokenIntrospectionResult, error) {
	clientId, clientSecret := c.getClientAndSecretFromToken(ctx, token)

	rtr, err := c.keycloak.RetrospectToken(
		ctx,
		token.AccessToken,
		clientId,
		clientSecret,
		c.cfg.Realm,
	)
	if err != nil {
		return nil, err
	}

	result := &userlib.TokenIntrospectionResult{
		Active: pointer.ToBool(rtr.Active),
	}
	if result.Active {
		customClaims := &userlib.AccessTokenCustomClaims{}
		_, err := c.keycloak.DecodeAccessTokenCustomClaims(
			ctx,
			token.AccessToken,
			c.cfg.Realm,
			customClaims,
		)
		if err != nil {
			return nil, err
		}
		result.Subject = customClaims.Subject
		result.EmailVerified = customClaims.EmailVerified
		result.ExpiresAt = customClaims.ExpiresAt.Unix()
		result.RealmAccess = userlib.RealmAccess{
			Roles: customClaims.RealmAccess.Roles,
		}
		result.IdentityProvider = customClaims.IdentityProvider
	}

	return result, nil
}

func (c *keycloakClient) DeleteUserSessions(ctx context.Context, id string) error {
	token, err := c.getAdminToken(ctx)
	if err != nil {
		return err
	}

	if err := c.keycloak.LogoutAllSessions(ctx, token.AccessToken, c.cfg.Realm, id); err != nil {
		if aErr, ok := err.(*gocloak.APIError); ok && aErr.Code == http.StatusNotFound {
			return nil
		}
	}

	return err
}

func (c *keycloakClient) getRealmURL(realm string, path ...string) string {
	path = append([]string{c.cfg.BaseUrl, "realms", realm}, path...)
	return strings.Join(path, "/")
}

func (c *keycloakClient) getAdminToken(ctx context.Context) (oauth2.Token, error) {
	var err error
	if c.adminTokenIsExpired() {
		if err := c.loginAsAdmin(ctx); err != nil {
			return oauth2.Token{}, err
		}
	}

	c.adminTokenLock.RLock()
	defer c.adminTokenLock.RUnlock()
	return *c.adminToken, err
}

func (c *keycloakClient) loginAsAdmin(ctx context.Context) error {
	jwt, err := c.keycloak.LoginAdmin(
		ctx,
		c.cfg.AdminUsername,
		c.cfg.AdminPassword,
		masterRealm,
	)
	if err != nil {
		return err
	}

	c.adminTokenLock.Lock()
	defer c.adminTokenLock.Unlock()
	c.adminToken = c.jwtToAccessToken(jwt)
	expiration := time.Now().Add(time.Duration(jwt.ExpiresIn)*time.Second - time.Second*5) // check if adding a small buffer to expire time to allow earlier refresh still results in a time in the future
	if expiration.After(time.Now()) {
		c.adminTokenRefreshExpires = expiration
	} else {
		c.adminTokenRefreshExpires = time.Now().Add(time.Duration(jwt.ExpiresIn) * time.Second)
	}
	return nil
}

func (c *keycloakClient) adminTokenIsExpired() bool {
	c.adminTokenLock.RLock()
	defer c.adminTokenLock.RUnlock()
	return c.adminToken == nil || time.Now().After(c.adminTokenRefreshExpires)
}

func (c *keycloakClient) updateRolesForUser(ctx context.Context, user *userlib.User) error {
	token, err := c.getAdminToken(ctx)
	if err != nil {
		return err
	}
	userID := pointer.ToString(user.UserID)

	realmRoles, err := c.keycloak.GetRealmRoles(ctx, token.AccessToken, c.cfg.Realm, gocloak.GetRoleParams{
		Max: gocloak.IntP(1000),
	})
	if err != nil {
		return err
	}
	currentUserRoles, err := c.keycloak.GetRealmRolesByUserID(ctx, token.AccessToken, c.cfg.Realm, userID)
	if err != nil {
		return err
	}

	var rolesToAdd []gocloak.Role
	var rolesToDelete []gocloak.Role

	targetRoles := make(map[string]struct{})
	if user.Roles != nil && len(*user.Roles) > 0 {
		for _, targetRoleName := range *user.Roles {
			targetRoles[targetRoleName] = struct{}{}
		}
	}

	for targetRoleName := range targetRoles {
		realmRole := getRealmRoleByName(realmRoles, targetRoleName)
		if realmRole != nil {
			rolesToAdd = append(rolesToAdd, *realmRole)
		}
	}

	if len(currentUserRoles) > 0 {
		for _, currentRole := range currentUserRoles {
			if currentRole == nil || currentRole.Name == nil || *currentRole.Name == "" {
				continue
			}

			if _, ok := targetRoles[*currentRole.Name]; !ok {
				// Only remove roles managed by shoreline
				if _, ok := userlib.ShorelineManagedRoles[*currentRole.Name]; ok {
					rolesToDelete = append(rolesToDelete, *currentRole)
				}
			}
		}
	}

	if len(rolesToAdd) > 0 {
		if err = c.keycloak.AddRealmRoleToUser(ctx, token.AccessToken, c.cfg.Realm, userID, rolesToAdd); err != nil {
			return err
		}
	}
	if len(rolesToDelete) > 0 {
		if err = c.keycloak.DeleteRealmRoleFromUser(ctx, token.AccessToken, c.cfg.Realm, userID, rolesToDelete); err != nil {
			return err
		}
	}

	return nil
}

func (c *keycloakClient) GetRolesForUser(ctx context.Context, userID string) ([]string, error) {
	token, err := c.getAdminToken(ctx)
	if err != nil {
		return nil, err
	}

	realmRoles, err := c.keycloak.GetRealmRolesByUserID(ctx, token.AccessToken, c.cfg.Realm, userID)
	if err != nil {
		return nil, err
	}

	roles := make([]string, 0, len(realmRoles))
	for _, role := range realmRoles {
		if role == nil || strings.TrimSpace(pointer.ToString(role.Name)) == "" {
			continue
		}
		roleName := strings.TrimSpace(pointer.ToString(role.Name))
		roles = append(roles, roleName)
	}

	return roles, nil
}

func (c *keycloakClient) getClientAndSecretFromToken(ctx context.Context, token oauth2.Token) (string, string) {
	clientId := c.cfg.ClientID
	clientSecret := c.cfg.ClientSecret

	customClaims := &jwx.Claims{}
	_, err := c.keycloak.DecodeAccessTokenCustomClaims(
		ctx,
		token.AccessToken,
		c.cfg.Realm,
		customClaims,
	)

	if err == nil && customClaims.Azp == c.cfg.LongLivedClientID {
		clientId = c.cfg.LongLivedClientID
		clientSecret = c.cfg.LongLivedClientSecret
	}

	return clientId, clientSecret
}

func newUserFromGocloakUser(gocloakUser *gocloak.User) *userlib.User {
	user := &userlib.User{
		UserID:        gocloakUser.ID,
		Username:      gocloakUser.Username,
		Emails:        []string{},
		Roles:         gocloakUser.RealmRoles,
		EmailVerified: gocloakUser.EmailVerified,
		IsMigrated:    true,
		Enabled:       pointer.ToBool(gocloakUser.Enabled),
	}
	if gocloakUser.Attributes != nil {
		attrs := *gocloakUser.Attributes
		if termsAttrs, ok := attrs[termsAcceptedAttribute]; ok && len(termsAttrs) > 0 {
			if ts, err := userlib.UnixStringToTimestamp(termsAttrs[0]); err == nil {
				user.TermsAccepted = &ts
			}
		}
		var roles []string
		if gocloakUser.RealmRoles != nil {
			roles = *gocloakUser.RealmRoles
		}
		if profile := userlib.ProfileFromAttributes(pointer.ToString(gocloakUser.Username), attrs, roles); profile != nil {
			user.Profile = profile
		}
	}

	// All non-custodial users have a password and it's important to set the hash to a non-empty value.
	// When users are serialized by this service, the payload contains a flag `passwordExists` that
	// is computed based on the presence of a password hash in the user struct. This flag is used by
	// other services (e.g. hydrophone) to determine whether the user is custodial or not.
	if !user.IsCustodialAccount() {
		user.PwHash = "true"
	}

	return user
}

func getRealmRoleByName(realmRoles []*gocloak.Role, name string) *gocloak.Role {
	for _, realmRole := range realmRoles {
		if realmRole.Name != nil && *realmRole.Name == name {
			return realmRole
		}
	}

	return nil
}

// checkForError Copied from gocloak - used for sending requests to custom endpoints
func checkForError(resp *resty.Response, err error, errMessage string) error {
	if err != nil {
		return &gocloak.APIError{
			Code:    0,
			Message: fmt.Errorf("%w: %s", err, errMessage).Error(),
		}
	}

	if resp == nil {
		return &gocloak.APIError{
			Message: "empty response",
		}
	}

	if resp.IsError() {
		var msg string

		if e, ok := resp.Error().(*gocloak.HTTPErrorResponse); ok && e.NotEmpty() {
			msg = fmt.Sprintf("%s: %s", resp.Status(), e)
		} else {
			msg = resp.Status()
		}

		return &gocloak.APIError{
			Code:    resp.StatusCode(),
			Message: msg,
		}
	}

	return nil
}
