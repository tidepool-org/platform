package tool

import (
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli"

	"github.com/tidepool-org/platform/application"
)

const (
	HelpFlag    = "help"
	VersionFlag = "version"
)

type Tool struct {
	*application.Application
	cli  *cli.App
	args []string
}

func New() *Tool {
	return &Tool{
		Application: application.New(),
	}
}

func (t *Tool) Initialize(provider application.Provider) error {
	if err := t.Application.Initialize(provider); err != nil {
		return err
	}

	t.cli = cli.NewApp()

	t.CLI().Version = t.VersionReporter().Long()
	t.CLI().Copyright = fmt.Sprintf("Copyright \u00A9 %d, Tidepool Project", time.Now().Year())
	t.CLI().HideHelp = true
	t.CLI().HideVersion = true
	t.CLI().Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  fmt.Sprintf("%s,%s,%s", HelpFlag, "h", "?"),
			Usage: "print this page and exit",
		},
		cli.BoolFlag{
			Name:  VersionFlag,
			Usage: "print version and exit",
		},
	}

	return nil
}

func (t *Tool) Terminate() {
	t.args = nil
	t.cli = nil

	t.Application.Terminate()
}

func (t *Tool) CLI() *cli.App {
	return t.cli
}

func (t *Tool) Args() []string {
	return t.args
}

func (t *Tool) Run() error {
	return t.CLI().Run(os.Args)
}

func (t *Tool) ParseContext(ctx *cli.Context) bool {
	if ctx.Bool(HelpFlag) {
		cli.ShowAppHelp(ctx)
		return false
	}

	if ctx.Bool(VersionFlag) {
		fmt.Fprintln(t.CLI().Writer, t.VersionReporter().Long())
		return false
	}

	t.args = ctx.Args()

	return true
}
