package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli"

	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/data"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTypesBasal "github.com/tidepool-org/platform/data/types/basal"
	dataTypesBasalAutomated "github.com/tidepool-org/platform/data/types/basal/automated"
	dataTypesBasalScheduled "github.com/tidepool-org/platform/data/types/basal/scheduled"
	dataTypesBasalSuspend "github.com/tidepool-org/platform/data/types/basal/suspend"
	dataTypesBasalTemporary "github.com/tidepool-org/platform/data/types/basal/temporary"
	dataTypesBolus "github.com/tidepool-org/platform/data/types/bolus"
	dataTypesBolusAutomated "github.com/tidepool-org/platform/data/types/bolus/automated"
	dataTypesBolusCombination "github.com/tidepool-org/platform/data/types/bolus/combination"
	dataTypesBolusExtended "github.com/tidepool-org/platform/data/types/bolus/extended"
	dataTypesBolusNormal "github.com/tidepool-org/platform/data/types/bolus/normal"
	dataTypesFactory "github.com/tidepool-org/platform/data/types/factory"
	dataTypesInsulin "github.com/tidepool-org/platform/data/types/insulin"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/tool"
)

// WARNING: THIS TOOL IS NOT COMPLETELY TESTED!

// This tool will analyze the provided Tidepool data and determine the total insulin delivered for each datum type
// for the specified time period.
//
// To build the tool:
//
//   BUILD=tools/total_insulin make build
//
// To execute the tool:
//
//   _bin/tools/total_insulin/total_insulin -s '2024-03-01T00:00:00Z' -e '2024-03-02T00:00:00Z' data.json
//
// The single, optional argument is the path to the file containing the Tidepool data. If not specified or specified as "-",
// then the tool will read from standard input.
//
// The two optional flags (-s and -e) indicate the start time and end time, respectively, for the time range to calculate
// the total insulin. If either of these flags are not specified then they default to the start time and end time of
// the Tidepool data.

const (
	flagStartTime = "start-time"
	flagEndTime   = "end-time"

	argStdin = "-"
)

var insulinTypes = []string{
	dataTypesBasal.Type,
	dataTypesBolus.Type,
	dataTypesInsulin.Type,
}

var insulinSubTypeMap = map[string][]string{
	dataTypesBasal.Type: {
		dataTypesBasalScheduled.DeliveryType,
		dataTypesBasalTemporary.DeliveryType,
		dataTypesBasalAutomated.DeliveryType,
		dataTypesBasalSuspend.DeliveryType,
	},
	dataTypesBolus.Type: {
		dataTypesBolusNormal.SubType,
		dataTypesBolusCombination.SubType,
		dataTypesBolusExtended.SubType,
		dataTypesBolusAutomated.SubType,
	},
}

func Key(parts ...string) string {
	parts = slices.DeleteFunc(parts, func(p string) bool { return p == "" })
	return strings.Join(parts, "/")
}

type Stats struct {
	Count    int
	Duration time.Duration
	Amount   float64
}

func (s Stats) Add(a Stats) Stats {
	return Stats{
		Count:    s.Count + a.Count,
		Duration: s.Duration + a.Duration,
		Amount:   s.Amount + a.Amount,
	}
}

type Tool struct {
	*tool.Tool
	startTimeFlag string
	endTimeFlag   string
	startTime     *time.Time
	endTime       *time.Time
	data          data.Data
	stats         map[string]Stats
}

func NewTool() *Tool {
	return &Tool{
		Tool:  tool.New(),
		stats: map[string]Stats{},
	}
}

func (t *Tool) Initialize(provider application.Provider) error {
	if err := t.Tool.Initialize(provider); err != nil {
		return err
	}
	t.CLI().Usage = "Total Insulin"
	t.CLI().Authors = []cli.Author{
		{
			Name:  "Darin Krauss",
			Email: "darin.krauss@tidepool.org",
		},
	}
	t.CLI().Flags = append(t.CLI().Flags,
		cli.StringFlag{
			Name:  fmt.Sprintf("%s,%s", flagStartTime, "s"),
			Usage: "start time, optional, RFC3339Nano formatted",
		},
		cli.StringFlag{
			Name:  fmt.Sprintf("%s,%s", flagEndTime, "e"),
			Usage: "end time, optional, RFC3339Nano formatted",
		},
	)
	t.CLI().Action = func(ctx *cli.Context) error {
		if !t.ParseContext(ctx) {
			return nil
		}
		return t.execute()
	}
	return nil
}

