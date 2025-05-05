package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/urfave/cli/v3"
)

func main() {

	var filePath string
	var outputDir string
	var urlBase string

	cmd := &cli.Command{
		Usage: "Load test the work system",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "filePath",
				Aliases: []string{"f"},
				Usage:   "path to the load test file",
				Action: func(ctx context.Context, cmd *cli.Command, v string) error {
					if _, err := os.Stat(v); err != nil {
						return fmt.Errorf("filePath %s does not exist", v)
					}
					return nil
				},
				Destination: &filePath,
			},
			&cli.StringFlag{
				Name:        "outputDir",
				Aliases:     []string{"o"},
				Usage:       "directory to save the test output",
				Destination: &outputDir,
				Action: func(ctx context.Context, cmd *cli.Command, v string) error {
					if _, err := os.Stat(v); err != nil {
						if os.IsNotExist(err) {
							os.Mkdir(v, 0755)
						}
						return nil
					}
					return nil
				},
			},
			&cli.StringFlag{
				Name:        "urlBase",
				Aliases:     []string{"u"},
				Value:       "https://qa2.development.tidepool.org",
				Usage:       "base URL for environment we are testing against",
				Destination: &urlBase,
				Required:    true,
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "run",
				Aliases: []string{"r"},
				Usage:   "run the load test",
				Action: func(ctx context.Context, cmd *cli.Command) error {

					testData, err := os.ReadFile(filePath)
					if err != nil {
						return fmt.Errorf("error loading %s %s", filePath, err.Error())
					}

					url := fmt.Sprintf("%s/v1/work/load", urlBase)

					req, err := http.NewRequest("POST", url, bytes.NewReader(testData))
					if err != nil {
						return fmt.Errorf("error sending data to load work %s", err.Error())
					}
					req.Header.Add("content-type", "application/json")
					req.Header.Add("accept", "application/json")

					res, err := http.DefaultClient.Do(req)
					if err != nil {
						return fmt.Errorf("unable to issue work load test API request: %s", err.Error())
					}
					defer res.Body.Close()

					bodyData, err := io.ReadAll(res.Body)
					if err != nil {
						log.Printf("could not read the response body %s", err.Error())
					}

					if res.StatusCode != http.StatusCreated {
						return fmt.Errorf("unsuccessful work load test API response: %v: %v", res.Status, string(bodyData))
					}
					if outputDir != "" {
						outputFile := fmt.Sprintf("%s/work_%s_created.json", outputDir, time.Now().Format(time.DateTime))
						os.WriteFile(outputFile, bodyData, os.ModePerm)
					}
					fmt.Printf("%s", res.Status)
					return nil
				},
			},
			{
				Name:    "verify",
				Aliases: []string{"v"},
				Usage:   "is load test running",
				Action: func(ctx context.Context, cmd *cli.Command) error {

					url := fmt.Sprintf("%s/v1/work/load/ok", urlBase)

					req, err := http.NewRequest("GET", url, nil)
					if err != nil {
						return fmt.Errorf("error checking %s", err.Error())
					}

					res, err := http.DefaultClient.Do(req)
					if err != nil {
						return fmt.Errorf("unable to issue work load test API request: %s", err.Error())
					}
					defer res.Body.Close()

					bodyData, err := io.ReadAll(res.Body)
					if err != nil {
						log.Printf("could not read the response body %s", err.Error())
					}

					if res.StatusCode != http.StatusOK {
						return fmt.Errorf("unsuccessful work test API response: %v: %v", res.Status, string(bodyData))
					}

					fmt.Printf("%s", string(bodyData))

					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
