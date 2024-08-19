package mongo

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	"github.com/tidepool-org/platform/tool"
)

const (
	AddressesFlag = "addresses"
	TLSFlag       = "tls"
)

type Tool struct {
	*tool.Tool
	mongoConfig *storeStructuredMongo.Config
}

func NewTool() *Tool {
	return &Tool{
		Tool:        tool.New(),
		mongoConfig: storeStructuredMongo.NewConfig(),
	}
}

func (t *Tool) Initialize(provider application.Provider) error {
	if err := t.Tool.Initialize(provider); err != nil {
		return err
	}

	if err := t.mongoConfig.Load(); err != nil {
		return errors.Wrap(err, "unable to load store config")
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

func (t *Tool) ParseContext(ctx *cli.Context) bool {
	if parsed := t.Tool.ParseContext(ctx); !parsed {
		return parsed
	}

	if ctx.IsSet(AddressesFlag) {
		t.mongoConfig.Addresses = config.SplitTrimCompact(ctx.String(AddressesFlag))
	}
	if ctx.IsSet(TLSFlag) {
		t.mongoConfig.TLS = ctx.Bool(TLSFlag)
	}

	return true
}

func (t *Tool) NewMongoConfig() *storeStructuredMongo.Config {
	mongoConfig := storeStructuredMongo.NewConfig()
	mongoConfig.Scheme = t.mongoConfig.Scheme
	if t.mongoConfig.Addresses != nil {
		mongoConfig.Addresses = append([]string{}, t.mongoConfig.Addresses...)
	}
	mongoConfig.TLS = t.mongoConfig.TLS
	mongoConfig.Database = t.mongoConfig.Database
	mongoConfig.CollectionPrefix = t.mongoConfig.CollectionPrefix
	if t.mongoConfig.Username != nil {
		mongoConfig.Username = pointer.FromString(*t.mongoConfig.Username)
	}
	if t.mongoConfig.Password != nil {
		mongoConfig.Password = pointer.FromString(*t.mongoConfig.Password)
	}
	mongoConfig.Timeout = t.mongoConfig.Timeout
	mongoConfig.OptParams = t.mongoConfig.OptParams
	return mongoConfig
}
