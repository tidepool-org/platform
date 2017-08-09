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
	application.Run(NewTool())
}

type Tool struct {
	*tool.Tool
	secret string
	encode bool
	decode bool
}

func NewTool() (*Tool, error) {
	tuel, err := tool.New("TIDEPOOL")
	if err != nil {
		return nil, err
	}

	return &Tool{
		Tool: tuel,
	}, nil
}

func (t *Tool) Initialize() error {
	if err := t.Tool.Initialize(); err != nil {
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

	t.secret = t.ConfigReporter().WithScopes("permission", "store").GetWithDefault("secret", "")

	if context.IsSet(SecretFlag) {
		t.secret = context.String(SecretFlag)
	}
	t.encode = context.Bool(EncodeFlag)
	t.decode = context.Bool(DecodeFlag)

	return true
}

func (t *Tool) execute() error {
	if t.secret == "" {
		return errors.New("main", "secret is missing")
	}
	if t.encode == t.decode {
		return errors.New("main", "must specify only one of --encode or --decode")
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
			return errors.Wrap(err, "main", "unable to process input")
		}
		fmt.Println(result)
	}
	if err := scanner.Err(); err != nil {
		return errors.Wrap(err, "main", "unable to read input")
	}

	return nil
}
