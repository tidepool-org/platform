package log

import (
	"context"
	"time"
)

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

func ContextWithField(ctx context.Context, key string, value interface{}) context.Context {
	ctx, _ = ContextAndLoggerWithField(ctx, key, value)
	return ctx
}

func ContextAndLoggerWithField(ctx context.Context, key string, value interface{}) (context.Context, Logger) {
	if logger := LoggerFromContext(ctx); logger != nil {
		logger = logger.WithField(key, value)
		return NewContextWithLogger(ctx, logger), logger
	}
	return ctx, nil
}

func ContextWithFields(ctx context.Context, fields Fields) context.Context {
	ctx, _ = ContextAndLoggerWithFields(ctx, fields)
	return ctx
}

func ContextAndLoggerWithFields(ctx context.Context, fields Fields) (context.Context, Logger) {
	if logger := LoggerFromContext(ctx); logger != nil {
		logger = logger.WithFields(fields)
		return NewContextWithLogger(ctx, logger), logger
	}
	return ctx, nil
}

func WarnIfDurationExceedsMaximum(ctx context.Context, durationMaximum time.Duration, operation string, fn func(ctx context.Context) error) error {
	start := time.Now()
	err := fn(ctx)
	if duration := time.Since(start).Truncate(time.Microsecond); duration > durationMaximum {
		LoggerFromContext(ctx).WithField("exceeds", Fields{"duration": duration.Seconds(), "durationMaximum": durationMaximum.Seconds(), "operation": operation}).Warn("Duration exceeds maximum")
	}
	return err
}
