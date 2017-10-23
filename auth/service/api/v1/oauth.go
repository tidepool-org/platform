package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/oauth"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/provider"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service/api"
)

func (r *Router) OAuthRoutes() []*rest.Route {
	return []*rest.Route{
		rest.Get("/v1/oauth/:name/authorize", r.OAuthProviderAuthorizeGet),
		rest.Delete("/v1/oauth/:name/authorize", api.RequireUser(r.OAuthProviderAuthorizeDelete)),
		rest.Get("/v1/oauth/:name/redirect", r.OAuthProviderRedirectGet),
	}
}

func (r *Router) OAuthProviderAuthorizeGet(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	ctx := req.Context()
	details := request.DetailsFromContext(ctx)

	if details == nil || details.Method() != request.MethodRestrictedToken {
		r.htmlOnError(res, req, request.ErrorUnauthenticated())
		return
	}

	prvdr, err := r.oauthProvider(req)
	if err != nil {
		r.htmlOnError(res, req, err)
		return
	}

	restrictedToken, err := r.AuthClient().GetRestrictedToken(ctx, details.Token())
	if err != nil {
		r.htmlOnError(res, req, err)
		return
	}

	maxAge := restrictedToken.ExpirationTime.Sub(time.Now()) / time.Second
	if maxAge <= 0 {
		r.htmlOnError(res, req, request.ErrorUnauthenticated())
		return
	}

	responder.SetCookie(r.providerCookie(prvdr, details.Token(), int(maxAge)))
	responder.Redirect(http.StatusTemporaryRedirect, prvdr.Config().AuthCodeURL(prvdr.State(details.Token())))
}

func (r *Router) OAuthProviderAuthorizeDelete(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	ctx := req.Context()
	details := request.DetailsFromContext(ctx)

	prvdr, err := r.oauthProvider(req)
	if err != nil {
		responder.Error(request.StatusCodeForError(err), err)
		return
	}

	providerSessionFilter := auth.NewProviderSessionFilter()
	providerSessionFilter.Type = pointer.String(prvdr.Type())
	providerSessionFilter.Name = pointer.String(prvdr.Name())
	providerSessions, err := r.AuthClient().ListUserProviderSessions(ctx, details.UserID(), providerSessionFilter, page.NewPagination())
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	if len(providerSessions) > 1 {
		r.Logger().WithFields(log.Fields{"userId": details.UserID(), "filter": providerSessionFilter, "providerSessions": providerSessions}).Warn("Deleting multiple provider sessions")
	}

	for _, providerSession := range providerSessions {
		if err = r.AuthClient().DeleteProviderSession(ctx, providerSession.ID); err != nil {
			responder.Error(http.StatusInternalServerError, err)
			return
		}
	}

	responder.Empty(http.StatusOK)
}

func (r *Router) OAuthProviderRedirectGet(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	ctx := req.Context()
	query := req.URL.Query()

	prvdr, err := r.oauthProvider(req)
	if err != nil {
		r.htmlOnError(res, req, err)
		return
	}

	restrictedToken, err := r.oauthProviderRestrictedToken(req.Request, prvdr)
	if err != nil {
		r.htmlOnError(res, req, err)
		return
	}

	responder.SetCookie(r.providerCookie(prvdr, restrictedToken.ID, -1))

	if err = r.AuthClient().DeleteRestrictedToken(ctx, restrictedToken.ID); err != nil {
		log.LoggerFromContext(ctx).WithError(err).Error("unable to delete restricted token after oauth redirect")
	}

	if errorCode := query.Get("error"); errorCode == oauth.ErrorAccessDenied {
		r.htmlOnRedirect(res, req)
		return
	} else if errorCode != "" {
		r.htmlOnError(res, req, errors.Newf("oauth provider return unexpected error %q", errorCode))
		return
	}

	filter := auth.NewProviderSessionFilter()
	filter.Type = pointer.String(prvdr.Type())
	filter.Name = pointer.String(prvdr.Name())
	providerSessions, err := r.AuthClient().ListUserProviderSessions(ctx, restrictedToken.UserID, filter, nil)
	if err != nil {
		r.htmlOnError(res, req, err)
		return
	} else if len(providerSessions) > 0 {
		r.htmlOnError(res, req, errors.Newf("provider session already exists for user, type, and name"))
		return
	}

	token, err := prvdr.Config().Exchange(ctx, query.Get("code"))
	if err != nil {
		r.htmlOnError(res, req, err)
		return
	}

	// HACK: Dexcom - expires_in=600000 (should be 600) - force to immediate expiration
	token.Expiry = time.Now()

	oauthToken, err := oauth.NewTokenFromRawToken(token)
	if err != nil {
		r.htmlOnError(res, req, err)
		return
	}

	providerSessionCreate := auth.NewProviderSessionCreate()
	providerSessionCreate.Type = prvdr.Type()
	providerSessionCreate.Name = prvdr.Name()
	providerSessionCreate.OAuthToken = oauthToken
	_, err = r.AuthClient().CreateUserProviderSession(ctx, restrictedToken.UserID, providerSessionCreate)
	if err != nil {
		r.htmlOnError(res, req, err)
		return
	}

	r.htmlOnRedirect(res, req)
}

