package v1

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/golang-jwt/jwt/v4"

	"github.com/tidepool-org/platform/data"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/data/summary"
	"github.com/tidepool-org/platform/data/summary/types"
	dataTypesFactory "github.com/tidepool-org/platform/data/types/factory"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func DataSetsDataCreate(dataServiceContext dataService.Context) {
	req := dataServiceContext.Request()
	ctx := req.Context()
	lgr := log.LoggerFromContext(ctx)

	dataSetID := dataServiceContext.Request().PathParam("dataSetId")
	if dataSetID == "" {
		dataServiceContext.RespondWithError(ErrorDataSetIDMissing())
		return
	}

	dataSet, err := dataServiceContext.DataRepository().GetDataSetByID(ctx, dataSetID)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to get data set by id", err)
		return
	}
	if dataSet == nil {
		dataServiceContext.RespondWithError(ErrorDataSetIDNotFound(dataSetID))
		return
	}

	var authDetails request.AuthDetails
	if authDetails = request.GetAuthDetails(ctx); !authDetails.IsService() {
		var permissions permission.Permissions
		permissions, err = dataServiceContext.PermissionClient().GetUserPermissions(ctx, authDetails.UserID(), *dataSet.UserID)
		if err != nil {
			if request.IsErrorUnauthorized(err) {
				dataServiceContext.RespondWithError(service.ErrorUnauthorized())
			} else {
				dataServiceContext.RespondWithInternalServerFailure("Unable to get user permissions", err)
			}
			return
		}
		if _, ok := permissions[permission.Write]; !ok {
			dataServiceContext.RespondWithError(service.ErrorUnauthorized())
			return
		}
	}

	if (dataSet.State != nil && *dataSet.State == "closed") || (dataSet.DataState != nil && *dataSet.DataState == "closed") { // TODO: Deprecated DataState (after data migration)
		dataServiceContext.RespondWithError(ErrorDataSetClosed(dataSetID))
		return
	}

	var rawDatumArray []interface{}
	if err = dataServiceContext.Request().DecodeJsonPayload(&rawDatumArray); err != nil {
		dataServiceContext.RespondWithError(service.ErrorJSONMalformed())
		return
	}

	parser := structureParser.NewArray(&rawDatumArray)
	validator := structureValidator.New()
	normalizer := dataNormalizer.New()

	datumArray := []data.Datum{}
	for _, reference := range parser.References() {
		if datum := dataTypesFactory.ParseDatum(parser.WithReferenceObjectParser(reference)); datum != nil && *datum != nil {
			(*datum).Validate(validator.WithReference(strconv.Itoa(reference)))
			(*datum).Normalize(normalizer.WithReference(strconv.Itoa(reference)))
			datumArray = append(datumArray, *datum)
		}
	}
	parser.NotParsed()

	if err = parser.Error(); err != nil {
		request.MustNewResponder(dataServiceContext.Response(), dataServiceContext.Request()).Error(http.StatusBadRequest, err)
		return
	}
	if err = validator.Error(); err != nil {
		request.MustNewResponder(dataServiceContext.Response(), dataServiceContext.Request()).Error(http.StatusBadRequest, err)
		return
	}
	if err = normalizer.Error(); err != nil {
		request.MustNewResponder(dataServiceContext.Response(), dataServiceContext.Request()).Error(http.StatusBadRequest, err)
		return
	}

	datumArray = append(datumArray, normalizer.Data()...)
	for _, datum := range datumArray {
		datum.SetUserID(dataSet.UserID)
		datum.SetDataSetID(dataSet.UploadID)
		datum.SetProvenance(CollectProvenanceInfo(ctx, req, authDetails))
	}

	if deduplicator, getErr := dataServiceContext.DataDeduplicatorFactory().Get(dataSet); getErr != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to get deduplicator", getErr)
		return
	} else if deduplicator == nil {
		dataServiceContext.RespondWithInternalServerFailure("Deduplicator not found")
		return
	} else if err = deduplicator.AddData(ctx, dataServiceContext.DataRepository(), dataSet, datumArray); err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to add data", err)
		return
	}

	updatesSummary := make(map[string]struct{})
	for _, datum := range datumArray {
		summary.CheckDatumUpdatesSummary(updatesSummary, datum)
	}
	summary.MaybeUpdateSummary(ctx, dataServiceContext.SummarizerRegistry(), updatesSummary, *dataSet.UserID, types.OutdatedReasonDataAdded)

	if err = dataServiceContext.MetricClient().RecordMetric(ctx, "data_sets_data_create", map[string]string{"count": strconv.Itoa(len(datumArray))}); err != nil {
		lgr.WithError(err).Error("Unable to record metric")
	}

	dataServiceContext.RespondWithStatusAndData(http.StatusOK, []struct{}{})
}

// CollectProvenanceInfo from a request and its auth details.
//
// All work is done as a best effort right now.
func CollectProvenanceInfo(ctx context.Context, req *rest.Request, authDetails request.AuthDetails) *data.Provenance {
	lgr := log.LoggerFromContext(ctx)
	provenance := &data.Provenance{}

	token := authDetails.Token()
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = token[len("bearer "):]
	}

	if token != "" {
		claims := &TokenClaims{}
		if _, _, err := jwt.NewParser().ParseUnverified(token, claims); err != nil {
			lgr.WithError(err).Warn("Unable to parse access token for provenance")
		} else {
			provenance.ClientID = claims.ClientID
		}
	} else if !authDetails.IsService() {
		lgr.Warn("Unable to read ClientID: The request's access token is blank")
	}

	if xff := SelectXFF(req.Header); xff != "" {
		provenance.SourceIP = xff
	} else {
		if host, _, err := net.SplitHostPort(req.RemoteAddr); err != nil {
			lgr.WithError(err).Warnf("Unable to read SourceIP from request for provenance")
		} else {
			provenance.SourceIP = host
		}
	}

	if userID := authDetails.UserID(); userID == "" {
		lgr.Warnf("Unable to read the request's userID for provenance: userID is empty")
	} else {
		provenance.ByUserID = userID
	}

	return provenance
}

// SelectXFF is the first public IP from the X-Forwarded-For request header.
//
// If no suitable IPs are found, the empty string is returned.
//
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For#selecting_an_ip_address
func SelectXFF(header http.Header) string {
	all := []string{}
	for _, h := range header.Values("X-Forwarded-For") {
		all = append(all, strings.Split(h, ",")...)
	}
	for _, rawAddr := range all {
		addr := strings.TrimSpace(rawAddr)
		if ip := net.ParseIP(addr); ip != nil {
			if !ip.IsPrivate() && !ip.IsLoopback() && ip.IsGlobalUnicast() {
				return addr
			}
		}
	}
	return ""
}

// TokenClaims retrieves claims of interest in a JWT access token.
type TokenClaims struct {
	*jwt.RegisteredClaims

	// ClientID in the "azp" claim for the "Authorized Party".
	//
	// If coming from Keycloak, this will be the Keycloak client
	// id. https://openid.net/specs/openid-connect-core-1_0.html#IDToken
	ClientID string `json:"azp"`
}

var _ jwt.Claims = (*TokenClaims)(nil)
