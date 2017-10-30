package mongo

import (
	"fmt"

	"github.com/urfave/cli"

	mongoTool "github.com/tidepool-org/platform/tool/mongo"
)

const (
	DryRunFlag = "dry-run"
)

type Migration struct {
	*mongoTool.Tool
	dryRun bool
}

func NewMigration(prefix string) (*Migration, error) {
	tuel, err := mongoTool.NewTool(prefix)
	if err != nil {
		return nil, err
	}

	return &Migration{
		Tool: tuel,
	}, nil
}

func (m *Migration) Initialize() error {
	if err := m.Tool.Initialize(); err != nil {
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

func (m *Migration) ParseContext(context *cli.Context) bool {
	if parsed := m.Tool.ParseContext(context); !parsed {
		return parsed
	}

	m.dryRun = context.Bool(DryRunFlag)

	return true
}

func (m *Migration) DryRun() bool {
	return m.dryRun
}
