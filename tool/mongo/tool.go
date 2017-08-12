package mongo

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
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

	if err := t.mongoConfig.Load(t.ConfigReporter().WithScopes("store")); err != nil {
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
		t.mongoConfig.Addresses = mongo.SplitAddresses(context.String(AddressesFlag))
	}
	if context.IsSet(TLSFlag) {
		t.mongoConfig.TLS = context.Bool(TLSFlag)
	}

	return true
}

func (t *Tool) NewMongoConfig() *mongo.Config {
	mongoConfig := mongo.NewConfig()
	if t.mongoConfig.Addresses != nil {
		mongoConfig.Addresses = append([]string{}, t.mongoConfig.Addresses...)
	}
	mongoConfig.TLS = t.mongoConfig.TLS
	mongoConfig.Database = t.mongoConfig.Database
	mongoConfig.Collection = t.mongoConfig.Collection
	if t.mongoConfig.Username != nil {
		mongoConfig.Username = pointer.String(*t.mongoConfig.Username)
	}
	if t.mongoConfig.Password != nil {
		mongoConfig.Password = pointer.String(*t.mongoConfig.Password)
	}
	return mongoConfig
}
