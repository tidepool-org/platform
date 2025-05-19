package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"slices"
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

func randomWorkResult(includeErrors bool) string {
	if !includeErrors {
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
	outputDir         string
	durationMS        int64
	failureMS         int64
	failureSimulation bool
	count             int
	workActions       string
	actionData        map[string]any
	includeErrors     bool
	processingTimeout int
	offsetMS          int64
	items             []workLoad.LoadItem
}

func newGenerate() *generate {
	return &generate{
		failureSimulation: false,
		items:             []workLoad.LoadItem{},
		actionData:        map[string]any{},
	}
}

func (g *generate) setActionData() error {
	allowedActions := []string{workLoad.CreateAction, workLoad.FailureAction, workLoad.RegisterAction, workLoad.SleepAction, workLoad.ResultAction}
	actions := strings.Split(g.workActions, ",")
	var err error

	for _, action := range actions {
		name := strings.Trim(action, " ")
		var data any
		if strings.Contains(name, ":") {
			parts := strings.Split(name, ":")
			name = parts[0]
			if len(parts) > 1 {
				data = parts[1]
			}
		}
		if !slices.Contains(allowedActions, name) {
			err = errors.Join(err, fmt.Errorf("invalid action %s", name))
		}
		g.actionData[name] = data
	}
	return err
}

func (g *generate) buildActions() workLoad.Actions {
	actions := workLoad.Actions{}

	for name, data := range g.actionData {

		action := workLoad.Action{
			"action": name,
		}
		switch name {

		case workLoad.FailureAction:
			g.failureSimulation = true
			// default start failing 1 min into process
			action[workLoad.FailureOffsetMS] = 60 * int(time.Second/time.Millisecond)
			failureOffset, err := strconv.Atoi(data.(string))
			if err == nil {
				action[workLoad.FailureOffsetMS] = failureOffset
			}
			action[workLoad.FailureDurationMS] = g.failureMS
		case workLoad.SleepAction:
			action[workLoad.SleepDelayMS] = 1000
			if data != nil {
				delay, err := strconv.Atoi(data.(string))
				if err == nil {
					action[workLoad.SleepDelayMS] = delay
				}
			}
		case workLoad.ResultAction:
			action[workLoad.ResultAction] = resultData(data)
		case workLoad.CreateAction:
			action["create"] = work.Create{
				Type: randomWorkType(),
				Metadata: map[string]any{
					//TODO: need ability to specify createWork actions
					"actions": workLoad.Actions{
						workLoad.Action{"action": workLoad.SleepAction, "delay": rand.IntN(4000)},
						workLoad.Action{"action": workLoad.ResultAction, "result": randomWorkResult(g.includeErrors)},
					},
				},
				ProcessingTimeout: g.processingTimeout,
			}
		case workLoad.RegisterAction:
			action[workLoad.RegisterType] = workLoad.DomainName("other")
			if data != nil {
				subdomain, ok := data.(string)
				if ok {
					action[workLoad.ResultAction] = workLoad.DomainName(subdomain)
				}
			}
		}
		actions = append(actions, action)
	}
	return actions
}

func (g *generate) calcOffsetFromStart() int64 {
	if g.durationMS == 0 {
		return g.offsetMS
	}
	currentCount := len(g.items)
	interval := g.durationMS / int64(g.count)
	return interval*int64(currentCount) + g.offsetMS
}

func (g *generate) saveTestData(data []byte) error {

	fileName := fmt.Sprintf("items[%d]_duration[%dms]_offset[%dms]_result-errs[%t]_sim-failure[%t]",
		len(g.items),
		g.durationMS,
		g.offsetMS,
		g.includeErrors,
		g.failureSimulation,
	)

	file, err := os.Create(fmt.Sprintf("%s/%s.json", g.outputDir, fileName))
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

func (g *generate) commandBefore(ctx *cli.Context) error {
	if err := g.setActionData(); err != nil {
		return err
	}
	return nil
}

func (g *generate) commandAction(ctx *cli.Context) error {
	for range int(g.count) {
		g.items = append(g.items, workLoad.LoadItem{
			OffsetMilliseconds: g.calcOffsetFromStart(),
			Create: &work.Create{
				Type: randomWorkType(),
				Metadata: map[string]any{
					"actions": g.buildActions(),
				},
				ProcessingTimeout: g.processingTimeout,
			},
		})
	}
	jsonData, err := json.Marshal(g.items)
	if err != nil {
		return err
	}

	return g.saveTestData(jsonData)
}

func (g *generate) GetCommand() cli.Command {
	return cli.Command{
		Name:    "generate",
		Aliases: []string{"g"},
		Usage:   "generate a load test",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "outputDir",
				Usage:       "directory to save the test output",
				Destination: &g.outputDir,
				Required:    true,
			},
			&cli.IntFlag{
				Name:        "count",
				Usage:       "number of top level work items to create",
				Destination: &g.count,
				Value:       100,
				Required:    false,
			},
			&cli.IntFlag{
				Name:        "timeout",
				Usage:       "processing timeout",
				Destination: &g.processingTimeout,
				Value:       15,
				Required:    false,
			},
			&cli.StringFlag{
				Name:        "actions",
				Usage:       "comma sperated list of actions for each top level workitem",
				Destination: &g.workActions,
				Value:       "sleep",
				Required:    false,
			},
			&cli.BoolFlag{
				Name:        "errors",
				Usage:       "include error results",
				Destination: &g.includeErrors,
				Required:    false,
			},
			&cli.Int64Flag{
				Name:        "offsetMs",
				Usage:       "offset from which all items will start in the future e.g. --offsetMs 3600000",
				Destination: &g.offsetMS,
				Value:       0,
				Required:    false,
			},
			&cli.Int64Flag{
				Name:        "durationMs",
				Usage:       "duration to spread the work load over. Default means all will be available at the same time",
				Destination: &g.durationMS,
				Value:       0,
				Required:    false,
			},
			&cli.Int64Flag{
				Name:        "failureMs",
				Usage:       "if failure action is included this is how long we will simulate failure for",
				Destination: &g.failureMS,
				Value:       180000,
				Required:    false,
			},
		},
		Before: g.commandBefore,
		Action: g.commandAction,
	}
}
