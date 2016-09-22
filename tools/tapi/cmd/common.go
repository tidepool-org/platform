package cmd

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	EndpointFlag = "endpoint"
	EnvFlag      = "env"
	PrettyFlag   = "pretty"
	ProxyFlag    = "proxy"
	VerboseFlag  = "verbose"
)

func CommandFlags(flags ...cli.Flag) []cli.Flag {
	return append(flags,
		cli.StringFlag{
			Name:   EndpointFlag,
			Usage:  "Tidepool API `ENDPOINT` (eg. 'https://api.tidepool.org')",
			EnvVar: "TIDEPOOL_ENDPOINT",
		},
		cli.StringFlag{
			Name:   EnvFlag,
			Usage:  "Tidepool `ENVIRONMENT` (ie. 'prd', 'int', 'stg', 'dev', 'local')",
			EnvVar: "TIDEPOOL_ENV",
		},
		cli.StringFlag{
			Name:   ProxyFlag,
			Value:  "",
			Usage:  "proxy `URL`",
			EnvVar: "HTTP_PROXY",
		},
		cli.BoolFlag{
			Name:  fmt.Sprintf("%s,%s", PrettyFlag, "p"),
			Usage: "pretty print JSON",
		},
		cli.BoolFlag{
			Name:  fmt.Sprintf("%s,%s", VerboseFlag, "v"),
			Usage: "include info output",
		},
	)
}

func mergeCommands(left cli.Commands, rights ...cli.Commands) cli.Commands {
	merged := cli.Commands{}
	merged = append(merged, left...)
	for _, right := range rights {
		merged = append(merged, right...)
	}
	return merged
}

func wrapCommands(commands cli.Commands) cli.Commands {
	wrapped := cli.Commands{}
	for _, command := range commands {
		wrapped = append(wrapped, wrapCommand(command))
	}
	return wrapped
}

func wrapCommand(command cli.Command) cli.Command {
	command.Before = wrapCommandFunc(command.Before)
	command.After = wrapCommandFunc(command.After)
	if actionFunc, actionOK := command.Action.(cli.ActionFunc); actionOK {
		command.Action = wrapCommandFunc(actionFunc)
	} else if commandFunc, commandOK := command.Action.(func(*cli.Context) error); commandOK {
		command.Action = wrapCommandFunc(commandFunc)
	}
	command.Subcommands = wrapCommands(command.Subcommands)
	return command
}

func wrapCommandFunc(commandFunc func(c *cli.Context) error) func(c *cli.Context) error {
	if commandFunc == nil {
		return nil
	}
	return func(c *cli.Context) error {
		if err := commandFunc(c); err != nil {
			reportErrorAndDie(c, err)
		}
		return nil
	}
}

func reportErrorAndDie(c *cli.Context, err error) {
	fmt.Fprintf(c.App.Writer, "ERROR: %s\n", err.Error())
	os.Exit(1)
}

func reportMessage(c *cli.Context, messages ...interface{}) error {
	fmt.Fprintln(c.App.Writer, messages...)
	return nil
}

func reportMessageWithJSON(c *cli.Context, data interface{}) error {
	var messageBytes []byte
	var err error

	if c.Bool(PrettyFlag) {
		messageBytes, err = json.MarshalIndent(data, "", "    ")
	} else {
		messageBytes, err = json.Marshal(data)
	}
	if err != nil {
		return err
	}

	return reportMessage(c, string(messageBytes))
}

func ensureNoArgs(c *cli.Context) error {
	if c.Args().Present() {
		return fmt.Errorf("Unexpected arguments: %s", strings.Join(c.Args(), " "))
	}
	return nil
}

func readFromConsole(prompt string) (string, error) {
	fmt.Print(prompt)
	result, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		fmt.Println()
		return "", err
	}
	return strings.TrimRight(result, "\r\n"), nil
}

func readFromConsoleNoEcho(prompt string) (string, error) {
	fmt.Print(prompt)
	bytes, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", err
	}
	result := string(bytes)
	return result, nil
}
