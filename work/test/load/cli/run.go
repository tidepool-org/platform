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
	ID                string
	groupID           string
	duplicates        bool
	duplicateIDPrefix string
	serialize         bool
	serialIDPrefix    string
}

func newRun() *run {
	id := data.NewID()
	return &run{
		ID:      id,
		groupID: fmt.Sprintf("group-id-%s", id),
	}
}

func (r *run) saveTestRun(data []byte) error {
	if r.outputDir == "" {
		return nil
	}
	_, fileName := filepath.Split(r.inputFile)
	outputFile := fmt.Sprintf("%s%s_duplicates[%t]_serialize[%t]_%s",
		r.outputDir,
		time.Now().Format(time.DateOnly),
		r.duplicates,
		r.serialize,
		fileName,
	)
	return os.WriteFile(outputFile, data, os.ModePerm)
}

func (r *run) commandBefore(ctx *cli.Context) error {
	if _, err := os.Stat(r.inputFile); err != nil {
		return fmt.Errorf("filePath %s does not exist", r.inputFile)
	}
	if _, err := os.Stat(r.outputDir); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(r.outputDir, 0755)
		}
	}
	if r.duplicates {
		r.duplicateIDPrefix = fmt.Sprintf("%s-deduplication-id", r.ID)
	}
	if r.serialize {
		r.serialIDPrefix = fmt.Sprintf("%s-serial-id", r.ID)
	}
	return nil
}

func (r *run) commandAction(ctx *cli.Context) error {
	testDataContent, err := os.ReadFile(r.inputFile)
	if err != nil {
		return fmt.Errorf("error opening %s %s", r.inputFile, err.Error())
	}

	var items []workLoad.LoadItem
	err = json.Unmarshal(testDataContent, &items)
	if err != nil {
		return fmt.Errorf("unable to load testing data: %s", err.Error())
	}

	for i := range items {
		items[i].Create.GroupID = &r.groupID
		if r.duplicates {
			items[i].Create.DeduplicationID = pointer.FromString(fmt.Sprintf("%s-%d", r.duplicateIDPrefix, i))
		}
		if r.serialize {
			items[i].Create.SerialID = pointer.FromString(fmt.Sprintf("%s-%d", r.serialIDPrefix, i))
		}
	}

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(items)
	if err != nil {
		return fmt.Errorf("unable to load testing data: %s", err.Error())
	}

	res, err := http.Post(fmt.Sprintf("%s/v1/work/load", r.apiURLBase), "application/json", &buf)
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
	return r.saveTestRun(bodyData)
}

func (r *run) GetCommand() cli.Command {
	return cli.Command{
		Name:    "run",
		Aliases: []string{"r"},
		Usage:   "run the load test",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "urlBase",
				Value:       "https://qa2.development.tidepool.org",
				Usage:       "base URL for environment we are testing against",
				Destination: &r.apiURLBase,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "filePath",
				Usage:       "path to the load test file",
				Destination: &r.inputFile,
				Required:    true,
			},
			&cli.BoolFlag{
				Name:        "duplicates",
				Usage:       "set work items to be duplicates",
				Destination: &r.duplicates,
				Required:    false,
			},
			&cli.BoolFlag{
				Name:        "serialize",
				Usage:       "serialize work per data source",
				Destination: &r.serialize,
				Required:    false,
			},
			&cli.StringFlag{
				Name:        "outputDir",
				Usage:       "directory to save the test output",
				Destination: &r.outputDir,
				Required:    false,
			},
		},
		Before: r.commandBefore,
		Action: r.commandAction,
	}
}
