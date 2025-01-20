package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/golang-jwt/jwt/v4"

	"github.com/tidepool-org/platform/data"
	dataRaw "github.com/tidepool-org/platform/data/raw"
	dataService "github.com/tidepool-org/platform/data/service"
	dataTypesFactory "github.com/tidepool-org/platform/data/types/factory"
	dataWorkIngest "github.com/tidepool-org/platform/data/work/ingest"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	bodyParseValidateLimit = 8 * 1024 * 1024
)

func DataSetsDataCreate(dataServiceContext dataService.Context) {
	req := dataServiceContext.Request()
	res := dataServiceContext.Response()
	ctx := req.Context()
	lgr := log.LoggerFromContext(ctx)
	responder := request.MustNewResponder(res, req)

	dataSetID := dataServiceContext.Request().PathParam("dataSetId")
	if dataSetID == "" {
		dataServiceContext.RespondWithError(ErrorDataSetIDMissing())
		return
	}

	dataSet, err := dataServiceContext.DataRepository().GetDataSet(ctx, dataSetID)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to get data set by id", err)
		return
	} else if dataSet == nil {
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

	// TODO: Switch to "Content-Digest" - support only sha-512, sha-256

	digestMD5, err := request.ParseDigestMD5Header(req.Header, "Digest")
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}
	mediaType, err := request.ParseMediaTypeHeader(req.Header, "Content-Type")
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(req.Body, bodyParseValidateLimit+1))
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	var body io.Reader = bytes.NewReader(bodyBytes)
	if len(bodyBytes) <= bodyParseValidateLimit {
		var array []any
		if err = json.Unmarshal(bodyBytes, &array); err != nil {
			dataServiceContext.RespondWithError(service.ErrorJSONMalformed())
			return
		} else if len(array) == 0 {
			dataServiceContext.RespondWithStatusAndData(http.StatusOK, []struct{}{})
			return
		}

		parser := structureParser.NewArray(lgr, &array)
		validator := structureValidator.New(lgr)

		for _, reference := range parser.References() {
			if datum := dataTypesFactory.ParseDatum(parser.WithReferenceObjectParser(reference)); datum != nil && *datum != nil {
				(*datum).Validate(validator.WithReference(strconv.Itoa(reference)))
			}
		}
		parser.NotParsed()

		if err = errors.Append(parser.Error(), validator.Error()); err != nil {
			responder.Error(http.StatusBadRequest, err)
			return
		}
	} else {
		body = io.MultiReader(body, req.Body)
	}

	create := dataRaw.NewCreate()
	create.DigestMD5 = digestMD5
	create.MediaType = mediaType
	create.Metadata = metadata.NewMetadata()
	create.Metadata.Set("provenance", CollectProvenanceInfo(ctx, req, authDetails))

	raw, err := dataServiceContext.DataRawClient().Create(ctx, *dataSet.UserID, *dataSet.ID, create, body)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to create raw data", err)
		return
	}
	req.Body.Close()

	if dataSet.HasDataSetTypeContinuous() {
		create, err := dataWorkIngest.NewCreateForDataSetTypeContinuous(dataSet, raw)
		if err != nil {
			dataServiceContext.RespondWithInternalServerFailure("Unable to create work create", err)
			return
		}
		work, err := dataServiceContext.WorkClient().Create(ctx, create)
		if err != nil {
			dataServiceContext.RespondWithInternalServerFailure("Unable to create work", err)
			return
		}
		lgr.WithFields(log.Fields{"rawId": raw.ID, "workId": work.ID}).Debug("Created work for raw")
	}

	if err = dataServiceContext.MetricClient().RecordMetric(ctx, "data_sets_data_create"); err != nil {
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

	// Allow for legacy Shoreline tokens
	if parts := strings.Split(token, ":"); len(parts) == 3 && parts[0] == "kc" {
		token = parts[1]
	}

	if token != "" && shouldHaveJWT(authDetails) {
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

	if userID := authDetails.UserID(); userID != "" {
		provenance.ByUserID = userID
	} else if shouldHaveJWT(authDetails) && !authDetails.IsService() {
		lgr.Warnf("Unable to read the request's userID for provenance: userID is empty")
	}

	return provenance
}

// shouldHaveJWT indicates if it is expected that this token is a JWT.
//
// Of the current authentication methods, three of the four provide token
// information, but only two of those three, use a JSON Web Token (JWT).
func shouldHaveJWT(authDetails request.AuthDetails) bool {
	switch authDetails.Method() {
	case request.MethodAccessToken:
		return true
	case request.MethodSessionToken:
		return true
	}
	return false
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
