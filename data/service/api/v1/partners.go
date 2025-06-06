package v1

import (
	"net/http"

	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/request"
)

func PartnersSector(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()
	responder := request.MustNewResponder(res, req)

	if environmentsSectorIdentifierURLs, ok := partnersEnvironmentsSectorIdentifierURLs[req.PathParam("partner")]; ok {
		if sectorIdentifierURLs, ok := environmentsSectorIdentifierURLs[req.PathParam("environment")]; ok {
			responder.Data(http.StatusOK, sectorIdentifierURLs)
			return
		}
	}

	responder.Error(http.StatusNotFound, errors.New("partner environment not found"))
}

var partnersEnvironmentsSectorIdentifierURLs = map[string]map[string][]string{
	"abbott": {
		"production": {
			"https://app.tidepool.org/v1/oauth/abbott/redirect",
			"https://app.tidepool.org/v1/oauth/abbott-private-1/redirect",
			"https://app.tidepool.org/v1/oauth/abbott-private-2/redirect",
			"https://external.integration.tidepool.org/v1/oauth/abbott/redirect",
			"https://external.integration.tidepool.org/v1/oauth/abbott-private-1/redirect",
			"https://external.integration.tidepool.org/v1/oauth/abbott-private-2/redirect",
			"https://qa1.development.tidepool.org/v1/oauth/abbott/redirect",
			"https://qa1.development.tidepool.org/v1/oauth/abbott-private-1/redirect",
			"https://qa1.development.tidepool.org/v1/oauth/abbott-private-2/redirect",
			"https://qa2.development.tidepool.org/v1/oauth/abbott/redirect",
			"https://qa2.development.tidepool.org/v1/oauth/abbott-private-1/redirect",
			"https://qa2.development.tidepool.org/v1/oauth/abbott-private-2/redirect",
			"https://qa3.development.tidepool.org/v1/oauth/abbott/redirect",
			"https://qa3.development.tidepool.org/v1/oauth/abbott-private-1/redirect",
			"https://qa3.development.tidepool.org/v1/oauth/abbott-private-2/redirect",
			"https://qa4.development.tidepool.org/v1/oauth/abbott/redirect",
			"https://qa4.development.tidepool.org/v1/oauth/abbott-private-1/redirect",
			"https://qa4.development.tidepool.org/v1/oauth/abbott-private-2/redirect",
			"https://qa5.development.tidepool.org/v1/oauth/abbott/redirect",
			"https://qa5.development.tidepool.org/v1/oauth/abbott-private-1/redirect",
			"https://qa5.development.tidepool.org/v1/oauth/abbott-private-2/redirect",
			"https://dev1.dev.tidepool.org/v1/oauth/abbott/redirect",
			"https://dev1.dev.tidepool.org/v1/oauth/abbott-private-1/redirect",
			"https://dev1.dev.tidepool.org/v1/oauth/abbott-private-2/redirect",
		},
		"sandbox": {
			"https://external.integration.tidepool.org/v1/oauth/abbott/redirect",
			"https://external.integration.tidepool.org/v1/oauth/abbott-private-1/redirect",
			"https://external.integration.tidepool.org/v1/oauth/abbott-private-2/redirect",
			"https://qa1.development.tidepool.org/v1/oauth/abbott/redirect",
			"https://qa1.development.tidepool.org/v1/oauth/abbott-private-1/redirect",
			"https://qa1.development.tidepool.org/v1/oauth/abbott-private-2/redirect",
			"https://qa2.development.tidepool.org/v1/oauth/abbott/redirect",
			"https://qa2.development.tidepool.org/v1/oauth/abbott-private-1/redirect",
			"https://qa2.development.tidepool.org/v1/oauth/abbott-private-2/redirect",
			"https://qa3.development.tidepool.org/v1/oauth/abbott/redirect",
			"https://qa3.development.tidepool.org/v1/oauth/abbott-private-1/redirect",
			"https://qa3.development.tidepool.org/v1/oauth/abbott-private-2/redirect",
			"https://qa4.development.tidepool.org/v1/oauth/abbott/redirect",
			"https://qa4.development.tidepool.org/v1/oauth/abbott-private-1/redirect",
			"https://qa4.development.tidepool.org/v1/oauth/abbott-private-2/redirect",
			"https://qa5.development.tidepool.org/v1/oauth/abbott/redirect",
			"https://qa5.development.tidepool.org/v1/oauth/abbott-private-1/redirect",
			"https://qa5.development.tidepool.org/v1/oauth/abbott-private-2/redirect",
			"https://dev1.dev.tidepool.org/v1/oauth/abbott/redirect",
			"https://dev1.dev.tidepool.org/v1/oauth/abbott-private-1/redirect",
			"https://dev1.dev.tidepool.org/v1/oauth/abbott-private-2/redirect",
		},
		"": { // DEPRECATED: Remove once Abbott sandbox client uses environment in path
			"https://external.integration.tidepool.org/v1/oauth/abbott/redirect",
			"https://external.integration.tidepool.org/v1/oauth/abbott-private-1/redirect",
			"https://external.integration.tidepool.org/v1/oauth/abbott-private-2/redirect",
			"https://qa1.development.tidepool.org/v1/oauth/abbott/redirect",
			"https://qa2.development.tidepool.org/v1/oauth/abbott/redirect",
			"https://qa3.development.tidepool.org/v1/oauth/abbott/redirect",
			"https://qa4.development.tidepool.org/v1/oauth/abbott/redirect",
			"https://qa5.development.tidepool.org/v1/oauth/abbott/redirect",
			"https://dev1.dev.tidepool.org/v1/oauth/abbott/redirect",
		},
	},
}