func (t *Tool) ParseContext(ctx *cli.Context) bool {
	if parsed := t.Tool.ParseContext(ctx); !parsed {
		return parsed
	}
	t.startTimeFlag = ctx.String(flagStartTime)
	t.endTimeFlag = ctx.String(flagEndTime)
	return true
}

func (t *Tool) execute() error {
	if err := t.parseFlags(); err != nil {
		return err
	}
	if err := t.loadData(); err != nil {
		return err
	}
	if err := t.analyzeData(); err != nil {
		return err
	}
	if err := t.outputStats(); err != nil {
		return err
	}
	return nil
}

func (t *Tool) parseFlags() error {
	if t.startTimeFlag != "" {
		if startTime, err := time.Parse(time.RFC3339Nano, t.startTimeFlag); err != nil {
			return errors.Wrap(err, "unable to parse start time")
		} else {
			t.startTime = pointer.FromTime(startTime)
		}
	}
	if t.endTimeFlag != "" {
		if endTime, err := time.Parse(time.RFC3339Nano, t.endTimeFlag); err != nil {
			return errors.Wrap(err, "unable to parse end time")
		} else {
			t.endTime = pointer.FromTime(endTime)
		}
	}
	if t.startTime != nil && t.endTime != nil && t.startTime.After(*t.endTime) {
		return errors.Newf("start time '%s' is after end time '%s'", t.startTime.Format(time.RFC3339Nano), t.endTime.Format(time.RFC3339Nano))
	}
	return nil
}

func (t *Tool) loadData() error {
	args := t.Args()
	if len(args) == 0 {
		args = append(args, argStdin)
	}
	for _, arg := range args {
		if arg == argStdin {
			if err := t.loadDataFromReader(os.Stdin); err != nil {
				return errors.Wrap(err, "unable to load data from stdin")
			}
		} else {
			if err := t.loadDataFromFile(arg); err != nil {
				return errors.Wrapf(err, "unable to load data from file '%s'", arg)
			}
		}
	}
	return nil
}

func (t *Tool) loadDataFromFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return errors.New("unable to open file")
	}
	defer file.Close()
	return t.loadDataFromReader(file)
}

func (t *Tool) loadDataFromReader(reader io.Reader) error {
	var raw []any
	if err := json.NewDecoder(reader).Decode(&raw); err != nil {
		return errors.New("unable to decode data")
	}
	if err := t.parseDataFromRaw(raw); err != nil {
		return errors.Wrap(err, "unable to parse data")
	}
	return nil
}

func (t *Tool) parseDataFromRaw(raw []any) error {
	var data data.Data

	parser := structureParser.NewArray(t.Logger(), &raw)
	validator := structureValidator.New(t.Logger())
	normalizer := dataNormalizer.New(t.Logger())

	for _, reference := range parser.References() {
		if datum := dataTypesFactory.ParseDatum(parser.WithReferenceObjectParser(reference)); datum != nil && *datum != nil {
			(*datum).Validate(validator.WithReference(strconv.Itoa(reference)))
			(*datum).Normalize(normalizer.WithReference(strconv.Itoa(reference)))
			data = append(data, *datum)
		}
	}

	if err := parser.Error(); err != nil {
		return err
	}
	if err := validator.Error(); err != nil {
		return err
	}
	if err := normalizer.Error(); err != nil {
		return err
	}

	t.data = append(t.data, data...)
	return nil
}

