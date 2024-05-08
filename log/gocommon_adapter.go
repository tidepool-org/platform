package log

import (
	"context"
	"fmt"
	"log/slog"
)

// GoCommonAdapter implements gocommon's asyncevents.Logger interface.
//
// It adapts a Logger for the purpose.
type GoCommonAdapter struct {
	Logger Logger
}

func (a *GoCommonAdapter) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	logger := a.Logger
	if fields := a.fieldsFromArgs(args); len(fields) > 0 {
		logger = logger.WithFields(fields)
	}
	logger.Log(SlogLevelToLevel[level], msg)
}

// fieldsFromArgs builds a Fields following the same rules as slog.Log.
//
// As Fields is a map instead of a slice, !BADKEY becomes !BADKEY[x] where
// x is the index counter of the value. See the godoc for slog.Log for
// details.
func (a *GoCommonAdapter) fieldsFromArgs(args []any) Fields {
	fields := Fields{}
	for i := 0; i < len(args); i++ {
		switch v := args[i].(type) {
		case slog.Attr:
			fields[v.Key] = v.Value
		case string:
			if i+1 < len(args) {
				fields[v] = args[i+1]
				i++
			} else {
				fields[fmt.Sprintf("!BADKEY[%d]", i)] = v
			}
		default:
			fields[fmt.Sprintf("!BADKEY[%d]", i)] = v
		}
	}
	return fields
}

var SlogLevelToLevel = map[slog.Level]Level{
	slog.LevelDebug: DebugLevel,
	slog.LevelInfo:  InfoLevel,
	slog.LevelWarn:  WarnLevel,
	slog.LevelError: ErrorLevel,
}
