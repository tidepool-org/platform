// Package devlog provides a plain-text logger suitable for reading from the
// console.
//
// For production, the team prefers JSON logging, but plain-text can be useful
// for local development and tests.
//
// So while this package shouldn't be used for production, it can be enabled
// for your development use by setting TIDEPOOL_LOGGER_PACKGE=devlog as an
// environment variable. If you're using the development repo's helm chart,
// there's a global.loggerPackage override available to do the same.

package devlog

import (
	"fmt"
	"io"
	stdlog "log"
	"sort"
	"strings"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
)

// New provides a plain-text log.Logger suitable for development or testing.
func New(writer io.Writer, levelRanks log.LevelRanks, level log.Level) (log.Logger, error) {
	if writer == nil {
		return nil, errors.New("writer is missing")
	}
	logLogger := stdlog.New(writer, "", 0)
	return log.NewLogger(&serializer{logLogger}, levelRanks, level)
}

func NewWithDefaults(writer io.Writer) (log.Logger, error) {
	return New(writer, log.DefaultLevelRanks(), log.DebugLevel)
}

type serializer struct {
	*stdlog.Logger
}

func (s *serializer) Serialize(fields log.Fields) error {
	var pairs = make([]string, 0, len(fields))
	var msg, msgTime, msgLevel, msgCaller string
	var showCaller = false

	msgLevel = abbreviateLevel("")
	msgTime = time.Now().Format(time.Stamp)

	for key, value := range fields {
		switch key {
		case "caller":
			msgCaller = formatCaller(value)
		case "level":
			level := sOrPlusV(value)
			msgLevel = abbreviateLevel(level)
			if level == string(log.ErrorLevel) {
				showCaller = true
			}
		case "message":
			msg = sOrPlusV(value)
		case "process":
			// process isn't useful outside of sumo
		case "time":
			// time is terser and constant width for console logging
			tv, err := time.Parse(time.RFC3339Nano, sOrPlusV(value))
			if err != nil {
				msgTime = sOrPlusV(value)
				continue
			}
			msgTime = tv.Format(time.Stamp)
		default:
			pairs = append(pairs, key+"="+sOrPlusV(value))
		}
	}
	if showCaller {
		pairs = append(pairs, msgCaller)
	}
	sort.Strings(pairs)
	rest := ""
	if len(pairs) > 0 {
		rest = ": " + strings.Join(pairs, " ")
	}
	s.Logger.Printf(msgTime + " " + msgLevel + " " + msg + rest)
	return nil
}

func formatCaller(value any) string {
	cc, ok := value.(*errors.Caller)
	if !ok {
		return "caller=" + sOrPlusV(value)
	}
	return fmt.Sprintf("caller=%s:%d", cc.File, cc.Line)
}

func sOrPlusV(thing any) (s string) {
	if stringer, ok := thing.(fmt.Stringer); ok {
		return stringer.String()
	} else if s, ok := thing.(string); ok {
		return s
	}
	return fmt.Sprintf("%+v", thing)
}

var abbr map[string]string = map[string]string{
	string(log.DebugLevel): "DD",
	string(log.ErrorLevel): "EE",
	string(log.InfoLevel):  "II",
	string(log.WarnLevel):  "WW",
}

func abbreviateLevel(fullLevel string) string {
	if a, ok := abbr[fullLevel]; ok {
		return a
	}
	return "??"
}
