package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/work"
	"github.com/tidepool-org/platform/work/load"
	"github.com/urfave/cli/v3"
)

func main() {

	var filePath string
	var outputDir string
	var urlBase string

	var generateDurationSeconds int64
	var generateCount int64
	var generateResult string

	var baseURLFlag = &cli.StringFlag{
		Name:        "urlBase",
		Aliases:     []string{"u"},
		Value:       "https://qa2.development.tidepool.org",
		Usage:       "base URL for environment we are testing against",
		Destination: &urlBase,
		Required:    true,
	}

	var outputDirFlag = &cli.StringFlag{
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
	}

	cmd := &cli.Command{
		Usage: "Load test the work system",
		Commands: []*cli.Command{
			{
				Name:    "run",
				Aliases: []string{"r"},
				Usage:   "run the load test",
				Flags: []cli.Flag{
					baseURLFlag,
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
					outputDirFlag,
				},
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
				Name:     "generate",
				Aliases:  []string{"g"},
				Usage:    "generate a load test",
				Commands: []*cli.Command{},
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:        "seconds",
						Aliases:     []string{"s"},
						Usage:       "number of seconds the load test will run over",
						Destination: &generateDurationSeconds,
						Value:       60,
						Required:    false,
					},
					&cli.IntFlag{
						Name:        "count",
						Aliases:     []string{"c"},
						Usage:       "number of work items to run",
						Destination: &generateCount,
						Value:       100,
						Required:    false,
					},
					&cli.StringFlag{
						Name:        "result",
						Aliases:     []string{"r"},
						Usage:       "expected result of work item",
						Destination: &generateResult,
						Value:       work.ResultSuccess,
						Required:    false,
						Action: func(ctx context.Context, cmd *cli.Command, v string) error {
							if !slices.Contains(work.Results(), v) {
								return fmt.Errorf("results must be one of %s", work.Results())
							}
							return nil
						},
					},
					outputDirFlag,
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {

					items := []load.LoadItem{}
					offsetSeconds := 0

					for range int(generateCount) {
						items = append(items, load.LoadItem{
							SecondsOffsetFromStart: int64(offsetSeconds),
							Create: &work.Create{
								Type:    load.TypeSleepy,
								GroupID: pointer.FromString("group-1"),
								Metadata: map[string]any{
									load.MetadataProcessResult: generateResult,
								},
							},
						})
						if int64(offsetSeconds) >= generateDurationSeconds {
							offsetSeconds = int(generateDurationSeconds)
						} else {
							offsetSeconds++
						}
					}

					jsonData, err := json.Marshal(items)
					if err != nil {
						return err
					}

					file, err := os.Create(fmt.Sprintf("%s/count_%d_secs_%d_result_%s_load.json", outputDir, generateCount, generateDurationSeconds, generateResult))
					if err != nil {
						return err
					}
					defer file.Close()

					_, err = file.Write(jsonData)
					if err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:    "verify",
				Aliases: []string{"v"},
				Usage:   "is load test running",
				Flags: []cli.Flag{
					baseURLFlag,
				},
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
