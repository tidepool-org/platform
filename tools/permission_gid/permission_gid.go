package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/urfave/cli"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/version"
)

type Config struct {
	Secret string
	Encode bool
}

const (
	HelpFlag    = "help"
	VersionFlag = "version"
	EncodeFlag  = "encode"
	DecodeFlag  = "decode"
	SecretFlag  = "secret"
)

func main() {
	application, err := initializeApplication()
	if err != nil {
		fmt.Println("ERROR: Unable to initialize application:", err)
		os.Exit(1)
	}

	if err := application.Run(os.Args); err != nil {
		fmt.Println("ERROR: Unable to run application:", err)
		os.Exit(1)
	}
}

func initializeApplication() (*cli.App, error) {
	versionReporter, err := initializeVersionReporter()
	if err != nil {
		return nil, err
	}

	application := cli.NewApp()
	application.Usage = "Encode or decode permission group ids"
	application.Version = versionReporter.Long()
	application.Authors = []cli.Author{{"Darin Krauss", "darin@tidepool.org"}}
	application.Copyright = "Copyright \u00A9 2017, Tidepool Project"
	application.HideHelp = true
	application.HideVersion = true
	application.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  fmt.Sprintf("%s,%s,%s", HelpFlag, "h", "?"),
			Usage: "print this page and exit",
		},
		cli.BoolFlag{
			Name:  VersionFlag,
			Usage: "print version and exit",
		},
		cli.BoolFlag{
			Name:  fmt.Sprintf("%s,%s", EncodeFlag, "e"),
			Usage: "encode the specified user id to a group id",
		},
		cli.BoolFlag{
			Name:  fmt.Sprintf("%s,%s", DecodeFlag, "d"),
			Usage: "decode the specified group id to a user id",
		},
		cli.StringFlag{
			Name:   SecretFlag,
			Usage:  "gatekeeper secret",
			EnvVar: "GATEKEEPER_SECRET",
		},
	}
	application.Action = func(context *cli.Context) error {
		executeApplication(versionReporter, context)
		return nil
	}

	return application, nil
}

func initializeVersionReporter() (version.Reporter, error) {
	versionReporter, err := version.NewDefaultReporter()
	if err != nil {
		return nil, app.ExtError(err, "permission_gid", "unable to create version reporter")
	}

	return versionReporter, nil
}

func executeApplication(versionReporter version.Reporter, context *cli.Context) {
	if context.Bool(HelpFlag) {
		cli.ShowAppHelp(context)
		return
	}

	if context.Bool(VersionFlag) {
		fmt.Println(versionReporter.Long())
		return
	}

	config, err := buildConfigFromContext(context)
	if err != nil {
		fmt.Println("ERROR: Unable to build config from context:", err)
		os.Exit(1)
	}

	var reader io.Reader
	if len(context.Args()) > 0 {
		reader = strings.NewReader(strings.Join(context.Args(), "\n"))
	} else {
		reader = os.Stdin
	}

	if err = permissionGroupIDs(config, reader); err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}
}

func buildConfigFromContext(context *cli.Context) (*Config, error) {
	config := &Config{}

	config.Secret = context.String(SecretFlag)
	if config.Secret == "" {
		return nil, app.Error("permission_gid", "secret is missing")
	}

	config.Encode = context.Bool(EncodeFlag)
	if config.Encode == context.Bool(DecodeFlag) {
		return nil, app.Error("permission_gid", "must specify either encode or decode")
	}

	return config, nil
}

func permissionGroupIDs(config *Config, reader io.Reader) error {
	var coder func(userID string, secret string) (string, error)
	if config.Encode {
		coder = permission.GroupIDFromUserID
	} else {
		coder = permission.UserIDFromGroupID
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		result, err := coder(scanner.Text(), config.Secret)
		if err != nil {
			return app.ExtError(err, "permission_gid", "unable to process input")
		}
		fmt.Println(result)
	}
	if err := scanner.Err(); err != nil {
		return app.ExtError(err, "permission_gid", "unable to read input")
	}

	return nil
}
