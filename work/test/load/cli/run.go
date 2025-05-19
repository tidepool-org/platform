package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/urfave/cli"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/pointer"
	workLoad "github.com/tidepool-org/platform/work/load"
)

type run struct {
	apiURLBase        string
	inputFile         string
	outputDir         string
	runID             string
	groupID           string
	duplicates        bool
	duplicateIDPrefix string
	serialize         bool
	serialIDPrefix    string
}

func newRun() *run {
	runID := data.NewID()
	return &run{
		runID:   runID,
		groupID: fmt.Sprintf("group-id-%s", runID),
	}
}

func (r *run) saveTestRun(data []byte) error {
	if r.outputDir == "" {
		return nil
	}
	_, file := filepath.Split(r.inputFile)
	outputFile := fmt.Sprintf("%s-%s-duplicates[%t]-serialize[%t].json", time.Now().Format(time.DateOnly), file, r.duplicates, r.serialize)
	return os.WriteFile(outputFile, data, os.ModePerm)
}

func (c *run) commandBefore(ctx *cli.Context) error {
	if _, err := os.Stat(c.inputFile); err != nil {
		return fmt.Errorf("filePath %s does not exist", c.inputFile)
	}
	if _, err := os.Stat(c.outputDir); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(c.outputDir, 0755)
		}
	}
	if c.duplicates {
		c.duplicateIDPrefix = fmt.Sprintf("%s-deduplication-id", c.runID)
	}
	if c.serialize {
		c.serialIDPrefix = fmt.Sprintf("%s-serial-id", c.runID)
	}
	return nil
}

func (c *run) commandAction(ctx *cli.Context) error {
	testDataContent, err := os.ReadFile(c.inputFile)
	if err != nil {
		return fmt.Errorf("error opening %s %s", c.inputFile, err.Error())
	}

	var items []workLoad.LoadItem
	err = json.Unmarshal(testDataContent, &items)
	if err != nil {
		return fmt.Errorf("unable to load testing data: %s", err.Error())
	}

	for i := range items {
		items[i].Create.GroupID = &c.groupID
		if c.duplicates {
			items[i].Create.DeduplicationID = pointer.FromString(fmt.Sprintf("%s-%d", c.duplicateIDPrefix, i))
		}
		if c.serialize {
			items[i].Create.SerialID = pointer.FromString(fmt.Sprintf("%s-%d", c.serialIDPrefix, i))
		}
	}

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(items)
	if err != nil {
		return fmt.Errorf("unable to load testing data: %s", err.Error())
	}

	res, err := http.Post(fmt.Sprintf("%s/v1/work/load", c.apiURLBase), "application/json", &buf)
	if err != nil {
		return fmt.Errorf("unable to issue work load test API request: %s", err.Error())
	}
	defer res.Body.Close()

	bodyData, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("could not read the response body %w", err)
	}

	if res.StatusCode != http.StatusCreated {
		return fmt.Errorf("unsuccessful work load test API response: %v: %v", res.Status, string(bodyData))
	}
	return c.saveTestRun(bodyData)
}

func (c *run) GetCommand() cli.Command {
	return cli.Command{
		Name:    "run",
		Aliases: []string{"r"},
		Usage:   "run the load test",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "urlBase",
				Value:       "https://qa2.development.tidepool.org",
				Usage:       "base URL for environment we are testing against",
				Destination: &c.apiURLBase,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "filePath",
				Usage:       "path to the load test file",
				Destination: &c.inputFile,
			},
			&cli.BoolFlag{
				Name:        "duplicates",
				Usage:       "set work items to be duplicates",
				Destination: &c.duplicates,
				Required:    false,
			},
			&cli.BoolFlag{
				Name:        "serialize",
				Usage:       "serialize work per data source",
				Destination: &c.serialize,
				Required:    false,
			},
			&cli.StringFlag{
				Name:        "outputDir",
				Usage:       "directory to save the test output",
				Destination: &c.outputDir,
			},
		},
		Before: c.commandBefore,
		Action: c.commandAction,
	}
}
