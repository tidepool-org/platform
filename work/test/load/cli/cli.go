package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/work"
	workLoad "github.com/tidepool-org/platform/work/load"
	"github.com/urfave/cli"
)

func main() {

	var filePath string
	var groupID string
	var outputDir string
	var urlBase string

	var generateDurationSeconds int64
	var generateCount int64
	var generateResult string
	var generateActions string
	var generateErrors bool
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
				&cli.StringFlag{
					Name:        "groupId",
					Usage:       "groupId for this test run",
					Destination: &groupID,
					Required:    false,
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
				if groupID == "" {
					groupID = fmt.Sprintf("group-id-%s", data.NewID())
				}
				return nil
			},
			Action: func(ctx *cli.Context) error {

				testDataContent, err := os.ReadFile(filePath)
				if err != nil {
					return fmt.Errorf("error opening %s %s", filePath, err.Error())
				}

				var items []workLoad.LoadItem
				err = json.Unmarshal(testDataContent, &items)
				if err != nil {
					return fmt.Errorf("unable to load testing data: %s", err.Error())
				}

				for i := range items {
					items[i].Create.GroupID = pointer.FromString(groupID)
				}

				var buf bytes.Buffer
				err = json.NewEncoder(&buf).Encode(items)
				if err != nil {
					return fmt.Errorf("unable to load testing data: %s", err.Error())
				}

				res, err := http.Post(fmt.Sprintf("%s/v1/work/load", urlBase), "application/json", &buf)

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
					outputFile := fmt.Sprintf("%s/%s_work_%s_created.json", outputDir, groupID, time.Now().Format(time.DateTime))
					log.Printf("run data [%s]", outputFile)
					os.WriteFile(outputFile, bodyData, os.ModePerm)
				}
				log.Printf("%s", res.Status)
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
					Value:       5,
					Required:    false,
				},
				&cli.Int64Flag{
					Name:        "count",
					Usage:       "number of work items to run",
					Destination: &generateCount,
					Value:       100,
					Required:    false,
				},
				&cli.IntFlag{
					Name:        "timeout",
					Usage:       "processing timeout",
					Destination: &generateTimeout,
					Value:       15,
					Required:    false,
				},
				&cli.StringFlag{
					Name:        "actions",
					Usage:       "comma sperated list of actions",
					Destination: &generateActions,
					Value:       "sleep",
					Required:    false,
				},
				&cli.BoolFlag{
					Name:        "errors",
					Usage:       "include error results",
					Destination: &generateErrors,
					Required:    false,
				},
				outputDirFlag,
			},

			Action: func(ctx *cli.Context) error {

				items := []workLoad.LoadItem{}

				var getRandomType = func() string {
					if rand.IntN(2) == 0 {
						return workLoad.TypeDopey
					}
					return workLoad.TypeSleepy
				}

				var getRandomResult = func(errors bool) string {
					if !errors {
						if rand.IntN(2) == 0 {
							return work.ResultSuccess
						}
						return work.ResultDelete
					}
					i := rand.IntN(len(work.Results()))
					return work.Results()[i]
				}

				var getActions = func(actionNames []string) workLoad.Actions {
					actions := workLoad.Actions{}

					for _, name := range actionNames {

						actionName := strings.Trim(name, " ")
						var actionData any
						if strings.Contains(actionName, ":") {
							parts := strings.Split(actionName, ":")
							actionName = parts[0]
							if len(parts) > 1 {
								actionData = parts[1]
							}
						}

						action := workLoad.Action{
							"action": actionName,
						}
						switch actionName {
						case workLoad.SleepAction:
							action[workLoad.SleepDelay] = 1000
							if actionData != nil {
								delay, err := strconv.Atoi(actionData.(string))
								if err == nil {
									action[workLoad.SleepDelay] = delay
								}
							}
						case workLoad.ResultAction:
							action[workLoad.ResultAction] = work.ResultSuccess
							if actionData != nil {
								action[workLoad.ResultAction] = actionData
							}

						case workLoad.CreateAction:
							action["create"] = work.Create{
								Type: getRandomType(),
								Metadata: map[string]any{
									//TODO: need ability to specify createWork actions
									"actions": workLoad.Actions{
										workLoad.Action{"action": workLoad.SleepAction, "delay": rand.IntN(4000)},
										workLoad.Action{"action": workLoad.ResultAction, "result": getRandomResult(generateErrors)},
									},
								},
								ProcessingTimeout: generateTimeout,
							}
						case workLoad.RegisterAction:
							action[workLoad.RegisterType] = workLoad.DomainName("other")
							if actionData != nil {
								subdomain, ok := actionData.(string)
								if ok {
									action[workLoad.ResultAction] = workLoad.DomainName(subdomain)
								}
							}

						}
						actions = append(actions, action)
					}
					return actions
				}

				var calcOffset = func() int64 {
					currentCount := len(items)
					offsetMilliseconds := int(generateDurationSeconds * 1000)
					interval := offsetMilliseconds / int(generateCount)
					return int64(interval * currentCount)
				}

				for range int(generateCount) {
					actionNames := strings.Split(generateActions, ",")

					items = append(items, workLoad.LoadItem{
						OffsetMilliseconds: calcOffset(),
						Create: &work.Create{
							Type: getRandomType(),
							Metadata: map[string]any{
								"actions": getActions(actionNames),
							},
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
