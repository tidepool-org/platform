package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/urfave/cli"
)

func main() {

	var urlBase string
	var run = newRun()
	var generate = newGenerate()

	app := cli.NewApp()
	app.Usage = "Load test the work system"
	app.Commands = []cli.Command{
		run.GetCommand(),
		generate.GetCommand(),
		{
			Name:    "status",
			Aliases: []string{"s"},
			Usage:   "is load test running",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "urlBase",
					Value:       "https://qa2.development.tidepool.org",
					Usage:       "base URL for environment we are testing against",
					Destination: &urlBase,
					Required:    true,
				},
			},
			Action: func(ctx *cli.Context) error {
				res, err := http.Get(fmt.Sprintf("%s/v1/work/load/status", urlBase))
				if err != nil {
					return fmt.Errorf("unable to issue work load test API request: %w", err)
				}

				bodyData, err := io.ReadAll(res.Body)
				if err != nil {
					return fmt.Errorf("could not read the response body %w", err)
				}

				if res.StatusCode != http.StatusOK {
					return fmt.Errorf("unsuccessful work test API response: %v: %v", res.Status, string(bodyData))
				}

				fmt.Printf("%s", string(bodyData))
				return nil
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
