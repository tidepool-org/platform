package v1

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/ant0ine/go-json-rest/rest"

	confirmationClient "github.com/tidepool-org/hydrophone/client"

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
	details := request.GetAuthDetails(ctx)

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
	responder.Redirect(http.StatusTemporaryRedirect, prvdr.GetAuthorizationCodeURLWithState(prvdr.CalculateStateForRestrictedToken(details.Token())))
}

func (r *Router) OAuthProviderAuthorizeDelete(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	ctx := req.Context()
	details := request.GetAuthDetails(ctx)

	prvdr, err := r.oauthProvider(req)
	if err != nil {
		responder.Error(request.StatusCodeForError(err), err)
		return
	}

	providerSessionFilter := auth.NewProviderSessionFilter()
	providerSessionFilter.Type = pointer.FromString(prvdr.Type())
	providerSessionFilter.Name = pointer.FromString(prvdr.Name())
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

	redirectURLAuthorized := req.BaseUrl()
	redirectURLAuthorized.Path = path.Join(redirectURLAuthorized.Path, prvdr.Type(), prvdr.Name(), "authorized")

	redirectURLDeclined := req.BaseUrl()
	redirectURLDeclined.Path = path.Join(redirectURLDeclined.Path, prvdr.Type(), prvdr.Name(), "declined")

	restrictedToken, err := r.oauthProviderRestrictedToken(req.Request, prvdr)
	if err != nil {
		r.htmlOnError(res, req, err)
		return
	}

	// Include custodial account signup credentials in redirect URL query, if applicable
	confirmation, err := r.ConfirmationClient().GetAccountSignupConfirmationWithResponse(ctx, confirmationClient.UserId(restrictedToken.UserID))
	if err != nil {
		r.htmlOnError(res, req, err)
		return
	}

	signupParams := url.Values{}
	if confirmation.JSON200 != nil {
		if confirmation.JSON200.Email != "" {
			signupParams.Add("signupEmail", string(confirmation.JSON200.Email))
		}
		if confirmation.JSON200.Key != "" {
			signupParams.Add("signupKey", string(confirmation.JSON200.Key))
		}
	}

	if len(signupParams) > 0 {
		redirectURLAuthorized.RawQuery = signupParams.Encode()
		redirectURLDeclined.RawQuery = signupParams.Encode()
	}

	responder.SetCookie(r.providerCookie(prvdr, restrictedToken.ID, -1))

	if err = r.AuthClient().DeleteRestrictedToken(ctx, restrictedToken.ID); err != nil {
		log.LoggerFromContext(ctx).WithError(err).Error("unable to delete restricted token after oauth redirect")
	}

	if errorCode := query.Get("error"); errorCode == oauth.ErrorAccessDenied {
		html := fmt.Sprintf(htmlOnRedirect, redirectURLDeclined.String())
		r.htmlOnRedirect(res, req, html)
		return
	} else if errorCode != "" {
		r.htmlOnError(res, req, errors.Newf("oauth provider return unexpected error %q", errorCode))
		return
	}

	filter := auth.NewProviderSessionFilter()
	filter.Type = pointer.FromString(prvdr.Type())
	filter.Name = pointer.FromString(prvdr.Name())
	providerSessions, err := r.AuthClient().ListUserProviderSessions(ctx, restrictedToken.UserID, filter, nil)
	if err != nil {
		r.htmlOnError(res, req, err)
		return
	} else if len(providerSessions) > 0 {
		// Delete existing provider sessions and tasks if matching name and type found for user.
		// This operation will also reset the data source to a `disconnected` state, and remove any associated tasks
		// A new provider session and task will be created below which will update the existing data source state to `connected`.
		for _, session := range providerSessions {
			if deleteSessionErr := r.AuthClient().DeleteProviderSession(ctx, session.ID); deleteSessionErr != nil {
				r.htmlOnError(res, req, errors.Newf("could not remove existing provider session"), alreadyConnectedError)
				return
			}
		}
	}

	oauthToken, err := prvdr.ExchangeAuthorizationCodeForToken(ctx, query.Get("code"))
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

	html := fmt.Sprintf(htmlOnRedirect, redirectURLAuthorized.String())
	r.htmlOnRedirect(res, req, html)
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
	errorCode := req.URL.Query().Get("error")
	cookieName := r.providerCookieName(prvdr)
	for _, cookie := range req.Cookies() {
		if cookie.Name == cookieName {
			if restrictedToken, err := r.AuthClient().GetRestrictedToken(req.Context(), cookie.Value); err != nil {
				return nil, err
			} else if restrictedToken != nil && restrictedToken.Authenticates(req) && (errorCode == oauth.ErrorAccessDenied || state == prvdr.CalculateStateForRestrictedToken(restrictedToken.ID)) {
				return restrictedToken, nil
			}
		}
	}
	return nil, request.ErrorUnauthenticated()
}

