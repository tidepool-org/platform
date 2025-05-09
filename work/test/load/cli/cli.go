package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/work"
	workLoad "github.com/tidepool-org/platform/work/test/load"
	"github.com/urfave/cli"
)

func main() {

	var filePath string
	var outputDir string
	var urlBase string

	var generateDurationSeconds int64
	var generateCount int64
	var generateResult string
	var generateGroupID string
	var generateTimeout int

	var baseURLFlag = &cli.StringFlag{
		Name:        "urlBase",
		Value:       "https://qa2.development.tidepool.org",
		Usage:       "base URL for environment we are testing against",
		Destination: &urlBase,
		Required:    true,
	}

	var outputDirFlag = &cli.StringFlag{
		Name:        "outputDir",
		Usage:       "directory to save the test output",
		Destination: &outputDir,
	}

	app := cli.NewApp()
	app.Usage = "Load test the work system"
	app.Commands = []cli.Command{
		{
			Name:    "run",
			Aliases: []string{"r"},
			Usage:   "run the load test",
			Flags: []cli.Flag{
				baseURLFlag,
				&cli.StringFlag{
					Name:        "filePath",
					Usage:       "path to the load test file",
					Destination: &filePath,
				},
				outputDirFlag,
			},
			Before: func(ctx *cli.Context) error {
				if _, err := os.Stat(filePath); err != nil {
					return fmt.Errorf("filePath %s does not exist", filePath)
				}
				if _, err := os.Stat(outputDir); err != nil {
					if os.IsNotExist(err) {
						os.Mkdir(outputDir, 0755)
					}
				}
				return nil
			},
			Action: func(ctx *cli.Context) error {

				testData, err := os.Open(filePath)
				if err != nil {
					return fmt.Errorf("error opening %s %s", filePath, err.Error())
				}
				defer testData.Close()

				res, err := http.Post(fmt.Sprintf("%s/v1/work/load", urlBase), "application/json", testData)

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
			Name:    "generate",
			Aliases: []string{"g"},
			Usage:   "generate a load test",
			Flags: []cli.Flag{
				&cli.Int64Flag{
					Name:        "seconds",
					Usage:       "number of seconds the load test will run over",
					Destination: &generateDurationSeconds,
					Value:       60,
					Required:    false,
				},
				&cli.Int64Flag{
					Name:        "count",
					Usage:       "number of work items to run",
					Destination: &generateCount,
					Value:       100,
					Required:    false,
				},
				&cli.StringFlag{
					Name:        "result",
					Usage:       "expected result of work item",
					Destination: &generateResult,
					Value:       work.ResultSuccess,
					Required:    false,
				},
				&cli.StringFlag{
					Name:        "group",
					Usage:       "group id",
					Destination: &generateGroupID,
					Value:       "test_group_1",
					Required:    false,
				},
				&cli.IntFlag{
					Name:        "timeout",
					Usage:       "processing timeout",
					Destination: &generateTimeout,
					Value:       5,
					Required:    false,
				},
				outputDirFlag,
			},
			Before: func(ctx *cli.Context) error {
				if generateResult != "random" {
					if !slices.Contains(work.Results(), generateResult) {
						return fmt.Errorf("results must be one of %s", work.Results())
					}
				}
				return nil
			},
			Action: func(ctx *cli.Context) error {

				items := []workLoad.LoadItem{}

				var getRandomType = func() string {
					if rand.IntN(2) == 0 {
						return workLoad.TypeDopey
					}
					return workLoad.TypeSleepy
				}

				var getMetadata = func() map[string]any {
					metadata := map[string]any{
						workLoad.MetadataProcessResult: generateResult,
					}
					if rand.IntN(2) == 0 {
						metadata[workLoad.MetadataSleep] = true
					}
					return metadata
				}

				var calcOffset = func() int64 {
					currentCount := len(items)
					offsetMilliseconds := int(generateDurationSeconds * 1000)
					interval := offsetMilliseconds / int(generateCount)
					return int64(interval * currentCount)
				}

				for range int(generateCount) {

					items = append(items, workLoad.LoadItem{
						OffsetMilliseconds: calcOffset(),
						Create: &work.Create{
							Type:              getRandomType(),
							GroupID:           pointer.FromString(generateGroupID),
							Metadata:          getMetadata(),
							ProcessingTimeout: generateTimeout,
						},
					})
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
			Name:    "status",
			Aliases: []string{"s"},
			Usage:   "is load test running",
			Flags: []cli.Flag{
				baseURLFlag,
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
