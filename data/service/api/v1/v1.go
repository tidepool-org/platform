package v1

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/golang-jwt/jwt/v4"

	abbottServiceApiV1 "github.com/tidepool-org/platform-plugin-abbott/abbott/service/api/v1"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/service"
	ouraServiceApiV1 "github.com/tidepool-org/platform/oura/service/api/v1"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service/api"
)

func Routes() []service.Route {
	routes := []service.Route{
		service.Post("/v1/datasets/:dataSetId/data", DataSetsDataCreate, api.RequireAuth),
		service.Delete("/v1/datasets/:dataSetId", DataSetsDelete, api.RequireAuth),
		service.Put("/v1/datasets/:dataSetId", DataSetsUpdate, api.RequireAuth),
		service.Delete("/v1/users/:userId/data", UsersDataDelete, api.RequireAuth),
		service.Post("/v1/users/:userId/datasets", UsersDataSetsCreate, api.RequireAuth),
		service.Get("/v1/users/:userId/datasets", UsersDataSetsGet, api.RequireAuth),

		service.Post("/v1/data_sets/:dataSetId/data", DataSetsDataCreate, api.RequireAuth),
		service.Delete("/v1/data_sets/:dataSetId/data", DataSetsDataDelete, api.RequireAuth),
		service.Delete("/v1/data_sets/:dataSetId", DataSetsDelete, api.RequireAuth),
		service.Put("/v1/data_sets/:dataSetId", DataSetsUpdate, api.RequireAuth),
		service.Get("/v1/time", TimeGet),
		service.Post("/v1/users/:userId/data_sets", UsersDataSetsCreate, api.RequireAuth),

		service.Get("/v1/partners/:partner/sector/:environment", PartnersSector),
		service.Get("/v1/partners/:partner/sector", PartnersSector), // DEPRECATED: Remove once Abbott sandbox client uses environment in path

		service.Post("/v1/partners/twiist/data/:tidepoolLinkId", NewTwiistDataCreateHandler(DataSetsDataCreate), api.RequireAuth),
	}

	routes = append(routes, DataSetsRoutes()...)
	routes = append(routes, SourcesRoutes()...)
	routes = append(routes, SummaryRoutes()...)
	routes = append(routes, AlertsRoutes()...)
	routes = append(routes, NotificationsRoutes()...)
	routes = append(routes, abbottServiceApiV1.Routes()...)
	routes = append(routes, ouraServiceApiV1.Routes()...)

	return routes
}

// Get provenance from a request and auth details
func GetProvenanceFromRequest(ctx context.Context, req *rest.Request, authDetails request.AuthDetails) *data.Provenance {
	provenance := &data.Provenance{}

	switch authDetails.Method() {
	case request.MethodAccessToken, request.MethodSessionToken:
		claims := &tokenClaims{}
		if _, _, err := jwt.NewParser().ParseUnverified(authDetails.Token(), claims); err == nil {
			provenance.ClientID = claims.AuthorizedParty
		}
	case request.MethodServiceSecret, request.MethodRestrictedToken:
	}

	if userID := authDetails.UserID(); userID != "" {
		provenance.ByUserID = pointer.From(userID)
	}

	if xff := SelectXFF(req.Header); xff != nil {
		provenance.SourceIP = xff
	} else if host, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		provenance.SourceIP = pointer.From(host)
	}

	return provenance
}

// Get first X-Forwarded-For header value that is not private/loopback and is a global unicast.
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For#selecting_an_ip_address
func SelectXFF(header http.Header) *string {
	for address := range strings.SplitSeq(strings.Join(header.Values("X-Forwarded-For"), ","), ",") {
		if address = strings.TrimSpace(address); address != "" {
			if ip := net.ParseIP(address); ip != nil && !ip.IsPrivate() && !ip.IsLoopback() && ip.IsGlobalUnicast() {
				return pointer.From(address)
			}
		}
	}
	return nil
}

type tokenClaims struct {
	*jwt.RegisteredClaims

	// Keycloak client id
	// https://openid.net/specs/openid-connect-core-1_0.html#IDToken
	AuthorizedParty string `json:"azp"`
}

var _ jwt.Claims = (*tokenClaims)(nil)
