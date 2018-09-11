package mongo

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/tidepool-org/platform/application"
	mongoTool "github.com/tidepool-org/platform/tool/mongo"
)

const (
	DryRunFlag = "dry-run"
)

type Migration struct {
	*mongoTool.Tool
	dryRun bool
}

func NewMigration() *Migration {
	return &Migration{
		Tool: mongoTool.NewTool(),
	}
}

func (m *Migration) Initialize(provider application.Provider) error {
	if err := m.Tool.Initialize(provider); err != nil {
		return err
	}

	m.CLI().Flags = append(m.CLI().Flags,
		cli.BoolFlag{
			Name:  fmt.Sprintf("%s,%s", DryRunFlag, "n"),
			Usage: "dry run only; do not migrate",
		},
	)

	return nil
}

func (m *Migration) ParseContext(ctx *cli.Context) bool {
	if parsed := m.Tool.ParseContext(ctx); !parsed {
		return parsed
	}

	m.dryRun = ctx.Bool(DryRunFlag)

	return true
}

func (m *Migration) DryRun() bool {
	return m.dryRun
}
