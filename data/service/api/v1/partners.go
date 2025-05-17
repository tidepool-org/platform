package v1

import (
	"net/http"
	"os"

	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/request"
)

// TODO: BACK-3394 - Restrict partner OpenID sector information to environment variable and authorized access only
// This implementation is a temporary placeholder to allow bootstrapping of the Abbott OAuth client workflow.
// Will need to migrate this to environment variables and add minimal authorization. For now, though, this is
// acceptable since it isn't revealing anything that is not already available in other locations (i.e. other
// public repos).

func PartnersSector(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()
	responder := request.MustNewResponder(res, req)

	if partnerSectorIdentifiers, ok := namespacePartnerSectorIdentifiers[os.Getenv("POD_NAMESPACE")]; ok {
		if sectorIdentifier, ok := partnerSectorIdentifiers[req.PathParam("partner")]; ok {
			responder.Data(http.StatusOK, sectorIdentifier)
			return
		}
	}

	responder.Data(http.StatusOK, []string{})
}

var namespacePartnerSectorIdentifiers = map[string]map[string][]string{
	"external": {
		"abbott": {
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
