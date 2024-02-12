package user

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"strings"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/Nerzal/gocloak/v13/pkg/jwx"
	"github.com/go-resty/resty/v2"
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/oauth2"

	"github.com/tidepool-org/platform/pointer"
)

const (
	tokenPrefix            = "kc"
	tokenPartsSeparator    = ":"
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

// keycloakUser is an intermediate user representation from FullModel to gocloak's model - though is this actually needed? Can it be removed entirely?
type keycloakUser struct {
	ID            string                 `json:"id"`
	Username      string                 `json:"username,omitempty"`
	Email         string                 `json:"email,omitempty"`
	FirstName     string                 `json:"firstName,omitempty"`
	LastName      string                 `json:"lastName,omitempty"`
	Enabled       bool                   `json:"enabled,omitempty"`
	EmailVerified bool                   `json:"emailVerified,omitempty"`
	Roles         []string               `json:"roles,omitempty"`
	Attributes    keycloakUserAttributes `json:"attributes"`
}

type keycloakUserAttributes struct {
	TermsAcceptedDate []string     `json:"terms_and_conditions,omitempty"`
	Profile           *UserProfile `json:"profile"`
}

type keycloakClient struct {
	cfg                      *KeycloakConfig
	adminToken               *oauth2.Token
	adminTokenRefreshExpires time.Time
	keycloak                 *gocloak.GoCloak
}

func newKeycloakClient(config *KeycloakConfig) *keycloakClient {
	return &keycloakClient{
		cfg:      config,
		keycloak: gocloak.NewClient(config.BaseUrl),
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
	fmt.Println("GetBackendServiceToken LoginClient", c.cfg.BackendClientID, c.cfg.BackendClientSecret, c.cfg.Realm)
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

func (c *keycloakClient) GetUserById(ctx context.Context, id string) (*keycloakUser, error) {
	if id == "" {
		return nil, nil
	}

	users, err := c.FindUsersWithIds(ctx, []string{id})
	if err != nil || len(users) == 0 {
		return nil, err
	}

	return users[0], nil
}

func (c *keycloakClient) GetUserByEmail(ctx context.Context, email string) (*keycloakUser, error) {
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

func (c *keycloakClient) UpdateUser(ctx context.Context, user *keycloakUser) error {
	token, err := c.getAdminToken(ctx)
	if err != nil {
		return err
	}

	gocloakUser := gocloak.User{
		ID:            &user.ID,
		Username:      &user.Username,
		Enabled:       &user.Enabled,
		EmailVerified: &user.EmailVerified,
		FirstName:     &user.FirstName,
		LastName:      &user.LastName,
		Email:         &user.Email,
	}

	attrs := map[string][]string{
		termsAcceptedAttribute: user.Attributes.TermsAcceptedDate,
	}
	if user.Attributes.Profile != nil {
		profileAttrs := user.Attributes.Profile.ToAttributes()
		maps.Copy(attrs, profileAttrs)
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

func (c *keycloakClient) UpdateUserProfile(ctx context.Context, id string, p *UserProfile) error {
	user, err := c.GetUserById(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}
	user.Attributes.Profile = p
	return c.UpdateUser(ctx, user)
}

func (c *keycloakClient) UpdateUserPassword(ctx context.Context, id, password string) error {
	token, err := c.getAdminToken(ctx)
	if err != nil {
		return err
	}

	return c.keycloak.SetPassword(
		ctx,
		token.AccessToken,
		id,
		c.cfg.Realm,
		password,
		false,
	)
}

func (c *keycloakClient) CreateUser(ctx context.Context, user *keycloakUser) (*keycloakUser, error) {
	token, err := c.getAdminToken(ctx)
	if err != nil {
		return nil, err
	}

	model := gocloak.User{
		Username:      &user.Username,
		Email:         &user.Email,
		EmailVerified: &user.EmailVerified,
		Enabled:       &user.Enabled,
		RealmRoles:    &user.Roles,
	}

	if len(user.Attributes.TermsAcceptedDate) > 0 {
		attrs := map[string][]string{
			termsAcceptedAttribute: user.Attributes.TermsAcceptedDate,
		}
		model.Attributes = &attrs
	}

	user.ID, err = c.keycloak.CreateUser(ctx, token.AccessToken, c.cfg.Realm, model)
	if err != nil {
		if e, ok := err.(*gocloak.APIError); ok && e.Code == http.StatusConflict {
			err = ErrUserConflict
		}
		return nil, err
	}

	if err := c.updateRolesForUser(ctx, user); err != nil {
		return nil, err
	}

	return c.GetUserById(ctx, user.ID)
}

func (c *keycloakClient) FindUsersWithIds(ctx context.Context, ids []string) (users []*keycloakUser, err error) {
	const errMessage = "could not retrieve users by ids"

	token, err := c.getAdminToken(ctx)
	if err != nil {
		return
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
		return
	}

	users = make([]*keycloakUser, len(res))
	for i, u := range res {
		users[i] = newKeycloakUser(u)
	}

	return
}

func (c *keycloakClient) IntrospectToken(ctx context.Context, token oauth2.Token) (*TokenIntrospectionResult, error) {
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

	result := &TokenIntrospectionResult{
		Active: pointer.ToBool(rtr.Active),
	}
	if result.Active {
		customClaims := &AccessTokenCustomClaims{}
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
		result.ExpiresAt = customClaims.ExpiresAt
		result.RealmAccess = RealmAccess{
			Roles: customClaims.RealmAccess.Roles,
		}
		result.IdentityProvider = customClaims.IdentityProvider
	}

	return result, nil
}

func (c *keycloakClient) DeleteUser(ctx context.Context, id string) error {
	token, err := c.getAdminToken(ctx)
	if err != nil {
		return err
	}

	if err := c.keycloak.DeleteUser(ctx, token.AccessToken, c.cfg.Realm, id); err != nil {
		if aErr, ok := err.(*gocloak.APIError); ok && aErr.Code == http.StatusNotFound {
			return nil
		}
	}
	return err
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

func (c *keycloakClient) getAdminToken(ctx context.Context) (*oauth2.Token, error) {
	var err error
	if c.adminTokenIsExpired() {
		err = c.loginAsAdmin(ctx)
	}

	return c.adminToken, err
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

	c.adminToken = c.jwtToAccessToken(jwt)
	c.adminTokenRefreshExpires = time.Now().Add(time.Duration(jwt.ExpiresIn) * time.Second)
	return nil
}

func (c *keycloakClient) adminTokenIsExpired() bool {
	return c.adminToken == nil || time.Now().After(c.adminTokenRefreshExpires)
}

func (c *keycloakClient) updateRolesForUser(ctx context.Context, user *keycloakUser) error {
	token, err := c.getAdminToken(ctx)
	if err != nil {
		return err
	}

	realmRoles, err := c.keycloak.GetRealmRoles(ctx, token.AccessToken, c.cfg.Realm, gocloak.GetRoleParams{
		Max: gocloak.IntP(1000),
	})
	if err != nil {
		return err
	}
	currentUserRoles, err := c.keycloak.GetRealmRolesByUserID(ctx, token.AccessToken, c.cfg.Realm, user.ID)
	if err != nil {
		return err
	}

	var rolesToAdd []gocloak.Role
	var rolesToDelete []gocloak.Role

	targetRoles := make(map[string]struct{})
	if len(user.Roles) > 0 {
		for _, targetRoleName := range user.Roles {
			targetRoles[targetRoleName] = struct{}{}
		}
	}

	for targetRoleName, _ := range targetRoles {
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
				if _, ok := shorelineManagedRoles[*currentRole.Name]; ok {
					rolesToDelete = append(rolesToDelete, *currentRole)
				}
			}
		}
	}

	if len(rolesToAdd) > 0 {
		if err = c.keycloak.AddRealmRoleToUser(ctx, token.AccessToken, c.cfg.Realm, user.ID, rolesToAdd); err != nil {
			return err
		}
	}
	if len(rolesToDelete) > 0 {
		if err = c.keycloak.DeleteRealmRoleFromUser(ctx, token.AccessToken, c.cfg.Realm, user.ID, rolesToDelete); err != nil {
			return err
		}
	}

	return nil
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

func newKeycloakUser(gocloakUser *gocloak.User) *keycloakUser {
	if gocloakUser == nil {
		return nil
	}

	user := &keycloakUser{
		ID:            pointer.ToString(gocloakUser.ID),
		Username:      pointer.ToString(gocloakUser.Username),
		FirstName:     pointer.ToString(gocloakUser.FirstName),
		LastName:      pointer.ToString(gocloakUser.LastName),
		Email:         pointer.ToString(gocloakUser.Email),
		EmailVerified: pointer.ToBool(gocloakUser.EmailVerified),
		Enabled:       pointer.ToBool(gocloakUser.Enabled),
	}
	if gocloakUser.Attributes != nil {
		attrs := *gocloakUser.Attributes
		if ts, ok := attrs[termsAcceptedAttribute]; ok {
			user.Attributes.TermsAcceptedDate = ts
		}
		if prof, ok := profileFromAttributes(attrs); ok {
			user.Attributes.Profile = prof
		}
	}

	if gocloakUser.RealmRoles != nil {
		user.Roles = *gocloakUser.RealmRoles
	}

	return user
}

func newUserFromKeycloakUser(keycloakUser *keycloakUser) *FullUser {
	termsAcceptedDate := ""
	attrs := keycloakUser.Attributes
	if len(attrs.TermsAcceptedDate) > 0 {
		if ts, err := UnixStringToTimestamp(attrs.TermsAcceptedDate[0]); err == nil {
			termsAcceptedDate = ts
		}
	}

	user := &FullUser{
		Id:            keycloakUser.ID,
		Username:      keycloakUser.Username,
		Emails:        []string{keycloakUser.Email},
		Roles:         keycloakUser.Roles,
		TermsAccepted: termsAcceptedDate,
		EmailVerified: keycloakUser.EmailVerified,
		IsMigrated:    true,
		Enabled:       keycloakUser.Enabled,
		Profile:       attrs.Profile,
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

func userToKeycloakUser(u *FullUser) *keycloakUser {
	keycloakUser := &keycloakUser{
		ID:            u.Id,
		Username:      strings.ToLower(u.Username),
		Email:         strings.ToLower(u.Email()),
		Enabled:       u.IsEnabled(),
		EmailVerified: u.EmailVerified,
		Roles:         u.Roles,
		Attributes:    keycloakUserAttributes{},
	}
	if len(keycloakUser.Roles) == 0 {
		keycloakUser.Roles = []string{RolePatient}
	}
	if !u.IsMigrated && u.PwHash == "" && !u.HasRole(RoleCustodialAccount) {
		keycloakUser.Roles = append(keycloakUser.Roles, RoleCustodialAccount)
	}
	if termsAccepted, err := TimestampToUnixString(u.TermsAccepted); err == nil {
		keycloakUser.Attributes.TermsAcceptedDate = []string{termsAccepted}
	}
	if u.Profile != nil {
		keycloakUser.Attributes.Profile = u.Profile
	}

	return keycloakUser
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
