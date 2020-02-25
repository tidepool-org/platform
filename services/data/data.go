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

// @securityDefinitions.apikey TidepoolAuth
// @in header
// @name x-tidepool-session-token
package main

import (
	"github.com/tidepool-org/platform/application"
	dataServiceService "github.com/tidepool-org/platform/data/service/service"
)

func main() {
	application.RunAndExit(dataServiceService.NewStandard(), "service")
}
