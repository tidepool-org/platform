package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/urfave/cli"

	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/metric"
	"github.com/tidepool-org/platform/tool"
)

const SaltFlag = "salt"

func main() {
	application.RunAndExit(NewTool())
}

type Tool struct {
	*tool.Tool
	salt string
}

func NewTool() *Tool {
	return &Tool{
		Tool: tool.New(),
	}
}

func (t *Tool) Initialize(provider application.Provider) error {
	if err := t.Tool.Initialize(provider); err != nil {
		return err
	}

	t.CLI().Usage = "Generate metric hash"
	t.CLI().Authors = []cli.Author{
		{
			Name:  "Darin Krauss",
			Email: "darin@tidepool.org",
		},
	}
	t.CLI().Flags = append(t.CLI().Flags,
		cli.StringFlag{
			Name:  SaltFlag,
			Usage: "metric salt",
		},
	)

	t.CLI().Action = func(context *cli.Context) error {
		if !t.ParseContext(context) {
			return nil
		}
		return t.execute()
	}

	return nil
}

func (t *Tool) ParseContext(context *cli.Context) bool {
	if parsed := t.Tool.ParseContext(context); !parsed {
		return parsed
	}

	t.salt = t.ConfigReporter().WithScopes("metric").GetWithDefault("salt", "")

	if context.IsSet(SaltFlag) {
		t.salt = context.String(SaltFlag)
	}

	return true
}

func (t *Tool) execute() error {
	if t.salt == "" {
		return errors.New("salt is missing")
	}

	var reader io.Reader
	if len(t.Args()) > 0 {
		reader = strings.NewReader(strings.Join(t.Args(), "\n"))
	} else {
		reader = os.Stdin
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		result, err := metric.HashFromUserID(scanner.Text(), t.salt)
		if err != nil {
			return errors.Wrap(err, "unable to process input")
		}
		fmt.Println(result)
	}
	if err := scanner.Err(); err != nil {
		return errors.Wrap(err, "unable to read input")
	}

	return nil
}
