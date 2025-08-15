package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/urfave/cli"

	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/services/tools/tapi/api"
)

var environmentEndpointMap = map[string]string{
	"prd":   "https://api.tidepool.org",
	"int":   "https://int-api.tidepool.org",
	"stg":   "https://stg-api.tidepool.org",
	"dev":   "https://dev-api.tidepool.org",
	"local": "http://localhost:3000",
	"tilt":  "http://gateway-proxy",
}

var _API *api.API

func InitializeApplication() (*cli.App, error) {
	versionReporter, err := application.NewVersionReporter()
	if err != nil {
		return nil, err
	}

	app := cli.NewApp()
	app.Usage = "Command-line interface to interact with the Tidepool API"
	app.Version = versionReporter.Long()
	app.Authors = []cli.Author{{Name: "Darin Krauss", Email: "darin@tidepool.org"}}
	app.Copyright = "Copyright \u00A9 2016, Tidepool Project"
	app.HideVersion = true
	app.Commands = wrapCommands(mergeCommands(
		AuthCommands(),
		UserCommands(),
		DataSetCommands(),
		VersionCommands(versionReporter),
	))
	return app, nil
}

func initializeAPI(c *cli.Context) (*api.API, error) {
	name := fmt.Sprintf("%s-%s", c.App.Name, c.App.Version)
	endpoint := c.String(EndpointFlag)
	if endpoint == "" {
		environment := c.String(EnvFlag)
		if environment == "" {
			return nil, errors.New("Endpoint or environment must be specified")
		}

		var ok bool
		if endpoint, ok = environmentEndpointMap[strings.ToLower(environment)]; !ok {
			return nil, fmt.Errorf("Unknown environment: %s", environment)
		}
	}
	proxy := c.String(ProxyFlag)

	API, err := api.NewAPI(name, endpoint, proxy)
	if err != nil {
		return nil, err
	}
	API.Verbose = c.Bool(VerboseFlag)
	API.Writer = c.App.Writer

	return API, nil
}

func API(c *cli.Context) *api.API {
	if _API == nil {
		API, err := initializeAPI(c)
		if err != nil {
			reportErrorAndDie(c, err)
		}
		_API = API
	}
	return _API
}