func (r *Router) htmlOnRedirect(res rest.ResponseWriter, req *rest.Request, html string) {
	request.MustNewResponder(res, req).String(http.StatusOK, html, request.NewHeaderMutator("Content-Type", "text/html"))
}

func (r *Router) htmlOnError(res rest.ResponseWriter, req *rest.Request, err error, messages ...string) {
	if len(messages) == 0 {
		messages = append(messages, unexpectedError)
	}
	log.LoggerFromContext(req.Context()).WithError(err).WithField("messages", messages).Error("Unexpected failure during OAuth workflow")
	request.MustNewResponder(res, req).String(
		request.StatusCodeForError(err),
		strings.Replace(htmlOnError, "{{ MESSAGES }}", strings.Join(messages, " "), -1),
		request.NewHeaderMutator("Content-Type", "text/html"),
	)
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

const htmlOnRedirect = `
<html>
	<body onload="closeOrRedirect()">
		<script>
			function closeOrRedirect() {
				var isIframe = window.location !== window.parent.location;
				if (isIframe) {
					window.close();
				} else {
					window.location.replace('%s');
				}
			}
		</script>
	</body>
</html>
`

const htmlOnError = `
<!doctype html>
<html lang="" style="-ms-text-size-adjust: 100%; -webkit-text-size-adjust: 100%; color: #222; font-size: 1em; height: 100%; line-height: 1.4;">
<head>
    <meta charset="utf-8">
    <meta http-equiv="x-ua-compatible" content="ie=edge">
    <title>Error</title>
    <meta name="description" content="">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://fonts.googleapis.com/css?family=Open+Sans" rel="stylesheet">
</head>
<body style="background-color: #f7f7f8; display: flex; height: 100%; justify-content: center; margin: 0;">
    <div class="error-pane" style="align-self: center; display: flex; flex-direction: column; justify-content: center; max-width: 440px;">
        <div style="display:flex;flex-direction:row;">
            <div class="alert-icon" style="margin-right: 20px;">
                <svg xmlns="http://www.w3.org/2000/svg" fill="#ea3324" height="40" viewBox="0 0 24 24" width="40" style="overflow: hidden; vertical-align: middle;">
                    <path d="M0 0h24v24H0z" fill="none" />
                    <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z" />
                </svg>
            </div>
            <div class="Error-Description" style="color: #6d6d6d; font-family: Open Sans; font-size: 20px; font-weight: 600; letter-spacing: -0.1px; text-align: left;">
                {{ MESSAGES }}
            </div>
        </div>
        <div style="display:flex;flex-direction:column;align-items: center;">
            <button class="Button-Text" onclick="window.close()" style="-webkit-appearance: button; background-color: #6d6d6d; border: none; border-radius: 8px; color: #ffffff; font-family: Open Sans; font-size: 16px; font-weight: bold; height: 40px; line-height: 1.15; margin: 0; margin-top: 40px; overflow: visible; text-align: center; text-transform: none; width: 184px;">Close this window</button>
        </div>
    </div>
    <div class="logo" style="bottom: 40px; left: 40px; position: absolute;">
        <svg xmlns="http://www.w3.org/2000/svg" width="150" height="17" viewBox="0 0 150 17" style="overflow: hidden; vertical-align: middle;">
            <g fill="none" fill-rule="evenodd">
                <path fill="#F7F7F8" d="M-40-783h1080V57H-40z" />
                <g fill="#6D6D6D">
                    <path d="M4.749.41a1.17 1.17 0 1 0 1.17 1.17A1.17 1.17 0 0 0 4.749.41zm-3.457 0a1.17 1.17 0 1 0 0 2.341 1.17 1.17 0 0 0 0-2.34zm6.913 10.365a1.168 1.168 0 1 0 0 2.338 1.17 1.17 0 0 0 0-2.337zm0 3.452a1.169 1.169 0 1 0 1.17 1.166 1.17 1.17 0 0 0-1.17-1.166zM24.582.41a1.17 1.17 0 1 0 1.167 1.17A1.17 1.17 0 0 0 24.582.41zm0 3.455c-.649 0-1.173.525-1.173 1.17a1.171 1.171 0 0 0 2.34 0 1.17 1.17 0 0 0-1.167-1.17zM59.948.41a1.17 1.17 0 1 0 0 2.341 1.17 1.17 0 0 0 0-2.34zm3.457 0a1.17 1.17 0 1 0 0 2.34 1.17 1.17 0 0 0 0-2.34zm0 13.817a1.17 1.17 0 1 0 0 2.34 1.17 1.17 0 0 0 0-2.34zm3.457 0a1.174 1.174 0 0 0-1.17 1.172c0 .305.124.609.343.824.218.218.518.342.827.342a1.167 1.167 0 1 0 0-2.338zm0-13.817a1.17 1.17 0 1 0 0 2.341 1.17 1.17 0 0 0 0-2.34zM56.491 14.23a1.17 1.17 0 1 0 0 2.34 1.17 1.17 0 0 0 0-2.34zm3.457-.003a1.17 1.17 0 1 0 0 2.34 1.17 1.17 0 0 0 0-2.34zm-3.457-3.454a1.17 1.17 0 1 0-.001 2.338 1.17 1.17 0 0 0 0-2.338zm92.174 3.454a1.17 1.17 0 1 0 0 2.34 1.17 1.17 0 0 0 0-2.34zM138.295.41a1.17 1.17 0 1 0-.001 2.34 1.17 1.17 0 0 0 0-2.34zm0 3.455a1.17 1.17 0 1 0 0 0zm8.083 11.531a1.17 1.17 0 1 0-2.341.002 1.17 1.17 0 0 0 2.34-.002zM125.796 1.128a1.17 1.17 0 0 0-.967 2.129h.002a1.172 1.172 0 0 0 1.548-.583 1.167 1.167 0 0 0-.583-1.546zm-8.736 11.824a1.166 1.166 0 0 0 .883 1.933 1.17 1.17 0 1 0-.883-1.933zm-.106-2.835a1.17 1.17 0 1 0 0 0zm4.698 4.053h-.003a1.17 1.17 0 1 0-.327 2.317 1.17 1.17 0 0 0 1.323-.995 1.171 1.171 0 0 0-.993-1.322zm-.193-11.35a1.168 1.168 0 1 0-.168-2.326 1.168 1.168 0 0 0-.99 1.325 1.167 1.167 0 0 0 1.158 1.002zM117.92 4.45a1.17 1.17 0 1 0-.767-2.052 1.167 1.167 0 0 0-.112 1.65c.23.264.555.402.88.402zm-2.423 3.23a1.172 1.172 0 0 0 1.45-.796 1.17 1.17 0 1 0-1.45.796zm-30.903-.361a1.17 1.17 0 1 0 0 2.34 1.17 1.17 0 0 0 0-2.34zm-7.052-3.453a1.17 1.17 0 1 0 0 0zm0-3.455a1.17 1.17 0 1 0 0 0zm7.057 0a1.17 1.17 0 1 0 0 2.341 1.17 1.17 0 0 0 0-2.34zm2.875 5.145a1.17 1.17 0 1 0 0 2.34 1.17 1.17 0 0 0 0-2.34zm0-3.374a1.17 1.17 0 1 0 .001 2.342 1.17 1.17 0 0 0-.001-2.342zm19.826 3.72a1.168 1.168 0 0 0 .98-1.809 1.172 1.172 0 0 0-1.62-.343 1.172 1.172 0 0 0 .64 2.152zm-3.413 7.808a1.168 1.168 0 1 0 .973 2.127 1.17 1.17 0 1 0-.973-2.127zm4.062-2.473a1.174 1.174 0 0 0-1.62.353 1.169 1.169 0 0 0 .983 1.802c.387 0 .762-.19.987-.536a1.174 1.174 0 0 0-.35-1.62zm-7.263 2.934a1.171 1.171 0 0 0-.328 2.317 1.173 1.173 0 0 0 1.323-.995 1.17 1.17 0 0 0-.995-1.322zm-4.593-1.218a1.165 1.165 0 0 0 .883 1.933 1.167 1.167 0 0 0 .767-2.051 1.17 1.17 0 0 0-1.65.118zm12.311-5.663a1.174 1.174 0 0 0-1.166 1.175c0 .007 0 .015.003.022h-.003c0 .649.524 1.17 1.17 1.17a1.17 1.17 0 0 0 1.173-1.17v-.03a1.173 1.173 0 0 0-1.177-1.167zm-12.417 2.828a1.171 1.171 0 0 0-1.454-.792 1.172 1.172 0 0 0 .665 2.245c.62-.184.973-.833.789-1.453zm7.877-6.86a1.168 1.168 0 1 0-.583-1.547 1.17 1.17 0 0 0 .583 1.547zM36.524.41c-.647 0-1.168.524-1.168 1.17A1.17 1.17 0 1 0 36.523.41zm3.574 0c-.646 0-1.167.524-1.167 1.17A1.17 1.17 0 1 0 40.098.41zm3.408.964c-.646 0-1.167.524-1.167 1.17a1.17 1.17 0 1 0 1.167-1.17zm2.478 2.52c-.645 0-1.166.525-1.166 1.17a1.17 1.17 0 1 0 1.167-1.17zm-9.46-.029c-.647 0-1.168.525-1.168 1.17a1.17 1.17 0 1 0 1.167-1.17zm0 10.365c-.647 0-1.168.525-1.168 1.169a1.17 1.17 0 0 0 2.34 0c0-.645-.524-1.17-1.173-1.17zm0-3.457a1.17 1.17 0 0 0-1.168 1.172 1.17 1.17 0 1 0 1.167-1.172zm0-3.453c-.647 0-1.168.521-1.168 1.17a1.17 1.17 0 1 0 1.167-1.169zM81.073.41a1.17 1.17 0 1 0 0 2.341 1.17 1.17 0 0 0 0-2.34zM46.894 7.31h-.012c-.604.006-1.087.5-1.081 1.102v.076c0 3.179-2.586 5.788-5.766 5.817a1.092 1.092 0 0 0 .01 2.183h.01c4.372-.04 7.93-3.629 7.93-8V8.39a1.091 1.091 0 0 0-1.091-1.08zM63.483 7.396h-5.9V1.502a1.092 1.092 0 0 0-2.183 0v6.985c0 .603.489 1.092 1.091 1.092h6.992a1.092 1.092 0 1 0 0-2.183zM15.118.488H8.205c-.602 0-1.091.49-1.091 1.091v6.903a1.092 1.092 0 0 0 2.183 0V2.671h5.82a1.092 1.092 0 1 0 0-2.183zM24.581 7.318c-.603 0-1.092.489-1.092 1.092v7.064a1.092 1.092 0 0 0 2.184 0V8.41c0-.603-.49-1.092-1.092-1.092zM81.187 7.396h-3.646c-.604 0-1.092.489-1.092 1.091v6.987a1.091 1.091 0 1 0 2.184 0V9.579h2.555a1.092 1.092 0 1 0 0-2.183zM141.829 14.306h-2.442V8.41a1.092 1.092 0 0 0-2.184 0v6.987c0 .603.489 1.092 1.092 1.092h3.534a1.092 1.092 0 1 0 0-2.183zM100.382.563a8.014 8.014 0 0 0-6.555 5.633 1.091 1.091 0 1 0 2.092.623 5.826 5.826 0 0 1 4.765-4.094 1.092 1.092 0 0 0-.302-2.162zM127.758 3.916c-.513.319-.67.991-.35 1.503a5.818 5.818 0 0 1-2.62 8.404 1.092 1.092 0 1 0 .874 2.002 8.004 8.004 0 0 0 3.6-11.56 1.093 1.093 0 0 0-1.504-.35z"
                    />
                </g>
            </g>
        </svg>
    </div>
</body>
</html>
`
const unexpectedError = `Looks like an unexpected error occurred. You can try again, or send an email to support@tidepool.org for help.`
const alreadyConnectedError = `This Tidepool account has already been connected to a Dexcom account. If this doesn't sound right, please send an email to support@tidepool.org and we'll help you out.`
