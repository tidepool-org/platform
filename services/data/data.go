// @title Platform Data API
// @version 0.6.4
// @description The Tidepool API is an HTTP REST API served by Tidepool.
// @description It is the API that Tidepool clients use to communicate with the Tidepool Platform.
// @license.name BSD 2-Clause "Simplified" License
// @host localhost
// @BasePath /v1
// @accept json
// @produce json
// @schemes https

// @securityDefinitions.apikey TidepoolSessionToken
// @in header
// @name x-tidepool-session-token

// @securityDefinitions.apikey TidepoolServiceSecret
// @in header
// @name X-Tidepool-Service-Secret

// @securityDefinitions.apikey TidepoolAuthorization
// @in header
// @name Authorization

// @securityDefinitions.apikey TidepoolRestrictedToken
// @in header
// @name restricted_token

package main

import (
	"github.com/tidepool-org/platform/application"
	dataServiceService "github.com/tidepool-org/platform/data/service/service"

	// Automatically set GOMAXPROCS to match Linux container CPU quota.
	_ "go.uber.org/automaxprocs"
)

func main() {
	application.RunAndExit(dataServiceService.NewStandard(), "service")
}
