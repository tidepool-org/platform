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
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/tool"
)

const (
	EncodeFlag = "encode"
	DecodeFlag = "decode"
	SecretFlag = "secret"
)

func main() {
	application.RunAndExit(NewTool())
}

type Tool struct {
	*tool.Tool
	secret string
	encode bool
	decode bool
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

	t.CLI().Usage = "Encode or decode permission group ids"
	t.CLI().Authors = []cli.Author{
		{
			Name:  "Darin Krauss",
			Email: "darin@tidepool.org",
		},
	}
	t.CLI().Flags = append(t.CLI().Flags,
		cli.BoolFlag{
			Name:  fmt.Sprintf("%s,%s", EncodeFlag, "e"),
			Usage: "encode the specified user id to a group id",
		},
		cli.BoolFlag{
			Name:  fmt.Sprintf("%s,%s", DecodeFlag, "d"),
			Usage: "decode the specified group id to a user id",
		},
		cli.StringFlag{
			Name:  SecretFlag,
			Usage: "permission store secret",
		},
	)

	t.CLI().Action = func(ctx *cli.Context) error {
		if !t.ParseContext(ctx) {
			return nil
		}
		return t.execute()
	}

	return nil
}

func (t *Tool) ParseContext(ctx *cli.Context) bool {
	if parsed := t.Tool.ParseContext(ctx); !parsed {
		return parsed
	}

	t.secret = t.ConfigReporter().WithScopes("permission", "store").GetWithDefault("secret", "")

	if ctx.IsSet(SecretFlag) {
		t.secret = ctx.String(SecretFlag)
	}
	t.encode = ctx.Bool(EncodeFlag)
	t.decode = ctx.Bool(DecodeFlag)

	return true
}

func (t *Tool) execute() error {
	if t.secret == "" {
		return errors.New("secret is missing")
	}
	if t.encode == t.decode {
		return errors.New("must specify only one of --encode or --decode")
	}

	var reader io.Reader
	if len(t.Args()) > 0 {
		reader = strings.NewReader(strings.Join(t.Args(), "\n"))
	} else {
		reader = os.Stdin
	}

	var coder func(userID string, secret string) (string, error)
	if t.encode {
		coder = permission.GroupIDFromUserID
	} else {
		coder = permission.UserIDFromGroupID
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		result, err := coder(scanner.Text(), t.secret)
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