func (r *Router) oauthProvider(req *rest.Request) (oauth.Provider, error) {
	name := req.PathParams["name"]
	if name == "" {
		return nil, request.ErrorParameterMissing("name")
	}

	prvdr, err := r.ProviderFactory().Get(auth.ProviderTypeOAuth, name)
	if err != nil {
		return nil, request.ErrorResourceNotFoundWithID(name)
	}
	oauthProvider, ok := prvdr.(oauth.Provider)
	if !ok {
		return nil, request.ErrorResourceNotFoundWithID(name)
	}

	return oauthProvider, nil
}

func (r *Router) oauthProviderRestrictedToken(req *http.Request, prvdr oauth.Provider) (*auth.RestrictedToken, error) {
	state := req.URL.Query().Get("state")
	cookieName := r.providerCookieName(prvdr)
	for _, cookie := range req.Cookies() {
		if cookie.Name == cookieName {
			if restrictedToken, err := r.AuthClient().GetRestrictedToken(req.Context(), cookie.Value); err != nil {
				return nil, err
			} else if restrictedToken != nil && restrictedToken.Authenticates(req) && state == prvdr.State(restrictedToken.ID) {
				return restrictedToken, nil
			}
		}
	}
	return nil, request.ErrorUnauthenticated()
}

func (r *Router) htmlOnRedirect(res rest.ResponseWriter, req *rest.Request) {
	request.MustNewResponder(res, req).HTML(http.StatusOK, htmlOnRedirect)
}

func (r *Router) htmlOnError(res rest.ResponseWriter, req *rest.Request, err error) {
	log.LoggerFromContext(req.Context()).WithError(err).Error("Unexpected failure during OAuth workflow")
	request.MustNewResponder(res, req).HTML(request.StatusCodeForError(err), htmlOnError)
}

func (r *Router) providerCookie(prvdr provider.Provider, value string, maxAge int) *http.Cookie {
	name := r.providerCookieName(prvdr)
	path := r.providerCookiePath(prvdr)
	domain := r.Domain()
	secure := (domain != "localhost")

	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		Domain:   domain,
		MaxAge:   maxAge,
		Secure:   secure,
		HttpOnly: true,
	}

	return cookie
}

func (r *Router) providerCookieName(prvdr provider.Provider) string {
	return fmt.Sprintf("org.tidepool.provider.%s.%s", prvdr.Type(), prvdr.Name())
}

func (r *Router) providerCookiePath(prvdr provider.Provider) string {
	return fmt.Sprintf("/v1/%s/%s", prvdr.Type(), prvdr.Name())
}

// TODO: Improve HTML
const htmlOnRedirect = `<html><head/><body onLoad="window.close();"/></html>`
const htmlOnError = `<html><head/><body>An expected error occurred. Please dismiss window and try again.</body></html>`
