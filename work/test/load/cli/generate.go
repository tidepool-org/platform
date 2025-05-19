package main

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli"

	"github.com/tidepool-org/platform/work"
	workLoad "github.com/tidepool-org/platform/work/load"
)

func randomWorkType() string {
	if rand.IntN(2) == 0 {
		return workLoad.TypeDopey
	}
	return workLoad.TypeSleepy
}

func randomWorkResult(errors bool) string {
	if !errors {
		if rand.IntN(2) == 0 {
			return work.ResultSuccess
		}
		return work.ResultDelete
	}
	i := rand.IntN(len(work.Results()))
	return work.Results()[i]
}

func resultData(result any) map[string]any {
	if result != nil {
		resultStr := fmt.Sprintf("%v", result)
		switch resultStr {
		case work.ResultDelete, work.ResultSuccess, work.ResultFailed:
			return map[string]any{"result": resultStr}
		case work.ResultPending:
			return map[string]any{
				"result": resultStr,
				"metadata": map[string]any{
					workLoad.PendingOffsetMS: 1000,
				},
			}
		case work.ResultFailing:
			return map[string]any{
				"result": resultStr,
				"metadata": map[string]any{
					workLoad.FailingOffsetMS: 1000,
				},
			}
		}
	}
	return map[string]any{"result": work.ResultSuccess}
}

type generate struct {
	outputDir             string
	spreadDurationSeconds int64
	failureDurationMS     int64
	systemFailure         bool
	workItemCount         int
	workActions           string
	includeErrors         bool
	processingTimeout     int
	startOffsetSeconds    int
	items                 []workLoad.LoadItem
}

func newGenerate() *generate {
	return &generate{
		systemFailure: false,
		items:         []workLoad.LoadItem{},
	}
}

func (c *generate) buildActions(actionNames []string) workLoad.Actions {
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

		case workLoad.FailureAction:
			c.systemFailure = true
			// default start failing 1 min into process
			action[workLoad.FailureOffsetMS] = 60 * int(time.Second/time.Millisecond)
			failureOffset, err := strconv.Atoi(actionData.(string))
			if err == nil {
				action[workLoad.FailureOffsetMS] = failureOffset
			}
			action[workLoad.FailureDurationMS] = c.failureDurationMS
		case workLoad.SleepAction:
			action[workLoad.SleepDelayMS] = 1000
			if actionData != nil {
				delay, err := strconv.Atoi(actionData.(string))
				if err == nil {
					action[workLoad.SleepDelayMS] = delay
				}
			}
		case workLoad.ResultAction:
			action[workLoad.ResultAction] = resultData(actionData)
		case workLoad.CreateAction:
			action["create"] = work.Create{
				Type: randomWorkType(),
				Metadata: map[string]any{
					//TODO: need ability to specify createWork actions
					"actions": workLoad.Actions{
						workLoad.Action{"action": workLoad.SleepAction, "delay": rand.IntN(4000)},
						workLoad.Action{"action": workLoad.ResultAction, "result": randomWorkResult(c.includeErrors)},
					},
				},
				ProcessingTimeout: c.processingTimeout,
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

func (c *generate) calcOffsetFromStart() int64 {
	generateFutureMilliseconds := int64(c.startOffsetSeconds * 1000)

	if c.spreadDurationSeconds == 0 {
		if generateFutureMilliseconds == 0 {
			return 0
		}
		return generateFutureMilliseconds
	}
	currentCount := len(c.items)
	offsetMilliseconds := int(c.spreadDurationSeconds * 1000)
	interval := offsetMilliseconds / int(c.workItemCount)
	return int64(interval*currentCount) + generateFutureMilliseconds
}

func (c *generate) saveTestData(data []byte) error {

	fileName := fmt.Sprintf("items[%d]-duration[%ds]-offset[%ds]-result-errs[%t]-sys-failure[%t]",
		len(c.items),
		c.spreadDurationSeconds,
		c.startOffsetSeconds,
		c.includeErrors,
		c.systemFailure,
	)

	file, err := os.Create(fmt.Sprintf("%s/%s.json", c.outputDir, fileName))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (c *generate) commandAction(ctx *cli.Context) error {
	for range int(c.workItemCount) {
		actionNames := strings.Split(c.workActions, ",")

		c.items = append(c.items, workLoad.LoadItem{
			OffsetMilliseconds: c.calcOffsetFromStart(),
			Create: &work.Create{
				Type: randomWorkType(),
				Metadata: map[string]any{
					"actions": c.buildActions(actionNames),
				},
				ProcessingTimeout: c.processingTimeout,
			},
		})
	}
	jsonData, err := json.Marshal(c.items)
	if err != nil {
		return err
	}

	return c.saveTestData(jsonData)
}

func (c *generate) GetCommand() cli.Command {
	return cli.Command{
		Name:    "generate",
		Aliases: []string{"g"},
		Usage:   "generate a load test",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:        "count",
				Usage:       "number of top level work items to create",
				Destination: &c.workItemCount,
				Value:       100,
				Required:    false,
			},
			&cli.IntFlag{
				Name:        "timeout",
				Usage:       "processing timeout",
				Destination: &c.processingTimeout,
				Value:       15,
				Required:    false,
			},
			&cli.StringFlag{
				Name:        "actions",
				Usage:       "comma sperated list of actions for each top level workitem",
				Destination: &c.workActions,
				Value:       "sleep",
				Required:    false,
			},
			&cli.BoolFlag{
				Name:        "errors",
				Usage:       "include error results",
				Destination: &c.includeErrors,
				Required:    false,
			},
			&cli.IntFlag{
				Name:        "futureSecs",
				Usage:       "will add all items but start the specified seconds in the future e.g. --futureSecs 3600",
				Destination: &c.startOffsetSeconds,
				Value:       0,
				Required:    false,
			},
			&cli.Int64Flag{
				Name:        "duration",
				Usage:       "seconds to spread the work load over. 0 means all will be available at the same time",
				Destination: &c.spreadDurationSeconds,
				Value:       0,
				Required:    false,
			},
			&cli.Int64Flag{
				Name:        "failureDuration",
				Usage:       "if a failure action is included this is how long we will fail for",
				Destination: &c.failureDurationMS,
				Value:       180000,
				Required:    false,
			},
			&cli.StringFlag{
				Name:        "outputDir",
				Usage:       "directory to save the test output",
				Destination: &c.outputDir,
			},
		},
		Action: c.commandAction,
	}
}
