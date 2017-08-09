package mongo

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/store/mongo"
	"github.com/tidepool-org/platform/tool"
)

const (
	AddressesFlag = "addresses"
	TLSFlag       = "tls"
)

type Tool struct {
	*tool.Tool
	mongoConfig *mongo.Config
}

func NewTool(prefix string) (*Tool, error) {
	tuel, err := tool.New(prefix)
	if err != nil {
		return nil, err
	}

	return &Tool{
		Tool:        tuel,
		mongoConfig: mongo.NewConfig(),
	}, nil
}

func (t *Tool) Initialize() error {
	if err := t.Tool.Initialize(); err != nil {
		return err
	}

	if err := t.MongoConfig().Load(t.ConfigReporter().WithScopes("store")); err != nil {
		return errors.Wrap(err, "mongo", "unable to load store config")
	}

	t.CLI().Flags = append(t.CLI().Flags,
		cli.StringFlag{
			Name:  fmt.Sprintf("%s,%s", AddressesFlag, "a"),
			Usage: "comma-delimited list of address(es) to mongo database (host:port)",
		},
		cli.BoolFlag{
			Name:  fmt.Sprintf("%s,%s", TLSFlag, "t"),
			Usage: "use TLS to connect to mongo database",
		},
	)

	return nil
}

func (t *Tool) Terminate() {
	t.mongoConfig = nil

	t.Tool.Terminate()
}

func (t *Tool) ParseContext(context *cli.Context) bool {
	if parsed := t.Tool.ParseContext(context); !parsed {
		return parsed
	}

	if context.IsSet(AddressesFlag) {
		t.MongoConfig().Addresses = mongo.SplitAddresses(context.String(AddressesFlag))
	}
	if context.IsSet(TLSFlag) {
		t.MongoConfig().TLS = context.Bool(TLSFlag)
	}

	return true
}

func (t *Tool) MongoConfig() *mongo.Config {
	return t.mongoConfig
}
