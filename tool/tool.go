package tool

import (
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli"

	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/log"
)

const (
	HelpFlag    = "help"
	VersionFlag = "version"
	VerboseFlag = "verbose"
)

type Tool struct {
	*application.Application
	cli  *cli.App
	args []string
}

func New(name string, prefix string) (*Tool, error) {
	app, err := application.New(name, prefix)
	if err != nil {
		return nil, err
	}

	return &Tool{
		Application: app,
	}, nil
}

func (t *Tool) Initialize() error {
	if err := t.Application.Initialize(); err != nil {
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
		cli.BoolFlag{
			Name:  fmt.Sprintf("%s,%s", VerboseFlag, "v"),
			Usage: "increased verbosity",
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

func (t *Tool) ParseContext(context *cli.Context) bool {
	if context.Bool(HelpFlag) {
		cli.ShowAppHelp(context)
		return false
	}

	if context.Bool(VersionFlag) {
		fmt.Fprintln(t.CLI().Writer, t.VersionReporter().Long())
		return false
	}

	if context.Bool(VerboseFlag) {
		t.Logger().SetLevel(log.DebugLevel)
	} else {
		t.Logger().SetLevel(log.InfoLevel)
	}

	t.args = context.Args()

	return true
}
