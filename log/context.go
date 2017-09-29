package log

import "context"

type contextKey string

const loggerContextKey contextKey = "logger"

func NewContextWithLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey, logger)
}

func LoggerFromContext(ctx context.Context) Logger {
	if ctx != nil {
		if logger, ok := ctx.Value(loggerContextKey).(Logger); ok {
			return logger
		}
	}
	return nil
}