func (t *Tool) analyzeData() error {
	sort.Stable(data.DataByTime(t.data))

	var previousBasalEndTime time.Time
	for _, datum := range t.data {
		var key string
		var duration time.Duration
		var amount float64
		var rate float64

		switch datum := datum.(type) {
		case *dataTypesBasalAutomated.Automated:
			key = Key(datum.Type, datum.DeliveryType)
			duration = time.Duration(*datum.Duration) * time.Millisecond
			rate = *datum.Rate
		case *dataTypesBasalScheduled.Scheduled:
			key = Key(datum.Type, datum.DeliveryType)
			duration = time.Duration(*datum.Duration) * time.Millisecond
			rate = *datum.Rate
		case *dataTypesBasalSuspend.Suspend:
			key = Key(datum.Type, datum.DeliveryType)
			duration = time.Duration(*datum.Duration) * time.Millisecond
		case *dataTypesBasalTemporary.Temporary:
			key = Key(datum.Type, datum.DeliveryType)
			duration = time.Duration(*datum.Duration) * time.Millisecond
			rate = *datum.Rate
		case *dataTypesBolusAutomated.Automated:
			key = Key(datum.Type, datum.SubType)
			amount = *datum.Normal
		case *dataTypesBolusCombination.Combination:
			key = Key(datum.Type, datum.SubType)
			amount = *datum.Normal
			duration = time.Duration(*datum.Duration) * time.Millisecond
			rate = *datum.Extended / duration.Hours()
		case *dataTypesBolusExtended.Extended:
			key = Key(datum.Type, datum.SubType)
			duration = time.Duration(*datum.Duration) * time.Millisecond
			rate = *datum.Extended / duration.Hours()
		case *dataTypesBolusNormal.Normal:
			key = Key(datum.Type, datum.SubType)
			amount = *datum.Normal
		case *dataTypesInsulin.Insulin:
			key = Key(datum.Type)
			if datum.Dose != nil {
				amount = *datum.Dose.Total
			}
		default:
			continue
		}

		startTime := *datum.GetTime()
		endTime := startTime.Add(duration)

		if t.startTime != nil {
			if endTime.Before(*t.startTime) {
				continue
			} else if startTime.Before(*t.startTime) {
				startTime = *t.startTime
			}
		}
		if t.endTime != nil {
			if startTime.After(*t.endTime) {
				continue
			} else if endTime.After(*t.endTime) {
				endTime = *t.endTime
			}
		}

		if datum.GetType() == dataTypesBasal.Type {
			if !previousBasalEndTime.IsZero() {
				if delta := startTime.Sub(previousBasalEndTime); delta < 0 {
					t.Logger().Warnf("detected %s overlap between basal segments", -delta)
				} else if delta > 0 {
					t.Logger().Warnf("detected %s gap between basal segments", delta)
				}
			}
			previousBasalEndTime = endTime
		}

		duration = endTime.Sub(startTime)
		amount += rate * duration.Hours()

		stats := Stats{Count: 1, Duration: duration, Amount: amount}
		t.stats[key] = t.stats[key].Add(stats)
	}

	var stats Stats
	for _, insulinType := range insulinTypes {
		var insulinTypeStats Stats
		for _, insulinSubType := range insulinSubTypeMap[insulinType] {
			insulinTypeStats = insulinTypeStats.Add(t.stats[Key(insulinType, insulinSubType)])
		}
		t.stats[Key(insulinType)] = insulinTypeStats
		stats = stats.Add(insulinTypeStats)
	}
	t.stats[Key("total")] = stats

	return nil
}

func (t *Tool) outputStats() error {
	for _, insulinType := range insulinTypes {
		if stats := t.stats[Key(insulinType)]; stats.Count > 0 {
			fmt.Printf("%-20s %6d %10.4f %20s\n", insulinType, stats.Count, stats.Amount, stats.Duration)
			for _, insulinSubType := range insulinSubTypeMap[insulinType] {
				if stats := t.stats[Key(insulinType, insulinSubType)]; stats.Count > 0 {
					fmt.Printf("  %-18s %6d %10.4f %20s\n", insulinSubType, stats.Count, stats.Amount, stats.Duration)
				}
			}
			fmt.Println()
		}
	}
	if stats := t.stats[Key("total")]; stats.Count > 0 {
		fmt.Printf("%-20s %6d %10.4f %20s\n", "total", stats.Count, stats.Amount, stats.Duration)
	}

	return nil
}

func main() {
	application.RunAndExit(NewTool())
}
